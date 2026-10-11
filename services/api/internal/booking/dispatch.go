package booking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/fleet"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
)

// eligibleDriverSQL is the deterministic eligibility rule. AI or dispatchers may
// pick among candidates, never outside it. $1 = booking id.
const eligibleDriverSQL = `
	FROM drivers d
	JOIN driver_compliance c ON c.driver_id = d.id
	JOIN vehicles v ON v.id = d.vehicle_id
	JOIN bookings b ON b.id = $1
	JOIN booking_quotes q ON q.id = b.quote_id
	WHERE cardinality(c.issues) = 0
	  AND d.availability = 'available'
	  AND v.category_id = q.category_id
	  AND NOT EXISTS (SELECT 1 FROM dispatch_offers o WHERE o.driver_id = d.id AND o.status = 'pending')
	  AND NOT EXISTS (SELECT 1 FROM trips t WHERE t.driver_id = d.id AND t.status NOT IN ('completed','aborted'))`

// dispatchNext offers a searching booking to the nearest eligible driver with a
// fresh location who has not already been offered it. Returns whether an offer was made.
func dispatchNext(ctx context.Context, tx pgx.Tx, tenantID, bookingID string) (bool, error) {
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1 FOR UPDATE`, bookingID).Scan(&status); err != nil {
		return false, err
	}
	if status != "searching" {
		return false, nil
	}
	var pending bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM dispatch_offers WHERE booking_id=$1 AND status='pending')`, bookingID).Scan(&pending); err != nil || pending {
		return false, err
	}
	var driverID string
	var distance int
	err := tx.QueryRow(ctx, `SELECT d.id, ST_Distance(d.location::geography, q.pickup::geography)::int`+eligibleDriverSQL+`
		  AND d.location_at > now() - $2::interval
		  AND ST_DWithin(d.location::geography, q.pickup::geography, $3)
		  AND NOT EXISTS (SELECT 1 FROM dispatch_offers o WHERE o.booking_id = b.id AND o.driver_id = d.id)
		ORDER BY d.location::geography <-> q.pickup::geography
		LIMIT 1 FOR UPDATE OF d SKIP LOCKED`,
		bookingID, fmt.Sprintf("%d seconds", int(fleet.LocationFreshness.Seconds())), DispatchRadiusM).Scan(&driverID, &distance)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, offer(ctx, tx, tenantID, bookingID, driverID, "auto", &distance, OfferTTL, "")
}

func offer(ctx context.Context, tx pgx.Tx, tenantID, bookingID, driverID, source string, distance *int, ttl time.Duration, actor string) error {
	var offerID string
	err := tx.QueryRow(ctx, `INSERT INTO dispatch_offers(tenant_id, booking_id, driver_id, source, distance_m, expires_at)
		VALUES ($1,$2,$3,$4,$5, now() + $6::interval) RETURNING id`,
		tenantID, bookingID, driverID, source, distance, fmt.Sprintf("%d seconds", int(ttl.Seconds()))).Scan(&offerID)
	if err != nil {
		return err
	}
	return event(ctx, tx, tenantID, bookingID, "dispatch.offered", "", "", actor, "",
		map[string]any{"offerId": offerID, "driverId": driverID, "source": source, "distanceM": distance})
}

// --- driver: offers ---

type Offer struct {
	ID         string    `json:"id"`
	BookingID  string    `json:"bookingId"`
	ExpiresAt  time.Time `json:"expiresAt"`
	DistanceM  *int      `json:"distanceToPickupM"`
	Pickup     string    `json:"pickupAddress"`
	Dropoff    string    `json:"dropoffAddress"`
	EstTotal   int64     `json:"estimatedTotalMinor"`
	Currency   string    `json:"currency"`
	PickupLat  float64   `json:"pickupLat"`
	PickupLng  float64   `json:"pickupLng"`
	Scheduled  *string   `json:"scheduledFor"`
	CustomerNm string    `json:"customerFirstName"`
}

func (s *Service) driverWork(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	driverID, err := fleet.DriverIDFor(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	var o *Offer
	var x Offer
	err = tx.QueryRow(r.Context(), `SELECT o.id, o.booking_id, o.expires_at, o.distance_m, q.pickup_address, q.dropoff_address,
		q.est_total_minor, q.currency, ST_Y(q.pickup), ST_X(q.pickup), to_char(b.scheduled_for, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
		split_part(cu.full_name, ' ', 1)
		FROM dispatch_offers o JOIN bookings b ON b.id = o.booking_id JOIN booking_quotes q ON q.id = b.quote_id
		JOIN customers cu ON cu.id = b.customer_id
		WHERE o.driver_id=$1 AND o.status='pending' AND o.expires_at > now()`, driverID).
		Scan(&x.ID, &x.BookingID, &x.ExpiresAt, &x.DistanceM, &x.Pickup, &x.Dropoff, &x.EstTotal, &x.Currency,
			&x.PickupLat, &x.PickupLng, &x.Scheduled, &x.CustomerNm)
	if err == nil {
		o = &x
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Result{}, err
	}
	var trip any
	var bookingID string
	err = tx.QueryRow(r.Context(), `SELECT booking_id FROM trips WHERE driver_id=$1 AND status NOT IN ('completed','aborted')`, driverID).Scan(&bookingID)
	if err == nil {
		if trip, err = bookingView(r.Context(), tx, bookingID, viewDriver); err != nil {
			return Result{}, err
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Result{}, err
	}
	return Result{Status: http.StatusOK, Body: map[string]any{"offer": o, "activeTrip": trip}}, nil
}

func (s *Service) acceptOffer(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	offerID, err := pathID(r, "offerID")
	if err != nil {
		return Result{}, err
	}
	driverID, err := fleet.DriverIDFor(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	ctx := r.Context()
	var bookingID string
	if err := tx.QueryRow(ctx, `SELECT booking_id FROM dispatch_offers WHERE id=$1 AND driver_id=$2`, offerID, driverID).Scan(&bookingID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Result{}, httpx.ErrNotFound
		}
		return Result{}, err
	}
	var bookingStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1 FOR UPDATE`, bookingID).Scan(&bookingStatus); err != nil {
		return Result{}, err
	}
	var offerStatus string
	var expired bool
	if err := tx.QueryRow(ctx, `SELECT status, expires_at <= now() FROM dispatch_offers WHERE id=$1 FOR UPDATE`, offerID).Scan(&offerStatus, &expired); err != nil {
		return Result{}, err
	}
	if offerStatus != "pending" || expired || bookingStatus != "searching" {
		return Result{}, conflict("offer_unavailable", "this offer is no longer available")
	}
	var vehicleID string
	var eligible bool
	err = tx.QueryRow(ctx, `SELECT d.vehicle_id, cardinality(c.issues) = 0 AND d.availability = 'available'
		FROM drivers d JOIN driver_compliance c ON c.driver_id = d.id WHERE d.id=$1 FOR UPDATE OF d`, driverID).Scan(&vehicleID, &eligible)
	if err != nil || !eligible {
		if err == nil || errors.Is(err, pgx.ErrNoRows) {
			return Result{}, conflict("not_eligible", "you are not currently eligible for trips")
		}
		return Result{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE dispatch_offers SET status='accepted', responded_at=now() WHERE id=$1`, offerID); err != nil {
		return Result{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE bookings SET status='assigned', driver_id=$2, vehicle_id=$3, assigned_at=now(), updated_at=now()
		WHERE id=$1`, bookingID, driverID, vehicleID); err != nil {
		return Result{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE drivers SET availability='busy' WHERE id=$1`, driverID); err != nil {
		return Result{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO trips(booking_id, tenant_id, driver_id, vehicle_id, status) VALUES ($1,$2,$3,$4,'driver_assigned')`,
		bookingID, m.TenantID, driverID, vehicleID); err != nil {
		return Result{}, err
	}
	if err := event(ctx, tx, m.TenantID, bookingID, "booking.assigned", "searching", "assigned", m.UserID, "",
		map[string]any{"offerId": offerID, "driverId": driverID, "vehicleId": vehicleID}); err != nil {
		return Result{}, err
	}
	v, err := bookingView(ctx, tx, bookingID, viewDriver)
	return Result{Status: http.StatusOK, Body: map[string]any{"booking": v}}, err
}

func (s *Service) rejectOffer(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	offerID, err := pathID(r, "offerID")
	if err != nil {
		return Result{}, err
	}
	driverID, err := fleet.DriverIDFor(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	var bookingID string
	err = tx.QueryRow(r.Context(), `UPDATE dispatch_offers SET status='rejected', responded_at=now()
		WHERE id=$1 AND driver_id=$2 AND status='pending' RETURNING booking_id`, offerID, driverID).Scan(&bookingID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, conflict("offer_unavailable", "this offer is no longer pending")
	}
	if err != nil {
		return Result{}, err
	}
	if err := event(r.Context(), tx, m.TenantID, bookingID, "dispatch.rejected", "", "", m.UserID, "",
		map[string]any{"offerId": offerID, "driverId": driverID}); err != nil {
		return Result{}, err
	}
	_, err = dispatchNext(r.Context(), tx, m.TenantID, bookingID)
	return Result{Status: http.StatusOK, Body: map[string]any{"status": "rejected"}}, err
}

// --- dispatcher ---

// manualAssign offers the booking to a chosen driver. The same eligibility rule
// applies; stale location is allowed (the dispatcher may know better) but reported.
func (s *Service) manualAssign(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	bookingID, err := pathID(r, "bookingID")
	if err != nil {
		return Result{}, err
	}
	var in struct {
		DriverID string `json:"driverId"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	ctx := r.Context()
	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1 FOR UPDATE`, bookingID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, httpx.ErrNotFound
	}
	if err != nil {
		return Result{}, err
	}
	if status != "searching" {
		return Result{}, conflict("invalid_state", "only bookings that are searching for a driver can be assigned (status: %s)", status)
	}
	if _, err := tx.Exec(ctx, `UPDATE dispatch_offers SET status='withdrawn', responded_at=now() WHERE booking_id=$1 AND status='pending'`, bookingID); err != nil {
		return Result{}, err
	}
	var distance *int
	var fresh bool
	err = tx.QueryRow(ctx, `SELECT ST_Distance(d.location::geography, q.pickup::geography)::int,
		coalesce(d.location_at > now() - $3::interval, false)`+eligibleDriverSQL+` AND d.id = $2 FOR UPDATE OF d`,
		bookingID, in.DriverID, fmt.Sprintf("%d seconds", int(fleet.LocationFreshness.Seconds()))).Scan(&distance, &fresh)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, conflict("driver_not_eligible", "driver is not eligible (approved, documents valid, matching vehicle, available, no live offer or trip)")
	}
	if err != nil {
		return Result{}, err
	}
	if err := offer(ctx, tx, m.TenantID, bookingID, in.DriverID, "dispatcher", distance, ManualOfferTTL, m.UserID); err != nil {
		return Result{}, err
	}
	res, err := s.bookingResult(r, tx, bookingID, http.StatusOK)
	if body, ok := res.Body.(map[string]any); ok {
		body["driverLocationFresh"] = fresh
	}
	return res, err
}

// --- background worker ---

// Tick runs one pass of time-driven dispatch work: expire offers and re-offer,
// activate scheduled bookings, retry searching bookings, and time out searches.
// It scans across tenants in platform scope, then acts per booking in that tenant's scope.
func (s *Service) Tick(ctx context.Context) error {
	type item struct{ tenantID, bookingID, offerID, kind string }
	var work []item
	err := db.InTx(ctx, s.Pool, db.Scope{Platform: true}, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT tenant_id, booking_id, id::text, 'expire_offer' FROM dispatch_offers WHERE status='pending' AND expires_at <= now()
			UNION ALL
			SELECT tenant_id, id, '', 'activate' FROM bookings WHERE status='scheduled' AND scheduled_for - $1::interval <= now()
			UNION ALL
			SELECT b.tenant_id, b.id, '', CASE WHEN b.search_started_at < now() - $2::interval THEN 'timeout' ELSE 'retry' END
			FROM bookings b WHERE b.status='searching'
			  AND NOT EXISTS (SELECT 1 FROM dispatch_offers o WHERE o.booking_id=b.id AND o.status='pending')
			LIMIT 500`,
			fmt.Sprintf("%d seconds", int(ScheduledLeadTime.Seconds())), fmt.Sprintf("%d seconds", int(SearchTimeout.Seconds())))
		if err != nil {
			return err
		}
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.tenantID, &it.bookingID, &it.offerID, &it.kind); err != nil {
				return err
			}
			work = append(work, it)
		}
		return rows.Err()
	})
	if err != nil {
		return err
	}
	for _, it := range work {
		err := db.InTx(ctx, s.Pool, db.Scope{TenantID: it.tenantID}, func(tx pgx.Tx) error {
			switch it.kind {
			case "expire_offer":
				var st string
				if err := tx.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1 FOR UPDATE`, it.bookingID).Scan(&st); err != nil {
					return err
				}
				tag, err := tx.Exec(ctx, `UPDATE dispatch_offers SET status='expired', responded_at=now()
					WHERE id=$1 AND status='pending' AND expires_at <= now()`, it.offerID)
				if err != nil || tag.RowsAffected() == 0 {
					return err
				}
				if err := event(ctx, tx, it.tenantID, it.bookingID, "dispatch.expired", "", "", "", "", map[string]any{"offerId": it.offerID}); err != nil {
					return err
				}
				_, err = dispatchNext(ctx, tx, it.tenantID, it.bookingID)
				return err
			case "activate":
				tag, err := tx.Exec(ctx, `UPDATE bookings SET status='searching', search_started_at=now(), updated_at=now()
					WHERE id=$1 AND status='scheduled'`, it.bookingID)
				if err != nil || tag.RowsAffected() == 0 {
					return err
				}
				if err := event(ctx, tx, it.tenantID, it.bookingID, "booking.search_started", "scheduled", "searching", "", "", nil); err != nil {
					return err
				}
				_, err = dispatchNext(ctx, tx, it.tenantID, it.bookingID)
				return err
			case "timeout":
				tag, err := tx.Exec(ctx, `UPDATE bookings SET status='expired', updated_at=now() WHERE id=$1 AND status='searching'
					AND NOT EXISTS (SELECT 1 FROM dispatch_offers o WHERE o.booking_id=$1 AND o.status='pending')`, it.bookingID)
				if err != nil || tag.RowsAffected() == 0 {
					return err
				}
				return event(ctx, tx, it.tenantID, it.bookingID, "booking.expired", "searching", "expired", "", "no driver accepted in time", nil)
			default: // retry: new drivers may have come online
				_, err := dispatchNext(ctx, tx, it.tenantID, it.bookingID)
				return err
			}
		})
		if err != nil {
			slog.ErrorContext(ctx, "dispatch tick item failed", "kind", it.kind, "booking_id", it.bookingID, "err", err)
		}
	}
	return nil
}

// Run ticks every interval until ctx ends.
func (s *Service) Run(ctx context.Context, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := s.Tick(ctx); err != nil {
				slog.ErrorContext(ctx, "dispatch tick failed", "err", err)
			}
		}
	}
}
