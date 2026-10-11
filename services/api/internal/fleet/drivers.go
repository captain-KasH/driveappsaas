package fleet

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
	"github.com/captain-KasH/driveappsaas/services/api/internal/tenancy"
)

// LocationFreshness is how old a driver position may be and still count as live.
const LocationFreshness = 2 * time.Minute

var phoneRe = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// NormalizePhone accepts E.164 or a local Sri Lankan number (07XXXXXXXX).
func NormalizePhone(p string) (string, bool) {
	p = strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(p))
	if strings.HasPrefix(p, "0") && len(p) == 10 {
		p = "+94" + p[1:]
	}
	return p, phoneRe.MatchString(p)
}

// --- vehicles ---

type Vehicle struct {
	ID         string `json:"id"`
	CategoryID string `json:"categoryId"`
	Plate      string `json:"plate"`
	Make       string `json:"make"`
	Model      string `json:"model"`
	Colour     string `json:"colour"`
	Status     string `json:"status"`
}

const vehicleCols = `id, category_id, plate, make, model, colour, status`

func scanVehicle(row pgx.Row) (Vehicle, error) {
	var v Vehicle
	return v, row.Scan(&v.ID, &v.CategoryID, &v.Plate, &v.Make, &v.Model, &v.Colour, &v.Status)
}

func (s *Service) createVehicle(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in Vehicle
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if !tenancy.UUID(in.CategoryID) {
		return Result{}, invalid("invalid_vehicle", "categoryId is required")
	}
	plate := strings.ToUpper(strings.Join(strings.Fields(in.Plate), " "))
	v, err := scanVehicle(tx.QueryRow(r.Context(), `INSERT INTO vehicles(tenant_id, category_id, plate, make, model, colour)
		SELECT $1, c.id, $3, $4, $5, $6 FROM vehicle_categories c WHERE c.id=$2 RETURNING `+vehicleCols,
		m.TenantID, in.CategoryID, plate, in.Make, in.Model, in.Colour))
	switch code := tenancy.PgCode(err); {
	case errors.Is(err, pgx.ErrNoRows):
		return Result{}, invalid("invalid_vehicle", "category not found")
	case code == "23505":
		return Result{}, httpx.Errorf(http.StatusConflict, "plate_exists", "vehicle %s already exists", plate)
	case code == "23514":
		return Result{}, invalid("invalid_vehicle", "plate must be 2-20 characters")
	case err != nil:
		return Result{}, err
	}
	return Result{Status: http.StatusCreated, Body: map[string]any{"vehicle": v}}, record(r, tx, m, "vehicle.created", "vehicle", v.ID, map[string]any{"plate": plate})
}

func (s *Service) listVehicles(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	rows, _ := tx.Query(r.Context(), `SELECT `+vehicleCols+` FROM vehicles ORDER BY plate`)
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Vehicle, error) { return scanVehicle(row) })
	return Result{Status: http.StatusOK, Body: map[string]any{"vehicles": list}}, err
}

func (s *Service) setVehicleStatus(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	id, err := pathID(r, "vehicleID")
	if err != nil {
		return Result{}, err
	}
	var in struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if in.Status != "approved" && in.Status != "suspended" {
		return Result{}, invalid("invalid_status", "status must be approved or suspended")
	}
	v, err := scanVehicle(tx.QueryRow(r.Context(), `UPDATE vehicles SET status=$2 WHERE id=$1 RETURNING `+vehicleCols, id, in.Status))
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, httpx.ErrNotFound
	}
	if err != nil {
		return Result{}, err
	}
	return Result{Status: http.StatusOK, Body: map[string]any{"vehicle": v}},
		record(r, tx, m, "vehicle.status_changed", "vehicle", id, map[string]any{"status": in.Status, "reason": in.Reason})
}

// --- drivers ---

type Driver struct {
	ID              string     `json:"id"`
	FullName        string     `json:"fullName"`
	Phone           string     `json:"phone"`
	Email           string     `json:"email"`
	Status          string     `json:"status"`
	Availability    string     `json:"availability"`
	VehicleID       *string    `json:"vehicleId"`
	HasAccount      bool       `json:"hasAccount"`
	Lat             *float64   `json:"lat"`
	Lng             *float64   `json:"lng"`
	LocationAt      *time.Time `json:"locationAt"`
	LocationFresh   bool       `json:"locationFresh"`
	ComplianceIssue []string   `json:"complianceIssues"`
}

const driverCols = `d.id, d.full_name, d.phone, d.email, d.status, d.availability, d.vehicle_id, d.user_id IS NOT NULL,
	ST_Y(d.location), ST_X(d.location), d.location_at, c.issues`

const driverFrom = ` FROM drivers d JOIN driver_compliance c ON c.driver_id = d.id`

func scanDriver(row pgx.Row) (Driver, error) {
	var d Driver
	err := row.Scan(&d.ID, &d.FullName, &d.Phone, &d.Email, &d.Status, &d.Availability, &d.VehicleID, &d.HasAccount,
		&d.Lat, &d.Lng, &d.LocationAt, &d.ComplianceIssue)
	d.LocationFresh = d.LocationAt != nil && time.Since(*d.LocationAt) < LocationFreshness
	if d.ComplianceIssue == nil {
		d.ComplianceIssue = []string{}
	}
	return d, err
}

func (s *Service) createDriver(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in struct {
		FullName string `json:"fullName"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	phone, ok := NormalizePhone(in.Phone)
	if !ok {
		return Result{}, invalid("invalid_phone", "phone must be E.164 (+947XXXXXXXX) or 07XXXXXXXX")
	}
	var id string
	err := tx.QueryRow(r.Context(), `INSERT INTO drivers(tenant_id, full_name, phone, email) VALUES ($1,$2,$3,$4) RETURNING id`,
		m.TenantID, strings.TrimSpace(in.FullName), phone, strings.TrimSpace(in.Email)).Scan(&id)
	switch tenancy.PgCode(err) {
	case "":
	case "23505":
		return Result{}, httpx.Errorf(http.StatusConflict, "driver_exists", "a driver with phone %s already exists", phone)
	case "23514":
		return Result{}, invalid("invalid_driver", "full name is required")
	default:
		return Result{}, err
	}
	inv, err := tenancy.CreateDriverInvitation(r.Context(), tx, m.TenantID, in.Email, m.UserID, id)
	if err != nil {
		return Result{}, err
	}
	d, err := scanDriver(tx.QueryRow(r.Context(), `SELECT `+driverCols+driverFrom+` WHERE d.id=$1`, id))
	if err != nil {
		return Result{}, err
	}
	return Result{Status: http.StatusCreated, Body: map[string]any{"driver": d, "invitation": inv}},
		record(r, tx, m, "driver.created", "driver", id, nil)
}

func (s *Service) listDrivers(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	rows, _ := tx.Query(r.Context(), `SELECT `+driverCols+driverFrom+` ORDER BY d.full_name`)
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Driver, error) { return scanDriver(row) })
	return Result{Status: http.StatusOK, Body: map[string]any{"drivers": list}}, err
}

type Document struct {
	ID         string     `json:"id"`
	DocType    string     `json:"docType"`
	Number     string     `json:"number"`
	ExpiresOn  *string    `json:"expiresOn"`
	Status     string     `json:"status"`
	ReviewNote *string    `json:"reviewNote"`
	ReviewedAt *time.Time `json:"reviewedAt"`
	FileID     *string    `json:"fileId"`
}

func documents(r *http.Request, tx pgx.Tx, driverID string) ([]Document, error) {
	rows, _ := tx.Query(r.Context(), `SELECT id, doc_type, number, to_char(expires_on, 'YYYY-MM-DD'), status, review_note, reviewed_at, file_id
		FROM driver_documents WHERE driver_id=$1 ORDER BY created_at DESC`, driverID)
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Document, error) {
		var d Document
		return d, row.Scan(&d.ID, &d.DocType, &d.Number, &d.ExpiresOn, &d.Status, &d.ReviewNote, &d.ReviewedAt, &d.FileID)
	})
}

func (s *Service) driverDetail(r *http.Request, tx pgx.Tx, where string, arg any) (map[string]any, error) {
	d, err := scanDriver(tx.QueryRow(r.Context(), `SELECT `+driverCols+driverFrom+` WHERE `+where, arg))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	docs, err := documents(r, tx, d.ID)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"driver": d, "documents": docs, "vehicle": nil}
	if d.VehicleID != nil {
		v, err := scanVehicle(tx.QueryRow(r.Context(), `SELECT `+vehicleCols+` FROM vehicles WHERE id=$1`, *d.VehicleID))
		if err != nil {
			return nil, err
		}
		out["vehicle"] = v
	}
	return out, nil
}

func (s *Service) getDriver(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	id, err := pathID(r, "driverID")
	if err != nil {
		return Result{}, err
	}
	out, err := s.driverDetail(r, tx, "d.id=$1", id)
	return Result{Status: http.StatusOK, Body: out}, err
}

func addDocument(r *http.Request, tx pgx.Tx, m Member, driverID string) (Result, error) {
	var in struct {
		DocType   string  `json:"docType"`
		Number    string  `json:"number"`
		ExpiresOn *string `json:"expiresOn"` // YYYY-MM-DD
		FileID    *string `json:"fileId"`    // optional scan/photo uploaded via POST /files
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if in.FileID != nil {
		if !tenancy.UUID(*in.FileID) {
			return Result{}, invalid("invalid_file", "fileId must be a file uploaded to this business")
		}
		// Drivers may only attach files they uploaded themselves; staff may attach any tenant file.
		var ok bool
		if err := tx.QueryRow(r.Context(), `SELECT EXISTS (SELECT 1 FROM files WHERE id=$1 AND ($2 OR uploaded_by=$3))`,
			*in.FileID, m.Role != "driver", m.UserID).Scan(&ok); err != nil {
			return Result{}, err
		}
		if !ok {
			return Result{}, invalid("invalid_file", "fileId must be a file uploaded to this business")
		}
	}
	var id string
	err := tx.QueryRow(r.Context(), `INSERT INTO driver_documents(tenant_id, driver_id, doc_type, number, expires_on, file_id)
		SELECT $1, d.id, $3, $4, $5::date, $6 FROM drivers d WHERE d.id=$2 RETURNING id`,
		m.TenantID, driverID, in.DocType, strings.TrimSpace(in.Number), in.ExpiresOn, in.FileID).Scan(&id)
	switch code := tenancy.PgCode(err); {
	case errors.Is(err, pgx.ErrNoRows):
		return Result{}, httpx.ErrNotFound
	case code == "23514" || code == "22007" || code == "22008":
		return Result{}, invalid("invalid_document", "docType must be driving_licence, national_id, vehicle_insurance, revenue_licence or police_report; number required; expiresOn YYYY-MM-DD")
	case err != nil:
		return Result{}, err
	}
	return Result{Status: http.StatusCreated, Body: map[string]any{"documentId": id, "status": "pending"}},
		record(r, tx, m, "driver_document.submitted", "driver_document", id, map[string]any{"docType": in.DocType})
}

func (s *Service) staffAddDocument(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	id, err := pathID(r, "driverID")
	if err != nil {
		return Result{}, err
	}
	return addDocument(r, tx, m, id)
}

func (s *Service) reviewDocument(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	driverID, err := pathID(r, "driverID")
	if err != nil {
		return Result{}, err
	}
	docID, err := pathID(r, "docID")
	if err != nil {
		return Result{}, err
	}
	var in struct {
		Decision string `json:"decision"`
		Note     string `json:"note"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if in.Decision != "approved" && in.Decision != "rejected" {
		return Result{}, invalid("invalid_decision", "decision must be approved or rejected")
	}
	tag, err := tx.Exec(r.Context(), `UPDATE driver_documents SET status=$3, review_note=nullif($4,''), reviewed_by=$5, reviewed_at=now()
		WHERE id=$1 AND driver_id=$2`, docID, driverID, in.Decision, in.Note, m.UserID)
	if err != nil {
		return Result{}, err
	}
	if tag.RowsAffected() == 0 {
		return Result{}, httpx.ErrNotFound
	}
	return Result{Status: http.StatusOK, Body: map[string]any{"documentId": docID, "status": in.Decision}},
		record(r, tx, m, "driver_document.reviewed", "driver_document", docID, map[string]any{"decision": in.Decision})
}

func (s *Service) assignVehicle(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	driverID, err := pathID(r, "driverID")
	if err != nil {
		return Result{}, err
	}
	var in struct {
		VehicleID string `json:"vehicleId"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	tag, err := tx.Exec(r.Context(), `UPDATE drivers d SET vehicle_id=v.id FROM vehicles v
		WHERE d.id=$1 AND v.id=$2 AND d.availability <> 'busy'`, driverID, in.VehicleID)
	if err != nil {
		return Result{}, err
	}
	if tag.RowsAffected() == 0 {
		return Result{}, invalid("invalid_assignment", "driver or vehicle not found, or driver is on a trip")
	}
	out, err := s.driverDetail(r, tx, "d.id=$1", driverID)
	if err != nil {
		return Result{}, err
	}
	return Result{Status: http.StatusOK, Body: out}, record(r, tx, m, "driver.vehicle_assigned", "driver", driverID, map[string]any{"vehicleId": in.VehicleID})
}

// setDriverStatus approves or suspends. Approval requires valid required documents.
func (s *Service) setDriverStatus(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	driverID, err := pathID(r, "driverID")
	if err != nil {
		return Result{}, err
	}
	var in struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if in.Status != "approved" && in.Status != "suspended" {
		return Result{}, invalid("invalid_status", "status must be approved or suspended")
	}
	if in.Status == "approved" {
		var docsBad bool
		err := tx.QueryRow(r.Context(), `SELECT 'documents_missing_or_expired' = ANY(issues) FROM driver_compliance WHERE driver_id=$1`, driverID).Scan(&docsBad)
		if errors.Is(err, pgx.ErrNoRows) {
			return Result{}, httpx.ErrNotFound
		}
		if err != nil {
			return Result{}, err
		}
		if docsBad {
			return Result{}, httpx.Errorf(http.StatusConflict, "documents_incomplete",
				"approve a valid driving_licence, national_id and vehicle_insurance first")
		}
	}
	// Suspension also takes the driver offline (a live trip, if any, continues; dispatch stops).
	tag, err := tx.Exec(r.Context(), `UPDATE drivers SET status=$2,
		availability = CASE WHEN $2 = 'suspended' AND availability = 'available' THEN 'offline' ELSE availability END
		WHERE id=$1`, driverID, in.Status)
	if err != nil {
		return Result{}, err
	}
	if tag.RowsAffected() == 0 {
		return Result{}, httpx.ErrNotFound
	}
	out, err := s.driverDetail(r, tx, "d.id=$1", driverID)
	if err != nil {
		return Result{}, err
	}
	return Result{Status: http.StatusOK, Body: out},
		record(r, tx, m, "driver.status_changed", "driver", driverID, map[string]any{"status": in.Status, "reason": in.Reason})
}

// --- driver app ---

func driverIDFor(r *http.Request, tx pgx.Tx, m Member) (string, error) {
	var id string
	err := tx.QueryRow(r.Context(), `SELECT id FROM drivers WHERE user_id=$1`, m.UserID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", httpx.ErrNotFound
	}
	return id, err
}

// DriverIDFor resolves the driver record of the signed-in driver member.
func DriverIDFor(r *http.Request, tx pgx.Tx, m Member) (string, error) { return driverIDFor(r, tx, m) }

func (s *Service) driverMe(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	out, err := s.driverDetail(r, tx, "d.user_id=$1", m.UserID)
	return Result{Status: http.StatusOK, Body: out}, err
}

func (s *Service) driverAddDocument(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	id, err := driverIDFor(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	return addDocument(r, tx, m, id)
}

// setAvailability toggles available/offline. "busy" is owned by the trip lifecycle.
func (s *Service) setAvailability(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in struct {
		Available bool `json:"available"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	// Lock order everywhere: booking → offer → driver. So: read the driver unlocked,
	// touch offers, then update the driver with a guarded UPDATE.
	var id, current string
	var issues []string
	err := tx.QueryRow(r.Context(), `SELECT d.id, d.availability, c.issues`+driverFrom+` WHERE d.user_id=$1`, m.UserID).
		Scan(&id, &current, &issues)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, httpx.ErrNotFound
	}
	if err != nil {
		return Result{}, err
	}
	onTrip := httpx.Errorf(http.StatusConflict, "on_trip", "finish the current trip first")
	if current == "busy" {
		return Result{}, onTrip
	}
	next := "offline"
	if in.Available {
		if len(issues) > 0 {
			return Result{}, httpx.Errorf(http.StatusConflict, "not_eligible", "cannot go online: %s", strings.Join(issues, ", "))
		}
		next = "available"
	} else if _, err := tx.Exec(r.Context(), `UPDATE dispatch_offers SET status='withdrawn', responded_at=now()
		WHERE driver_id=$1 AND status='pending'`, id); err != nil {
		return Result{}, err
	}
	tag, err := tx.Exec(r.Context(), `UPDATE drivers SET availability=$2 WHERE id=$1 AND availability <> 'busy'`, id, next)
	if err != nil {
		return Result{}, err
	}
	if tag.RowsAffected() == 0 {
		return Result{}, onTrip
	}
	return Result{Status: http.StatusOK, Body: map[string]any{"availability": next}}, nil
}

// updateLocation records the latest position; during a live trip it also feeds the distance log.
func (s *Service) updateLocation(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in struct {
		Lat        float64    `json:"lat"`
		Lng        float64    `json:"lng"`
		RecordedAt *time.Time `json:"recordedAt"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if in.Lat < -90 || in.Lat > 90 || in.Lng < -180 || in.Lng > 180 || (in.Lat == 0 && in.Lng == 0) {
		return Result{}, invalid("invalid_location", "lat/lng out of range")
	}
	now := time.Now()
	at := now
	if in.RecordedAt != nil && in.RecordedAt.Before(now) && now.Sub(*in.RecordedAt) < 10*time.Minute {
		at = *in.RecordedAt // offline-buffered points keep their capture time
	}
	id, err := driverIDFor(r, tx, m)
	if err != nil {
		return Result{}, err
	}
	if _, err := tx.Exec(r.Context(), `UPDATE drivers SET location=ST_SetSRID(ST_MakePoint($2,$3),4326), location_at=$4
		WHERE id=$1 AND (location_at IS NULL OR location_at <= $4)`, id, in.Lng, in.Lat, at); err != nil {
		return Result{}, err
	}
	if _, err := tx.Exec(r.Context(), `INSERT INTO location_updates(tenant_id, driver_id, booking_id, location, recorded_at)
		SELECT $1, $2, t.booking_id, ST_SetSRID(ST_MakePoint($3,$4),4326), $5 FROM trips t
		WHERE t.driver_id=$2 AND t.status='in_progress'`, m.TenantID, id, in.Lng, in.Lat, at); err != nil {
		return Result{}, err
	}
	return Result{Status: http.StatusOK, Body: map[string]any{"recordedAt": at}}, nil
}
