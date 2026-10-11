// Package booking owns quotes, bookings, dispatch, trip execution, cash
// records, receipts and operational reports.
//
// Lock order (deadlock avoidance): booking → dispatch offer → driver.
package booking

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/fare"
	"github.com/captain-KasH/driveappsaas/services/api/internal/fleet"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
	"github.com/captain-KasH/driveappsaas/services/api/internal/tenancy"
)

const (
	QuoteTTL          = 10 * time.Minute
	OfferTTL          = 30 * time.Second
	ManualOfferTTL    = 90 * time.Second
	DispatchRadiusM   = 8000
	SearchTimeout     = 10 * time.Minute
	ScheduledLeadTime = 15 * time.Minute
	MinScheduleAhead  = 20 * time.Minute
	MaxScheduleAhead  = 30 * 24 * time.Hour
)

type Service struct {
	Pool    *pgxpool.Pool
	Tenancy *tenancy.Service
}

type Result = tenancy.Result
type Member = tenancy.Member

func (s *Service) Register(mux *http.ServeMux) {
	t := s.Tenancy
	p := "/v1/tenants/{tenantID}"
	// passenger app
	mux.Handle("GET /v1/public/tenants/{slug}", httpx.Handler(s.publicTenant))
	mux.Handle("POST /v1/public/tenants/{slug}/passengers", httpx.Handler(s.registerPassenger))
	mux.Handle("POST "+p+"/quotes", t.RoleRoute("passenger", true, s.passengerQuote))
	mux.Handle("POST "+p+"/my/bookings", t.RoleRoute("passenger", true, s.passengerBook))
	mux.Handle("GET "+p+"/my/bookings", t.RoleRoute("passenger", false, s.myBookings))
	mux.Handle("GET "+p+"/my/bookings/{bookingID}", t.RoleRoute("passenger", false, s.myBooking))
	mux.Handle("POST "+p+"/my/bookings/{bookingID}/cancel", t.RoleRoute("passenger", true, s.passengerCancel))
	mux.Handle("GET "+p+"/my/bookings/{bookingID}/receipt", t.RoleRoute("passenger", false, s.myReceipt))
	mux.Handle("POST "+p+"/my/bookings/{bookingID}/rating", t.RoleRoute("passenger", true, s.rate))
	// driver app
	mux.Handle("GET "+p+"/driver/work", t.RoleRoute("driver", false, s.driverWork))
	mux.Handle("POST "+p+"/driver/offers/{offerID}/accept", t.RoleRoute("driver", true, s.acceptOffer))
	mux.Handle("POST "+p+"/driver/offers/{offerID}/reject", t.RoleRoute("driver", true, s.rejectOffer))
	mux.Handle("POST "+p+"/driver/trips/{bookingID}/{action}", t.RoleRoute("driver", true, s.driverTripAction))
	// operator / dispatcher console
	mux.Handle("POST "+p+"/bookings", t.TenantRoute(tenancy.PermDispatch, true, s.staffBook))
	mux.Handle("GET "+p+"/bookings", t.TenantRoute(tenancy.PermBookingRead, false, s.listBookings))
	mux.Handle("GET "+p+"/bookings/{bookingID}", t.TenantRoute(tenancy.PermBookingRead, false, s.getBooking))
	mux.Handle("POST "+p+"/bookings/{bookingID}/assign", t.TenantRoute(tenancy.PermDispatch, true, s.manualAssign))
	mux.Handle("POST "+p+"/bookings/{bookingID}/cancel", t.TenantRoute(tenancy.PermDispatch, true, s.staffCancel))
	mux.Handle("GET "+p+"/dispatch/board", t.TenantRoute(tenancy.PermDispatch, false, s.dispatchBoard))
	mux.Handle("GET "+p+"/reports/summary", t.TenantRoute(tenancy.PermReportsRead, false, s.summary))
	mux.Handle("GET "+p+"/customers", t.TenantRoute(tenancy.PermBookingRead, false, s.listCustomers))
	canExport := func(m tenancy.Member) bool { return tenancy.RoleHas(m.Role, tenancy.PermReportsRead) }
	mux.Handle("GET "+p+"/exports/customers.csv", t.MemberRouteRaw(canExport, s.exportCustomers))
	mux.Handle("GET "+p+"/exports/bookings.csv", t.MemberRouteRaw(canExport, s.exportBookings))
}

func invalid(code, format string, args ...any) error {
	return httpx.Errorf(http.StatusUnprocessableEntity, code, format, args...)
}

func conflict(code, format string, args ...any) error {
	return httpx.Errorf(http.StatusConflict, code, format, args...)
}

func pathID(r *http.Request, name string) (string, error) {
	id := strings.ToLower(r.PathValue(name))
	if !tenancy.UUID(id) {
		return "", httpx.ErrNotFound
	}
	return id, nil
}

// event appends to the booking's immutable history.
func event(ctx context.Context, tx pgx.Tx, tenantID, bookingID, name, from, to, actor, reason string, data map[string]any) error {
	if data == nil {
		data = map[string]any{}
	}
	_, err := tx.Exec(ctx, `INSERT INTO booking_events(tenant_id, booking_id, event, from_status, to_status, actor_user_id, reason, data)
		VALUES ($1,$2,$3,nullif($4,''),nullif($5,''),nullif($6,'')::uuid,nullif($7,''),$8)`,
		tenantID, bookingID, name, from, to, actor, reason, data)
	return err
}

// --- quotes ---

type Place struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address"`
}

func (p Place) valid() bool {
	return p.Lat >= -90 && p.Lat <= 90 && p.Lng >= -180 && p.Lng <= 180 && !(p.Lat == 0 && p.Lng == 0) &&
		strings.TrimSpace(p.Address) != "" && len(p.Address) <= 300
}

type QuoteInput struct {
	Pickup       Place      `json:"pickup"`
	Dropoff      Place      `json:"dropoff"`
	CategoryID   string     `json:"categoryId"`
	ScheduledFor *time.Time `json:"scheduledFor"`
}

type Quote struct {
	ID             string         `json:"id"`
	CategoryID     string         `json:"categoryId"`
	ZoneID         string         `json:"zoneId"`
	FareRuleVer    int            `json:"fareRuleVersion"`
	EstDistanceM   int64          `json:"estimatedDistanceM"`
	EstDurationS   int64          `json:"estimatedDurationS"`
	EstimateMethod string         `json:"estimateMethod"`
	Estimate       fare.Breakdown `json:"estimate"`
	ScheduledFor   *time.Time     `json:"scheduledFor"`
	ExpiresAt      time.Time      `json:"expiresAt"`
	IsEstimate     bool           `json:"isEstimate"`
}

// toRule converts a stored fare version into the engine's input.
func toRule(f fleet.FareRuleVersion) fare.Rule {
	parse := func(s *string) *time.Duration {
		if s == nil {
			return nil
		}
		t, err := time.Parse("15:04", *s)
		if err != nil {
			return nil
		}
		d := time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute
		return &d
	}
	return fare.Rule{Currency: f.Currency, BaseMinor: f.BaseMinor, PerKmMinor: f.PerKmMinor, PerMinuteMinor: f.PerMinuteMinor,
		MinimumMinor: f.MinimumMinor, BookingFeeMinor: f.BookingFeeMinor, ScheduledFeeMinor: f.ScheduledFeeMinor,
		WaitingPerMinMinor: f.WaitingPerMinMinor, FreeWaitingMinutes: f.FreeWaitingMinutes, NightSurchargeBps: f.NightSurchargeBps,
		NightStart: parse(f.NightStart), NightEnd: parse(f.NightEnd), TaxBps: f.TaxBps, RoundToMinor: f.RoundToMinor}
}

func tenantLocation(ctx context.Context, tx pgx.Tx, tenantID string) (*time.Location, error) {
	var tz string
	if err := tx.QueryRow(ctx, `SELECT timezone FROM tenants WHERE id=$1`, tenantID).Scan(&tz); err != nil {
		return nil, err
	}
	return time.LoadLocation(tz)
}

func createQuote(ctx context.Context, tx pgx.Tx, tenantID, customerID string, in QuoteInput) (Quote, error) {
	if !in.Pickup.valid() || !in.Dropoff.valid() {
		return Quote{}, invalid("invalid_location", "pickup and dropoff need lat, lng and an address")
	}
	if !tenancy.UUID(in.CategoryID) {
		return Quote{}, invalid("invalid_category", "categoryId is required")
	}
	now := time.Now()
	if in.ScheduledFor != nil {
		if in.ScheduledFor.Before(now.Add(MinScheduleAhead)) || in.ScheduledFor.After(now.Add(MaxScheduleAhead)) {
			return Quote{}, invalid("invalid_schedule", "scheduled rides must be 20 minutes to 30 days ahead")
		}
	}
	var zoneID string
	err := tx.QueryRow(ctx, `SELECT id FROM service_zones WHERE active AND ST_Covers(area, ST_SetSRID(ST_MakePoint($1,$2),4326))
		ORDER BY ST_Area(area) LIMIT 1`, in.Pickup.Lng, in.Pickup.Lat).Scan(&zoneID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Quote{}, invalid("outside_service_area", "pickup is outside this operator's service area")
	}
	if err != nil {
		return Quote{}, err
	}
	f, err := fleet.ScanFare(tx.QueryRow(ctx, fleet.FareVersionSQL+` JOIN vehicle_categories c ON c.id = fr.category_id
		WHERE fr.zone_id=$1 AND fr.category_id=$2 AND v.version = fr.current_version AND c.active`, zoneID, in.CategoryID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Quote{}, invalid("category_unavailable", "this vehicle category is not offered here")
	}
	if err != nil {
		return Quote{}, err
	}
	loc, err := tenantLocation(ctx, tx, tenantID)
	if err != nil {
		return Quote{}, err
	}
	meters, seconds := fare.EstimateRoute(in.Pickup.Lat, in.Pickup.Lng, in.Dropoff.Lat, in.Dropoff.Lng)
	start := now
	if in.ScheduledFor != nil {
		start = *in.ScheduledFor
	}
	b := fare.Compute(toRule(f), fare.Input{DistanceM: meters, DurationS: seconds, Scheduled: in.ScheduledFor != nil, LocalStart: start.In(loc)})
	q := Quote{CategoryID: in.CategoryID, ZoneID: zoneID, FareRuleVer: f.Version, EstDistanceM: meters, EstDurationS: seconds,
		EstimateMethod: fare.EstimateMethod, Estimate: b, ScheduledFor: in.ScheduledFor, ExpiresAt: now.Add(QuoteTTL).UTC(), IsEstimate: true}
	err = tx.QueryRow(ctx, `INSERT INTO booking_quotes(tenant_id, customer_id, zone_id, category_id, fare_rule_version_id,
		pickup, pickup_address, dropoff, dropoff_address, scheduled_for, est_distance_m, est_duration_s, estimate_method,
		currency, est_total_minor, breakdown, expires_at)
		VALUES ($1,$2,$3,$4,$5, ST_SetSRID(ST_MakePoint($6,$7),4326), $8, ST_SetSRID(ST_MakePoint($9,$10),4326), $11,
		$12,$13,$14,$15,$16,$17,$18,$19) RETURNING id`,
		tenantID, customerID, zoneID, in.CategoryID, f.ID, in.Pickup.Lng, in.Pickup.Lat, strings.TrimSpace(in.Pickup.Address),
		in.Dropoff.Lng, in.Dropoff.Lat, strings.TrimSpace(in.Dropoff.Address), in.ScheduledFor, meters, seconds,
		fare.EstimateMethod, b.Currency, b.TotalMinor, b, q.ExpiresAt).Scan(&q.ID)
	return q, err
}

// --- booking creation ---

const refAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func newReference() string {
	b := make([]byte, 7)
	rand.Read(b)
	for i := range b {
		b[i] = refAlphabet[int(b[i])%len(refAlphabet)]
	}
	return "TX-" + string(b)
}

func idempotencyKey(r *http.Request) (string, error) {
	k := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if len(k) < 8 || len(k) > 128 {
		return "", invalid("idempotency_key_required", "send an Idempotency-Key header (8-128 chars) so retries never double-book")
	}
	return k, nil
}

type createInput struct {
	TenantID, CustomerID, QuoteID, Channel, Key, ActorUserID, PaymentMethod, Note string
}

// createBooking turns a quote into a booking exactly once per idempotency key.
// Returns created=false when the key was already used for this same quote.
func createBooking(ctx context.Context, tx pgx.Tx, in createInput) (bookingID string, created bool, err error) {
	if in.PaymentMethod == "" {
		in.PaymentMethod = "cash"
	}
	if in.PaymentMethod != "cash" {
		return "", false, invalid("payment_method_unavailable", "only cash is available until an online payment provider is connected")
	}
	if len(in.Note) > 500 {
		return "", false, invalid("invalid_note", "note must be at most 500 characters")
	}
	var existingID, existingQuote string
	err = tx.QueryRow(ctx, `SELECT id, quote_id FROM bookings WHERE idempotency_key=$1`, in.Key).Scan(&existingID, &existingQuote)
	if err == nil {
		if existingQuote != in.QuoteID {
			return "", false, conflict("idempotency_key_reused", "this Idempotency-Key was used for a different booking")
		}
		return existingID, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", false, err
	}
	var scheduledFor *time.Time
	var expires time.Time
	var quoteCustomer string
	err = tx.QueryRow(ctx, `SELECT customer_id, scheduled_for, expires_at FROM booking_quotes WHERE id=$1`, in.QuoteID).
		Scan(&quoteCustomer, &scheduledFor, &expires)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && quoteCustomer != in.CustomerID) {
		return "", false, invalid("invalid_quote", "quote not found")
	}
	if err != nil {
		return "", false, err
	}
	if time.Now().After(expires) {
		return "", false, invalid("quote_expired", "this fare quote has expired; request a new one")
	}
	status := "searching"
	if scheduledFor != nil {
		status = "scheduled"
	}
	err = tx.QueryRow(ctx, `INSERT INTO bookings(tenant_id, reference, customer_id, quote_id, created_by, channel, idempotency_key,
		status, payment_method, scheduled_for, search_started_at, passenger_note)
		VALUES ($1,$2,$3,$4,nullif($5,'')::uuid,$6,$7,$8,$9,$10, CASE WHEN $8='searching' THEN now() END, nullif($11,''))
		ON CONFLICT (tenant_id, idempotency_key) DO NOTHING RETURNING id`,
		in.TenantID, newReference(), in.CustomerID, in.QuoteID, in.ActorUserID, in.Channel, in.Key, status,
		in.PaymentMethod, scheduledFor, in.Note).Scan(&bookingID)
	if errors.Is(err, pgx.ErrNoRows) { // a concurrent request with the same key won
		return createBooking(ctx, tx, in)
	}
	if tenancy.PgCode(err) == "23505" {
		return "", false, conflict("quote_already_used", "this quote has already been booked")
	}
	if err != nil {
		return "", false, err
	}
	if err := event(ctx, tx, in.TenantID, bookingID, "booking.requested", "", status, in.ActorUserID, "",
		map[string]any{"channel": in.Channel}); err != nil {
		return "", false, err
	}
	if status == "searching" {
		if _, err := dispatchNext(ctx, tx, in.TenantID, bookingID); err != nil {
			return "", false, err
		}
	}
	return bookingID, true, nil
}

// --- dispatcher phone bookings ---

func upsertCustomer(ctx context.Context, tx pgx.Tx, tenantID, name, phone string) (string, error) {
	p, ok := fleet.NormalizePhone(phone)
	if !ok {
		return "", invalid("invalid_phone", "customer phone must be E.164 or 07XXXXXXXX")
	}
	if strings.TrimSpace(name) == "" {
		return "", invalid("invalid_customer", "customer name is required")
	}
	var id string
	err := tx.QueryRow(ctx, `INSERT INTO customers(tenant_id, full_name, phone) VALUES ($1,$2,$3)
		ON CONFLICT (tenant_id, phone) DO UPDATE SET full_name = customers.full_name RETURNING id`,
		tenantID, strings.TrimSpace(name), p).Scan(&id)
	return id, err
}

func (s *Service) staffBook(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	key, err := idempotencyKey(r)
	if err != nil {
		return Result{}, err
	}
	var in struct {
		QuoteInput
		Customer struct {
			FullName string `json:"fullName"`
			Phone    string `json:"phone"`
		} `json:"customer"`
		Note string `json:"note"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	// A replayed request returns the original booking without re-quoting.
	var existing string
	if err := tx.QueryRow(r.Context(), `SELECT id FROM bookings WHERE idempotency_key=$1`, key).Scan(&existing); err == nil {
		return s.bookingResult(r, tx, existing, http.StatusOK)
	}
	customerID, err := upsertCustomer(r.Context(), tx, m.TenantID, in.Customer.FullName, in.Customer.Phone)
	if err != nil {
		return Result{}, err
	}
	q, err := createQuote(r.Context(), tx, m.TenantID, customerID, in.QuoteInput)
	if err != nil {
		return Result{}, err
	}
	id, _, err := createBooking(r.Context(), tx, createInput{TenantID: m.TenantID, CustomerID: customerID, QuoteID: q.ID,
		Channel: "dispatcher", Key: key, ActorUserID: m.UserID, Note: in.Note})
	if err != nil {
		return Result{}, err
	}
	return s.bookingResult(r, tx, id, http.StatusCreated)
}
