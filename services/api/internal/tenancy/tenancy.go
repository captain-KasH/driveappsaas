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
	return s.MemberRoute(func(m Member) bool { return RoleHas(m.Role, perm) }, write, h)
}

// RoleRoute is TenantRoute for one exact role (the driver and passenger apps).
func (s *Service) RoleRoute(role string, write bool, h func(*http.Request, pgx.Tx, Member) (Result, error)) httpx.Handler {
	return s.MemberRoute(func(m Member) bool { return m.Role == role }, write, h)
}

func (s *Service) MemberRoute(allow func(Member) bool, write bool, h func(*http.Request, pgx.Tx, Member) (Result, error)) httpx.Handler {
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
			if !allow(m) {
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

// PublicTenant resolves an active tenant by slug for unauthenticated brand
// endpoints. Platform scope is used for this single lookup only.
func PublicTenant(ctx context.Context, pool *pgxpool.Pool, slug string) (id, name string, err error) {
	var status string
	err = db.InTx(ctx, pool, db.Scope{Platform: true}, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT id, name, status FROM tenants WHERE slug = $1`, strings.ToLower(slug)).Scan(&id, &name, &status)
	})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && status != "active") {
		return "", "", httpx.ErrNotFound
	}
	return id, name, err
}

// UUID reports whether s is a canonical lowercase UUID (path parameter guard).
func UUID(s string) bool { return uuidRe.MatchString(s) }

// PgCode returns the Postgres SQLSTATE of err, or "".
func PgCode(err error) string { return pgCode(err) }

// CreateDriverInvitation invites the person behind an existing driver record.
func CreateDriverInvitation(ctx context.Context, tx pgx.Tx, tenantID, email, invitedBy, driverID string) (InvitationDTO, error) {
	return createInvitation(ctx, tx, tenantID, email, "driver", invitedBy, driverID)
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
	ID             string     `json:"id"`
	Slug           string     `json:"slug"`
	Name           string     `json:"name"`
	BusinessType   string     `json:"businessType"`
	CountryCode    string     `json:"countryCode"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	LegalName      *string    `json:"legalName"`
	RegistrationNo *string    `json:"registrationNo"`
	TaxNo          *string    `json:"taxNo"`
	ContactEmail   *string    `json:"contactEmail"`
	ContactPhone   *string    `json:"contactPhone"`
	BillingEmail   *string    `json:"billingEmail"`
	Address        *string    `json:"address"`
	Timezone       string     `json:"timezone"`
	Currency       string     `json:"currency"`
	Language       string     `json:"language"`
	StatusReason   *string    `json:"statusReason"`
	SubmittedAt    *time.Time `json:"submittedAt"`
	VerifiedAt     *time.Time `json:"verifiedAt"`
}

const tenantCols = `id, slug, name, business_type, country_code, status, created_at, legal_name, registration_no, tax_no,
	contact_email, contact_phone, billing_email, address, timezone, currency, language, status_reason, submitted_at, verified_at`

func scanTenant(row pgx.Row) (TenantDTO, error) {
	var t TenantDTO
	err := row.Scan(&t.ID, &t.Slug, &t.Name, &t.BusinessType, &t.CountryCode, &t.Status, &t.CreatedAt, &t.LegalName, &t.RegistrationNo,
		&t.TaxNo, &t.ContactEmail, &t.ContactPhone, &t.BillingEmail, &t.Address, &t.Timezone, &t.Currency, &t.Language,
		&t.StatusReason, &t.SubmittedAt, &t.VerifiedAt)
	return t, err
}

type InvitationDTO struct {
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Token     string    `json:"token"` // returned once; only its hash is stored
	ExpiresAt time.Time `json:"expiresAt"`
}

func createInvitation(ctx context.Context, tx pgx.Tx, tenantID, email, role, invitedBy, driverID string) (InvitationDTO, error) {
	email = strings.TrimSpace(email)
	if !strings.Contains(email, "@") || len(email) > 254 {
		return InvitationDTO{}, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_email", "a valid email is required")
	}
	if !validRole(role) {
		return InvitationDTO{}, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_role", "unknown role %q", role)
	}
	token, hash := identity.NewToken("txi_")
	inv := InvitationDTO{Email: email, Role: role, Token: token, ExpiresAt: time.Now().Add(InvitationTTL).UTC()}
	_, err := tx.Exec(ctx, `INSERT INTO tenant_invitations(tenant_id, email, role, token_hash, invited_by, expires_at, driver_id)
		VALUES ($1,$2,$3,$4,$5,$6,nullif($7,'')::uuid)`, tenantID, email, role, hash, invitedBy, inv.ExpiresAt, driverID)
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
