// Package fleet owns the operator's configuration: service zones, vehicle
// categories, versioned fare rules, vehicles, drivers and driver documents.
package fleet

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/audit"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
	"github.com/captain-KasH/driveappsaas/services/api/internal/tenancy"
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
	mux.Handle("POST "+p+"/zones", t.TenantRoute(tenancy.PermPricing, true, s.createZone))
	mux.Handle("GET "+p+"/zones", t.TenantRoute(tenancy.PermTenantRead, false, s.listZones))
	mux.Handle("POST "+p+"/vehicle-categories", t.TenantRoute(tenancy.PermPricing, true, s.createCategory))
	mux.Handle("GET "+p+"/vehicle-categories", t.TenantRoute(tenancy.PermTenantRead, false, s.listCategories))
	mux.Handle("POST "+p+"/fare-rules", t.TenantRoute(tenancy.PermPricing, true, s.publishFareRule))
	mux.Handle("GET "+p+"/fare-rules", t.TenantRoute(tenancy.PermTenantRead, false, s.listFareRules))

	mux.Handle("POST "+p+"/vehicles", t.TenantRoute(tenancy.PermFleetManage, true, s.createVehicle))
	mux.Handle("GET "+p+"/vehicles", t.TenantRoute(tenancy.PermFleetRead, false, s.listVehicles))
	mux.Handle("POST "+p+"/vehicles/{vehicleID}/status", t.TenantRoute(tenancy.PermFleetManage, true, s.setVehicleStatus))

	mux.Handle("POST "+p+"/drivers", t.TenantRoute(tenancy.PermFleetManage, true, s.createDriver))
	mux.Handle("GET "+p+"/drivers", t.TenantRoute(tenancy.PermFleetRead, false, s.listDrivers))
	mux.Handle("GET "+p+"/drivers/{driverID}", t.TenantRoute(tenancy.PermFleetRead, false, s.getDriver))
	mux.Handle("POST "+p+"/drivers/{driverID}/documents", t.TenantRoute(tenancy.PermFleetManage, true, s.staffAddDocument))
	mux.Handle("POST "+p+"/drivers/{driverID}/documents/{docID}/review", t.TenantRoute(tenancy.PermFleetManage, true, s.reviewDocument))
	mux.Handle("PUT "+p+"/drivers/{driverID}/vehicle", t.TenantRoute(tenancy.PermFleetManage, true, s.assignVehicle))
	mux.Handle("POST "+p+"/drivers/{driverID}/status", t.TenantRoute(tenancy.PermFleetManage, true, s.setDriverStatus))

	mux.Handle("GET "+p+"/driver/me", t.RoleRoute("driver", false, s.driverMe))
	mux.Handle("POST "+p+"/driver/documents", t.RoleRoute("driver", true, s.driverAddDocument))
	mux.Handle("POST "+p+"/driver/availability", t.RoleRoute("driver", true, s.setAvailability))
	mux.Handle("POST "+p+"/driver/location", t.RoleRoute("driver", true, s.updateLocation))
}

func invalid(code, format string, args ...any) error {
	return httpx.Errorf(http.StatusUnprocessableEntity, code, format, args...)
}

func pathID(r *http.Request, name string) (string, error) {
	id := strings.ToLower(r.PathValue(name))
	if !tenancy.UUID(id) {
		return "", httpx.ErrNotFound
	}
	return id, nil
}

func record(r *http.Request, tx pgx.Tx, m Member, action, targetType, targetID string, meta map[string]any) error {
	return audit.Record(r.Context(), tx, audit.Event{TenantID: m.TenantID, ActorUserID: m.UserID,
		Action: action, TargetType: targetType, TargetID: targetID, Metadata: meta})
}

// --- zones ---

type Zone struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Active bool            `json:"active"`
	Area   json.RawMessage `json:"area"`
}

func (s *Service) createZone(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in struct {
		Name string          `json:"name"`
		Area json.RawMessage `json:"area"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	var g struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(in.Area, &g) != nil || (g.Type != "Polygon" && g.Type != "MultiPolygon") {
		return Result{}, invalid("invalid_area", "area must be a GeoJSON Polygon or MultiPolygon in WGS84 (lng, lat)")
	}
	var z Zone
	err := tx.QueryRow(r.Context(), `INSERT INTO service_zones(tenant_id, name, area)
		VALUES ($1, $2, ST_Multi(ST_SetSRID(ST_GeomFromGeoJSON($3), 4326)))
		RETURNING id, name, active, ST_AsGeoJSON(area)::json`, m.TenantID, strings.TrimSpace(in.Name), string(in.Area)).
		Scan(&z.ID, &z.Name, &z.Active, &z.Area)
	switch code := tenancy.PgCode(err); {
	case code == "23505":
		return Result{}, httpx.Errorf(http.StatusConflict, "zone_exists", "a zone named %q already exists", in.Name)
	case code == "23514" || code == "XX000" || code == "22023" || code == "22P02":
		return Result{}, invalid("invalid_area", "area is not a valid polygon, or name is empty")
	case err != nil:
		return Result{}, err
	}
	return Result{Status: http.StatusCreated, Body: map[string]any{"zone": z}}, record(r, tx, m, "zone.created", "zone", z.ID, map[string]any{"name": z.Name})
}

func (s *Service) listZones(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	rows, _ := tx.Query(r.Context(), `SELECT id, name, active, ST_AsGeoJSON(area)::json FROM service_zones ORDER BY name`)
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Zone, error) {
		var z Zone
		return z, row.Scan(&z.ID, &z.Name, &z.Active, &z.Area)
	})
	return Result{Status: http.StatusOK, Body: map[string]any{"zones": list}}, err
}

// --- vehicle categories ---

type Category struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Seats  int    `json:"seats"`
	Active bool   `json:"active"`
}

func (s *Service) createCategory(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in Category
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	var c Category
	err := tx.QueryRow(r.Context(), `INSERT INTO vehicle_categories(tenant_id, code, name, seats) VALUES ($1,$2,$3,$4)
		RETURNING id, code, name, seats, active`, m.TenantID, in.Code, strings.TrimSpace(in.Name), in.Seats).
		Scan(&c.ID, &c.Code, &c.Name, &c.Seats, &c.Active)
	switch tenancy.PgCode(err) {
	case "":
	case "23505":
		return Result{}, httpx.Errorf(http.StatusConflict, "category_exists", "category code %q already exists", in.Code)
	case "23514":
		return Result{}, invalid("invalid_category", "code must be 2-32 of a-z 0-9 _, name 1-80 chars, seats 1-60")
	default:
		return Result{}, err
	}
	return Result{Status: http.StatusCreated, Body: map[string]any{"category": c}}, record(r, tx, m, "category.created", "vehicle_category", c.ID, nil)
}

func (s *Service) listCategories(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	list, err := Categories(r, tx)
	return Result{Status: http.StatusOK, Body: map[string]any{"categories": list}}, err
}

// Categories lists the active categories visible in the current tenant scope.
func Categories(r *http.Request, tx pgx.Tx) ([]Category, error) {
	rows, _ := tx.Query(r.Context(), `SELECT id, code, name, seats, active FROM vehicle_categories WHERE active ORDER BY seats, name`)
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Category, error) {
		var c Category
		return c, row.Scan(&c.ID, &c.Code, &c.Name, &c.Seats, &c.Active)
	})
}

// --- fare rules ---

type FareRuleVersion struct {
	ID                 string    `json:"id"`
	FareRuleID         string    `json:"fareRuleId"`
	ZoneID             string    `json:"zoneId"`
	CategoryID         string    `json:"categoryId"`
	Version            int       `json:"version"`
	Currency           string    `json:"currency"`
	BaseMinor          int64     `json:"baseMinor"`
	PerKmMinor         int64     `json:"perKmMinor"`
	PerMinuteMinor     int64     `json:"perMinuteMinor"`
	MinimumMinor       int64     `json:"minimumMinor"`
	BookingFeeMinor    int64     `json:"bookingFeeMinor"`
	ScheduledFeeMinor  int64     `json:"scheduledFeeMinor"`
	WaitingPerMinMinor int64     `json:"waitingPerMinMinor"`
	FreeWaitingMinutes int64     `json:"freeWaitingMinutes"`
	NightSurchargeBps  int64     `json:"nightSurchargeBps"`
	NightStart         *string   `json:"nightStart"` // "22:00"
	NightEnd           *string   `json:"nightEnd"`
	TaxBps             int64     `json:"taxBps"`
	RoundToMinor       int64     `json:"roundToMinor"`
	CreatedAt          time.Time `json:"createdAt"`
}

const fareCols = `v.id, v.fare_rule_id, fr.zone_id, fr.category_id, v.version, v.currency, v.base_minor, v.per_km_minor,
	v.per_minute_minor, v.minimum_minor, v.booking_fee_minor, v.scheduled_fee_minor, v.waiting_per_min_minor,
	v.free_waiting_minutes, v.night_surcharge_bps, to_char(v.night_start, 'HH24:MI'), to_char(v.night_end, 'HH24:MI'),
	v.tax_bps, v.round_to_minor, v.created_at`

func ScanFare(row pgx.Row) (FareRuleVersion, error) {
	var f FareRuleVersion
	err := row.Scan(&f.ID, &f.FareRuleID, &f.ZoneID, &f.CategoryID, &f.Version, &f.Currency, &f.BaseMinor, &f.PerKmMinor,
		&f.PerMinuteMinor, &f.MinimumMinor, &f.BookingFeeMinor, &f.ScheduledFeeMinor, &f.WaitingPerMinMinor,
		&f.FreeWaitingMinutes, &f.NightSurchargeBps, &f.NightStart, &f.NightEnd, &f.TaxBps, &f.RoundToMinor, &f.CreatedAt)
	return f, err
}

// FareVersionSQL selects one fare_rule_versions row (alias v) joined to its rule (fr).
const FareVersionSQL = `SELECT ` + fareCols + ` FROM fare_rule_versions v JOIN fare_rules fr ON fr.id = v.fare_rule_id`

// publishFareRule creates the next immutable version of the (zone, category) rule.
func (s *Service) publishFareRule(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	var in FareRuleVersion
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return Result{}, err
	}
	if !tenancy.UUID(in.ZoneID) || !tenancy.UUID(in.CategoryID) {
		return Result{}, invalid("invalid_fare_rule", "zoneId and categoryId are required")
	}
	if in.RoundToMinor == 0 {
		in.RoundToMinor = 100
	}
	var currency string
	if err := tx.QueryRow(r.Context(), `SELECT currency FROM tenants WHERE id=$1`, m.TenantID).Scan(&currency); err != nil {
		return Result{}, err
	}
	var ruleID string
	var version int
	err := tx.QueryRow(r.Context(), `INSERT INTO fare_rules(tenant_id, zone_id, category_id, current_version)
		SELECT $1, z.id, c.id, 1 FROM service_zones z, vehicle_categories c WHERE z.id=$2 AND c.id=$3
		ON CONFLICT (tenant_id, zone_id, category_id) DO UPDATE SET current_version = fare_rules.current_version + 1
		RETURNING id, current_version`, m.TenantID, in.ZoneID, in.CategoryID).Scan(&ruleID, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, invalid("invalid_fare_rule", "zone or category not found")
	}
	if err != nil {
		return Result{}, err
	}
	var id string
	err = tx.QueryRow(r.Context(), `INSERT INTO fare_rule_versions(tenant_id, fare_rule_id, version, currency, base_minor,
		per_km_minor, per_minute_minor, minimum_minor, booking_fee_minor, scheduled_fee_minor, waiting_per_min_minor,
		free_waiting_minutes, night_surcharge_bps, night_start, night_end, tax_bps, round_to_minor, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::time,$15::time,$16,$17,$18) RETURNING id`,
		m.TenantID, ruleID, version, currency, in.BaseMinor, in.PerKmMinor, in.PerMinuteMinor, in.MinimumMinor,
		in.BookingFeeMinor, in.ScheduledFeeMinor, in.WaitingPerMinMinor, in.FreeWaitingMinutes, in.NightSurchargeBps,
		in.NightStart, in.NightEnd, in.TaxBps, in.RoundToMinor, m.UserID).Scan(&id)
	switch tenancy.PgCode(err) {
	case "":
	case "23514", "22007", "22008":
		return Result{}, invalid("invalid_fare_rule", "amounts must be ≥ 0, bps 0-10000, night window needs both start and end (HH:MM)")
	default:
		return Result{}, err
	}
	f, err := ScanFare(tx.QueryRow(r.Context(), FareVersionSQL+` WHERE v.id=$1`, id))
	if err != nil {
		return Result{}, err
	}
	return Result{Status: http.StatusCreated, Body: map[string]any{"fareRule": f}},
		record(r, tx, m, "fare_rule.published", "fare_rule", ruleID, map[string]any{"version": version})
}

func (s *Service) listFareRules(r *http.Request, tx pgx.Tx, m Member) (Result, error) {
	rows, _ := tx.Query(r.Context(), FareVersionSQL+` WHERE v.version = fr.current_version ORDER BY fr.created_at`)
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (FareRuleVersion, error) { return ScanFare(row) })
	return Result{Status: http.StatusOK, Body: map[string]any{"fareRules": list}}, err
}
