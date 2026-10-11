package server_test

import (
	"testing"
)

// Prompt §18: runtime branding is tenant-owned, validated, public on the booking endpoint.
func TestBranding(t *testing.T) {
	f := operator(t)
	base := "/v1/tenants/" + f.ID
	if r := call(t, "PATCH", base+"/branding", f.OwnerToken, map[string]string{"color": "teal"}); errCode(r) != "invalid_color" {
		t.Fatalf("bad colour: %d %v", r.Status, r.Body)
	}
	if r := call(t, "PATCH", base+"/branding", f.OwnerToken, map[string]string{"websiteUrl": "javascript:alert(1)"}); errCode(r) != "invalid_website_url" {
		t.Fatalf("bad url: %d %v", r.Status, r.Body)
	}
	pdf := dig(uploadFile(t, base+"/files", f.OwnerToken, "x.pdf", pdfBytes).Body, "file", "id")
	if r := call(t, "PATCH", base+"/branding", f.OwnerToken, map[string]string{"logoFileId": pdf}); errCode(r) != "invalid_logo" {
		t.Fatalf("pdf logo: %d %v", r.Status, r.Body)
	}
	other := operator(t)
	foreign := dig(uploadFile(t, "/v1/tenants/"+other.ID+"/files", other.OwnerToken, "o.png", pngBytes).Body, "file", "id")
	if r := call(t, "PATCH", base+"/branding", f.OwnerToken, map[string]string{"logoFileId": foreign}); errCode(r) != "invalid_logo" {
		t.Fatalf("another tenant's file as logo: %d %v", r.Status, r.Body)
	}
	logo := dig(uploadFile(t, base+"/files", f.OwnerToken, "logo.png", pngBytes).Body, "file", "id")
	r := must(t, call(t, "PATCH", base+"/branding", f.OwnerToken, map[string]string{"color": "#FFD400", "logoFileId": logo, "tagline": "Kandy's friendliest cabs",
		"supportPhone": "0812223344", "supportEmail": "help@kandycabs.test", "termsUrl": "https://kandycabs.test/terms"}), 200, "branding")
	if dig(r.Body, "brand", "color") != "#ffd400" || dig(r.Body, "brand", "onColor") != "#000000" {
		t.Fatalf("yellow needs black text: %v", r.Body)
	}
	pub := must(t, call(t, "GET", "/v1/public/tenants/"+f.Slug, "", nil), 200, "public")
	if dig(pub.Body, "brand", "tagline") != "Kandy's friendliest cabs" || dig(pub.Body, "brand", "logoUrl") != "/v1/public/tenants/"+f.Slug+"/logo" || dig(pub.Body, "brand", "termsUrl") == "<nil>" {
		t.Fatalf("public brand: %v", pub.Body["brand"])
	}
	if s, ct, b := getRaw(t, "/v1/public/tenants/"+f.Slug+"/logo", ""); s != 200 || ct != "image/png" || len(b) != len(pngBytes) {
		t.Fatalf("public logo: %d %s %d", s, ct, len(b))
	}
	if s, _, _ := getRaw(t, "/v1/public/tenants/"+other.Slug+"/logo", ""); s != 404 {
		t.Fatalf("tenant without logo: %d", s)
	}
	must(t, call(t, "PATCH", base+"/branding", f.OwnerToken, map[string]string{"logoFileId": "", "color": "#0f766e"}), 200, "remove logo")
	if s, _, _ := getRaw(t, "/v1/public/tenants/"+f.Slug+"/logo", ""); s != 404 {
		t.Fatalf("logo removed: %d", s)
	}
	if r := call(t, "PATCH", base+"/branding", d2Token(t, f), map[string]string{"color": "#000000"}); r.Status != 403 {
		t.Fatalf("dispatcher cannot brand: %d", r.Status)
	}
}
