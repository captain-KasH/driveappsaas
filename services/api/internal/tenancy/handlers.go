package tenancy

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/captain-KasH/driveappsaas/services/api/internal/audit"
	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
	"github.com/captain-KasH/driveappsaas/services/api/internal/identity"
)

func (s *Service) Register(mux *http.ServeMux) {
	super := []string{"super_admin"}
	staff := []string{"super_admin", "support"}
	mux.Handle("POST /v1/platform/tenants", s.PlatformRoute(super, s.createTenant))
	mux.Handle("GET /v1/platform/tenants", s.PlatformRoute(staff, s.listTenants))
	mux.Handle("POST /v1/platform/tenants/{tenantID}/status", s.PlatformRoute(super, s.setTenantStatus))

	mux.Handle("POST /v1/auth/accept-invite", httpx.Handler(s.acceptInvite))
	mux.Handle("GET /v1/me", identity.RequireAuth(s.Pool, s.me))

	mux.Handle("GET /v1/tenants/{tenantID}", s.TenantRoute(PermTenantRead, false, s.getTenant))
	mux.Handle("PATCH /v1/tenants/{tenantID}", s.TenantRoute(PermTenantManage, true, s.updateTenant))
	mux.Handle("POST /v1/tenants/{tenantID}/onboarding/complete", s.TenantRoute(PermTenantManage, true, s.completeOnboarding))
	mux.Handle("GET /v1/tenants/{tenantID}/members", s.TenantRoute(PermStaffRead, false, s.listMembers))
	mux.Handle("POST /v1/tenants/{tenantID}/invitations", s.TenantRoute(PermStaffManage, true, s.inviteStaff))
	mux.Handle("GET /v1/tenants/{tenantID}/audit-events", s.TenantRoute(PermAuditRead, false, s.listAudit))
}

// --- platform (super-admin console) ---

func (s *Service) createTenant(r *http.Request, tx pgx.Tx, p identity.Principal) (Result, error) {
	var in struct {
		Slug         string `json:"slug"`
		Name         string `json:"name"`
		BusinessType string `json:"businessType"`
		CountryCode  string `json:"countryCode"`
		OwnerEmail   string `json:"ownerEmail"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if in.CountryCode == "" {
		in.CountryCode = "LK"
	}
	t, err := scanTenant(tx.QueryRow(r.Context(), `INSERT INTO tenants(slug, name, business_type, country_code)
		VALUES ($1,$2,$3,$4) RETURNING `+tenantCols,
		strings.ToLower(strings.TrimSpace(in.Slug)), strings.TrimSpace(in.Name), in.BusinessType, strings.ToUpper(in.CountryCode)))
	switch pgCode(err) {
	case "":
	case "23505":
		return Result{}, httpx.Errorf(http.StatusConflict, "slug_taken", "slug %q is already in use", in.Slug)
	case "23514", "22001":
		return Result{}, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_tenant",
			"slug must be 3-63 lowercase letters, digits or dashes; name 1-200 chars; businessType must be a supported type")
	default:
		return Result{}, err
	}
	if err := audit.Record(r.Context(), tx, audit.Event{TenantID: t.ID, ActorUserID: p.UserID,
		Action: "tenant.created", TargetType: "tenant", TargetID: t.ID}); err != nil {
		return Result{}, err
	}
	inv, err := createInvitation(r.Context(), tx, t.ID, in.OwnerEmail, "owner", p.UserID, "")
	if err != nil {
		return Result{}, err
	}
	return Result{http.StatusCreated, map[string]any{"tenant": t, "ownerInvitation": inv}}, nil
}

func (s *Service) listTenants(r *http.Request, tx pgx.Tx, _ identity.Principal) (Result, error) {
	rows, err := tx.Query(r.Context(), `SELECT `+tenantCols+` FROM tenants ORDER BY created_at DESC LIMIT 200`)
	if err != nil {
		return Result{}, err
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (TenantDTO, error) { return scanTenant(row) })
	return Result{http.StatusOK, map[string]any{"tenants": list}}, err
}

// Platform-driven status transitions. onboarding->active is the owner's job (completeOnboarding).
var platformTransitions = map[string][]string{
	"onboarding": {"suspended", "closed"},
	"active":     {"suspended", "closed"},
	"suspended":  {"active", "closed"},
}

func (s *Service) setTenantStatus(r *http.Request, tx pgx.Tx, p identity.Principal) (Result, error) {
	var in struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if len(strings.TrimSpace(in.Reason)) < 3 {
		return Result{}, httpx.Errorf(http.StatusUnprocessableEntity, "reason_required", "a reason is required for status changes")
	}
	tid := r.PathValue("tenantID")
	if !uuidRe.MatchString(tid) {
		return Result{}, httpx.ErrNotFound
	}
	var from string
	err := tx.QueryRow(r.Context(), `SELECT status FROM tenants WHERE id = $1 FOR UPDATE`, tid).Scan(&from)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, httpx.ErrNotFound
	}
	if err != nil {
		return Result{}, err
	}
	ok := false
	for _, to := range platformTransitions[from] {
		ok = ok || to == in.Status
	}
	if !ok {
		return Result{}, httpx.Errorf(http.StatusConflict, "invalid_transition", "cannot move tenant from %s to %s", from, in.Status)
	}
	t, err := scanTenant(tx.QueryRow(r.Context(), `UPDATE tenants SET status=$2, status_reason=$3, updated_at=now()
		WHERE id=$1 RETURNING `+tenantCols, tid, in.Status, in.Reason))
	if err != nil {
		return Result{}, err
	}
	// Suspension keeps read-only access. Closing signs out staff who belong to no other tenant.
	if in.Status == "closed" {
		if _, err := tx.Exec(r.Context(), `UPDATE sessions SET revoked_at=now() WHERE revoked_at IS NULL AND user_id IN
			(SELECT user_id FROM tenant_memberships WHERE tenant_id=$1) AND user_id NOT IN
			(SELECT user_id FROM tenant_memberships WHERE tenant_id<>$1) AND user_id NOT IN
			(SELECT id FROM users WHERE platform_role IS NOT NULL)`, tid); err != nil {
			return Result{}, err
		}
	}
	err = audit.Record(r.Context(), tx, audit.Event{TenantID: tid, ActorUserID: p.UserID, Action: "tenant.status_changed",
		TargetType: "tenant", TargetID: tid, Reason: in.Reason, Metadata: map[string]any{"from": from, "to": in.Status}})
	return Result{http.StatusOK, map[string]any{"tenant": t}}, err
}

// --- onboarding ---

func (s *Service) acceptInvite(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Token    string `json:"token"`
		FullName string `json:"fullName"`
		Password string `json:"password"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	invalid := httpx.Errorf(http.StatusNotFound, "invalid_invitation", "this invitation is invalid or has expired")
	var out map[string]any
	// Platform scope: the invitation is located by token hash before any tenant is known.
	err := db.InTx(r.Context(), s.Pool, db.Scope{Platform: true}, func(tx pgx.Tx) error {
		var invID, tenantID, email, role, tenantStatus string
		var driverID *string
		err := tx.QueryRow(r.Context(), `SELECT i.id, i.tenant_id, i.email, i.role, t.status, i.driver_id FROM tenant_invitations i
			JOIN tenants t ON t.id = i.tenant_id
			WHERE i.token_hash = $1 AND i.accepted_at IS NULL AND i.expires_at > now() FOR UPDATE OF i`,
			identity.HashToken(in.Token)).Scan(&invID, &tenantID, &email, &role, &tenantStatus, &driverID)
		if errors.Is(err, pgx.ErrNoRows) || tenantStatus == "closed" {
			return invalid
		}
		if err != nil {
			return err
		}
		var userID string
		var hash *string
		err = tx.QueryRow(r.Context(), `SELECT id, password_hash FROM users WHERE email=$1`, email).Scan(&userID, &hash)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			name := strings.TrimSpace(in.FullName)
			if name == "" {
				return httpx.Errorf(http.StatusUnprocessableEntity, "name_required", "full name is required")
			}
			pw, err := identity.HashPassword(in.Password)
			if err != nil {
				return err
			}
			if err := tx.QueryRow(r.Context(), `INSERT INTO users(email, full_name, password_hash) VALUES ($1,$2,$3) RETURNING id`,
				email, name, pw).Scan(&userID); err != nil {
				return err
			}
		case err != nil:
			return err
		case !identity.CheckPassword(hash, in.Password):
			return identity.ErrBadCredentials // existing account: prove ownership with its password
		}
		if _, err := tx.Exec(r.Context(), `INSERT INTO tenant_memberships(tenant_id, user_id, role) VALUES ($1,$2,$3)
			ON CONFLICT (tenant_id, user_id) DO NOTHING`, tenantID, userID, role); err != nil {
			return err
		}
		if driverID != nil {
			if _, err := tx.Exec(r.Context(), `UPDATE drivers SET user_id=$2 WHERE id=$1 AND user_id IS NULL`, *driverID, userID); err != nil {
				if PgCode(err) == "23505" {
					return httpx.Errorf(http.StatusConflict, "already_driver", "this account is already a driver for this business")
				}
				return err
			}
		}
		if _, err := tx.Exec(r.Context(), `UPDATE tenant_invitations SET accepted_at=now() WHERE id=$1`, invID); err != nil {
			return err
		}
		if err := audit.Record(r.Context(), tx, audit.Event{TenantID: tenantID, ActorUserID: userID,
			Action: "invitation.accepted", TargetType: "user", TargetID: userID, Metadata: map[string]any{"role": role}}); err != nil {
			return err
		}
		token, err := identity.CreateSession(r.Context(), tx, userID, r.UserAgent(), identity.RemoteIP(r))
		out = map[string]any{"token": token, "tenantId": tenantID, "role": role}
		return err
	})
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, out)
}

func (s *Service) me(w http.ResponseWriter, r *http.Request) error {
	p, _ := identity.FromContext(r.Context())
	type membership struct {
		TenantID     string `json:"tenantId"`
		TenantSlug   string `json:"tenantSlug"`
		TenantName   string `json:"tenantName"`
		TenantStatus string `json:"tenantStatus"`
		Role         string `json:"role"`
	}
	var list []membership
	var name string
	err := db.InTx(r.Context(), s.Pool, p.Scope(), func(tx pgx.Tx) error {
		if err := tx.QueryRow(r.Context(), `SELECT full_name FROM users WHERE id=$1`, p.UserID).Scan(&name); err != nil {
			return err
		}
		rows, err := tx.Query(r.Context(), `SELECT t.id, t.slug, t.name, t.status, m.role FROM tenant_memberships m
			JOIN tenants t ON t.id = m.tenant_id WHERE m.user_id = $1 ORDER BY t.name`, p.UserID)
		if err != nil {
			return err
		}
		list, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (membership, error) {
			var m membership
			return m, row.Scan(&m.TenantID, &m.TenantSlug, &m.TenantName, &m.TenantStatus, &m.Role)
		})
		return err
	})
	if err != nil {
		return err
	}
	if list == nil {
		list = []membership{}
	}
	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"user": map[string]any{
		"id": p.UserID, "email": p.Email, "fullName": name, "platformRole": p.PlatformRole}, "memberships": list})
}

// --- tenant workspace ---

func (s *Service) getTenant(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	t, err := scanTenant(tx.QueryRow(r.Context(), `SELECT `+tenantCols+` FROM tenants WHERE id=$1`, m.TenantID))
	return Result{http.StatusOK, map[string]any{"tenant": t, "myRole": m.Role}}, err
}

func (s *Service) updateTenant(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in struct {
		Name *string `json:"name"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if in.Name == nil || strings.TrimSpace(*in.Name) == "" || len(*in.Name) > 200 {
		return Result{}, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_name", "name must be 1-200 characters")
	}
	t, err := scanTenant(tx.QueryRow(r.Context(), `UPDATE tenants SET name=$2, updated_at=now() WHERE id=$1 RETURNING `+tenantCols,
		m.TenantID, strings.TrimSpace(*in.Name)))
	if err != nil {
		return Result{}, err
	}
	err = audit.Record(r.Context(), tx, audit.Event{TenantID: m.TenantID, ActorUserID: m.UserID,
		Action: "tenant.updated", TargetType: "tenant", TargetID: m.TenantID, Metadata: map[string]any{"name": t.Name}})
	return Result{http.StatusOK, map[string]any{"tenant": t}}, err
}

// completeOnboarding activates the tenant once it can actually price a ride:
// at least one active zone, active category and fare rule.
func (s *Service) completeOnboarding(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	if m.Role != "owner" {
		return Result{}, httpx.ErrForbidden
	}
	if m.TenantStatus != "onboarding" {
		return Result{}, httpx.Errorf(http.StatusConflict, "invalid_transition", "tenant is already %s", m.TenantStatus)
	}
	var zones, categories, fares int
	if err := tx.QueryRow(r.Context(), `SELECT (SELECT count(*) FROM service_zones WHERE active),
		(SELECT count(*) FROM vehicle_categories WHERE active), (SELECT count(*) FROM fare_rules)`).Scan(&zones, &categories, &fares); err != nil {
		return Result{}, err
	}
	var missing []string
	for name, n := range map[string]int{"service zone": zones, "vehicle category": categories, "fare rule": fares} {
		if n == 0 {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return Result{}, httpx.Errorf(http.StatusConflict, "onboarding_incomplete", "add at least one %s first", strings.Join(missing, ", "))
	}
	t, err := scanTenant(tx.QueryRow(r.Context(), `UPDATE tenants SET status='active', updated_at=now()
		WHERE id=$1 AND status='onboarding' RETURNING `+tenantCols, m.TenantID))
	if err != nil {
		return Result{}, err
	}
	err = audit.Record(r.Context(), tx, audit.Event{TenantID: m.TenantID, ActorUserID: m.UserID,
		Action: "tenant.onboarding_completed", TargetType: "tenant", TargetID: m.TenantID})
	return Result{http.StatusOK, map[string]any{"tenant": t}}, err
}

func (s *Service) listMembers(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	type member struct {
		UserID   string    `json:"userId"`
		Email    string    `json:"email"`
		FullName string    `json:"fullName"`
		Role     string    `json:"role"`
		Since    time.Time `json:"since"`
	}
	rows, err := tx.Query(r.Context(), `SELECT u.id, u.email, u.full_name, m.role, m.created_at FROM tenant_memberships m
		JOIN users u ON u.id = m.user_id WHERE m.tenant_id = $1 AND m.role <> 'passenger' ORDER BY m.created_at`, m.TenantID)
	if err != nil {
		return Result{}, err
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (member, error) {
		var x member
		return x, row.Scan(&x.UserID, &x.Email, &x.FullName, &x.Role, &x.Since)
	})
	return Result{http.StatusOK, map[string]any{"members": list}}, err
}

func (s *Service) inviteStaff(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if in.Role == "driver" {
		return Result{}, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_role", "drivers are invited from the drivers page")
	}
	if in.Role == "owner" && m.Role != "owner" {
		return Result{}, httpx.Errorf(http.StatusForbidden, "forbidden", "only an owner can invite another owner")
	}
	inv, err := createInvitation(r.Context(), tx, m.TenantID, in.Email, in.Role, m.UserID, "")
	return Result{http.StatusCreated, map[string]any{"invitation": inv}}, err
}

func (s *Service) listAudit(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	type event struct {
		ID         int64          `json:"id"`
		OccurredAt time.Time      `json:"occurredAt"`
		ActorID    *string        `json:"actorUserId"`
		Action     string         `json:"action"`
		TargetType string         `json:"targetType"`
		TargetID   string         `json:"targetId"`
		Reason     *string        `json:"reason"`
		Metadata   map[string]any `json:"metadata"`
	}
	rows, err := tx.Query(r.Context(), `SELECT id, occurred_at, actor_user_id, action, target_type, target_id, reason, metadata
		FROM audit_events WHERE tenant_id = $1 ORDER BY id DESC LIMIT 100`, m.TenantID)
	if err != nil {
		return Result{}, err
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (event, error) {
		var e event
		return e, row.Scan(&e.ID, &e.OccurredAt, &e.ActorID, &e.Action, &e.TargetType, &e.TargetID, &e.Reason, &e.Metadata)
	})
	return Result{http.StatusOK, map[string]any{"events": list}}, err
}
