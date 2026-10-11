package server_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

var (
	pdfBytes = []byte("%PDF-1.4\n1 0 obj << /Type /Catalog >> endobj\ntrailer << /Root 1 0 R >>\n%%EOF\n")
	pngBytes = append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 64)...)
)

// uploadFile posts one multipart file and returns the raw response.
func uploadFile(t *testing.T, path, token, name string, data []byte) resp {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", name)
	fw.Write(data)
	w.Close()
	req, _ := http.NewRequest("POST", srv.URL+path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	out := map[string]any{}
	decodeJSON(res.Body, &out)
	return resp{res.StatusCode, out}
}

func getRaw(t *testing.T, path, token string) (int, string, []byte) {
	t.Helper()
	req, _ := http.NewRequest("GET", srv.URL+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, res.Header.Get("Content-Type"), b
}

// Prompt §11/§30: document files are tenant-scoped, type-sniffed, size-capped and only visible to
// the uploader, the driver they belong to, fleet staff, and (audited) platform staff.
func TestFileUploadsAndAccess(t *testing.T) {
	f := operator(t)
	base := "/v1/tenants/" + f.ID
	d := addDriver(t, f, [2]float64{6.93, 79.85}, "2030-01-01")
	other := operator(t)
	pax := addPassenger(t, f)

	if r := uploadFile(t, base+"/files", d.Token, "licence.txt", []byte("hello, not an image")); errCode(r) != "unsupported_file_type" {
		t.Fatalf("text upload should be refused: %d %v", r.Status, r.Body)
	}
	if r := uploadFile(t, base+"/files", d.Token, "big.png", append(pngBytes, bytes.Repeat([]byte{1}, 10<<20)...)); r.Status != 413 {
		t.Fatalf("11 MB upload should be 413: %d %v", r.Status, r.Body)
	}
	if r := uploadFile(t, base+"/files", pax, "id.png", pngBytes); r.Status != 403 {
		t.Fatalf("passengers cannot upload: %d", r.Status)
	}
	r := uploadFile(t, base+"/files", d.Token, "licence.png", pngBytes)
	if r.Status != 201 || dig(r.Body, "file", "contentType") != "image/png" {
		t.Fatalf("driver upload: %d %v", r.Status, r.Body)
	}
	fileID := dig(r.Body, "file", "id")

	// Attach to a document; the driver may only attach their own uploads.
	staffFile := dig(uploadFile(t, base+"/files", f.OwnerToken, "ins.pdf", pdfBytes).Body, "file", "id")
	if r := call(t, "POST", base+"/driver/documents", d.Token, map[string]any{"docType": "police_report", "number": "PR-1", "fileId": staffFile}); errCode(r) != "invalid_file" {
		t.Fatalf("driver attaching a staff upload: %d %v", r.Status, r.Body)
	}
	must(t, call(t, "POST", base+"/driver/documents", d.Token, map[string]any{"docType": "police_report", "number": "PR-1", "fileId": fileID}), 201, "attach own file")
	me := must(t, call(t, "GET", base+"/driver/me", d.Token, nil), 200, "driver me")
	found := false
	for _, x := range me.Body["documents"].([]any) {
		found = found || x.(map[string]any)["fileId"] == fileID
	}
	if !found {
		t.Fatalf("document should carry the file id: %v", me.Body["documents"])
	}

	// Reads.
	if s, ct, b := getRaw(t, base+"/files/"+fileID, d.Token); s != 200 || ct != "image/png" || !bytes.Equal(b, pngBytes) {
		t.Fatalf("uploader read: %d %s %d bytes", s, ct, len(b))
	}
	if s, _, _ := getRaw(t, base+"/files/"+fileID, f.OwnerToken); s != 200 {
		t.Fatalf("fleet staff read: %d", s)
	}
	if s, _, _ := getRaw(t, base+"/files/"+fileID, pax); s != 403 {
		t.Fatalf("passenger read should be 403: %d", s)
	}
	if s, _, _ := getRaw(t, "/v1/tenants/"+other.ID+"/files/"+fileID, other.OwnerToken); s != 404 {
		t.Fatalf("another tenant's owner must not see the file: %d", s)
	}
	if s, _, _ := getRaw(t, base+"/files/"+fileID, other.OwnerToken); s != 404 {
		t.Fatalf("non-member must not see the file: %d", s)
	}
	// A second driver of the same tenant can't read the first driver's file.
	d2 := addDriver(t, f, [2]float64{6.94, 79.86}, "2030-01-01")
	if s, _, _ := getRaw(t, base+"/files/"+fileID, d2.Token); s != 404 {
		t.Fatalf("other driver read should be 404: %d", s)
	}
	// Platform staff read is audited.
	admin := superAdminToken(t)
	if s, _, _ := getRaw(t, "/v1/platform/files/"+fileID, admin); s != 200 {
		t.Fatalf("platform read: %d", s)
	}
	if s, _, _ := getRaw(t, "/v1/platform/files/"+fileID, f.OwnerToken); s != 403 {
		t.Fatalf("platform route for a tenant owner: %d", s)
	}
	audit := call(t, "GET", base+"/audit-events", f.OwnerToken, nil)
	seen := false
	for _, e := range audit.Body["events"].([]any) {
		seen = seen || e.(map[string]any)["action"] == "file.viewed_by_platform"
	}
	if !seen {
		t.Fatalf("platform file read must be audited")
	}
}

func TestBusinessDocumentsRequiredForSelfService(t *testing.T) {
	_, token, verify := register(t, "Docs Founder")
	must(t, call(t, "POST", "/v1/auth/verify-email", "", map[string]string{"token": verify}), 200, "verify")
	tid := dig(must(t, call(t, "POST", "/v1/businesses", token, map[string]string{"slug": uniq("docs-"), "name": "Docs Cabs", "businessType": "taxi"}), 201, "biz").Body, "tenant", "id")
	base := "/v1/tenants/" + tid
	f := tenantFixture{ID: tid, OwnerToken: token}
	completeProfile(t, f)
	configurePricing(t, &f)
	cl := must(t, call(t, "GET", base+"/onboarding", token, nil), 200, "checklist")
	if cl.Body["canSubmit"] != false {
		t.Fatalf("self-service business without a registration document must not be submittable: %v", cl.Body)
	}
	if r := call(t, "POST", base+"/documents", token, map[string]string{"docType": "business_registration", "fileId": "00000000-0000-0000-0000-000000000000"}); errCode(r) != "invalid_file" {
		t.Fatalf("unknown file: %d %v", r.Status, r.Body)
	}
	fileID := dig(uploadFile(t, base+"/files", token, "br.pdf", pdfBytes).Body, "file", "id")
	must(t, call(t, "POST", base+"/documents", token, map[string]string{"docType": "business_registration", "fileId": fileID, "note": "PV 777 certificate"}), 201, "attach")
	cl = must(t, call(t, "GET", base+"/onboarding", token, nil), 200, "checklist")
	if cl.Body["canSubmit"] != true {
		t.Fatalf("with the certificate the business should be submittable: %v", cl.Body)
	}
	must(t, call(t, "POST", base+"/onboarding/complete", token, nil), 200, "submit")
	admin := superAdminToken(t)
	det := must(t, call(t, "GET", "/v1/platform/tenants/"+tid, admin, nil), 200, "platform detail")
	docs := det.Body["documents"].([]any)
	if len(docs) != 1 || docs[0].(map[string]any)["fileId"] != fileID || !strings.Contains(dig(det.Body, "tenant", "status"), "pending") {
		t.Fatalf("platform detail should list the document for a pending tenant: %v", det.Body)
	}
}
