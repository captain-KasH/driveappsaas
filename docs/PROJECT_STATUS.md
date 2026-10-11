# Project status

_Last updated: 2026-10-11_

## Phase
Stage 1 (foundation) — identity, tenancy, RLS isolation and audit are done and tested.
**Next task: T-010 → T-012** (PostGIS, then service zones and vehicle categories), then the fare engine (T-013).

## Audit (Stage 0, 2026-10-11)
The repo `captain-KasH/driveappsaas` was empty (no commits). There was no existing code to preserve.
Local toolchain: Go 1.26.5, Node 24, pnpm, Homebrew PostgreSQL 16 (pgcrypto, citext; **no PostGIS**),
no Docker, no Redis.

## Working (verified)
| Capability | Evidence |
|---|---|
| Platform admin bootstrap, login/logout, sessions | `TestAuthSessions` |
| Tenant provisioning → owner invitation → owner account → onboarding complete | `TestOnboardingJourney` |
| Staff invitations, role permissions (dispatcher can't manage) | `TestOnboardingJourney` |
| Tenant isolation — API + RLS | `TestTenantIsolationAPI`, `TestTenantIsolationRLS` (mutation: opening the memberships policy makes it fail with "scope B sees 1 rows of tenant A") |
| Suspend / reactivate / close with reason; closure signs out staff | `TestSuspensionAndClosure` |
| Append-only audit trail | `TestAuditIsAppendOnly` |

Verification run (2026-10-11): `make test` →
```
--- PASS: TestOnboardingJourney · TestTenantIsolationAPI · TestTenantIsolationRLS · TestAuditIsAppendOnly
--- PASS: TestSuspensionAndClosure · TestAuthSessions · TestInvalidTenantInput
ok  github.com/captain-KasH/driveappsaas/services/api/internal/server
```

## Not built yet
All of Stages 2–11: fleet, zones, fares, bookings, dispatch, trips, payments, billing, comms, AI,
marketplace, white-label, clients (web/mobile), deployment. See `IMPLEMENTATION_PLAN.md`.

## MVP journey progress (spec Part 17)
Operational: 1 ✅ · 2 🟡 (prerequisites come with T-012) · 3–17 ⬜ · 18 🟡 (Stage 1 tables only).
Monetization: 1–11 ⬜.

## Outstanding external dependencies
PostGIS install (K-001), Docker (K-002), legal review (K-003), payment and SMS sandbox accounts (K-004).

## Change log
- 2026-10-11 — T-000…T-003 done; T-004 workflow written. Files: `services/api/**`, `Makefile`, `docs/**`, `CLAUDE.md`, CI, compose, `.env.example`.
