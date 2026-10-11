# Database design

PostgreSQL 16. Migrations: `services/api/internal/db/migrations/` (forward-only, embedded).

## Built (0001_foundation)
| Table | Ownership | Notes |
|---|---|---|
| `users` | global | `email citext UNIQUE`, bcrypt `password_hash`, optional `platform_role` |
| `sessions` | global (by user) | `token_hash` SHA-256 UNIQUE, `expires_at`, `revoked_at` |
| `tenants` | self | `slug` UNIQUE + format CHECK, `business_type`, `status` (onboarding/active/suspended/closed), `isolation` |
| `tenant_memberships` | tenant | PK (tenant_id, user_id), `role` CHECK |
| `tenant_invitations` | tenant | `token_hash` UNIQUE, single-use (`accepted_at`), 7-day expiry |
| `audit_events` | tenant or platform (`tenant_id` NULL) | identity PK, append-only via trigger + revoked grants |
| `schema_migrations` | system | applied versions |

RLS policies (all `FORCE`):
- `tenants`: platform scope, or `id = app_tenant_id()`, or the user is a member.
- `tenant_memberships`: platform, tenant match, or own rows (`user_id = app_user_id()`); writes need tenant match.
- `tenant_invitations`: platform or tenant match.
- `audit_events`: read/insert need platform or tenant match; platform rows (NULL tenant) only in platform scope.

## Conventions
- UUID PKs (`gen_random_uuid()`), except high-volume append logs (identity bigint).
- Money: `bigint` minor units + `char(3)` currency. Never float/numeric-in-float.
- Financial and audit records are append-only; corrections are new adjustment rows.
- Soft delete only where retention requires it (`*_at` timestamps); otherwise status columns.
- Every tenant-owned table: `tenant_id uuid NOT NULL REFERENCES tenants`, leading index on `tenant_id`, RLS + isolation test.

## Built (0002_mobility)
| Table | Notes |
|---|---|
| `service_zones` | `geometry(MultiPolygon,4326)` + GiST; `ST_IsValid` check |
| `vehicle_categories` | `(tenant_id, code)` unique |
| `fare_rules` / `fare_rule_versions` | one rule per (zone, category); versions **immutable** (trigger), referenced by quotes |
| `vehicles` | `(tenant_id, plate)` unique; pending/approved/suspended |
| `drivers` | linked `user_id` after invite; `availability` offline/available/busy; `location` point + `location_at` |
| `driver_documents` | type, number, expiry, review status |
| `driver_compliance` (view, security_invoker) | `issues[]`: not approved / vehicle / missing-or-expired required docs (tenant-local date) |
| `customers` | per-tenant, `(tenant_id, phone)` unique; optional `user_id` (app passengers) |
| `booking_quotes` | pickup/dropoff points, estimate breakdown, fare version, expiry |
| `bookings` | status, `(tenant_id, idempotency_key)` unique, `quote_id` unique, reference `TX-XXXXXXX` |
| `booking_events` | append-only history of every transition |
| `dispatch_offers` | partial unique: one pending per booking, one pending per driver |
| `trips` | 1:1 booking; timestamps per phase; final distance/source/duration/waiting/breakdown; partial unique: one live trip per driver |
| `location_updates` | GPS points during `in_progress` (final distance = `ST_Length` of the track) |
| `payments` | append-only; `cash_collected` once per booking; corrections are `adjustment` rows |
| `ratings` | one per booking |
`tenants` gained `timezone` (default Asia/Colombo) and `currency` (LKR).

## Planned entities (per spec Part 12, normalized as built)
Operations: drivers, driver_documents, vehicles, driver_vehicles, service_zones (PostGIS polygon — T-010), service_categories, fare_rules + immutable fare_rule_versions.
Bookings: bookings, booking_quotes, booking_events (append-only), dispatch_offers, trips, trip_events, location_updates (partitioned by day), cancellations, ratings, support_cases.
Finance: payments, payment_attempts, refunds, ledger_entries (double-entry, append-only), driver_settlements, invoices/lines, subscriptions/items, billing_events.
Add-ons: modules, module_versions, module_dependencies, entitlements, module_subscriptions, usage_events (idempotency key UNIQUE per tenant), credit_packs, credit_ledger.
Integrations, communications, AI, deployment/licensing, corporate — designed when their stage starts.

## Migration & rollback
Forward-only. Rollback = deploy previous binary (migrations must stay backward-compatible for one
release: add columns nullable, backfill, then constrain) or restore from backup for destructive failures.
