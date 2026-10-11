package server_test

import (
	"testing"
	"time"
)

func TestPassengerAppSignupIsMobileSession(t *testing.T) {
	f := operator(t)
	r := must(t, call(t, "POST", "/v1/public/tenants/"+f.Slug+"/passengers", "", map[string]any{"fullName": "App Rider",
		"email": uniq("app-") + "@taxios.test", "phone": phone(), "password": password, "client": "mobile"}), 201, "app signup")
	list := must(t, call(t, "GET", "/v1/auth/sessions", r.Body["token"].(string), nil), 200, "sessions")
	if s := list.Body["sessions"].([]any)[0].(map[string]any); s["client"] != "mobile" {
		t.Fatalf("app sign-up should open a mobile session: %v", s)
	}
}

func TestSessionManagement(t *testing.T) {
	email, web, _ := register(t, "Multi Device")
	r := must(t, call(t, "POST", "/v1/auth/login", "", map[string]string{"email": email, "password": password, "client": "mobile"}), 200, "mobile login")
	mobile := r.Body["token"].(string)
	must(t, call(t, "POST", "/v1/auth/login", "", map[string]string{"email": email, "password": password}), 200, "second web login")

	list := must(t, call(t, "GET", "/v1/auth/sessions", mobile, nil), 200, "sessions")
	sessions := list.Body["sessions"].([]any)
	if len(sessions) != 3 {
		t.Fatalf("expected 3 live sessions, got %d", len(sessions))
	}
	var currentID, otherID string
	for _, x := range sessions {
		s := x.(map[string]any)
		if s["current"] == true {
			currentID = s["id"].(string)
			if s["client"] != "mobile" {
				t.Fatalf("current session should be mobile: %v", s)
			}
			exp, _ := time.Parse(time.RFC3339Nano, s["expiresAt"].(string))
			if time.Until(exp) < 170*24*time.Hour {
				t.Fatalf("mobile session should live ~180 days, expires %v", exp)
			}
		} else if otherID == "" {
			otherID = s["id"].(string)
		}
	}
	// Revoke one specific session; can't revoke another user's.
	_, stranger, _ := register(t, "Stranger")
	if r := call(t, "DELETE", "/v1/auth/sessions/"+otherID, stranger, nil); r.Status != 404 {
		t.Fatalf("revoking someone else's session: %d", r.Status)
	}
	if r := call(t, "DELETE", "/v1/auth/sessions/"+otherID, mobile, nil); r.Status != 204 {
		t.Fatalf("revoke own other session: %d %v", r.Status, r.Body)
	}
	// Sign out everywhere else keeps the current session alive.
	r = must(t, call(t, "POST", "/v1/auth/logout-all", mobile, nil), 200, "logout-all")
	if r.Body["revoked"].(float64) != 1 {
		t.Fatalf("expected 1 more session revoked, got %v", r.Body)
	}
	if r := call(t, "GET", "/v1/me", web, nil); r.Status != 401 {
		t.Fatalf("web session should be revoked: %d", r.Status)
	}
	must(t, call(t, "GET", "/v1/me", mobile, nil), 200, "mobile still signed in")
	list = must(t, call(t, "GET", "/v1/auth/sessions", mobile, nil), 200, "sessions after")
	if n := len(list.Body["sessions"].([]any)); n != 1 || list.Body["sessions"].([]any)[0].(map[string]any)["id"] != currentID {
		t.Fatalf("only the current session should remain, got %d", n)
	}
}
