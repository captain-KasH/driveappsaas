# Taxios — Claude instructions

Taxios is a multi-tenant, white-label ride-hailing SaaS (Sri Lanka first). The brief is
`docs/MASTER_PROMPT.md` (v2, 14 milestones). Owner decisions that shape the build: D-021 self-signup +
super-admin approval, D-022 per-tenant store apps under the tenant's own store accounts, D-023 manual
invoicing first, D-024 Expo, D-025 build order M1 → M4 → M7 → M9 → M10 → M8 → M11 → M14.

## Session start (required)
Read `docs/PROJECT_STATUS.md`, `docs/FEATURE_STATUS.md` and `docs/IMPLEMENTATION_PLAN.md` before doing
anything. After every meaningful change: update them (task status + verification evidence = the exact
command and result), and add to `docs/DECISIONS.md` / `docs/KNOWN_ISSUES.md` when relevant. Never mark a
feature done without an executed test or a live check named in the doc (prompt §35).

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
Mobile (`apps/mobile`, Expo SDK 57): `npm run passenger` / `npm run driver` (Metro on 8081/8082), open in Expo Go;
`APP_VARIANT` picks the app, `TAXIOS_TENANT=<slug>` pins an operator build, `TAXIOS_API_URL` overrides the API.
Device testing runs against the SERVER dev stack, not the Mac (owner rule: the Mac only runs simulators — low disk):
`make dev-remote` (builds the working tree on the server as project `taxios-dev`, migrates, seeds) → `make dev-tunnel`
(localhost:8080 → server 127.0.0.1:18080) → start the apps with `TAXIOS_API_URL=http://localhost:8080`. Test logins live in
`scripts/dev-seed.py`. `scripts/dev-local.sh` is a fallback only; never install DBs/toolchains on the Mac.
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
Ride engine (M2, M3, M5, M6 core) verified end to end (`TestMVPJourney`); web console/booking/driver pages live.
**Now: M4** — self-service onboarding (T-120…T-123) + platform console v2 (T-124) — then M7 Expo apps, M9 white-label.
Audit: `docs/PROJECT_AUDIT.md`. Competitors: `docs/COMPETITOR_FEATURE_MATRIX.md`.
