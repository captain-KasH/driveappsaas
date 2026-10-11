package files

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/audit"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
	"github.com/captain-KasH/driveappsaas/services/api/internal/identity"
	"github.com/captain-KasH/driveappsaas/services/api/internal/tenancy"
)

const MaxSize = 10 << 20 // 10 MiB, matches the CHECK constraint

var allowedTypes = map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "application/pdf": ".pdf"}

type Service struct {
	Pool    *pgxpool.Pool
	Tenancy *tenancy.Service
	Store   Store
}

type File struct {
	ID          string    `json:"id"`
	ContentType string    `json:"contentType"`
	SizeBytes   int64     `json:"sizeBytes"`
	FileName    string    `json:"fileName"`
	CreatedAt   time.Time `json:"createdAt"`
}

const fileCols = `id, content_type, size_bytes, file_name, created_at`

func scanFile(row pgx.Row) (File, error) {
	var f File
	return f, row.Scan(&f.ID, &f.ContentType, &f.SizeBytes, &f.FileName, &f.CreatedAt)
}

func (s *Service) Register(mux *http.ServeMux) {
	p := "/v1/tenants/{tenantID}"
	notPassenger := func(m tenancy.Member) bool { return m.Role != "passenger" }
	mux.Handle("POST "+p+"/files", s.Tenancy.MemberRoute(notPassenger, true, s.upload))
	mux.Handle("GET "+p+"/files/{fileID}", s.Tenancy.MemberRouteRaw(notPassenger, s.download))
	mux.Handle("GET /v1/platform/files/{fileID}", s.Tenancy.PlatformRouteRaw([]string{"super_admin", "support"}, s.platformDownload))
}

// upload accepts one multipart file (field "file"). The type is sniffed from the bytes, never
// trusted from the client; the row and the bytes are written together (row first, so a failed
// store write rolls the row back).
func (s *Service) upload(r *http.Request, tx pgx.Tx, m tenancy.Member) (tenancy.Result, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, MaxSize+64<<10)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			return tenancy.Result{}, httpx.Errorf(http.StatusRequestEntityTooLarge, "file_too_large", "files must be 10 MB or smaller")
		}
		return tenancy.Result{}, httpx.Errorf(http.StatusBadRequest, "invalid_upload", "send one file in a multipart field named \"file\"")
	}
	part, hdr, err := r.FormFile("file")
	if err != nil {
		return tenancy.Result{}, httpx.Errorf(http.StatusBadRequest, "invalid_upload", "send one file in a multipart field named \"file\"")
	}
	defer part.Close()
	data, err := io.ReadAll(io.LimitReader(part, MaxSize+1))
	if err != nil {
		return tenancy.Result{}, err
	}
	if len(data) > MaxSize {
		return tenancy.Result{}, httpx.Errorf(http.StatusRequestEntityTooLarge, "file_too_large", "files must be 10 MB or smaller")
	}
	if len(data) == 0 {
		return tenancy.Result{}, httpx.Errorf(http.StatusUnprocessableEntity, "empty_file", "the file is empty")
	}
	ctype := http.DetectContentType(data)
	if i := strings.Index(ctype, ";"); i > 0 {
		ctype = ctype[:i]
	}
	ext, ok := allowedTypes[ctype]
	if !ok {
		return tenancy.Result{}, httpx.Errorf(http.StatusUnsupportedMediaType, "unsupported_file_type", "upload a JPEG, PNG, WebP or PDF")
	}
	name := strings.TrimSpace(hdr.Filename)
	if name == "" || len(name) > 200 {
		name = "upload" + ext
	}
	sum := sha256.Sum256(data)
	var f File
	var key string
	err = tx.QueryRow(r.Context(), `WITH new AS (SELECT gen_random_uuid() AS id)
		INSERT INTO files(id, tenant_id, uploaded_by, content_type, size_bytes, sha256, storage_key, file_name)
		SELECT id, $1::uuid, $2, $3, $4, $5, $1::text || '/' || id::text, $6 FROM new
		RETURNING `+fileCols+`, storage_key`, m.TenantID, m.UserID, ctype, len(data), sum[:], name).
		Scan(&f.ID, &f.ContentType, &f.SizeBytes, &f.FileName, &f.CreatedAt, &key)
	if err != nil {
		return tenancy.Result{}, err
	}
	if err := s.Store.Put(r.Context(), key, bytes.NewReader(data)); err != nil {
		return tenancy.Result{}, fmt.Errorf("store: %w", err)
	}
	err = audit.Record(r.Context(), tx, audit.Event{TenantID: m.TenantID, ActorUserID: m.UserID, Action: "file.uploaded",
		TargetType: "file", TargetID: f.ID, Metadata: map[string]any{"contentType": ctype, "sizeBytes": len(data)}})
	return tenancy.Result{Status: http.StatusCreated, Body: map[string]any{"file": f}}, err
}

// canRead: staff who may read the fleet or manage the tenant see every file; anyone sees what
// they uploaded; a driver sees files attached to their own documents.
func canRead(r *http.Request, tx pgx.Tx, m tenancy.Member, fileID string) (File, string, error) {
	var f File
	var key, uploadedBy string
	err := tx.QueryRow(r.Context(), `SELECT `+fileCols+`, storage_key, uploaded_by FROM files WHERE id=$1`, fileID).
		Scan(&f.ID, &f.ContentType, &f.SizeBytes, &f.FileName, &f.CreatedAt, &key, &uploadedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return f, "", httpx.ErrNotFound
	}
	if err != nil {
		return f, "", err
	}
	if uploadedBy == m.UserID || tenancy.RoleHas(m.Role, tenancy.PermFleetRead) || tenancy.RoleHas(m.Role, tenancy.PermTenantManage) {
		return f, key, nil
	}
	var mine bool
	if err := tx.QueryRow(r.Context(), `SELECT EXISTS (SELECT 1 FROM driver_documents dd JOIN drivers d ON d.id = dd.driver_id
		WHERE dd.file_id = $1 AND d.user_id = $2)`, fileID, m.UserID).Scan(&mine); err != nil {
		return f, "", err
	}
	if !mine {
		return f, "", httpx.ErrNotFound
	}
	return f, key, nil
}

func (s *Service) download(w http.ResponseWriter, r *http.Request, tx pgx.Tx, m tenancy.Member) error {
	id := strings.ToLower(r.PathValue("fileID"))
	if !tenancy.UUID(id) {
		return httpx.ErrNotFound
	}
	f, key, err := canRead(r, tx, m, id)
	if err != nil {
		return err
	}
	return s.serve(w, r, f, key)
}

// platformDownload is the verification-queue view; every read is audited (prompt §9: minimal PII exposure).
func (s *Service) platformDownload(w http.ResponseWriter, r *http.Request, tx pgx.Tx, p identity.Principal) error {
	id := strings.ToLower(r.PathValue("fileID"))
	if !tenancy.UUID(id) {
		return httpx.ErrNotFound
	}
	var f File
	var key, tenantID string
	err := tx.QueryRow(r.Context(), `SELECT `+fileCols+`, storage_key, tenant_id FROM files WHERE id=$1`, id).
		Scan(&f.ID, &f.ContentType, &f.SizeBytes, &f.FileName, &f.CreatedAt, &key, &tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.ErrNotFound
	}
	if err != nil {
		return err
	}
	if err := audit.Record(r.Context(), tx, audit.Event{TenantID: tenantID, ActorUserID: p.UserID, Action: "file.viewed_by_platform",
		TargetType: "file", TargetID: f.ID}); err != nil {
		return err
	}
	return s.serve(w, r, f, key)
}

func (s *Service) serve(w http.ResponseWriter, r *http.Request, f File, key string) error {
	rc, err := s.Store.Open(r.Context(), key)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	defer rc.Close()
	w.Header().Set("Content-Type", f.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(f.SizeBytes, 10))
	w.Header().Set("Content-Disposition", `inline; filename="`+strings.ReplaceAll(f.FileName, `"`, "")+`"`)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, f.FileName, f.CreatedAt, rc)
	return nil
}
