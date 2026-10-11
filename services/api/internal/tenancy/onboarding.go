package tenancy

import (
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/captain-KasH/driveappsaas/services/api/internal/audit"
	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
	"github.com/captain-KasH/driveappsaas/services/api/internal/identity"
)

// MaxDraftBusinesses caps how many unverified businesses one account can open (abuse control).
const MaxDraftBusinesses = 3

type newTenantInput struct {
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	BusinessType string `json:"businessType"`
	CountryCode  string `json:"countryCode"`
}

// insertTenant creates the tenant row. verified = created by platform staff (trusted).
func insertTenant(r *http.Request, tx pgx.Tx, in newTenantInput, createdBy string, verified bool) (TenantDTO, error) {
	if in.CountryCode == "" {
		in.CountryCode = "LK"
	}
	t, err := scanTenant(tx.QueryRow(r.Context(), `INSERT INTO tenants(slug, name, business_type, country_code, created_by, verified_at, verified_by)
		VALUES ($1,$2,$3,$4,$5, CASE WHEN $6 THEN now() END, CASE WHEN $6 THEN $5::uuid END) RETURNING `+tenantCols,
		strings.ToLower(strings.TrimSpace(in.Slug)), strings.TrimSpace(in.Name), in.BusinessType, strings.ToUpper(in.CountryCode), createdBy, verified))
	switch pgCode(err) {
	case "":
	case "23505":
		return t, httpx.Errorf(http.StatusConflict, "slug_taken", "slug %q is already in use", in.Slug)
	case "23514", "22001":
		return t, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_tenant",
			"slug must be 3-63 lowercase letters, digits or dashes; name 1-200 chars; businessType must be a supported type")
	default:
		return t, err
	}
	return t, audit.Record(r.Context(), tx, audit.Event{TenantID: t.ID, ActorUserID: createdBy,
		Action: "tenant.created", TargetType: "tenant", TargetID: t.ID, Metadata: map[string]any{"selfService": !verified}})
}

// createBusiness is the self-service path (D-021): a verified account opens a draft business and
// becomes its owner. Platform scope is needed to insert the first membership (D-006 lists this).
func (s *Service) createBusiness(w http.ResponseWriter, r *http.Request) error {
	p, _ := identity.FromContext(r.Context())
	var in newTenantInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	var t TenantDTO
	err := db.InTx(r.Context(), s.Pool, db.Scope{UserID: p.UserID, Platform: true}, func(tx pgx.Tx) error {
		var verified bool
		var drafts int
		if err := tx.QueryRow(r.Context(), `SELECT email_verified_at IS NOT NULL,
			(SELECT count(*) FROM tenants WHERE created_by = $1 AND status IN ('onboarding','pending_verification'))
			FROM users WHERE id = $1`, p.UserID).Scan(&verified, &drafts); err != nil {
			return err
		}
		if !verified {
			return httpx.Errorf(http.StatusForbidden, "email_unverified", "verify your email address before creating a business")
		}
		if drafts >= MaxDraftBusinesses {
			return httpx.Errorf(http.StatusConflict, "too_many_drafts", "finish setting up your existing businesses first")
		}
		var err error
		if t, err = insertTenant(r, tx, in, p.UserID, false); err != nil {
			return err
		}
		_, err = tx.Exec(r.Context(), `INSERT INTO tenant_memberships(tenant_id, user_id, role) VALUES ($1,$2,'owner')`, t.ID, p.UserID)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusCreated, map[string]any{"tenant": t})
}

// --- business profile ---

var phoneRe = regexp.MustCompile(`^\+?[0-9 ()-]{7,20}$`)

func (s *Service) updateTenant(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in struct {
		Name           *string `json:"name"`
		LegalName      *string `json:"legalName"`
		RegistrationNo *string `json:"registrationNo"`
		TaxNo          *string `json:"taxNo"`
		ContactEmail   *string `json:"contactEmail"`
		ContactPhone   *string `json:"contactPhone"`
		BillingEmail   *string `json:"billingEmail"`
		Address        *string `json:"address"`
		Timezone       *string `json:"timezone"`
		Currency       *string `json:"currency"`
		Language       *string `json:"language"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	invalid := func(field, msg string) (Result, error) {
		return Result{}, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_"+field, "%s", msg)
	}
	sets, args := []string{"updated_at=now()"}, []any{m.TenantID}
	changed := map[string]any{}
	// optional text: "" clears the field (NULL); nil leaves it untouched.
	opt := func(col, field string, v *string, max int, check func(string) bool) bool {
		if v == nil {
			return true
		}
		s := strings.TrimSpace(*v)
		if s != "" && (len(s) > max || (check != nil && !check(s))) {
			return false
		}
		args = append(args, s)
		sets = append(sets, col+"=nullif($"+itoa(len(args))+",'')")
		changed[field] = s
		return true
	}
	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" || len(n) > 200 {
			return invalid("name", "name must be 1-200 characters")
		}
		args = append(args, n)
		sets = append(sets, "name=$"+itoa(len(args)))
		changed["name"] = n
	}
	isEmail := func(s string) bool { a, err := mail.ParseAddress(s); return err == nil && a.Address == s }
	if !opt("legal_name", "legalName", in.LegalName, 200, nil) {
		return invalid("legal_name", "legal name must be at most 200 characters")
	}
	if !opt("registration_no", "registrationNo", in.RegistrationNo, 64, nil) {
		return invalid("registration_no", "registration number must be at most 64 characters")
	}
	if !opt("tax_no", "taxNo", in.TaxNo, 64, nil) {
		return invalid("tax_no", "tax number must be at most 64 characters")
	}
	if !opt("contact_email", "contactEmail", in.ContactEmail, 254, isEmail) {
		return invalid("contact_email", "contact email must be a valid address")
	}
	if !opt("billing_email", "billingEmail", in.BillingEmail, 254, isEmail) {
		return invalid("billing_email", "billing email must be a valid address")
	}
	if !opt("contact_phone", "contactPhone", in.ContactPhone, 20, phoneRe.MatchString) {
		return invalid("contact_phone", "contact phone must be 7-20 digits, optionally starting with +")
	}
	if !opt("address", "address", in.Address, 500, nil) {
		return invalid("address", "address must be at most 500 characters")
	}
	if in.Timezone != nil {
		if _, err := time.LoadLocation(*in.Timezone); err != nil || *in.Timezone == "" {
			return invalid("timezone", "timezone must be an IANA name such as Asia/Colombo")
		}
		args = append(args, *in.Timezone)
		sets = append(sets, "timezone=$"+itoa(len(args)))
		changed["timezone"] = *in.Timezone
	}
	if in.Currency != nil {
		c := strings.ToUpper(strings.TrimSpace(*in.Currency))
		if !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(c) {
			return invalid("currency", "currency must be a 3-letter ISO code")
		}
		args = append(args, c)
		sets = append(sets, "currency=$"+itoa(len(args)))
		changed["currency"] = c
	}
	if in.Language != nil {
		if l := *in.Language; l != "en" && l != "si" && l != "ta" {
			return invalid("language", "language must be en, si or ta")
		}
		args = append(args, *in.Language)
		sets = append(sets, "language=$"+itoa(len(args)))
		changed["language"] = *in.Language
	}
	if len(changed) == 0 {
		return invalid("input", "nothing to update")
	}
	t, err := scanTenant(tx.QueryRow(r.Context(), `UPDATE tenants SET `+strings.Join(sets, ", ")+` WHERE id=$1 RETURNING `+tenantCols, args...))
	if err != nil {
		if pgCode(err) == "23514" {
			return invalid("input", "a field is out of range")
		}
		return Result{}, err
	}
	err = audit.Record(r.Context(), tx, audit.Event{TenantID: m.TenantID, ActorUserID: m.UserID,
		Action: "tenant.updated", TargetType: "tenant", TargetID: m.TenantID, Metadata: changed})
	return Result{http.StatusOK, map[string]any{"tenant": t}}, err
}

func itoa(i int) string { return strconv.Itoa(i) }

// --- checklist + submission ---

type checklistStep struct {
	Key      string   `json:"key"`
	Done     bool     `json:"done"`
	Required bool     `json:"required"`
	Missing  []string `json:"missing,omitempty"`
}

type checklist struct {
	Status    string          `json:"status"`
	Verified  bool            `json:"verified"`
	CanSubmit bool            `json:"canSubmit"`
	Steps     []checklistStep `json:"steps"`
}

// buildChecklist computes onboarding progress from real rows — nothing is stored client-side.
func buildChecklist(r *http.Request, tx pgx.Tx, tenantID string) (checklist, error) {
	var c checklist
	var legal, reg, cemail, cphone *string
	var verifiedAt *time.Time
	var zones, cats, fares, vehicles, drivers, staff int
	err := tx.QueryRow(r.Context(), `SELECT t.status, t.legal_name, t.registration_no, t.contact_email, t.contact_phone, t.verified_at,
		(SELECT count(*) FROM service_zones WHERE tenant_id=t.id AND active),
		(SELECT count(*) FROM vehicle_categories WHERE tenant_id=t.id AND active),
		(SELECT count(*) FROM fare_rules WHERE tenant_id=t.id),
		(SELECT count(*) FROM vehicles WHERE tenant_id=t.id),
		(SELECT count(*) FROM drivers WHERE tenant_id=t.id),
		(SELECT count(*) FROM tenant_memberships WHERE tenant_id=t.id AND role NOT IN ('owner','driver','passenger'))
		FROM tenants t WHERE t.id=$1`, tenantID).Scan(&c.Status, &legal, &reg, &cemail, &cphone, &verifiedAt, &zones, &cats, &fares, &vehicles, &drivers, &staff)
	if err != nil {
		return c, err
	}
	c.Verified = verifiedAt != nil
	profile := checklistStep{Key: "profile", Required: true}
	for name, v := range map[string]*string{"legalName": legal, "registrationNo": reg, "contactEmail": cemail, "contactPhone": cphone} {
		if v == nil {
			profile.Missing = append(profile.Missing, name)
		}
	}
	sort.Strings(profile.Missing)
	profile.Done = len(profile.Missing) == 0
	c.Steps = []checklistStep{
		profile,
		{Key: "zone", Required: true, Done: zones > 0},
		{Key: "category", Required: true, Done: cats > 0},
		{Key: "fare", Required: true, Done: fares > 0},
		{Key: "vehicle", Done: vehicles > 0},
		{Key: "driver", Done: drivers > 0},
		{Key: "staff", Done: staff > 0},
		{Key: "verification", Required: true, Done: c.Verified},
		{Key: "live", Required: true, Done: c.Status == "active"},
	}
	c.CanSubmit = c.Status == "onboarding"
	for _, s := range c.Steps {
		if s.Required && !s.Done && s.Key != "verification" && s.Key != "live" {
			c.CanSubmit = false
		}
	}
	return c, nil
}

func (s *Service) getChecklist(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	c, err := buildChecklist(r, tx, m.TenantID)
	return Result{http.StatusOK, c}, err
}

// completeOnboarding is the owner's "publish" step. A tenant verified by platform staff goes live;
// a self-registered one enters the verification queue (D-021).
func (s *Service) completeOnboarding(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	if m.Role != "owner" {
		return Result{}, httpx.ErrForbidden
	}
	if m.TenantStatus != "onboarding" {
		return Result{}, httpx.Errorf(http.StatusConflict, "invalid_transition", "tenant is already %s", m.TenantStatus)
	}
	c, err := buildChecklist(r, tx, m.TenantID)
	if err != nil {
		return Result{}, err
	}
	if !c.CanSubmit {
		var missing []string
		for _, st := range c.Steps {
			if st.Required && !st.Done && st.Key != "verification" && st.Key != "live" {
				missing = append(missing, st.Key)
			}
		}
		return Result{}, httpx.Errorf(http.StatusConflict, "onboarding_incomplete", "complete these steps first: %s", strings.Join(missing, ", "))
	}
	next, action := "pending_verification", "tenant.submitted_for_verification"
	if c.Verified {
		next, action = "active", "tenant.onboarding_completed"
	}
	t, err := scanTenant(tx.QueryRow(r.Context(), `UPDATE tenants SET status=$2, submitted_at=now(), status_reason=NULL, updated_at=now()
		WHERE id=$1 AND status='onboarding' RETURNING `+tenantCols, m.TenantID, next))
	if err != nil {
		return Result{}, err
	}
	err = audit.Record(r.Context(), tx, audit.Event{TenantID: m.TenantID, ActorUserID: m.UserID,
		Action: action, TargetType: "tenant", TargetID: m.TenantID})
	return Result{http.StatusOK, map[string]any{"tenant": t, "checklist": c}}, err
}

// --- platform verification decision ---

func (s *Service) verifyTenant(r *http.Request, tx pgx.Tx, p identity.Principal) (Result, error) {
	var in struct {
		Decision string `json:"decision"` // approve | reject
		Note     string `json:"note"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	note := strings.TrimSpace(in.Note)
	if in.Decision != "approve" && in.Decision != "reject" {
		return Result{}, httpx.Errorf(http.StatusUnprocessableEntity, "invalid_decision", "decision must be approve or reject")
	}
	if in.Decision == "reject" && len(note) < 3 {
		return Result{}, httpx.Errorf(http.StatusUnprocessableEntity, "reason_required", "tell the owner what to fix")
	}
	tid := r.PathValue("tenantID")
	if !uuidRe.MatchString(tid) {
		return Result{}, httpx.ErrNotFound
	}
	var from string
	err := tx.QueryRow(r.Context(), `SELECT status FROM tenants WHERE id=$1 FOR UPDATE`, tid).Scan(&from)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, httpx.ErrNotFound
	}
	if err != nil {
		return Result{}, err
	}
	if from == "closed" {
		return Result{}, httpx.Errorf(http.StatusConflict, "invalid_transition", "tenant is closed")
	}
	var t TenantDTO
	if in.Decision == "approve" {
		t, err = scanTenant(tx.QueryRow(r.Context(), `UPDATE tenants SET verified_at=now(), verified_by=$2, verification_note=nullif($3,''),
			status = CASE WHEN status='pending_verification' THEN 'active' ELSE status END, status_reason=NULL, updated_at=now()
			WHERE id=$1 RETURNING `+tenantCols, tid, p.UserID, note))
	} else {
		t, err = scanTenant(tx.QueryRow(r.Context(), `UPDATE tenants SET verified_at=NULL, verified_by=NULL, verification_note=$2,
			status = CASE WHEN status='pending_verification' THEN 'onboarding' ELSE status END, status_reason=$2, updated_at=now()
			WHERE id=$1 RETURNING `+tenantCols, tid, note))
	}
	if err != nil {
		return Result{}, err
	}
	err = audit.Record(r.Context(), tx, audit.Event{TenantID: tid, ActorUserID: p.UserID, Action: map[string]string{"approve": "tenant.verification_approved", "reject": "tenant.verification_rejected"}[in.Decision],
		TargetType: "tenant", TargetID: tid, Reason: note, Metadata: map[string]any{"from": from, "to": t.Status}})
	return Result{http.StatusOK, map[string]any{"tenant": t}}, err
}
