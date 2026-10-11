package server_test

import (
	"strings"
	"testing"
)

func register(t *testing.T, name string) (email, token, verifyToken string) {
	t.Helper()
	email = uniq("founder-") + "@taxios.test"
	r := must(t, call(t, "POST", "/v1/auth/register", "", map[string]string{"email": email, "fullName": name, "password": password}), 201, "register")
	if r.Body["emailVerified"] != false {
		t.Fatalf("new account should be unverified: %v", r.Body)
	}
	return email, r.Body["token"].(string), r.Body["verifyToken"].(string)
}

// Master prompt §8 + §9 (D-021): self-signup → verify email → draft business → checklist → submit →
// super admin rejects/approves → live booking page.
func TestSelfServiceOnboarding(t *testing.T) {
	email, token, verifyToken := register(t, "Founder")
	if r := call(t, "POST", "/v1/auth/register", "", map[string]string{"email": strings.ToUpper(email), "fullName": "Again", "password": password}); errCode(r) != "email_taken" {
		t.Fatalf("duplicate email (case-insensitive) should conflict: %d %v", r.Status, r.Body)
	}
	biz := map[string]string{"slug": uniq("self-"), "name": "Kandy Cabs", "businessType": "taxi"}
	if r := call(t, "POST", "/v1/businesses", token, biz); errCode(r) != "email_unverified" {
		t.Fatalf("unverified account must not create a business: %d %v", r.Status, r.Body)
	}
	must(t, call(t, "POST", "/v1/auth/verify-email", "", map[string]string{"token": verifyToken}), 200, "verify")
	if r := call(t, "POST", "/v1/auth/verify-email", "", map[string]string{"token": verifyToken}); errCode(r) != "invalid_token" {
		t.Fatalf("verification token must be single-use: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", "/v1/auth/resend-verification", token, nil); errCode(r) != "already_verified" {
		t.Fatalf("resend after verification: %d %v", r.Status, r.Body)
	}
	if r := call(t, "GET", "/v1/me", token, nil); dig(r.Body, "user", "emailVerified") != "true" {
		t.Fatalf("me should show verified: %v", r.Body)
	}

	r := must(t, call(t, "POST", "/v1/businesses", token, biz), 201, "create business")
	tid := dig(r.Body, "tenant", "id")
	if dig(r.Body, "tenant", "status") != "onboarding" || dig(r.Body, "tenant", "verifiedAt") != "<nil>" {
		t.Fatalf("self-service tenant should start unverified in onboarding: %v", r.Body)
	}
	if r := call(t, "POST", "/v1/businesses", token, biz); errCode(r) != "slug_taken" {
		t.Fatalf("slug reuse: %d %v", r.Status, r.Body)
	}
	me := call(t, "GET", "/v1/me", token, nil)
	if ms := me.Body["memberships"].([]any); len(ms) != 1 || ms[0].(map[string]any)["role"] != "owner" || ms[0].(map[string]any)["tenantId"] != tid {
		t.Fatalf("founder should be owner: %v", me.Body)
	}
	base := "/v1/tenants/" + tid
	cl := must(t, call(t, "GET", base+"/onboarding", token, nil), 200, "checklist")
	if cl.Body["canSubmit"] != false {
		t.Fatalf("empty business must not be submittable: %v", cl.Body)
	}
	if r := call(t, "POST", base+"/onboarding/complete", token, nil); errCode(r) != "onboarding_incomplete" {
		t.Fatalf("submit incomplete: %d %v", r.Status, r.Body)
	}
	if r := call(t, "PATCH", base, token, map[string]string{"contactPhone": "abc"}); errCode(r) != "invalid_contact_phone" {
		t.Fatalf("bad phone: %d %v", r.Status, r.Body)
	}
	if r := call(t, "PATCH", base, token, map[string]string{"timezone": "Mars/Olympus"}); errCode(r) != "invalid_timezone" {
		t.Fatalf("bad timezone: %d %v", r.Status, r.Body)
	}
	r = must(t, call(t, "PATCH", base, token, map[string]string{"legalName": "Kandy Cabs (Pvt) Ltd", "registrationNo": "PV 777",
		"contactEmail": "ops@kandycabs.test", "contactPhone": "0812345678", "language": "si", "timezone": "Asia/Colombo", "currency": "lkr", "taxNo": ""}), 200, "profile")
	if dig(r.Body, "tenant", "currency") != "LKR" || dig(r.Body, "tenant", "language") != "si" || dig(r.Body, "tenant", "taxNo") != "<nil>" {
		t.Fatalf("profile not stored as expected: %v", r.Body)
	}
	f := tenantFixture{ID: tid, OwnerToken: token}
	configurePricing(t, &f)
	cl = must(t, call(t, "GET", base+"/onboarding", token, nil), 200, "checklist")
	if cl.Body["canSubmit"] != true || cl.Body["verified"] != false {
		t.Fatalf("ready business should be submittable and unverified: %v", cl.Body)
	}

	r = must(t, call(t, "POST", base+"/onboarding/complete", token, nil), 200, "submit")
	if dig(r.Body, "tenant", "status") != "pending_verification" {
		t.Fatalf("self-service submit must queue for verification: %v", r.Body)
	}
	if r := call(t, "GET", "/v1/public/tenants/"+biz["slug"], "", nil); r.Status != 404 {
		t.Fatalf("unverified business must not be public: %d", r.Status)
	}
	// The owner can keep configuring while waiting.
	must(t, call(t, "POST", base+"/vehicle-categories", token, map[string]any{"code": "van", "name": "Van", "seats": 8}), 201, "write while pending")

	admin := superAdminToken(t)
	queue := must(t, call(t, "GET", "/v1/platform/tenants?status=pending_verification", admin, nil), 200, "queue")
	found := false
	for _, x := range queue.Body["tenants"].([]any) {
		found = found || x.(map[string]any)["id"] == tid
	}
	if !found {
		t.Fatalf("submitted business missing from the verification queue")
	}
	if r := call(t, "POST", "/v1/platform/tenants/"+tid+"/verification", token, map[string]string{"decision": "approve"}); r.Status != 403 {
		t.Fatalf("owner must not verify their own business: %d", r.Status)
	}
	if r := call(t, "POST", "/v1/platform/tenants/"+tid+"/verification", admin, map[string]string{"decision": "reject"}); errCode(r) != "reason_required" {
		t.Fatalf("reject needs a note: %d %v", r.Status, r.Body)
	}
	r = must(t, call(t, "POST", "/v1/platform/tenants/"+tid+"/verification", admin, map[string]string{"decision": "reject", "note": "registration number does not match the certificate"}), 200, "reject")
	if dig(r.Body, "tenant", "status") != "onboarding" || !strings.Contains(dig(r.Body, "tenant", "statusReason"), "registration") {
		t.Fatalf("reject should return the business to onboarding with the note: %v", r.Body)
	}
	must(t, call(t, "PATCH", base, token, map[string]string{"registrationNo": "PV 778"}), 200, "fix")
	must(t, call(t, "POST", base+"/onboarding/complete", token, nil), 200, "resubmit")
	r = must(t, call(t, "POST", "/v1/platform/tenants/"+tid+"/verification", admin, map[string]string{"decision": "approve", "note": "certificate checked"}), 200, "approve")
	if dig(r.Body, "tenant", "status") != "active" || dig(r.Body, "tenant", "verifiedAt") == "<nil>" {
		t.Fatalf("approve should activate: %v", r.Body)
	}
	must(t, call(t, "GET", "/v1/public/tenants/"+biz["slug"], "", nil), 200, "public after approval")

	r = call(t, "GET", base+"/audit-events", token, nil)
	want := map[string]bool{"tenant.created": false, "tenant.submitted_for_verification": false, "tenant.verification_rejected": false, "tenant.verification_approved": false}
	for _, e := range r.Body["events"].([]any) {
		want[e.(map[string]any)["action"].(string)] = true
	}
	for action, seen := range want {
		if !seen {
			t.Errorf("audit trail missing %s", action)
		}
	}
}

func TestAdminCreatedTenantIsTrusted(t *testing.T) {
	f := newTenant(t, superAdminToken(t))
	r := must(t, call(t, "GET", "/v1/tenants/"+f.ID, f.OwnerToken, nil), 200, "get")
	if dig(r.Body, "tenant", "verifiedAt") == "<nil>" {
		t.Fatalf("tenants created by platform staff are verified: %v", r.Body)
	}
	configurePricing(t, &f)
	completeProfile(t, f)
	r = must(t, call(t, "POST", "/v1/tenants/"+f.ID+"/onboarding/complete", f.OwnerToken, nil), 200, "publish")
	if dig(r.Body, "tenant", "status") != "active" {
		t.Fatalf("verified tenant publishes straight to active: %v", r.Body)
	}
}

func TestDraftBusinessLimit(t *testing.T) {
	_, token, verifyToken := register(t, "Serial Founder")
	must(t, call(t, "POST", "/v1/auth/verify-email", "", map[string]string{"token": verifyToken}), 200, "verify")
	for i := 0; i < 3; i++ {
		must(t, call(t, "POST", "/v1/businesses", token, map[string]string{"slug": uniq("draft-"), "name": "Draft", "businessType": "taxi"}), 201, "draft")
	}
	if r := call(t, "POST", "/v1/businesses", token, map[string]string{"slug": uniq("draft-"), "name": "Draft", "businessType": "taxi"}); errCode(r) != "too_many_drafts" {
		t.Fatalf("4th draft should be refused: %d %v", r.Status, r.Body)
	}
}

func TestPasswordReset(t *testing.T) {
	email, token, _ := register(t, "Forgetful")
	if r := must(t, call(t, "POST", "/v1/auth/forgot-password", "", map[string]string{"email": uniq("nobody-") + "@taxios.test"}), 200, "forgot unknown"); r.Body["resetToken"] != nil {
		t.Fatalf("unknown email must not get a token: %v", r.Body)
	}
	r := must(t, call(t, "POST", "/v1/auth/forgot-password", "", map[string]string{"email": email}), 200, "forgot")
	reset := r.Body["resetToken"].(string)
	if r := call(t, "POST", "/v1/auth/reset-password", "", map[string]string{"token": "txr_bogus", "password": "another-good-password"}); errCode(r) != "invalid_token" {
		t.Fatalf("bogus reset token: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", "/v1/auth/reset-password", "", map[string]string{"token": reset, "password": "short"}); errCode(r) != "weak_password" {
		t.Fatalf("weak password must be rejected before the token is spent: %d %v", r.Status, r.Body)
	}
	r = must(t, call(t, "POST", "/v1/auth/reset-password", "", map[string]string{"token": reset, "password": "another-good-password"}), 200, "reset")
	fresh := r.Body["token"].(string)
	if r := call(t, "GET", "/v1/me", token, nil); r.Status != 401 {
		t.Fatalf("old session must be revoked after reset: %d", r.Status)
	}
	if r := call(t, "GET", "/v1/me", fresh, nil); r.Status != 200 || dig(r.Body, "user", "emailVerified") != "true" {
		t.Fatalf("reset proves email ownership: %d %v", r.Status, r.Body)
	}
	if r := call(t, "POST", "/v1/auth/login", "", map[string]string{"email": email, "password": password}); r.Status != 401 {
		t.Fatalf("old password must stop working: %d", r.Status)
	}
	must(t, call(t, "POST", "/v1/auth/login", "", map[string]string{"email": email, "password": "another-good-password"}), 200, "login new password")
	if r := call(t, "POST", "/v1/auth/reset-password", "", map[string]string{"token": reset, "password": "yet-another-password"}); errCode(r) != "invalid_token" {
		t.Fatalf("reset token must be single-use: %d %v", r.Status, r.Body)
	}
}
