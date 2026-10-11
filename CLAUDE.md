# Taxios — Claude instructions

Taxios is a multi-tenant, white-label ride-hailing SaaS (Sri Lanka first). The master spec is the
"MobilityOS master build specification" (working title); the product name is **Taxios**.

## Session start (required)
Read `docs/PROJECT_STATUS.md` and `docs/IMPLEMENTATION_PLAN.md` before doing anything. After every
meaningful change: update both (task status + verification evidence = the exact command and result),
and add to `docs/DECISIONS.md` / `docs/KNOWN_ISSUES.md` when relevant.

## Architecture (confirmed)
- `services/api` — Go 1.26 modular monolith, stdlib `net/http` routing, `pgx/v5`, no ORM.
  Modules: `identity`, `tenancy` (routes/RLS scope), `fleet` (zones, categories, fares, drivers, vehicles),
  `booking` (quotes, bookings, dispatch + worker, trips, cash, receipts, reports), `fare` (pure engine), `audit`.
  `internal/server` wires them. No circular imports. Lock order: booking → offer → driver.
- PostgreSQL 16 + PostGIS with row-level security. Web (`apps/web`, Next.js) and mobile
  (`apps/mobile`, React Native) clients are next (T-019); Redis not used yet.
- Details: `docs/ARCHITECTURE.md`, `docs/DATABASE_DESIGN.md`, `docs/API_CONTRACTS.md`.

## Commands
```
make test-remote   # rsync tree to the packline server, run the full suite in Docker (PostGIS) — the default
make test          # same, locally — needs Postgres 16 + PostGIS (dev Mac has none)
```
Deploy (manual, no CI): `ssh appuser@164.68.97.219` then `cd /data/apps/taxios && ./deploy/deploy.sh`
(tests → build → migrate → recreate → smoke). Live at https://taxios.packline.lk. See `docs/DEPLOYMENT_RUNBOOK.md`.
Server access is ONLY `appuser@164.68.97.219` — never root/sudo. Taxios has its own stack/DB; never touch packline's
containers except the shared nginx vhost (packline repo `ops/nginx/nginx.conf`).
Bootstrap a platform admin on the server: see "Bootstrap a platform admin" in `docs/DEPLOYMENT_RUNBOOK.md`.

## Security invariants (release-blocking)
- The API connects as `taxios_app` (no superuser, no BYPASSRLS); `serve` refuses otherwise.
- Every DB access goes through `db.InTx(ctx, pool, scope, …)`; scope settings are transaction-local.
- Tenant routes use `tenancy.TenantRoute` (membership resolved server-side, permission checked,
  writes blocked on suspended/closed tenants). Never trust a client tenant ID on its own.
- `Platform: true` scope only in `PlatformRoute` handlers and invitation acceptance.
- Unknown tenant or non-member → 404 (never reveal existence). Missing permission → 403.
- `audit_events` is append-only (trigger + revoked grants). Record audit in the same transaction.
- Money: integer minor units only. No float arithmetic on amounts. Never fabricate provider APIs.
- Secrets never in git, logs or test output. Tokens are stored only as SHA-256 hashes.

## Database rules
- Forward-only SQL migrations in `services/api/internal/db/migrations/NNNN_name.sql`, embedded and
  applied in order, one transaction each. Never edit an applied migration — add a new one.
- Every tenant-owned table: `tenant_id` + `ENABLE` and `FORCE ROW LEVEL SECURITY` + a policy using
  `app_tenant_id()`, plus an isolation assertion in `TestMobilityTablesRLS` / `TestTenantIsolationRLS`.
- Driver eligibility = the `driver_compliance` view only. Money/history tables (`payments`, `booking_events`,
  `fare_rule_versions`, `audit_events`) are append-only (trigger + revoked grants).
- App-role grants are re-applied by `db.Migrate` (`grantAppRole`); append-only tables revoke UPDATE/DELETE there.

## Conventions
- Errors: return `*httpx.Error` (code + message); anything else is a 500 with no detail leaked.
- Handlers return `tenancy.Result`; the response is written only after commit.
- Tests: each feature gets an integration test in `internal/server` (or the module) against real
  Postgres. Bug fix → failing test first. A test that never failed isn't evidence — mutate once.
- Comments only for non-obvious WHY. Small increments; commit after green `make test`.

## Current phase
Phase 1 API complete: spec Part 17 operational journey steps 1–18 pass end to end (`TestMVPJourney`).
Next: T-019 clients (operator/dispatcher console + passenger web booking), then RN apps.
