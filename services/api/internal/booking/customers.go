package booking

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/captain-KasH/driveappsaas/services/api/internal/tenancy"
)

// Customer is the operator's view of a passenger: contact + booking history aggregates.
// Phone is visible to staff with booking.read (they already see it on every booking).
type Customer struct {
	ID            string     `json:"id"`
	FullName      string     `json:"fullName"`
	Phone         string     `json:"phone"`
	HasAccount    bool       `json:"hasAccount"`
	PhoneVerified bool       `json:"phoneVerified"`
	Bookings      int        `json:"bookings"`
	Completed     int        `json:"completed"`
	Cancelled     int        `json:"cancelled"`
	SpentMinor    int64      `json:"spentMinor"`
	LastBookingAt *time.Time `json:"lastBookingAt"`
	CreatedAt     time.Time  `json:"createdAt"`
}

const customerSQL = `SELECT c.id, c.full_name, c.phone, c.user_id IS NOT NULL, c.phone_verified_at IS NOT NULL,
	count(b.id), count(b.id) FILTER (WHERE b.status='completed'), count(b.id) FILTER (WHERE b.status='cancelled'),
	coalesce(sum(t.final_total_minor) FILTER (WHERE b.status='completed'), 0), max(b.created_at), c.created_at
	FROM customers c LEFT JOIN bookings b ON b.customer_id=c.id LEFT JOIN trips t ON t.booking_id=b.id
	WHERE ($1 = '' OR c.full_name ILIKE '%' || $1 || '%' OR c.phone LIKE '%' || $1 || '%')
	GROUP BY c.id ORDER BY max(b.created_at) DESC NULLS LAST, c.created_at DESC`

func scanCustomers(rows pgx.Rows) ([]Customer, error) {
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Customer, error) {
		var c Customer
		return c, row.Scan(&c.ID, &c.FullName, &c.Phone, &c.HasAccount, &c.PhoneVerified, &c.Bookings, &c.Completed, &c.Cancelled, &c.SpentMinor, &c.LastBookingAt, &c.CreatedAt)
	})
	if list == nil {
		list = []Customer{}
	}
	return list, err
}

func (s *Service) listCustomers(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	rows, _ := tx.Query(r.Context(), customerSQL+` LIMIT 500`, strings.TrimSpace(r.URL.Query().Get("q")))
	list, err := scanCustomers(rows)
	return Result{Status: http.StatusOK, Body: map[string]any{"customers": list}}, err
}

// --- CSV exports (prompt §10 "reports and exports"); money as decimal strings, times RFC 3339 ---

func writeCSV(w http.ResponseWriter, name string, header []string, rows [][]string) error {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Cache-Control", "private, no-store")
	cw := csv.NewWriter(w)
	cw.Write(header)
	for _, row := range rows {
		cw.Write(row)
	}
	cw.Flush()
	return cw.Error()
}

func dec(minor int64) string { return strconv.FormatFloat(float64(minor)/100, 'f', 2, 64) }

func ts(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func (s *Service) exportCustomers(w http.ResponseWriter, r *http.Request, tx pgx.Tx, m tenancy.Member) error {
	rows, _ := tx.Query(r.Context(), customerSQL, "")
	list, err := scanCustomers(rows)
	if err != nil {
		return err
	}
	out := make([][]string, 0, len(list))
	for _, c := range list {
		out = append(out, []string{c.ID, c.FullName, c.Phone, strconv.FormatBool(c.HasAccount), strconv.Itoa(c.Bookings), strconv.Itoa(c.Completed),
			strconv.Itoa(c.Cancelled), dec(c.SpentMinor), ts(c.LastBookingAt), ts(&c.CreatedAt)})
	}
	return writeCSV(w, "customers-"+time.Now().UTC().Format("2006-01-02")+".csv",
		[]string{"id", "name", "phone", "has_account", "bookings", "completed", "cancelled", "spent", "last_booking_at", "customer_since"}, out)
}

func (s *Service) exportBookings(w http.ResponseWriter, r *http.Request, tx pgx.Tx, m tenancy.Member) error {
	q := r.URL.Query()
	rows, _ := tx.Query(r.Context(), rowSQL+` WHERE ($1 = '' OR b.status = $1) AND ($2 = '' OR b.created_at >= $2::date)
		AND ($3 = '' OR b.created_at < ($3::date + 1)) ORDER BY b.created_at DESC LIMIT 10000`, q.Get("status"), q.Get("from"), q.Get("to"))
	list, err := scanRows(rows)
	if err != nil {
		return err
	}
	out := make([][]string, 0, len(list))
	for _, b := range list {
		fare := dec(b.TotalMinor)
		kind := "estimate"
		if b.IsFinal {
			kind = "final"
		}
		driver := ""
		if b.DriverName != nil {
			driver = *b.DriverName
		}
		trip := ""
		if b.TripStatus != nil {
			trip = *b.TripStatus
		}
		out = append(out, []string{b.Reference, b.Status, trip, ts(&b.CreatedAt), ts(b.Scheduled), b.Customer, driver, b.Pickup, b.Dropoff, fare, kind, b.Currency})
	}
	return writeCSV(w, "bookings-"+time.Now().UTC().Format("2006-01-02")+".csv",
		[]string{"reference", "status", "trip_status", "created_at", "scheduled_for", "customer", "driver", "pickup", "dropoff", "fare", "fare_kind", "currency"}, out)
}
