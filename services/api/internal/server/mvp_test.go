package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/fare"
)

// Colombo test geography (lat, lng).
var (
	fort         = [2]float64{6.9344, 79.8428}
	bambalapitya = [2]float64{6.8906, 79.8562}
	kandy        = [2]float64{7.2906, 80.6337} // outside the zone
	colomboZone  = map[string]any{"type": "Polygon", "coordinates": [][][]float64{{
		{79.80, 6.80}, {79.95, 6.80}, {79.95, 7.00}, {79.80, 7.00}, {79.80, 6.80}}}}
	carRule = map[string]any{"baseMinor": 10000, "perKmMinor": 8500, "perMinuteMinor": 500, "minimumMinor": 25000,
		"bookingFeeMinor": 2000, "waitingPerMinMinor": 300, "freeWaitingMinutes": 3, "roundToMinor": 100}
)

func callH(t *testing.T, method, path, token string, headers map[string]string, body any) resp {
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
	for k, v := range headers {
		req.Header.Set(k, v)
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

func must(t *testing.T, r resp, status int, what string) resp {
	t.Helper()
	if r.Status != status {
		t.Fatalf("%s: got %d want %d: %v", what, r.Status, status, r.Body)
	}
	return r
}

func configurePricing(t *testing.T, f *tenantFixture) {
	t.Helper()
	base := "/v1/tenants/" + f.ID
	f.ZoneID = dig(must(t, call(t, "POST", base+"/zones", f.OwnerToken, map[string]any{"name": "Colombo", "area": colomboZone}), 201, "zone").Body, "zone", "id")
	f.CarID = dig(must(t, call(t, "POST", base+"/vehicle-categories", f.OwnerToken, map[string]any{"code": "car", "name": "Car", "seats": 4}), 201, "car").Body, "category", "id")
	f.TukID = dig(must(t, call(t, "POST", base+"/vehicle-categories", f.OwnerToken, map[string]any{"code": "tuk", "name": "Tuk", "seats": 3}), 201, "tuk").Body, "category", "id")
	rule := map[string]any{"zoneId": f.ZoneID, "categoryId": f.CarID}
	for k, v := range carRule {
		rule[k] = v
	}
	must(t, call(t, "POST", base+"/fare-rules", f.OwnerToken, rule), 201, "fare rule")
}

// completeProfile fills the business details the publish step requires.
func completeProfile(t *testing.T, f tenantFixture) {
	t.Helper()
	must(t, call(t, "PATCH", "/v1/tenants/"+f.ID, f.OwnerToken, map[string]any{"legalName": "Colombo Cabs (Pvt) Ltd",
		"registrationNo": "PV 12345", "contactEmail": "hello@colombocabs.test", "contactPhone": "0112345678"}), 200, "profile")
}

// operator is a fully onboarded, active tenant with pricing.
func operator(t *testing.T) tenantFixture {
	t.Helper()
	f := newTenant(t, superAdminToken(t))
	configurePricing(t, &f)
	completeProfile(t, f)
	must(t, call(t, "POST", "/v1/tenants/"+f.ID+"/onboarding/complete", f.OwnerToken, nil), 200, "onboarding")
	return f
}

type driverFixture struct{ ID, Token, VehicleID string }

func phone() string { return fmt.Sprintf("+9477%07d", rand.IntN(10_000_000)) }

// addDriver registers, documents, approves and brings online a driver at pos.
func addDriver(t *testing.T, f tenantFixture, pos [2]float64, docExpiry string) driverFixture {
	t.Helper()
	base := "/v1/tenants/" + f.ID
	var d driverFixture
	d.VehicleID = dig(must(t, call(t, "POST", base+"/vehicles", f.OwnerToken, map[string]any{
		"categoryId": f.CarID, "plate": uniq("CAB-"), "make": "Toyota", "model": "Axio", "colour": "White"}), 201, "vehicle").Body, "vehicle", "id")
	must(t, call(t, "POST", base+"/vehicles/"+d.VehicleID+"/status", f.OwnerToken, map[string]any{"status": "approved"}), 200, "approve vehicle")
	email := uniq("driver-") + "@taxios.test"
	r := must(t, call(t, "POST", base+"/drivers", f.OwnerToken, map[string]any{"fullName": "Nimal Perera", "phone": phone(), "email": email}), 201, "driver")
	d.ID = dig(r.Body, "driver", "id")
	d.Token = must(t, call(t, "POST", "/v1/auth/accept-invite", "", map[string]string{
		"token": dig(r.Body, "invitation", "token"), "fullName": "Nimal Perera", "password": password}), 200, "driver invite").Body["token"].(string)
	for _, doc := range []string{"driving_licence", "national_id", "vehicle_insurance"} {
		must(t, call(t, "POST", base+"/driver/documents", d.Token, map[string]any{"docType": doc, "number": uniq("N"), "expiresOn": docExpiry}), 201, "doc "+doc)
	}
	// Approval before review is refused: documents must be approved first.
	if r := call(t, "POST", base+"/drivers/"+d.ID+"/status", f.OwnerToken, map[string]any{"status": "approved"}); errCode(r) != "documents_incomplete" {
		t.Fatalf("approval before document review: %d %v", r.Status, r.Body)
	}
	detail := must(t, call(t, "GET", base+"/drivers/"+d.ID, f.OwnerToken, nil), 200, "driver detail")
	for _, doc := range detail.Body["documents"].([]any) {
		id := doc.(map[string]any)["id"].(string)
		must(t, call(t, "POST", base+"/drivers/"+d.ID+"/documents/"+id+"/review", f.OwnerToken, map[string]any{"decision": "approved"}), 200, "review")
	}
	must(t, call(t, "PUT", base+"/drivers/"+d.ID+"/vehicle", f.OwnerToken, map[string]any{"vehicleId": d.VehicleID}), 200, "assign vehicle")
	approve := call(t, "POST", base+"/drivers/"+d.ID+"/status", f.OwnerToken, map[string]any{"status": "approved"})
	if docExpiry < time.Now().Format("2006-01-02") {
		if errCode(approve) != "documents_incomplete" {
			t.Fatalf("driver with expired documents was approved: %d %v", approve.Status, approve.Body)
		}
		return d
	}
	must(t, approve, 200, "approve driver")
	must(t, call(t, "POST", base+"/driver/availability", d.Token, map[string]any{"available": true}), 200, "go online")
	moveDriver(t, f, d, pos)
	return d
}

func moveDriver(t *testing.T, f tenantFixture, d driverFixture, pos [2]float64) {
	t.Helper()
	must(t, call(t, "POST", "/v1/tenants/"+f.ID+"/driver/location", d.Token, map[string]any{"lat": pos[0], "lng": pos[1]}), 200, "location")
}

func addPassenger(t *testing.T, f tenantFixture) string {
	t.Helper()
	r := must(t, call(t, "POST", "/v1/public/tenants/"+f.Slug+"/passengers", "", map[string]any{
		"fullName": "Amaya Silva", "email": uniq("pax-") + "@taxios.test", "phone": phone(), "password": password}), 201, "passenger register")
	return r.Body["token"].(string)
}

func place(p [2]float64, addr string) map[string]any {
	return map[string]any{"lat": p[0], "lng": p[1], "address": addr}
}

// book quotes and books Fort → Bambalapitiya for a passenger.
func book(t *testing.T, f tenantFixture, pax string) (bookingID string, quote map[string]any) {
	t.Helper()
	q := must(t, call(t, "POST", "/v1/tenants/"+f.ID+"/quotes", pax, map[string]any{
		"pickup": place(fort, "Colombo Fort Railway Station"), "dropoff": place(bambalapitya, "Bambalapitiya Junction"),
		"categoryId": f.CarID}), 201, "quote").Body["quote"].(map[string]any)
	r := must(t, callH(t, "POST", "/v1/tenants/"+f.ID+"/my/bookings", pax, map[string]string{"Idempotency-Key": uniq("idem-")},
		map[string]any{"quoteId": q["id"], "paymentMethod": "cash"}), 201, "book")
	return dig(r.Body, "booking", "id"), q
}

func pendingOffer(t *testing.T, f tenantFixture, d driverFixture) map[string]any {
	t.Helper()
	r := must(t, call(t, "GET", "/v1/tenants/"+f.ID+"/driver/work", d.Token, nil), 200, "driver work")
	o, _ := r.Body["offer"].(map[string]any)
	return o
}

func trip(t *testing.T, f tenantFixture, d driverFixture, bookingID, action string, body any) resp {
	t.Helper()
	return call(t, "POST", "/v1/tenants/"+f.ID+"/driver/trips/"+bookingID+"/"+action, d.Token, body)
}

// Spec Part 17: the first commercial MVP journey, steps 1–18, end to end.
func TestMVPJourney(t *testing.T) {
	f := operator(t)                                                   // 1–3: admin creates tenant; owner onboards with zone, category, fare rule
	near := addDriver(t, f, [2]float64{6.9300, 79.8450}, "2030-12-31") // 4–5: driver + vehicle, documents approved
	far := addDriver(t, f, [2]float64{6.9150, 79.8520}, "2030-12-31")
	pax := addPassenger(t, f)
	base := "/v1/tenants/" + f.ID

	// 6–7: passenger books; the server stores the estimate with its fare version.
	bookingID, quote := book(t, f, pax)
	est := quote["estimate"].(map[string]any)
	if quote["isEstimate"] != true || est["totalMinor"].(float64) < 25000 || quote["fareRuleVersion"].(float64) != 1 {
		t.Fatalf("quote: %v", quote)
	}

	// 8: dispatch offers the nearest eligible driver, not the farther one.
	o := pendingOffer(t, f, near)
	if o == nil || o["bookingId"] != bookingID {
		t.Fatalf("nearest driver has no offer: %v", o)
	}
	if pendingOffer(t, f, far) != nil {
		t.Fatal("farther driver was offered the same booking")
	}

	// 9–10: driver accepts; passenger sees assignment and permitted driver info.
	must(t, call(t, "POST", base+"/driver/offers/"+o["id"].(string)+"/accept", near.Token, nil), 200, "accept")
	pv := must(t, call(t, "GET", base+"/my/bookings/"+bookingID, pax, nil), 200, "passenger view").Body["booking"].(map[string]any)
	drv, _ := pv["driver"].(map[string]any)
	if pv["status"] != "assigned" || drv == nil || drv["vehiclePlate"] == "" || drv["firstName"] != "Nimal" || drv["locationFresh"] != true {
		t.Fatalf("passenger view after assignment: %v", pv)
	}
	if _, leaked := drv["phone"]; leaked {
		t.Fatal("driver phone exposed to passenger")
	}

	// 11–12: permitted trip lifecycle, with a GPS track while in progress.
	if r := trip(t, f, near, bookingID, "start", nil); errCode(r) != "invalid_transition" {
		t.Fatalf("start before arrival should be refused: %d %v", r.Status, r.Body)
	}
	for _, a := range []string{"en-route", "arrived", "start"} {
		must(t, trip(t, f, near, bookingID, a, nil), 200, a)
	}
	for _, p := range [][2]float64{fort, {6.9120, 79.8500}, bambalapitya} {
		moveDriver(t, f, near, p)
	}
	done := must(t, trip(t, f, near, bookingID, "complete", nil), 200, "complete").Body["booking"].(map[string]any)

	// 13: final fare = engine(fare version 1, tracked distance, measured duration).
	tr := done["trip"].(map[string]any)
	if done["status"] != "completed" || tr["distanceSource"] != "gps_track" || done["isFinalFare"] != true {
		t.Fatalf("completed trip: %v", done)
	}
	want := fare.Compute(fare.Rule{Currency: "LKR", BaseMinor: 10000, PerKmMinor: 8500, PerMinuteMinor: 500, MinimumMinor: 25000,
		BookingFeeMinor: 2000, WaitingPerMinMinor: 300, FreeWaitingMinutes: 3, RoundToMinor: 100},
		fare.Input{DistanceM: int64(tr["distanceM"].(float64)), DurationS: int64(tr["durationS"].(float64))})
	total := int64(tr["finalTotalMinor"].(float64))
	if total != want.TotalMinor || tr["distanceM"].(float64) < 4500 {
		t.Fatalf("final fare %d (distance %v), engine says %d", total, tr["distanceM"], want.TotalMinor)
	}

	// 14: cash recorded once; a mismatched amount needs a note; a second record is refused.
	if r := trip(t, f, near, bookingID, "cash", map[string]any{"amountMinor": total - 100}); errCode(r) != "note_required" {
		t.Fatalf("short cash without note: %d %v", r.Status, r.Body)
	}
	must(t, trip(t, f, near, bookingID, "cash", map[string]any{"amountMinor": total}), 201, "cash")
	if r := trip(t, f, near, bookingID, "cash", map[string]any{"amountMinor": total}); errCode(r) != "already_recorded" {
		t.Fatalf("double cash record: %d %v", r.Status, r.Body)
	}

	// 15: receipt shows the final server fare and payment.
	rc := must(t, call(t, "GET", base+"/my/bookings/"+bookingID+"/receipt", pax, nil), 200, "receipt").Body["receipt"].(map[string]any)
	if rc["totalMinor"].(float64) != float64(total) || rc["paymentStatus"] != "paid_cash" || rc["reference"] == "" {
		t.Fatalf("receipt: %v", rc)
	}
	must(t, call(t, "POST", base+"/my/bookings/"+bookingID+"/rating", pax, map[string]any{"stars": 5, "comment": "Smooth ride"}), 201, "rating")

	// 16: operator sees the completed trip, its payment and full history.
	ov := must(t, call(t, "GET", base+"/bookings/"+bookingID, f.OwnerToken, nil), 200, "operator view").Body["booking"].(map[string]any)
	if ov["status"] != "completed" || len(ov["payments"].([]any)) != 1 || len(ov["events"].([]any)) < 8 {
		t.Fatalf("operator view: payments=%v events=%d", ov["payments"], len(ov["events"].([]any)))
	}
	// The driver is free again and the farther driver never got this booking.
	if r := must(t, call(t, "GET", base+"/drivers/"+near.ID, f.OwnerToken, nil), 200, "driver"); dig(r.Body, "driver", "availability") != "available" {
		t.Fatalf("driver not released: %v", r.Body)
	}

	// 17: analytics reflect it.
	m := must(t, call(t, "GET", base+"/reports/summary", f.OwnerToken, nil), 200, "summary").Body["metrics"].(map[string]any)
	if m["completed"].(float64) != 1 || m["faresCompletedMinor"].(float64) != float64(total) || m["cashRecordedMinor"].(float64) != float64(total) || m["avgRating"].(float64) != 5 {
		t.Fatalf("summary: %v", m)
	}

	// 18: another tenant sees none of it.
	other := operator(t)
	for _, p := range []string{"/bookings/" + bookingID, "/drivers/" + near.ID} {
		if r := call(t, "GET", "/v1/tenants/"+f.ID+p, other.OwnerToken, nil); r.Status != 404 {
			t.Errorf("other tenant reading %s: %d", p, r.Status)
		}
	}
	if r := call(t, "GET", "/v1/tenants/"+other.ID+"/bookings/"+bookingID, other.OwnerToken, nil); r.Status != 404 {
		t.Errorf("booking id through other tenant's path: %d", r.Status)
	}
	otherPax := addPassenger(t, f)
	if r := call(t, "GET", base+"/my/bookings/"+bookingID, otherPax, nil); r.Status != 404 {
		t.Errorf("another passenger reading the booking: %d", r.Status)
	}
}

// Mandatory scenario 3: retrying a booking request does not create a duplicate.
func TestBookingIdempotency(t *testing.T) {
	f := operator(t)
	pax := addPassenger(t, f)
	base := "/v1/tenants/" + f.ID
	quote := func() string {
		return dig(must(t, call(t, "POST", base+"/quotes", pax, map[string]any{"pickup": place(fort, "Fort"),
			"dropoff": place(bambalapitya, "Bamba"), "categoryId": f.CarID}), 201, "quote").Body, "quote", "id")
	}
	q1 := quote()
	key := map[string]string{"Idempotency-Key": uniq("idem-")}
	if r := call(t, "POST", base+"/my/bookings", pax, map[string]any{"quoteId": q1}); errCode(r) != "idempotency_key_required" {
		t.Fatalf("booking without key: %d %v", r.Status, r.Body)
	}
	// Concurrent retries of the same request.
	var wg sync.WaitGroup
	ids := make([]string, 5)
	codes := make([]int, 5)
	for i := range ids {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r := callH(t, "POST", base+"/my/bookings", pax, key, map[string]any{"quoteId": q1})
			codes[i] = r.Status
			if b, ok := r.Body["booking"].(map[string]any); ok {
				ids[i] = b["id"].(string)
			}
		}(i)
	}
	wg.Wait()
	created := 0
	for i := range ids {
		if ids[i] != ids[0] || ids[i] == "" {
			t.Fatalf("retries produced different bookings: %v (codes %v)", ids, codes)
		}
		if codes[i] == 201 {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("want exactly one 201, got codes %v", codes)
	}
	if r := callH(t, "POST", base+"/my/bookings", pax, key, map[string]any{"quoteId": quote()}); errCode(r) != "idempotency_key_reused" {
		t.Fatalf("same key, different quote: %d %v", r.Status, r.Body)
	}
	if r := callH(t, "POST", base+"/my/bookings", pax, map[string]string{"Idempotency-Key": uniq("idem-")}, map[string]any{"quoteId": q1}); errCode(r) != "quote_already_used" {
		t.Fatalf("new key, used quote: %d %v", r.Status, r.Body)
	}
	list := must(t, call(t, "GET", base+"/my/bookings", pax, nil), 200, "list").Body["bookings"].([]any)
	if len(list) != 1 {
		t.Fatalf("passenger has %d bookings, want 1", len(list))
	}
}

// Mandatory scenario 2: an exclusive assignment can't be taken twice.
func TestExclusiveAssignmentConcurrency(t *testing.T) {
	f := operator(t)
	d := addDriver(t, f, [2]float64{6.9300, 79.8450}, "2030-12-31")
	other := addDriver(t, f, [2]float64{6.9300, 79.8460}, "2030-12-31")
	pax := addPassenger(t, f)
	bookingID, _ := book(t, f, pax)
	var offered driverFixture
	var o map[string]any
	for _, x := range []driverFixture{d, other} {
		if got := pendingOffer(t, f, x); got != nil {
			offered, o = x, got
		}
	}
	if o == nil {
		t.Fatal("no offer made")
	}
	notOffered := d
	if offered.ID == d.ID {
		notOffered = other
	}
	if r := call(t, "POST", "/v1/tenants/"+f.ID+"/driver/offers/"+o["id"].(string)+"/accept", notOffered.Token, nil); r.Status != 404 {
		t.Fatalf("another driver accepting someone else's offer: %d", r.Status)
	}
	var wg sync.WaitGroup
	codes := make(chan int, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codes <- call(t, "POST", "/v1/tenants/"+f.ID+"/driver/offers/"+o["id"].(string)+"/accept", offered.Token, nil).Status
		}()
	}
	wg.Wait()
	close(codes)
	ok := 0
	for c := range codes {
		switch c {
		case 200:
			ok++
		case 409:
		default:
			t.Errorf("unexpected status %d", c)
		}
	}
	if ok != 1 {
		t.Fatalf("%d concurrent accepts succeeded, want exactly 1", ok)
	}
	// Dispatcher can't hand an assigned booking to another driver.
	if r := call(t, "POST", "/v1/tenants/"+f.ID+"/bookings/"+bookingID+"/assign", f.OwnerToken, map[string]any{"driverId": notOffered.ID}); errCode(r) != "invalid_state" {
		t.Fatalf("reassigning an assigned booking: %d %v", r.Status, r.Body)
	}
	var trips int
	ownerPool.QueryRow(context.Background(), `SELECT count(*) FROM trips WHERE booking_id=$1`, bookingID).Scan(&trips)
	if trips != 1 {
		t.Fatalf("%d trips for one booking", trips)
	}
}

// Mandatory scenario 11: invalid or expired documents block assignment.
func TestExpiredDocumentsBlockDispatch(t *testing.T) {
	f := operator(t)
	expired := addDriver(t, f, [2]float64{6.9340, 79.8430}, "2020-01-01") // closest, but documents expired
	base := "/v1/tenants/" + f.ID
	if r := call(t, "POST", base+"/driver/availability", expired.Token, map[string]any{"available": true}); errCode(r) != "not_eligible" {
		t.Fatalf("expired driver going online: %d %v", r.Status, r.Body)
	}
	valid := addDriver(t, f, [2]float64{6.9200, 79.8500}, "2030-12-31")
	pax := addPassenger(t, f)
	bookingID, _ := book(t, f, pax)
	if o := pendingOffer(t, f, valid); o == nil || o["bookingId"] != bookingID {
		t.Fatalf("valid driver should get the offer: %v", o)
	}
	// Force the expired driver online behind the API's back: dispatch must still refuse them.
	ownerPool.Exec(context.Background(), `UPDATE drivers SET availability='available', status='approved',
		location=ST_SetSRID(ST_MakePoint(79.8430, 6.9340),4326), location_at=now() WHERE id=$1`, expired.ID)
	must(t, call(t, "POST", base+"/bookings/"+bookingID+"/cancel", f.OwnerToken, map[string]any{"reason": "test reset"}), 200, "cancel")
	b2, _ := book(t, f, pax)
	if o := pendingOffer(t, f, expired); o != nil {
		t.Fatalf("expired-document driver was offered %v", o["bookingId"])
	}
	if r := call(t, "POST", base+"/bookings/"+b2+"/assign", f.OwnerToken, map[string]any{"driverId": expired.ID}); errCode(r) != "driver_not_eligible" {
		t.Fatalf("manual assign to expired driver: %d %v", r.Status, r.Body)
	}
}

// Mandatory scenario 12: a completed trip can't be reopened or cancelled.
func TestCompletedTripIsFinal(t *testing.T) {
	f := operator(t)
	d := addDriver(t, f, [2]float64{6.9300, 79.8450}, "2030-12-31")
	pax := addPassenger(t, f)
	bookingID, _ := book(t, f, pax)
	o := pendingOffer(t, f, d)
	must(t, call(t, "POST", "/v1/tenants/"+f.ID+"/driver/offers/"+o["id"].(string)+"/accept", d.Token, nil), 200, "accept")
	for _, a := range []string{"arrived", "start"} {
		must(t, trip(t, f, d, bookingID, a, nil), 200, a)
	}
	if r := call(t, "POST", "/v1/tenants/"+f.ID+"/my/bookings/"+bookingID+"/cancel", pax, nil); errCode(r) != "invalid_transition" {
		t.Fatalf("cancelling a trip in progress: %d %v", r.Status, r.Body)
	}
	done := must(t, trip(t, f, d, bookingID, "complete", nil), 200, "complete").Body["booking"].(map[string]any)
	if src := done["trip"].(map[string]any)["distanceSource"]; src != "estimate_fallback" {
		t.Fatalf("no GPS points should fall back to the estimate, got %v", src)
	}
	for _, a := range []string{"start", "arrived", "complete", "abort", "en-route"} {
		if r := trip(t, f, d, bookingID, a, map[string]any{"reason": "reopen attempt"}); errCode(r) != "invalid_transition" {
			t.Errorf("%s on completed trip: %d %v", a, r.Status, r.Body)
		}
	}
	for _, path := range []string{"/my/bookings/" + bookingID + "/cancel"} {
		if r := call(t, "POST", "/v1/tenants/"+f.ID+path, pax, map[string]any{"reason": "x"}); errCode(r) != "invalid_transition" {
			t.Errorf("passenger cancel completed: %d %v", r.Status, r.Body)
		}
	}
	if r := call(t, "POST", "/v1/tenants/"+f.ID+"/bookings/"+bookingID+"/cancel", f.OwnerToken, map[string]any{"reason": "reopen"}); errCode(r) != "invalid_transition" {
		t.Errorf("staff cancel completed: %d %v", r.Status, r.Body)
	}
	// Booking history is append-only even for the app role.
	if _, err := appPool.Exec(context.Background(), `DELETE FROM booking_events WHERE booking_id=$1`, bookingID); err == nil {
		t.Error("app role deleted booking events")
	}
}

func TestRejectAndExpiryRedispatch(t *testing.T) {
	f := operator(t)
	a := addDriver(t, f, [2]float64{6.9300, 79.8450}, "2030-12-31")
	b := addDriver(t, f, [2]float64{6.9150, 79.8520}, "2030-12-31")
	pax := addPassenger(t, f)
	bookingID, _ := book(t, f, pax)

	// a (nearest) rejects → b is offered.
	o := pendingOffer(t, f, a)
	must(t, call(t, "POST", "/v1/tenants/"+f.ID+"/driver/offers/"+o["id"].(string)+"/reject", a.Token, nil), 200, "reject")
	ob := pendingOffer(t, f, b)
	if ob == nil || ob["bookingId"] != bookingID {
		t.Fatalf("after rejection, next driver should be offered: %v", ob)
	}
	// b lets it expire → nobody left who hasn't seen it → booking stays searching.
	ownerPool.Exec(context.Background(), `UPDATE dispatch_offers SET expires_at = now() - interval '1 second' WHERE id=$1`, ob["id"])
	if err := bookings.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if pendingOffer(t, f, b) != nil || pendingOffer(t, f, a) != nil {
		t.Fatal("drivers who already saw the booking were re-offered it")
	}
	if r := call(t, "POST", "/v1/tenants/"+f.ID+"/driver/offers/"+ob["id"].(string)+"/accept", b.Token, nil); errCode(r) != "offer_unavailable" {
		t.Fatalf("accepting an expired offer: %d %v", r.Status, r.Body)
	}
	// A dispatcher can still assign manually to a driver who already declined.
	must(t, call(t, "POST", "/v1/tenants/"+f.ID+"/bookings/"+bookingID+"/assign", f.OwnerToken, map[string]any{"driverId": a.ID}), 200, "manual assign")
	if o := pendingOffer(t, f, a); o == nil || o["bookingId"] != bookingID {
		t.Fatalf("manual offer missing: %v", o)
	}
	// Search timeout expires the booking.
	ownerPool.Exec(context.Background(), `UPDATE dispatch_offers SET status='withdrawn' WHERE booking_id=$1 AND status='pending'`, bookingID)
	ownerPool.Exec(context.Background(), `UPDATE bookings SET search_started_at = now() - interval '11 minutes' WHERE id=$1`, bookingID)
	bookings.Tick(context.Background())
	if v := must(t, call(t, "GET", "/v1/tenants/"+f.ID+"/my/bookings/"+bookingID, pax, nil), 200, "view"); dig(v.Body, "booking", "status") != "expired" {
		t.Fatalf("search timeout: %v", dig(v.Body, "booking", "status"))
	}
}

func TestQuoteAndBookingRules(t *testing.T) {
	f := operator(t)
	pax := addPassenger(t, f)
	base := "/v1/tenants/" + f.ID
	cases := []struct {
		body map[string]any
		code string
	}{
		{map[string]any{"pickup": place(kandy, "Kandy"), "dropoff": place(fort, "Fort"), "categoryId": f.CarID}, "outside_service_area"},
		{map[string]any{"pickup": place(fort, "Fort"), "dropoff": place(bambalapitya, "B"), "categoryId": f.TukID}, "category_unavailable"},
		{map[string]any{"pickup": place(fort, ""), "dropoff": place(bambalapitya, "B"), "categoryId": f.CarID}, "invalid_location"},
		{map[string]any{"pickup": place(fort, "Fort"), "dropoff": place(bambalapitya, "B"), "categoryId": f.CarID,
			"scheduledFor": time.Now().Add(5 * time.Minute)}, "invalid_schedule"},
	}
	for _, c := range cases {
		if r := call(t, "POST", base+"/quotes", pax, c.body); errCode(r) != c.code {
			t.Errorf("want %s, got %d %v", c.code, r.Status, r.Body)
		}
	}
	// Scheduled booking waits, then activates at lead time.
	when := time.Now().Add(40 * time.Minute)
	q := must(t, call(t, "POST", base+"/quotes", pax, map[string]any{"pickup": place(fort, "Fort"), "dropoff": place(bambalapitya, "B"),
		"categoryId": f.CarID, "scheduledFor": when}), 201, "scheduled quote").Body["quote"].(map[string]any)
	lines := q["estimate"].(map[string]any)["lines"].([]any)
	if len(lines) == 0 {
		t.Fatal("empty breakdown")
	}
	r := must(t, callH(t, "POST", base+"/my/bookings", pax, map[string]string{"Idempotency-Key": uniq("idem-")},
		map[string]any{"quoteId": q["id"]}), 201, "scheduled booking")
	id := dig(r.Body, "booking", "id")
	if s := dig(r.Body, "booking", "status"); s != "scheduled" {
		t.Fatalf("scheduled booking status %s", s)
	}
	ownerPool.Exec(context.Background(), `UPDATE bookings SET scheduled_for = now() + interval '10 minutes' WHERE id=$1`, id)
	bookings.Tick(context.Background())
	if s := dig(must(t, call(t, "GET", base+"/my/bookings/"+id, pax, nil), 200, "view").Body, "booking", "status"); s != "searching" {
		t.Fatalf("scheduled booking not activated at lead time: %s", s)
	}
	// Online payment isn't offered until a provider is verified.
	q2 := dig(must(t, call(t, "POST", base+"/quotes", pax, map[string]any{"pickup": place(fort, "Fort"), "dropoff": place(bambalapitya, "B"),
		"categoryId": f.CarID}), 201, "quote").Body, "quote", "id")
	if r := callH(t, "POST", base+"/my/bookings", pax, map[string]string{"Idempotency-Key": uniq("idem-")},
		map[string]any{"quoteId": q2, "paymentMethod": "card"}); errCode(r) != "payment_method_unavailable" {
		t.Fatalf("card payment: %d %v", r.Status, r.Body)
	}
	// Dispatcher phone booking with the same idempotency semantics.
	key := map[string]string{"Idempotency-Key": uniq("phone-")}
	body := map[string]any{"customer": map[string]any{"fullName": "Kamal", "phone": "0771234567"},
		"pickup": place(fort, "Fort"), "dropoff": place(bambalapitya, "B"), "categoryId": f.CarID}
	r1 := must(t, callH(t, "POST", base+"/bookings", f.OwnerToken, key, body), 201, "phone booking")
	r2 := must(t, callH(t, "POST", base+"/bookings", f.OwnerToken, key, body), 200, "phone booking retry")
	if dig(r1.Body, "booking", "id") != dig(r2.Body, "booking", "id") {
		t.Fatal("phone booking retry created a duplicate")
	}
	board := must(t, call(t, "GET", base+"/dispatch/board", f.OwnerToken, nil), 200, "board").Body
	if len(board["bookings"].([]any)) < 2 {
		t.Fatalf("dispatch board missing open bookings: %v", board["bookings"])
	}
}

// RLS covers every new tenant table, not only the API paths.
func TestMobilityTablesRLS(t *testing.T) {
	a := operator(t)
	addDriver(t, a, [2]float64{6.9300, 79.8450}, "2030-12-31")
	book(t, a, addPassenger(t, a))
	b := operator(t)
	ctx := context.Background()
	for _, table := range []string{"service_zones", "vehicle_categories", "fare_rules", "fare_rule_versions", "vehicles", "drivers",
		"driver_documents", "customers", "booking_quotes", "bookings", "booking_events", "dispatch_offers", "driver_compliance"} {
		var n int
		err := appPoolScoped(ctx, b.ID, fmt.Sprintf(`SELECT count(*) FROM %s WHERE tenant_id=$1`, table), a.ID, &n)
		if err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		if n != 0 {
			t.Errorf("tenant B scope sees %d rows of tenant A in %s", n, table)
		}
	}
}

func appPoolScoped(ctx context.Context, tenantID, sql string, arg any, dst any) error {
	return db.InTx(ctx, appPool, db.Scope{TenantID: tenantID}, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, sql, arg).Scan(dst)
	})
}
