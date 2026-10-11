package booking

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/captain-KasH/driveappsaas/services/api/internal/fare"
	"github.com/captain-KasH/driveappsaas/services/api/internal/fleet"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
)

// tripTransitions is the driver-driven trip state machine: action → allowed from-states → to-state.
// completed and aborted are terminal; nothing reopens them.
var tripTransitions = map[string]struct {
	from []string
	to   string
}{
	"en-route": {[]string{"driver_assigned"}, "en_route"},
	"arrived":  {[]string{"driver_assigned", "en_route"}, "arrived"},
	"start":    {[]string{"arrived"}, "in_progress"},
	"complete": {[]string{"in_progress"}, "completed"},
	"abort":    {[]string{"driver_assigned", "en_route", "arrived", "in_progress"}, "aborted"},
}

func allowed(from string, list []string) bool {
	for _, s := range list {
		if s == from {
			return true
		}
	}
	return false
}

func (s *Service) driverTripAction(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	bookingID, err := pathID(r, "bookingID")
	if err != nil {
		return Result{}, err
	}
	driverID, err := fleet.DriverIDFor(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	action := r.PathValue("action")
	if action == "cash" {
		return s.recordCash(r, tx, m, bookingID, driverID)
	}
	tr, ok := tripTransitions[action]
	if !ok {
		return Result{}, httpx.ErrNotFound
	}
	var in struct {
		Reason string `json:"reason"`
	}
	if r.ContentLength > 0 {
		if err := httpx.DecodeJSON(r, &in); err != nil {
			return Result{}, err
		}
	}
	if action == "abort" && len(strings.TrimSpace(in.Reason)) < 3 {
		return Result{}, invalid("reason_required", "a reason is required to abort a trip")
	}
	ctx := r.Context()
	var bookingStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1 AND driver_id=$2 FOR UPDATE`, bookingID, driverID).Scan(&bookingStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Result{}, httpx.ErrNotFound // not this driver's trip
		}
		return Result{}, err
	}
	var from string
	if err := tx.QueryRow(ctx, `SELECT status FROM trips WHERE booking_id=$1 FOR UPDATE`, bookingID).Scan(&from); err != nil {
		return Result{}, err
	}
	if !allowed(from, tr.from) {
		return Result{}, conflict("invalid_transition", "cannot %s a trip that is %s", action, from)
	}
	switch tr.to {
	case "en_route":
		_, err = tx.Exec(ctx, `UPDATE trips SET status='en_route', en_route_at=now() WHERE booking_id=$1`, bookingID)
	case "arrived":
		_, err = tx.Exec(ctx, `UPDATE trips SET status='arrived', arrived_at=now() WHERE booking_id=$1`, bookingID)
	case "in_progress":
		_, err = tx.Exec(ctx, `UPDATE trips SET status='in_progress', started_at=now() WHERE booking_id=$1`, bookingID)
	case "completed":
		err = completeTrip(ctx, tx, m.TenantID, bookingID, driverID)
	case "aborted":
		err = endTrip(ctx, tx, bookingID, driverID, in.Reason)
	}
	if err != nil {
		return Result{}, err
	}
	if err := event(ctx, tx, m.TenantID, bookingID, "trip."+tr.to, from, tr.to, m.UserID, in.Reason, nil); err != nil {
		return Result{}, err
	}
	v, err := bookingView(ctx, tx, bookingID, viewDriver)
	return Result{Status: http.StatusOK, Body: map[string]any{"booking": v}}, err
}

// endTrip aborts a live trip, cancels its booking and frees the driver.
func endTrip(ctx context.Context, tx pgx.Tx, bookingID, driverID, reason string) error {
	if _, err := tx.Exec(ctx, `UPDATE trips SET status='aborted', aborted_at=now(), abort_reason=$2 WHERE booking_id=$1`, bookingID, reason); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE bookings SET status='cancelled', cancel_reason=$2, updated_at=now() WHERE id=$1`, bookingID, "trip aborted: "+reason); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `UPDATE drivers SET availability='available' WHERE id=$1 AND availability='busy'`, driverID)
	return err
}

// completeTrip computes the final fare from the fare version locked in at quote
// time, using the GPS track when there is one and the estimate otherwise.
func completeTrip(ctx context.Context, tx pgx.Tx, tenantID, bookingID, driverID string) error {
	var f struct {
		versionID      string
		estDistance    int64
		scheduled      bool
		startedAt      time.Time
		arrivedAt      *time.Time
		trackedM       *float64
		trackPoints    int
		durationS      int64
		currentTenantZ string
	}
	err := tx.QueryRow(ctx, `
		SELECT q.fare_rule_version_id, q.est_distance_m, b.scheduled_for IS NOT NULL, t.started_at, t.arrived_at,
		       (SELECT ST_Length(ST_MakeLine(l.location ORDER BY l.recorded_at)::geography) FROM location_updates l
		         WHERE l.booking_id = b.id AND l.recorded_at >= t.started_at),
		       (SELECT count(*) FROM location_updates l WHERE l.booking_id = b.id AND l.recorded_at >= t.started_at),
		       extract(epoch FROM now() - t.started_at)::bigint, tn.timezone
		FROM bookings b JOIN booking_quotes q ON q.id = b.quote_id JOIN trips t ON t.booking_id = b.id
		JOIN tenants tn ON tn.id = b.tenant_id WHERE b.id = $1`, bookingID).
		Scan(&f.versionID, &f.estDistance, &f.scheduled, &f.startedAt, &f.arrivedAt, &f.trackedM, &f.trackPoints, &f.durationS, &f.currentTenantZ)
	if err != nil {
		return err
	}
	rule, err := fleet.ScanFare(tx.QueryRow(ctx, fleet.FareVersionSQL+` WHERE v.id=$1`, f.versionID))
	if err != nil {
		return err
	}
	distance, source := f.estDistance, "estimate_fallback"
	if f.trackPoints >= 2 && f.trackedM != nil {
		distance, source = int64(*f.trackedM+0.5), "gps_track"
	}
	var waiting int64
	if f.arrivedAt != nil {
		waiting = int64(f.startedAt.Sub(*f.arrivedAt).Seconds())
	}
	loc, err := time.LoadLocation(f.currentTenantZ)
	if err != nil {
		return err
	}
	b := fare.Compute(toRule(rule), fare.Input{DistanceM: distance, DurationS: f.durationS, WaitingS: waiting,
		Scheduled: f.scheduled, LocalStart: f.startedAt.In(loc)})
	if _, err := tx.Exec(ctx, `UPDATE trips SET status='completed', completed_at=now(), distance_m=$2, duration_s=$3, waiting_s=$4,
		distance_source=$5, final_total_minor=$6, currency=$7, final_breakdown=$8 WHERE booking_id=$1`,
		bookingID, distance, f.durationS, waiting, source, b.TotalMinor, b.Currency, b); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE bookings SET status='completed', updated_at=now() WHERE id=$1`, bookingID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE drivers SET availability='available' WHERE id=$1 AND availability='busy'`, driverID)
	return err
}

// recordCash stores the driver's confirmation of cash collected, once per booking.
func (s *Service) recordCash(r *http.Request, tx pgx.Tx, m Member, bookingID, driverID string) (Result, error) {
	var in struct {
		AmountMinor int64  `json:"amountMinor"`
		Note        string `json:"note"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if in.AmountMinor < 0 {
		return Result{}, invalid("invalid_amount", "amountMinor must be ≥ 0")
	}
	ctx := r.Context()
	var tripStatus, currency string
	var total *int64
	err := tx.QueryRow(ctx, `SELECT t.status, t.final_total_minor, coalesce(t.currency, '') FROM trips t JOIN bookings b ON b.id=t.booking_id
		WHERE t.booking_id=$1 AND t.driver_id=$2 AND b.payment_method='cash' FOR UPDATE OF b`, bookingID, driverID).Scan(&tripStatus, &total, &currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, httpx.ErrNotFound
	}
	if err != nil {
		return Result{}, err
	}
	if tripStatus != "completed" || total == nil {
		return Result{}, conflict("invalid_state", "cash can only be recorded after the trip is completed")
	}
	note := strings.TrimSpace(in.Note)
	if in.AmountMinor != *total && note == "" {
		return Result{}, invalid("note_required", "amount differs from the fare (%d); add a note explaining why", *total)
	}
	_, err = tx.Exec(ctx, `INSERT INTO payments(tenant_id, booking_id, kind, amount_minor, currency, recorded_by, note)
		VALUES ($1,$2,'cash_collected',$3,$4,$5,nullif($6,''))`, m.TenantID, bookingID, in.AmountMinor, currency, m.UserID, note)
	if tenantCode(err) == "23505" {
		return Result{}, conflict("already_recorded", "cash for this trip was already recorded")
	}
	if err != nil {
		return Result{}, err
	}
	if err := event(ctx, tx, m.TenantID, bookingID, "payment.cash_collected", "", "", m.UserID, note,
		map[string]any{"amountMinor": in.AmountMinor, "fareMinor": *total}); err != nil {
		return Result{}, err
	}
	return Result{Status: http.StatusCreated, Body: map[string]any{"amountMinor": in.AmountMinor, "fareMinor": *total,
		"matchesFare": in.AmountMinor == *total}}, nil
}

// cancel applies the shared cancellation rule: allowed before pickup, never once in progress.
func cancel(ctx context.Context, tx pgx.Tx, tenantID, bookingID, actor, reason string) error {
	var status string
	var driverID *string
	if err := tx.QueryRow(ctx, `SELECT status, driver_id FROM bookings WHERE id=$1 FOR UPDATE`, bookingID).Scan(&status, &driverID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		return err
	}
	switch status {
	case "scheduled", "searching":
	case "assigned":
		var tripStatus string
		if err := tx.QueryRow(ctx, `SELECT status FROM trips WHERE booking_id=$1 FOR UPDATE`, bookingID).Scan(&tripStatus); err != nil {
			return err
		}
		if tripStatus == "in_progress" {
			return conflict("invalid_transition", "a trip in progress cannot be cancelled; the driver must complete or abort it")
		}
	default:
		return conflict("invalid_transition", "a %s booking cannot be cancelled", status)
	}
	if _, err := tx.Exec(ctx, `UPDATE dispatch_offers SET status='withdrawn', responded_at=now() WHERE booking_id=$1 AND status='pending'`, bookingID); err != nil {
		return err
	}
	if status == "assigned" && driverID != nil {
		if _, err := tx.Exec(ctx, `UPDATE trips SET status='aborted', aborted_at=now(), abort_reason='booking cancelled' WHERE booking_id=$1`, bookingID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE drivers SET availability='available' WHERE id=$1 AND availability='busy'`, *driverID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE bookings SET status='cancelled', cancelled_by=nullif($2,'')::uuid, cancel_reason=$3, updated_at=now()
		WHERE id=$1`, bookingID, actor, reason); err != nil {
		return err
	}
	return event(ctx, tx, tenantID, bookingID, "booking.cancelled", status, "cancelled", actor, reason, nil)
}

func (s *Service) staffCancel(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	id, err := pathID(r, "bookingID")
	if err != nil {
		return Result{}, err
	}
	var in struct {
		Reason string `json:"reason"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if len(strings.TrimSpace(in.Reason)) < 3 {
		return Result{}, invalid("reason_required", "a cancellation reason is required")
	}
	if err := cancel(r.Context(), tx, m.TenantID, id, m.UserID, in.Reason); err != nil {
		return Result{}, err
	}
	return s.bookingResult(r, tx, id, http.StatusOK)
}

// --- views ---

type viewMode int

const (
	viewPassenger viewMode = iota
	viewDriver
	viewStaff
)

type BookingView struct {
	ID            string          `json:"id"`
	Reference     string          `json:"reference"`
	Status        string          `json:"status"`
	Channel       string          `json:"channel"`
	PaymentMethod string          `json:"paymentMethod"`
	ScheduledFor  *time.Time      `json:"scheduledFor"`
	CreatedAt     time.Time       `json:"createdAt"`
	Pickup        Place           `json:"pickup"`
	Dropoff       Place           `json:"dropoff"`
	CategoryID    string          `json:"categoryId"`
	Estimate      json.RawMessage `json:"estimate"`
	Customer      *struct {
		Name  string `json:"name"`
		Phone string `json:"phone,omitempty"`
	} `json:"customer,omitempty"`
	Note        *string        `json:"note"`
	CancelNote  *string        `json:"cancelReason"`
	Trip        *TripView      `json:"trip"`
	Driver      *DriverPublic  `json:"driver"`
	Payments    []PaymentView  `json:"payments,omitempty"`
	Events      []EventView    `json:"events,omitempty"`
	Offers      []OfferHistory `json:"offers,omitempty"`
	IsFinalFare bool           `json:"isFinalFare"`
}

type TripView struct {
	Status         string          `json:"status"`
	EnRouteAt      *time.Time      `json:"enRouteAt"`
	ArrivedAt      *time.Time      `json:"arrivedAt"`
	StartedAt      *time.Time      `json:"startedAt"`
	CompletedAt    *time.Time      `json:"completedAt"`
	DistanceM      *int            `json:"distanceM"`
	DurationS      *int            `json:"durationS"`
	DistanceSource *string         `json:"distanceSource"`
	FinalTotal     *int64          `json:"finalTotalMinor"`
	FinalBreakdown json.RawMessage `json:"finalBreakdown"`
}

// DriverPublic is what a passenger may see about their driver.
type DriverPublic struct {
	FirstName     string     `json:"firstName"`
	VehiclePlate  string     `json:"vehiclePlate"`
	VehicleDesc   string     `json:"vehicleDescription"`
	Lat           *float64   `json:"lat"`
	Lng           *float64   `json:"lng"`
	LocationAt    *time.Time `json:"locationAt"`
	LocationFresh bool       `json:"locationFresh"`
}

type PaymentView struct {
	Kind        string    `json:"kind"`
	AmountMinor int64     `json:"amountMinor"`
	Currency    string    `json:"currency"`
	Note        *string   `json:"note"`
	At          time.Time `json:"at"`
}

type EventView struct {
	Event  string         `json:"event"`
	From   *string        `json:"from"`
	To     *string        `json:"to"`
	Reason *string        `json:"reason"`
	Data   map[string]any `json:"data"`
	At     time.Time      `json:"at"`
}

type OfferHistory struct {
	DriverID  string    `json:"driverId"`
	Source    string    `json:"source"`
	Status    string    `json:"status"`
	DistanceM *int      `json:"distanceM"`
	OfferedAt time.Time `json:"offeredAt"`
}

func bookingView(ctx context.Context, tx pgx.Tx, id string, mode viewMode) (*BookingView, error) {
	v := &BookingView{}
	var custName, custPhone string
	err := tx.QueryRow(ctx, `SELECT b.id, b.reference, b.status, b.channel, b.payment_method, b.scheduled_for, b.created_at,
		ST_Y(q.pickup), ST_X(q.pickup), q.pickup_address, ST_Y(q.dropoff), ST_X(q.dropoff), q.dropoff_address,
		q.category_id, q.breakdown, c.full_name, c.phone, b.passenger_note, b.cancel_reason
		FROM bookings b JOIN booking_quotes q ON q.id=b.quote_id JOIN customers c ON c.id=b.customer_id WHERE b.id=$1`, id).
		Scan(&v.ID, &v.Reference, &v.Status, &v.Channel, &v.PaymentMethod, &v.ScheduledFor, &v.CreatedAt,
			&v.Pickup.Lat, &v.Pickup.Lng, &v.Pickup.Address, &v.Dropoff.Lat, &v.Dropoff.Lng, &v.Dropoff.Address,
			&v.CategoryID, &v.Estimate, &custName, &custPhone, &v.Note, &v.CancelNote)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	switch mode {
	case viewStaff:
		v.Customer = &struct {
			Name  string `json:"name"`
			Phone string `json:"phone,omitempty"`
		}{custName, custPhone}
	case viewDriver: // first name only; no phone until a masked-calling provider exists
		v.Customer = &struct {
			Name  string `json:"name"`
			Phone string `json:"phone,omitempty"`
		}{Name: strings.Fields(custName + " ")[0]}
	}
	var t TripView
	var fullName, plate, make_, model, colour string
	var dLat, dLng *float64
	var dAt *time.Time
	err = tx.QueryRow(ctx, `SELECT t.status, t.en_route_at, t.arrived_at, t.started_at, t.completed_at, t.distance_m, t.duration_s,
		t.distance_source, t.final_total_minor, t.final_breakdown, d.full_name, v.plate, v.make, v.model, v.colour,
		ST_Y(d.location), ST_X(d.location), d.location_at
		FROM trips t JOIN drivers d ON d.id=t.driver_id JOIN vehicles v ON v.id=t.vehicle_id WHERE t.booking_id=$1`, id).
		Scan(&t.Status, &t.EnRouteAt, &t.ArrivedAt, &t.StartedAt, &t.CompletedAt, &t.DistanceM, &t.DurationS,
			&t.DistanceSource, &t.FinalTotal, &t.FinalBreakdown, &fullName, &plate, &make_, &model, &colour, &dLat, &dLng, &dAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		v.Trip = &t
		v.IsFinalFare = t.FinalTotal != nil
		live := t.Status != "completed" && t.Status != "aborted"
		dp := &DriverPublic{FirstName: strings.Fields(fullName + " ")[0], VehiclePlate: plate,
			VehicleDesc: strings.TrimSpace(colour + " " + make_ + " " + model)}
		if live || mode == viewStaff { // a passenger stops seeing the driver's position once the trip ends
			dp.Lat, dp.Lng, dp.LocationAt = dLat, dLng, dAt
			dp.LocationFresh = dAt != nil && time.Since(*dAt) < fleet.LocationFreshness
		}
		if mode == viewDriver {
			dp = nil
		}
		v.Driver = dp
	}
	rows, _ := tx.Query(ctx, `SELECT kind, amount_minor, currency, note, created_at FROM payments WHERE booking_id=$1 ORDER BY created_at`, id)
	if v.Payments, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (PaymentView, error) {
		var p PaymentView
		return p, row.Scan(&p.Kind, &p.AmountMinor, &p.Currency, &p.Note, &p.At)
	}); err != nil {
		return nil, err
	}
	if mode == viewStaff {
		rows, _ := tx.Query(ctx, `SELECT event, from_status, to_status, reason, data, occurred_at FROM booking_events WHERE booking_id=$1 ORDER BY id`, id)
		if v.Events, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (EventView, error) {
			var e EventView
			return e, row.Scan(&e.Event, &e.From, &e.To, &e.Reason, &e.Data, &e.At)
		}); err != nil {
			return nil, err
		}
		rows, _ = tx.Query(ctx, `SELECT driver_id, source, status, distance_m, offered_at FROM dispatch_offers WHERE booking_id=$1 ORDER BY offered_at`, id)
		if v.Offers, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (OfferHistory, error) {
			var o OfferHistory
			return o, row.Scan(&o.DriverID, &o.Source, &o.Status, &o.DistanceM, &o.OfferedAt)
		}); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (s *Service) bookingResult(r *http.Request, tx pgx.Tx, id string, status int) (Result, error) {
	v, err := bookingView(r.Context(), tx, id, viewStaff)
	return Result{Status: status, Body: map[string]any{"booking": v}}, err
}

func (s *Service) getBooking(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	id, err := pathID(r, "bookingID")
	if err != nil {
		return Result{}, err
	}
	return s.bookingResult(r, tx, id, http.StatusOK)
}

type BookingRow struct {
	ID          string     `json:"id"`
	Reference   string     `json:"reference"`
	Status      string     `json:"status"`
	TripStatus  *string    `json:"tripStatus"`
	Pickup      string     `json:"pickupAddress"`
	Dropoff     string     `json:"dropoffAddress"`
	Customer    string     `json:"customerName"`
	DriverName  *string    `json:"driverName"`
	Scheduled   *time.Time `json:"scheduledFor"`
	TotalMinor  int64      `json:"totalMinor"`
	IsFinal     bool       `json:"isFinal"`
	Currency    string     `json:"currency"`
	CreatedAt   time.Time  `json:"createdAt"`
	SearchSince *time.Time `json:"searchStartedAt"`
}

const rowSQL = `SELECT b.id, b.reference, b.status, t.status, q.pickup_address, q.dropoff_address, c.full_name, d.full_name,
	b.scheduled_for, coalesce(t.final_total_minor, q.est_total_minor), t.final_total_minor IS NOT NULL, q.currency, b.created_at, b.search_started_at
	FROM bookings b JOIN booking_quotes q ON q.id=b.quote_id JOIN customers c ON c.id=b.customer_id
	LEFT JOIN trips t ON t.booking_id=b.id LEFT JOIN drivers d ON d.id=b.driver_id`

func scanRows(rows pgx.Rows) ([]BookingRow, error) {
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (BookingRow, error) {
		var b BookingRow
		return b, row.Scan(&b.ID, &b.Reference, &b.Status, &b.TripStatus, &b.Pickup, &b.Dropoff, &b.Customer, &b.DriverName,
			&b.Scheduled, &b.TotalMinor, &b.IsFinal, &b.Currency, &b.CreatedAt, &b.SearchSince)
	})
}

func (s *Service) listBookings(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	status := r.URL.Query().Get("status")
	rows, _ := tx.Query(r.Context(), rowSQL+` WHERE ($1 = '' OR b.status = $1) ORDER BY b.created_at DESC LIMIT 200`, status)
	list, err := scanRows(rows)
	return Result{Status: http.StatusOK, Body: map[string]any{"bookings": list}}, err
}

// dispatchBoard is the dispatcher's live view: open bookings and every non-offline driver.
func (s *Service) dispatchBoard(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	ctx := r.Context()
	rows, _ := tx.Query(ctx, rowSQL+` WHERE b.status IN ('searching','assigned')
		OR (b.status='scheduled' AND b.scheduled_for < now() + interval '2 hours') ORDER BY b.created_at`)
	open, err := scanRows(rows)
	if err != nil {
		return Result{}, err
	}
	type boardDriver struct {
		ID            string     `json:"id"`
		Name          string     `json:"name"`
		Availability  string     `json:"availability"`
		VehiclePlate  *string    `json:"vehiclePlate"`
		CategoryID    *string    `json:"categoryId"`
		Lat           *float64   `json:"lat"`
		Lng           *float64   `json:"lng"`
		LocationAt    *time.Time `json:"locationAt"`
		LocationFresh bool       `json:"locationFresh"`
		Issues        []string   `json:"complianceIssues"`
		PendingOffer  bool       `json:"hasPendingOffer"`
	}
	rows, _ = tx.Query(ctx, `SELECT d.id, d.full_name, d.availability, v.plate, v.category_id, ST_Y(d.location), ST_X(d.location),
		d.location_at, c.issues, EXISTS(SELECT 1 FROM dispatch_offers o WHERE o.driver_id=d.id AND o.status='pending')
		FROM drivers d JOIN driver_compliance c ON c.driver_id=d.id LEFT JOIN vehicles v ON v.id=d.vehicle_id
		WHERE d.availability <> 'offline' ORDER BY d.full_name`)
	drivers, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (boardDriver, error) {
		var d boardDriver
		err := row.Scan(&d.ID, &d.Name, &d.Availability, &d.VehiclePlate, &d.CategoryID, &d.Lat, &d.Lng, &d.LocationAt, &d.Issues, &d.PendingOffer)
		d.LocationFresh = d.LocationAt != nil && time.Since(*d.LocationAt) < fleet.LocationFreshness
		return d, err
	})
	return Result{Status: http.StatusOK, Body: map[string]any{"bookings": open, "drivers": drivers,
		"locationFreshnessSeconds": int(fleet.LocationFreshness.Seconds()), "generatedAt": time.Now().UTC()}}, err
}
