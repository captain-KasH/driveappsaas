package server_test

import (
	"strings"
	"testing"
)

func TestStaffRemoval(t *testing.T) {
	f := operator(t)
	base := "/v1/tenants/" + f.ID
	invite := func(role string) (userID, token string) {
		r := must(t, call(t, "POST", base+"/invitations", f.OwnerToken, map[string]string{"email": uniq(role+"-") + "@taxios.test", "role": role}), 201, "invite")
		r = must(t, call(t, "POST", "/v1/auth/accept-invite", "", map[string]string{"token": dig(r.Body, "invitation", "token"), "fullName": "Staff", "password": password}), 200, "accept")
		token = r.Body["token"].(string)
		me := call(t, "GET", "/v1/me", token, nil)
		return dig(me.Body, "user", "id"), token
	}
	dispID, dispToken := invite("dispatcher")
	adminID, adminToken := invite("admin")
	owner2ID, _ := invite("owner")

	if r := call(t, "DELETE", base+"/members/"+adminID, adminToken, nil); errCode(r) != "cannot_remove_self" {
		t.Fatalf("self removal: %d %v", r.Status, r.Body)
	}
	if r := call(t, "DELETE", base+"/members/"+owner2ID, adminToken, nil); r.Status != 403 {
		t.Fatalf("admin removing an owner should be 403: %d", r.Status)
	}
	if r := call(t, "DELETE", base+"/members/"+dispID, dispToken, nil); r.Status != 403 {
		t.Fatalf("dispatcher lacks staff.manage: %d", r.Status)
	}
	must(t, call(t, "DELETE", base+"/members/"+dispID, adminToken, nil), 200, "admin removes dispatcher")
	if r := call(t, "GET", base, dispToken, nil); r.Status != 401 && r.Status != 404 {
		t.Fatalf("removed dispatcher should lose access (401 session revoked or 404): %d", r.Status)
	}
	must(t, call(t, "DELETE", base+"/members/"+owner2ID, f.OwnerToken, nil), 200, "owner removes second owner")
	me := call(t, "GET", "/v1/me", f.OwnerToken, nil)
	ownerID := dig(me.Body, "user", "id")
	if r := call(t, "DELETE", base+"/members/"+ownerID, f.OwnerToken, nil); errCode(r) != "cannot_remove_self" {
		t.Fatalf("last owner self: %d %v", r.Status, r.Body)
	}
	d := addDriver(t, f, [2]float64{6.93, 79.85}, "2030-01-01")
	dm := call(t, "GET", "/v1/me", d.Token, nil)
	if r := call(t, "DELETE", base+"/members/"+dig(dm.Body, "user", "id"), f.OwnerToken, nil); errCode(r) != "not_staff" {
		t.Fatalf("drivers are not removed here: %d %v", r.Status, r.Body)
	}
	r := call(t, "GET", base+"/audit-events", f.OwnerToken, nil)
	n := 0
	for _, e := range r.Body["events"].([]any) {
		if e.(map[string]any)["action"] == "member.removed" {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("expected 2 member.removed audit events, got %d", n)
	}
}

func TestCustomersAndExports(t *testing.T) {
	f := operator(t)
	base := "/v1/tenants/" + f.ID
	pax := addPassenger(t, f)
	book(t, f, pax)
	must(t, callH(t, "POST", base+"/bookings", f.OwnerToken, map[string]string{"Idempotency-Key": uniq("idem-")}, map[string]any{
		"customer":   map[string]string{"fullName": "Walk In", "phone": phone()},
		"categoryId": f.CarID, "pickup": place([2]float64{6.9271, 79.8612}, "Fort"), "dropoff": place([2]float64{6.9, 79.86}, "Galle Face")},
	), 201, "phone booking")

	r := must(t, call(t, "GET", base+"/customers", f.OwnerToken, nil), 200, "customers")
	list := r.Body["customers"].([]any)
	if len(list) != 2 {
		t.Fatalf("expected 2 customers, got %d: %v", len(list), list)
	}
	first := list[0].(map[string]any)
	if first["bookings"].(float64) != 1 {
		t.Fatalf("customer aggregate: %v", first)
	}
	r = must(t, call(t, "GET", base+"/customers?q=walk", f.OwnerToken, nil), 200, "search")
	if len(r.Body["customers"].([]any)) != 1 {
		t.Fatalf("search should match one: %v", r.Body)
	}
	other := operator(t)
	if r := call(t, "GET", base+"/customers", other.OwnerToken, nil); r.Status != 404 {
		t.Fatalf("other tenant: %d", r.Status)
	}
	if r := call(t, "GET", base+"/customers", pax, nil); r.Status != 403 {
		t.Fatalf("passenger listing customers: %d", r.Status)
	}

	status, ct, body := getRaw(t, base+"/exports/customers.csv", f.OwnerToken)
	if status != 200 || !strings.HasPrefix(ct, "text/csv") || !strings.HasPrefix(string(body), "id,name,phone,") || strings.Count(string(body), "\n") != 3 {
		t.Fatalf("customers csv: %d %s\n%s", status, ct, body)
	}
	status, _, body = getRaw(t, base+"/exports/bookings.csv", f.OwnerToken)
	if status != 200 || !strings.Contains(string(body), "Walk In") || strings.Count(string(body), "\n") != 3 {
		t.Fatalf("bookings csv: %d\n%s", status, body)
	}
	if s, _, _ := getRaw(t, base+"/exports/bookings.csv", d2Token(t, f)); s != 403 {
		t.Fatalf("dispatcher lacks reports.read: %d", s)
	}
}

// d2Token returns a dispatcher session for f.
func d2Token(t *testing.T, f tenantFixture) string {
	t.Helper()
	r := must(t, call(t, "POST", "/v1/tenants/"+f.ID+"/invitations", f.OwnerToken, map[string]string{"email": uniq("disp-") + "@taxios.test", "role": "dispatcher"}), 201, "invite")
	r = must(t, call(t, "POST", "/v1/auth/accept-invite", "", map[string]string{"token": dig(r.Body, "invitation", "token"), "fullName": "Disp", "password": password}), 200, "accept")
	return r.Body["token"].(string)
}
