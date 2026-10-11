package booking

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/captain-KasH/driveappsaas/services/api/internal/audit"
	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/fleet"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
	"github.com/captain-KasH/driveappsaas/services/api/internal/identity"
	"github.com/captain-KasH/driveappsaas/services/api/internal/tenancy"
)

var tenantCode = tenancy.PgCode

// publicTenant is the brand bootstrap for passenger apps and booking sites.
func (s *Service) publicTenant(w http.ResponseWriter, r *http.Request) error {
	id, brand, err := tenancy.PublicBrand(r.Context(), s.Pool, r.PathValue("slug"))
	if err != nil {
		return err
	}
	name := brand.Name
	var cats []fleet.Category
	err = db.InTx(r.Context(), s.Pool, db.Scope{TenantID: id}, func(tx pgx.Tx) error {
		cats, err = fleet.Categories(r, tx)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusOK, map[string]any{"tenant": map[string]string{"id": id, "name": name}, "brand": brand, "categories": cats})
}

// registerPassenger creates (or links) an account as a passenger of one operator.
// Phone ownership is not verified until an SMS provider is connected (KNOWN_ISSUES K-009).
func (s *Service) registerPassenger(w http.ResponseWriter, r *http.Request) error {
	tenantID, _, err := tenancy.PublicTenant(r.Context(), s.Pool, r.PathValue("slug"))
	if err != nil {
		return err
	}
	var in struct {
		FullName string `json:"fullName"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	phone, ok := fleet.NormalizePhone(in.Phone)
	if !ok {
		return invalid("invalid_phone", "phone must be E.164 (+947XXXXXXXX) or 07XXXXXXXX")
	}
	email := strings.TrimSpace(in.Email)
	if !strings.Contains(email, "@") || len(email) > 254 || strings.TrimSpace(in.FullName) == "" {
		return invalid("invalid_registration", "full name and a valid email are required")
	}
	var out map[string]any
	err = db.InTx(r.Context(), s.Pool, db.Scope{TenantID: tenantID}, func(tx pgx.Tx) error {
		ctx := r.Context()
		var userID string
		var hash *string
		err := tx.QueryRow(ctx, `SELECT id, password_hash FROM users WHERE email=$1`, email).Scan(&userID, &hash)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			pw, err := identity.HashPassword(in.Password)
			if err != nil {
				return err
			}
			if err := tx.QueryRow(ctx, `INSERT INTO users(email, full_name, password_hash) VALUES ($1,$2,$3) RETURNING id`,
				email, strings.TrimSpace(in.FullName), pw).Scan(&userID); err != nil {
				return err
			}
		case err != nil:
			return err
		case !identity.CheckPassword(hash, in.Password):
			return identity.ErrBadCredentials
		}
		tag, err := tx.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id, user_id, role) VALUES ($1,$2,'passenger') ON CONFLICT DO NOTHING`, tenantID, userID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return conflict("already_registered", "this account already belongs to this business; sign in instead")
		}
		var customerID string
		// A phone booking may already have created this customer; claim it if unclaimed.
		err = tx.QueryRow(ctx, `INSERT INTO customers(tenant_id, user_id, full_name, phone) VALUES ($1,$2,$3,$4)
			ON CONFLICT (tenant_id, phone) DO UPDATE SET user_id = EXCLUDED.user_id WHERE customers.user_id IS NULL
			RETURNING id`, tenantID, userID, strings.TrimSpace(in.FullName), phone).Scan(&customerID)
		if errors.Is(err, pgx.ErrNoRows) {
			return conflict("phone_taken", "this phone number is registered to another account")
		}
		if err != nil {
			return err
		}
		if err := audit.Record(ctx, tx, audit.Event{TenantID: tenantID, ActorUserID: userID, Action: "passenger.registered",
			TargetType: "customer", TargetID: customerID}); err != nil {
			return err
		}
		token, err := identity.CreateSession(ctx, tx, userID, r.UserAgent(), identity.RemoteIP(r))
		out = map[string]any{"token": token, "tenantId": tenantID, "customerId": customerID}
		return err
	})
	if err != nil {
		return err
	}
	return httpx.WriteJSON(w, http.StatusCreated, out)
}

func customerFor(r *http.Request, tx pgx.Tx, m Member) (string, error) {
	var id string
	err := tx.QueryRow(r.Context(), `SELECT id FROM customers WHERE user_id=$1`, m.UserID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", httpx.ErrNotFound
	}
	return id, err
}

func (s *Service) passengerQuote(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	customerID, err := customerFor(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	var in QuoteInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	q, err := createQuote(r.Context(), tx, m.TenantID, customerID, in)
	return Result{Status: http.StatusCreated, Body: map[string]any{"quote": q}}, err
}

func (s *Service) passengerBook(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	key, err := idempotencyKey(r)
	if err != nil {
		return Result{}, err
	}
	customerID, err := customerFor(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	var in struct {
		QuoteID       string `json:"quoteId"`
		PaymentMethod string `json:"paymentMethod"`
		Note          string `json:"note"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if !tenancy.UUID(in.QuoteID) {
		return Result{}, invalid("invalid_quote", "quoteId is required")
	}
	id, created, err := createBooking(r.Context(), tx, createInput{TenantID: m.TenantID, CustomerID: customerID, QuoteID: in.QuoteID,
		Channel: "passenger_app", Key: key, ActorUserID: m.UserID, PaymentMethod: in.PaymentMethod, Note: in.Note})
	if err != nil {
		return Result{}, err
	}
	v, err := bookingView(r.Context(), tx, id, viewPassenger)
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	return Result{Status: status, Body: map[string]any{"booking": v}}, err
}

// ownBooking resolves a booking only if it belongs to the signed-in passenger.
func ownBooking(r *http.Request, tx pgx.Tx, m Member) (string, error) {
	id, err := pathID(r, "bookingID")
	if err != nil {
		return "", err
	}
	var ok bool
	err = tx.QueryRow(r.Context(), `SELECT true FROM bookings b JOIN customers c ON c.id=b.customer_id WHERE b.id=$1 AND c.user_id=$2`, id, m.UserID).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", httpx.ErrNotFound
	}
	return id, err
}

func (s *Service) myBookings(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	rows, _ := tx.Query(r.Context(), rowSQL+` WHERE c.user_id=$1 ORDER BY b.created_at DESC LIMIT 100`, m.UserID)
	list, err := scanRows(rows)
	return Result{Status: http.StatusOK, Body: map[string]any{"bookings": list}}, err
}

func (s *Service) myBooking(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	id, err := ownBooking(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	v, err := bookingView(r.Context(), tx, id, viewPassenger)
	return Result{Status: http.StatusOK, Body: map[string]any{"booking": v}}, err
}

func (s *Service) passengerCancel(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	id, err := ownBooking(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	var in struct {
		Reason string `json:"reason"`
	}
	if r.ContentLength > 0 {
		if err := httpx.DecodeJSON(r, &in); err != nil {
			return Result{}, err
		}
	}
	if strings.TrimSpace(in.Reason) == "" {
		in.Reason = "cancelled by passenger"
	}
	if err := cancel(r.Context(), tx, m.TenantID, id, m.UserID, in.Reason); err != nil {
		return Result{}, err
	}
	v, err := bookingView(r.Context(), tx, id, viewPassenger)
	return Result{Status: http.StatusOK, Body: map[string]any{"booking": v}}, err
}

// Receipt is only issued for completed trips and only shows the server's final fare.
func (s *Service) myReceipt(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	id, err := ownBooking(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	v, err := bookingView(r.Context(), tx, id, viewPassenger)
	if err != nil {
		return Result{}, err
	}
	if v.Status != "completed" || v.Trip == nil || v.Trip.FinalTotal == nil {
		return Result{}, conflict("no_receipt", "a receipt is available once the trip is completed")
	}
	var business string
	if err := tx.QueryRow(r.Context(), `SELECT name FROM tenants WHERE id=$1`, m.TenantID).Scan(&business); err != nil {
		return Result{}, err
	}
	var paid int64
	for _, p := range v.Payments {
		paid += p.AmountMinor
	}
	payment := "unpaid"
	if len(v.Payments) > 0 {
		payment = "paid_cash"
	}
	return Result{Status: http.StatusOK, Body: map[string]any{"receipt": map[string]any{
		"business": business, "reference": v.Reference, "issuedAt": time.Now().UTC(),
		"pickup": v.Pickup.Address, "dropoff": v.Dropoff.Address,
		"startedAt": v.Trip.StartedAt, "completedAt": v.Trip.CompletedAt,
		"distanceM": v.Trip.DistanceM, "durationS": v.Trip.DurationS, "distanceSource": v.Trip.DistanceSource,
		"driver": v.Driver, "fare": v.Trip.FinalBreakdown, "totalMinor": *v.Trip.FinalTotal,
		"paymentMethod": v.PaymentMethod, "paymentStatus": payment, "paidMinor": paid,
	}}}, nil
}

func (s *Service) rate(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	id, err := ownBooking(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	var in struct {
		Stars   int    `json:"stars"`
		Comment string `json:"comment"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO ratings(booking_id, tenant_id, stars, comment)
		SELECT id, tenant_id, $2, nullif($3,'') FROM bookings WHERE id=$1 AND status='completed'`, id, in.Stars, strings.TrimSpace(in.Comment))
	switch tenantCode(err) {
	case "":
	case "23505":
		return Result{}, conflict("already_rated", "this trip was already rated")
	case "23514":
		return Result{}, invalid("invalid_rating", "stars must be 1-5, comment at most 1000 characters")
	default:
		return Result{}, err
	}
	var ok bool
	if err := tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM ratings WHERE booking_id=$1)`, id).Scan(&ok); err != nil || !ok {
		if err == nil {
			return Result{}, conflict("not_completed", "only completed trips can be rated")
		}
		return Result{}, err
	}
	return Result{Status: http.StatusCreated, Body: map[string]any{"stars": in.Stars}}, nil
}

// --- reports ---

// summary answers "how did we do" for a date range in the tenant's timezone,
// with the definitions next to the numbers.
func (s *Service) summary(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	ctx := r.Context()
	loc, err := tenantLocation(ctx, tx, m.TenantID)
	if err != nil {
		return Result{}, err
	}
	today := time.Now().In(loc)
	from, to := today.AddDate(0, 0, -6), today
	parse := func(k string, def time.Time) (time.Time, error) {
		v := r.URL.Query().Get(k)
		if v == "" {
			return def, nil
		}
		t, err := time.ParseInLocation("2006-01-02", v, loc)
		if err != nil {
			return t, invalid("invalid_date", "%s must be YYYY-MM-DD", k)
		}
		return t, nil
	}
	if from, err = parse("from", from); err != nil {
		return Result{}, err
	}
	if to, err = parse("to", to); err != nil {
		return Result{}, err
	}
	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, loc)
	end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
	if !end.After(start) || end.Sub(start) > 366*24*time.Hour {
		return Result{}, invalid("invalid_range", "from must be on or before to, at most 366 days")
	}
	var x struct {
		Total, Completed, Cancelled, Expired int64
		FareMinor, CashMinor                 int64
		AvgPickupS                           *float64
		AvgRating                            *float64
		Currency                             string
	}
	err = tx.QueryRow(ctx, `
		SELECT count(*), count(*) FILTER (WHERE b.status='completed'), count(*) FILTER (WHERE b.status='cancelled'),
		       count(*) FILTER (WHERE b.status='expired'),
		       coalesce(sum(t.final_total_minor) FILTER (WHERE b.status='completed'), 0),
		       coalesce((SELECT sum(p.amount_minor) FROM payments p JOIN bookings b2 ON b2.id=p.booking_id
		                 WHERE b2.created_at >= $1 AND b2.created_at < $2), 0),
		       avg(extract(epoch FROM t.arrived_at - b.assigned_at)) FILTER (WHERE t.arrived_at IS NOT NULL),
		       (SELECT avg(stars) FROM ratings rt JOIN bookings b3 ON b3.id=rt.booking_id WHERE b3.created_at >= $1 AND b3.created_at < $2),
		       (SELECT currency FROM tenants WHERE id=$3)
		FROM bookings b LEFT JOIN trips t ON t.booking_id=b.id
		WHERE b.created_at >= $1 AND b.created_at < $2`, start, end, m.TenantID).
		Scan(&x.Total, &x.Completed, &x.Cancelled, &x.Expired, &x.FareMinor, &x.CashMinor, &x.AvgPickupS, &x.AvgRating, &x.Currency)
	if err != nil {
		return Result{}, err
	}
	var rate *float64
	if x.Total > 0 {
		v := float64(x.Completed) / float64(x.Total)
		rate = &v
	}
	return Result{Status: http.StatusOK, Body: map[string]any{
		"period":      map[string]any{"from": start.Format("2006-01-02"), "to": to.Format("2006-01-02"), "timezone": loc.String()},
		"generatedAt": time.Now().UTC(), "dataFreshness": "live",
		"currency": x.Currency,
		"metrics": map[string]any{
			"bookings": x.Total, "completed": x.Completed, "cancelled": x.Cancelled, "expired": x.Expired,
			"completionRate": rate, "faresCompletedMinor": x.FareMinor, "cashRecordedMinor": x.CashMinor,
			"avgPickupSeconds": x.AvgPickupS, "avgRating": x.AvgRating,
		},
		"definitions": map[string]string{
			"bookings":            "bookings created in the period (by creation time, tenant timezone)",
			"completionRate":      "completed ÷ bookings",
			"faresCompletedMinor": "sum of server-computed final fares of completed trips, in minor units",
			"cashRecordedMinor":   "cash the drivers confirmed collecting for bookings in the period",
			"avgPickupSeconds":    "average time from driver assignment to arrival at pickup",
		},
	}}, nil
}
