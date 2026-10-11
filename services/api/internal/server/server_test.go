package server_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/booking"
	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/identity"
	"github.com/captain-KasH/driveappsaas/services/api/internal/server"
)

// Integration tests need a real Postgres: TAXIOS_TEST_MIGRATE_URL (owner) and
// TAXIOS_TEST_DATABASE_URL (taxios_app role). `make test` sets both up.
var (
	ownerPool *pgxpool.Pool
	appPool   *pgxpool.Pool
	srv       *httptest.Server
	bookings  *booking.Service
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	migrateURL, appURL := os.Getenv("TAXIOS_TEST_MIGRATE_URL"), os.Getenv("TAXIOS_TEST_DATABASE_URL")
	if migrateURL == "" || appURL == "" {
		fmt.Fprintln(os.Stderr, "FAIL: TAXIOS_TEST_MIGRATE_URL and TAXIOS_TEST_DATABASE_URL are required (run `make test`)")
		os.Exit(1)
	}
	var err error
	if ownerPool, err = db.Open(ctx, migrateURL); err != nil {
		panic(err)
	}
	if _, err = db.Migrate(ctx, ownerPool); err != nil {
		panic(err)
	}
	if appPool, err = db.Open(ctx, appURL); err != nil {
		panic(err)
	}
	var bypass bool
	appPool.QueryRow(ctx, `SELECT rolsuper OR rolbypassrls FROM pg_roles WHERE rolname = current_user`).Scan(&bypass)
	if bypass {
		panic("TAXIOS_TEST_DATABASE_URL must not bypass RLS, or isolation tests prove nothing")
	}
	os.Setenv("TAXIOS_DEV_TOKENS", "1") // tokens come back in responses so flows complete without a mail provider
	bookings = server.NewBookings(appPool)
	srv = httptest.NewServer(server.New(appPool, bookings))
	code := m.Run()
	srv.Close()
	os.Exit(code)
}

func uniq(prefix string) string {
	b := make([]byte, 4)
	rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

type resp struct {
	Status int
	Body   map[string]any
}

func call(t *testing.T, method, path, token string, body any) resp {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, srv.URL+path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	out := map[string]any{}
	json.NewDecoder(res.Body).Decode(&out)
	return resp{res.StatusCode, out}
}

func errCode(r resp) string {
	e, _ := r.Body["error"].(map[string]any)
	s, _ := e["code"].(string)
	return s
}

func dig(m map[string]any, keys ...string) string {
	var cur any = m
	for _, k := range keys {
		cur = cur.(map[string]any)[k]
	}
	return fmt.Sprint(cur)
}

const password = "correct-horse-battery"

func superAdminToken(t *testing.T) string {
	t.Helper()
	email := uniq("admin-") + "@taxios.test"
	hash, _ := identity.HashPassword(password)
	if _, err := ownerPool.Exec(context.Background(), `INSERT INTO users(email, full_name, password_hash, platform_role)
		VALUES ($1,'Platform Admin',$2,'super_admin')`, email, hash); err != nil {
		t.Fatal(err)
	}
	r := call(t, "POST", "/v1/auth/login", "", map[string]string{"email": email, "password": password})
	if r.Status != 200 {
		t.Fatalf("admin login: %d %v", r.Status, r.Body)
	}
	return r.Body["token"].(string)
}

type tenantFixture struct{ ID, Slug, OwnerToken, OwnerEmail, ZoneID, CarID, TukID string }

// newTenant provisions a tenant as a super admin and has its owner accept the invite.
func newTenant(t *testing.T, admin string) tenantFixture {
	t.Helper()
	f := tenantFixture{OwnerEmail: uniq("owner-") + "@taxios.test"}
	f.Slug = uniq("fleet-")
	r := call(t, "POST", "/v1/platform/tenants", admin, map[string]string{
		"slug": f.Slug, "name": "Colombo Cabs", "businessType": "taxi", "ownerEmail": f.OwnerEmail})
	if r.Status != 201 {
		t.Fatalf("create tenant: %d %v", r.Status, r.Body)
	}
	f.ID = dig(r.Body, "tenant", "id")
	if s := dig(r.Body, "tenant", "status"); s != "onboarding" {
		t.Fatalf("new tenant status = %s, want onboarding", s)
	}
	r = call(t, "POST", "/v1/auth/accept-invite", "", map[string]string{
		"token": dig(r.Body, "ownerInvitation", "token"), "fullName": "Owner", "password": password})
	if r.Status != 200 {
		t.Fatalf("accept invite: %d %v", r.Status, r.Body)
	}
	f.OwnerToken = r.Body["token"].(string)
	return f
}

func TestOnboardingJourney(t *testing.T) {
	admin := superAdminToken(t)
	a := newTenant(t, admin)

	me := call(t, "GET", "/v1/me", a.OwnerToken, nil)
	ms := me.Body["memberships"].([]any)
	if me.Status != 200 || len(ms) != 1 || ms[0].(map[string]any)["role"] != "owner" {
		t.Fatalf("me: %d %v", me.Status, me.Body)
	}

	if r := call(t, "PATCH", "/v1/tenants/"+a.ID, a.OwnerToken, map[string]string{"name": "Colombo Cabs Ltd"}); r.Status != 200 {
		t.Fatalf("rename: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", "/v1/tenants/"+a.ID+"/onboarding/complete", a.OwnerToken, nil); errCode(r) != "onboarding_incomplete" {
		t.Fatalf("onboarding without pricing should be refused: %d %v", r.Status, r.Body)
	}
	configurePricing(t, &a)
	if r := call(t, "POST", "/v1/tenants/"+a.ID+"/onboarding/complete", a.OwnerToken, nil); errCode(r) != "onboarding_incomplete" {
		t.Fatalf("onboarding without a business profile should be refused: %d %v", r.Status, r.Body)
	}
	completeProfile(t, a)
	r := call(t, "POST", "/v1/tenants/"+a.ID+"/onboarding/complete", a.OwnerToken, nil)
	if r.Status != 200 || dig(r.Body, "tenant", "status") != "active" {
		t.Fatalf("complete onboarding: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", "/v1/tenants/"+a.ID+"/onboarding/complete", a.OwnerToken, nil); errCode(r) != "invalid_transition" {
		t.Fatalf("second completion should conflict, got %d %v", r.Status, r.Body)
	}

	// Owner invites a dispatcher, who can read but not manage.
	r = call(t, "POST", "/v1/tenants/"+a.ID+"/invitations", a.OwnerToken, map[string]string{"email": uniq("disp-") + "@taxios.test", "role": "dispatcher"})
	if r.Status != 201 {
		t.Fatalf("invite: %d %v", r.Status, r.Body)
	}
	inv := dig(r.Body, "invitation", "token")
	r = call(t, "POST", "/v1/auth/accept-invite", "", map[string]string{"token": inv, "fullName": "Disp", "password": password})
	disp := r.Body["token"].(string)
	if r := call(t, "GET", "/v1/tenants/"+a.ID, disp, nil); r.Status != 200 || r.Body["myRole"] != "dispatcher" {
		t.Fatalf("dispatcher read: %d %v", r.Status, r.Body)
	}
	if r := call(t, "PATCH", "/v1/tenants/"+a.ID, disp, map[string]string{"name": "Hijack"}); r.Status != 403 {
		t.Fatalf("dispatcher rename should be 403, got %d", r.Status)
	}
	if r := call(t, "POST", "/v1/tenants/"+a.ID+"/invitations", disp, map[string]string{"email": "x@taxios.test", "role": "owner"}); r.Status != 403 {
		t.Fatalf("dispatcher invite should be 403, got %d", r.Status)
	}
	// Invitations are single-use.
	if r := call(t, "POST", "/v1/auth/accept-invite", "", map[string]string{"token": inv, "fullName": "Again", "password": password}); errCode(r) != "invalid_invitation" {
		t.Fatalf("reused invite: %d %v", r.Status, r.Body)
	}

	// Every step above is in the tenant's audit trail.
	r = call(t, "GET", "/v1/tenants/"+a.ID+"/audit-events", a.OwnerToken, nil)
	want := map[string]bool{"zone.created": false, "fare_rule.published": false, "tenant.created": false, "invitation.accepted": false, "tenant.updated": false, "tenant.onboarding_completed": false, "invitation.created": false}
	for _, e := range r.Body["events"].([]any) {
		want[e.(map[string]any)["action"].(string)] = true
	}
	for action, seen := range want {
		if !seen {
			t.Errorf("audit trail missing %s", action)
		}
	}
}

// Mandatory scenario 1: a tenant cannot access another tenant's data.
func TestTenantIsolationAPI(t *testing.T) {
	admin := superAdminToken(t)
	a, b := newTenant(t, admin), newTenant(t, admin)
	for _, path := range []string{"", "/members", "/audit-events"} {
		if r := call(t, "GET", "/v1/tenants/"+a.ID+path, b.OwnerToken, nil); r.Status != 404 {
			t.Errorf("tenant B reading A%s: got %d, want 404", path, r.Status)
		}
	}
	if r := call(t, "PATCH", "/v1/tenants/"+a.ID, b.OwnerToken, map[string]string{"name": "pwned"}); r.Status != 404 {
		t.Errorf("tenant B renaming A: got %d, want 404", r.Status)
	}
	if r := call(t, "GET", "/v1/platform/tenants", b.OwnerToken, nil); r.Status != 403 {
		t.Errorf("tenant owner listing platform tenants: got %d, want 403", r.Status)
	}
	if r := call(t, "GET", "/v1/tenants/not-a-uuid", b.OwnerToken, nil); r.Status != 404 {
		t.Errorf("malformed tenant id: got %d, want 404", r.Status)
	}
	if r := call(t, "GET", "/v1/tenants/"+a.ID, "", nil); r.Status != 401 {
		t.Errorf("anonymous: got %d, want 401", r.Status)
	}
}

// RLS must hold even if application code forgets a WHERE clause.
func TestTenantIsolationRLS(t *testing.T) {
	admin := superAdminToken(t)
	a, b := newTenant(t, admin), newTenant(t, admin)
	ctx := context.Background()
	err := db.InTx(ctx, appPool, db.Scope{TenantID: b.ID}, func(tx pgx.Tx) error {
		for _, table := range []string{"tenants", "tenant_memberships", "tenant_invitations", "audit_events"} {
			col := "tenant_id"
			if table == "tenants" {
				col = "id"
			}
			var n int
			if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s = $1`, table, col), a.ID).Scan(&n); err != nil {
				return err
			}
			if n != 0 {
				t.Errorf("scope B sees %d rows of tenant A in %s", n, table)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.InTx(ctx, appPool, db.Scope{TenantID: b.ID}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO audit_events(tenant_id, action, target_type, target_id) VALUES ($1,'x','x','x')`, a.ID)
		return err
	})
	if err == nil {
		t.Error("scope B wrote an audit event into tenant A")
	}
	// No scope at all sees nothing.
	db.InTx(ctx, appPool, db.Scope{}, func(tx pgx.Tx) error {
		var n int
		tx.QueryRow(ctx, `SELECT count(*) FROM tenant_memberships`).Scan(&n)
		if n != 0 {
			t.Errorf("unscoped transaction sees %d memberships", n)
		}
		return nil
	})
}

func TestAuditIsAppendOnly(t *testing.T) {
	admin := superAdminToken(t)
	a := newTenant(t, admin)
	ctx := context.Background()
	err := db.InTx(ctx, appPool, db.Scope{TenantID: a.ID}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE audit_events SET action='tampered' WHERE tenant_id=$1`, a.ID)
		return err
	})
	if err == nil {
		t.Fatal("app role updated audit_events")
	}
	if _, err := ownerPool.Exec(ctx, `DELETE FROM audit_events WHERE tenant_id=$1`, a.ID); err == nil {
		t.Fatal("even the owner must not delete audit events (trigger)")
	}
}

func TestSuspensionAndClosure(t *testing.T) {
	admin := superAdminToken(t)
	a := newTenant(t, admin)
	path := "/v1/platform/tenants/" + a.ID + "/status"
	if r := call(t, "POST", path, admin, map[string]string{"status": "suspended"}); errCode(r) != "reason_required" {
		t.Fatalf("suspend without reason: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", path, a.OwnerToken, map[string]string{"status": "suspended", "reason": "self"}); r.Status != 403 {
		t.Fatalf("owner suspending own tenant via platform route: %d", r.Status)
	}
	if r := call(t, "POST", path, admin, map[string]string{"status": "suspended", "reason": "unpaid invoice"}); r.Status != 200 {
		t.Fatalf("suspend: %d %v", r.Status, r.Body)
	}
	if r := call(t, "GET", "/v1/tenants/"+a.ID, a.OwnerToken, nil); r.Status != 200 {
		t.Fatalf("suspended tenant should stay readable, got %d", r.Status)
	}
	if r := call(t, "PATCH", "/v1/tenants/"+a.ID, a.OwnerToken, map[string]string{"name": "x"}); errCode(r) != "tenant_inactive" {
		t.Fatalf("write on suspended tenant: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", path, admin, map[string]string{"status": "onboarding", "reason": "undo"}); errCode(r) != "invalid_transition" {
		t.Fatalf("suspended->onboarding should be invalid: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", path, admin, map[string]string{"status": "closed", "reason": "customer request"}); r.Status != 200 {
		t.Fatalf("close: %d %v", r.Status, r.Body)
	}
	if r := call(t, "GET", "/v1/me", a.OwnerToken, nil); r.Status != 401 {
		t.Fatalf("closed tenant's sole-tenant owner should be signed out, got %d", r.Status)
	}
}

func TestAuthSessions(t *testing.T) {
	admin := superAdminToken(t)
	a := newTenant(t, admin)
	if r := call(t, "POST", "/v1/auth/login", "", map[string]string{"email": a.OwnerEmail, "password": "wrong-password-1"}); errCode(r) != "invalid_credentials" {
		t.Fatalf("bad password: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", "/v1/auth/login", "", map[string]string{"email": "nobody@taxios.test", "password": password}); errCode(r) != "invalid_credentials" {
		t.Fatalf("unknown email must look identical: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", "/v1/auth/logout", a.OwnerToken, nil); r.Status != 204 {
		t.Fatalf("logout: %d", r.Status)
	}
	if r := call(t, "GET", "/v1/me", a.OwnerToken, nil); r.Status != 401 {
		t.Fatalf("revoked token still works: %d", r.Status)
	}
	if r := call(t, "GET", "/v1/me", "txs_forged", nil); r.Status != 401 {
		t.Fatalf("forged token: %d", r.Status)
	}
}

func TestInvalidTenantInput(t *testing.T) {
	admin := superAdminToken(t)
	if r := call(t, "POST", "/v1/platform/tenants", admin, map[string]string{"slug": "Bad Slug!", "name": "x", "businessType": "taxi", "ownerEmail": "o@taxios.test"}); errCode(r) != "invalid_tenant" {
		t.Fatalf("bad slug: %d %v", r.Status, r.Body)
	}
	slug := uniq("dup-")
	body := map[string]string{"slug": slug, "name": "x", "businessType": "taxi", "ownerEmail": "o@taxios.test"}
	call(t, "POST", "/v1/platform/tenants", admin, body)
	if r := call(t, "POST", "/v1/platform/tenants", admin, body); errCode(r) != "slug_taken" {
		t.Fatalf("duplicate slug: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", "/v1/platform/tenants", admin, map[string]any{"slug": uniq("x-"), "surprise": true}); errCode(r) != "invalid_json" {
		t.Fatalf("unknown field: %d %v", r.Status, r.Body)
	}
}
