// Package tenancy owns tenants, memberships, invitations and the tenant
// request context. Every tenant route resolves membership server-side; the
// tenant ID in the path is never trusted on its own.
package tenancy

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/audit"
	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
	"github.com/captain-KasH/driveappsaas/services/api/internal/identity"
)

const InvitationTTL = 7 * 24 * time.Hour

type Service struct{ Pool *pgxpool.Pool }

// Member is the resolved tenant context for one request.
type Member struct {
	identity.Principal
	TenantID     string
	Role         string
	TenantStatus string
}

// Result is written only after the transaction commits.
type Result struct {
	Status int
	Body   any
}

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// TenantRoute authenticates, resolves membership under RLS, checks the
// permission, blocks writes on inactive tenants, and runs h in one transaction.
func (s *Service) TenantRoute(perm Permission, write bool, h func(*http.Request, pgx.Tx, Member) (Result, error)) httpx.Handler {
	return identity.RequireAuth(s.Pool, func(w http.ResponseWriter, r *http.Request) error {
		p, _ := identity.FromContext(r.Context())
		tid := strings.ToLower(r.PathValue("tenantID"))
		if !uuidRe.MatchString(tid) {
			return httpx.ErrNotFound
		}
		var res Result
		err := db.InTx(r.Context(), s.Pool, db.Scope{UserID: p.UserID, TenantID: tid}, func(tx pgx.Tx) error {
			m := Member{Principal: p, TenantID: tid}
			err := tx.QueryRow(r.Context(), `SELECT m.role, t.status FROM tenant_memberships m
				JOIN tenants t ON t.id = m.tenant_id WHERE m.tenant_id = $1 AND m.user_id = $2`,
				tid, p.UserID).Scan(&m.Role, &m.TenantStatus)
			if errors.Is(err, pgx.ErrNoRows) {
				return httpx.ErrNotFound // do not reveal whether the tenant exists
			}
			if err != nil {
				return err
			}
			if !RoleHas(m.Role, perm) {
				return httpx.ErrForbidden
			}
			if write && (m.TenantStatus == "suspended" || m.TenantStatus == "closed") {
				return httpx.Errorf(http.StatusForbidden, "tenant_inactive", "this business is %s", m.TenantStatus)
			}
			res, err = h(r, tx, m)
			return err
		})
		if err != nil {
			return err
		}
		return httpx.WriteJSON(w, res.Status, res.Body)
	})
}

// PlatformRoute runs h in cross-tenant platform scope for the given platform roles only.
func (s *Service) PlatformRoute(roles []string, h func(*http.Request, pgx.Tx, identity.Principal) (Result, error)) httpx.Handler {
	return identity.RequireAuth(s.Pool, func(w http.ResponseWriter, r *http.Request) error {
		p, _ := identity.FromContext(r.Context())
		allowed := false
		for _, role := range roles {
			allowed = allowed || (p.PlatformRole != "" && p.PlatformRole == role)
		}
		if !allowed {
			return httpx.ErrForbidden
		}
		var res Result
		err := db.InTx(r.Context(), s.Pool, db.Scope{UserID: p.UserID, Platform: true}, func(tx pgx.Tx) error {
			var err error
			res, err = h(r, tx, p)
			return err
		})
		if err != nil {
			return err
		}
		return httpx.WriteJSON(w, res.Status, res.Body)
	})
}

type TenantDTO struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	BusinessType string    `json:"businessType"`
	CountryCode  string    `json:"countryCode"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

const tenantCols = `id, slug, name, business_type, country_code, status, created_at`

func scanTenant(row pgx.Row) (TenantDTO, error) {
	var t TenantDTO
	err := row.Scan(&t.ID, &t.Slug, &t.Name, &t.BusinessType, &t.CountryCode, &t.Status, &t.CreatedAt)
	return t, err
}

type InvitationDTO struct {
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Token     string    `json:"token"` // returned once; only its hash is stored
	ExpiresAt time.Time `json:"expiresAt"`
}

func createInvitation(ctx context.Context, tx pgx.Tx, tenantID, email, role, invitedBy string) (InvitationDTO, error) {
	email = strings.TrimSpace(email)
	if !strings.Contains(email, "@") || len(email) > 254 {
		return InvitationDTO{}, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_email", "a valid email is required")
	}
	if !validRole(role) {
		return InvitationDTO{}, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_role", "unknown role %q", role)
	}
	token, hash := identity.NewToken("txi_")
	inv := InvitationDTO{Email: email, Role: role, Token: token, ExpiresAt: time.Now().Add(InvitationTTL).UTC()}
	_, err := tx.Exec(ctx, `INSERT INTO tenant_invitations(tenant_id, email, role, token_hash, invited_by, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6)`, tenantID, email, role, hash, invitedBy, inv.ExpiresAt)
	if err != nil {
		return inv, err
	}
	return inv, audit.Record(ctx, tx, audit.Event{TenantID: tenantID, ActorUserID: invitedBy,
		Action: "invitation.created", TargetType: "invitation", TargetID: email, Metadata: map[string]any{"role": role}})
}

func pgCode(err error) string {
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		return pe.Code
	}
	return ""
}
