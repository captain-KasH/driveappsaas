# Taxios — Claude instructions

Taxios is a multi-tenant, white-label ride-hailing SaaS (Sri Lanka first). The master spec is the
"MobilityOS master build specification" (working title); the product name is **Taxios**.

## Session start (required)
Read `docs/PROJECT_STATUS.md` and `docs/IMPLEMENTATION_PLAN.md` before doing anything. After every
meaningful change: update both (task status + verification evidence = the exact command and result),
and add to `docs/DECISIONS.md` / `docs/KNOWN_ISSUES.md` when relevant.

## Architecture (confirmed)
- `services/api` — Go 1.26 modular monolith, stdlib `net/http` routing, `pgx/v5`, no ORM.
  Modules live in `internal/<module>`; `internal/server` wires them. No circular imports.
- PostgreSQL 16 with row-level security. Redis, PostGIS, web (`apps/web`, Next.js) and mobile
  (`apps/mobile`, React Native) are planned, not yet present.
- Details: `docs/ARCHITECTURE.md`, `docs/DATABASE_DESIGN.md`, `docs/API_CONTRACTS.md`.

## Commands
```
make test      # vet + gofmt check + fresh taxios_test DB + go test ./... (real Postgres, app role)
make migrate   # apply migrations to taxios_dev
make serve     # run API on :8080 as taxios_app
```
Bootstrap a platform admin: `TAXIOS_ADMIN_PASSWORD=... MIGRATE_DATABASE_URL=postgres://localhost/taxios_dev go run ./cmd/taxios-api create-platform-admin EMAIL "NAME"` (from `services/api`).

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
  `app_tenant_id()`, plus an isolation assertion in `TestTenantIsolationRLS`.
- App-role grants are re-applied by `db.Migrate` (`grantAppRole`); append-only tables revoke UPDATE/DELETE there.

## Conventions
- Errors: return `*httpx.Error` (code + message); anything else is a 500 with no detail leaked.
- Handlers return `tenancy.Result`; the response is written only after commit.
- Tests: each feature gets an integration test in `internal/server` (or the module) against real
  Postgres. Bug fix → failing test first. A test that never failed isn't evidence — mutate once.
- Comments only for non-obvious WHY. Small increments; commit after green `make test`.

## Current phase
Stage 1 (foundation) complete for identity/tenants/audit. Next: Stage 2 core mobility — see plan.
