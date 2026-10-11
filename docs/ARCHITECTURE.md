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
   ├─ tenancy    tenants, memberships, invitations, roles, TenantRoute/PlatformRoute
   └─ audit      append-only audit events in the caller's transaction
PostgreSQL 16 (RLS FORCE on tenant tables)
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

## Planned modules (dependency order)
fleet (drivers, vehicles, documents) → zones/categories → fares → booking → dispatch → trips →
payments/ledger → settlements → notifications → billing/entitlements → integrations → AI → reporting → marketplace → licensing.

## Async work (planned, T-030)
Transactional outbox table written in the same transaction as the state change; a worker in the
same binary delivers events with stable IDs, schema versions and tenant context; consumers dedupe
on event ID. Redis is introduced only for real-time fan-out / location freshness, not as a queue (D-008).

## Booking state machine (draft — finalized with tests in T-014)
| From | To | Actor | Notes |
|---|---|---|---|
| — | quoted | passenger/dispatcher/voice tool | server fare quote stored with rule version |
| quoted | requested | passenger/dispatcher | idempotency key required |
| requested | searching | system | dispatch policy version recorded |
| searching | assigned | dispatch engine / dispatcher | row lock; one active assignment |
| searching | expired | system | offer timeout exhausted |
| requested/searching/assigned | cancelled | passenger/dispatcher/driver (assigned only) | reason + cancellation policy |
Trip: assigned → en_route → arrived → picked_up → in_progress → completed; any pre-completion → aborted
(authorized actor + reason). `completed` is terminal; no reopen.
