# Architecture

## Shape
A **Go modular monolith** (`services/api`) over **PostgreSQL**, with web and mobile clients to come.
No microservices until a measured need exists (D-001).

```
clients (Next.js admin/portal/web booking, React Native passenger + driver)
        │ HTTPS JSON, /v1, bearer session token
services/api ── internal/server (mux wiring, middleware)
   ├─ httpx      errors, JSON, request IDs, structured logs
   ├─ db         pool, embedded migrations, InTx(scope) → RLS context
   ├─ identity   users, passwords (bcrypt), opaque sessions
   ├─ tenancy    tenants, memberships, invitations, roles, TenantRoute/RoleRoute/PlatformRoute
   ├─ fleet      zones (PostGIS), vehicle categories, versioned fare rules, vehicles, drivers, documents
   ├─ fare       pure fare engine (integer minor units) + route estimate heuristic
   ├─ booking    quotes, idempotent bookings, dispatch + 5 s worker, trip state machine, cash, receipts, reports
   └─ audit      append-only audit events in the caller's transaction
PostgreSQL 16 + PostGIS (RLS FORCE on every tenant table)

Production: Docker stack on the packline server behind packline's nginx at taxios.packline.lk
(see DEPLOYMENT_RUNBOOK.md).
```

## Request path
1. `httpx.Middleware` assigns `X-Request-ID`, recovers panics, logs one JSON line.
2. `identity.RequireAuth` resolves `Authorization: Bearer txs_…` by SHA-256 hash → `Principal`.
3. `TenantRoute` opens a transaction with `app.user_id` + `app.tenant_id`, resolves membership
   under RLS (404 if none), checks permission, blocks writes on inactive tenants, runs the handler.
4. Handler returns `Result`; JSON is written only after commit.

## Tenant isolation (spec Part 6)
- Application layer: membership resolved from the session, never from client input alone.
- Database layer: `FORCE ROW LEVEL SECURITY` + policies on `app_tenant_id()` / `app_user_id()` /
  `app_platform_scope()`. Settings are transaction-local, so pooled connections can't leak scope.
- The API's DB role (`taxios_app`) cannot bypass RLS; `serve` refuses a bypassing role.
- Platform scope is set only by `PlatformRoute` and invitation acceptance.
- Isolation modes: `tenants.isolation` = `shared` (built) · `dedicated_db` · `dedicated_deployment` (T-090+).
  Planned: tenant-safe cache keys (`t:{tenant}:…`), WebSocket subscriptions authorized per tenant,
  object-storage keys prefixed by tenant, AI retrieval through the same scoped transactions.

## Dispatch
1. A booking enters `searching` (immediately, or 15 min before `scheduled_for` via the worker).
2. `dispatchNext` locks the booking, picks the nearest driver satisfying `driver_compliance` (no issues),
   `available`, location < 2 min old, within 8 km, matching vehicle category, no live offer/trip, not
   already offered this booking — `FOR UPDATE SKIP LOCKED` — and creates a 30 s offer.
3. Accept: lock booking → offer → driver, re-check everything, create the trip; unique partial
   indexes make a second assignment impossible even if a check regresses.
4. Reject/expiry → next driver. No driver for 10 min → `expired`. Dispatchers can offer to a chosen
   eligible driver (90 s), stale location allowed but flagged.

## Planned modules (dependency order)
notifications + realtime (WebSocket, push) → payments/ledger → settlements → billing/entitlements →
integrations → AI → marketplace → licensing.

## Async work (planned, T-030)
Transactional outbox table written in the same transaction as the state change; a worker in the
same binary delivers events with stable IDs, schema versions and tenant context; consumers dedupe
on event ID. Redis is introduced only for real-time fan-out / location freshness, not as a queue (D-008).

## Booking state machine (implemented: `booking/trip.go` `tripTransitions`, `cancel`; tests in `mvp_test.go`)
| From | To | Actor | Notes |
|---|---|---|---|
| — | (quote) | passenger / dispatcher | `booking_quotes` row: fare rule version, estimate, 10-min expiry |
| quote | scheduled \| searching | passenger / dispatcher | `Idempotency-Key` required; one booking per quote |
| scheduled | searching | worker | 15 min before pickup |
| searching | assigned | driver accepting an offer | booking → offer → driver locks; trip row created |
| searching | expired | worker | no acceptance within 10 min |
| scheduled/searching/assigned | cancelled | passenger (own) / dispatcher | reason; not once the trip is `in_progress`; frees the driver |
| assigned | completed | driver (`complete`) | final fare computed and stored |
| assigned | cancelled | driver (`abort`) | reason required |
Trip: `driver_assigned → en_route → arrived → in_progress → completed`; `arrived` may skip `en_route`;
any live state → `aborted` (driver, with reason). `completed`/`aborted` are terminal; no reopen.
Every transition appends a `booking_events` row (append-only).
