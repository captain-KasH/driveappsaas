# Implementation plan and backlog

Status: ✅ done (evidence in PROJECT_STATUS) · 🔄 in progress · ⬜ todo · ⛔ blocked.
Priority: P0 blocks the MVP journey · P1 needed for commercial MVP · P2 later.

## Stage 0–1 — Audit, specification, foundation
| ID | Task | Deps | P | Acceptance | Status |
|---|---|---|---|---|---|
| T-000 | Repo audit, docs set, CLAUDE.md, backlog | — | P0 | 16 docs + CLAUDE.md exist and reflect real state | ✅ |
| T-001 | Go API skeleton: httpx errors/request IDs/logs, health, migrations runner | T-000 | P0 | builds, vets, `/healthz` checks DB | ✅ |
| T-002 | Identity: users, bcrypt, opaque sessions, login/logout/me | T-001 | P0 | `TestAuthSessions` | ✅ |
| T-003 | Tenancy: provisioning, invitations, memberships, roles, status lifecycle, RLS, audit | T-002 | P0 | `TestOnboardingJourney`, `TestTenantIsolation*`, `TestSuspensionAndClosure`, `TestAuditIsAppendOnly` | ✅ |
| T-004 | CI: vet, gofmt, tests on Postgres service, `govulncheck` | T-001 | P1 | green run on GitHub Actions | 🔄 workflow written, first run pending |
| T-005 | Rate limiting on login + accept-invite (per IP + per email) | T-002 | P1 | test: 11th attempt in window → 429 | ⬜ |
| T-006 | MFA (TOTP) for platform roles | T-002 | P1 | platform login requires second factor | ⬜ |
| T-007 | Tenant branding config (logo, colours, languages, currency, contact, legal links) | T-003 | P1 | owner updates branding; validated; audited | ⬜ |

## Stage 2 — Core mobility
| ID | Task | Deps | P | Acceptance | Status |
|---|---|---|---|---|---|
| T-010 | Enable PostGIS (local + CI + compose) | — | P0 | `CREATE EXTENSION postgis` in migration passes `make test` | ⛔ K-001 |
| T-012 | Service zones + vehicle categories (+ onboarding prerequisites) | T-010 | P0 | owner creates zone polygon + category; onboarding completion requires ≥1 of each + a fare rule | ⬜ |
| T-013 | Fare engine: versioned rules, integer minor units, breakdown, estimate vs final | T-012 | P0 | table-driven unit tests incl. minimum fare, night surcharge, rounding; rule version stored | ⬜ |
| T-011 | OpenAPI generation from route table | T-003 | P1 | `/v1/openapi.json` served; CI diff check | ⬜ |
| T-014 | Booking state machine (formal table + transition function + events) | T-013 | P0 | every illegal transition rejected in tests; transitions audited | ⬜ |
| T-015 | Booking create with `Idempotency-Key`, tenant isolation on bookings | T-014 | P0 | scenario 1 + 3 tests | ⬜ |
| T-016 | Drivers, vehicles, documents + approval; dispatch (manual + nearest eligible, offers, expiry, locking) | T-015 | P0 | scenario 2 (parallel assignment) + 11 tests | ⬜ |
| T-017 | Trip lifecycle + final fare + cash collection record | T-016 | P0 | scenario 12 test; MVP steps 11–14 via API | ⬜ |
| T-020 | Reason-coded, time-limited, audited support access to tenant data | T-003 | P1 | grant expires; every read audited | ⬜ |

## Stage 3+ (outline — detailed when the stage starts)
| ID | Task | Stage |
|---|---|---|
| T-030 | Transactional outbox + worker + event dedupe | 3 |
| T-031 | Authenticated WebSocket updates (tenant-scoped subscriptions), driver location freshness | 3 |
| T-032 | Tenant portal + dispatcher console (Next.js) | 3 |
| T-033 | Passenger + driver apps (React Native) | 3 |
| T-040 | Email delivery (invitations, receipts) | 3 |
| T-041 | SMS/OTP provider research + adapter | 3/7 |
| T-042 | Push notifications | 3 |
| T-050 | Payment provider research (PayHere, Payable) + adapter + webhook intake + reconciliation | 4 |
| T-060 | Plans, subscriptions, invoices, module catalogue, entitlements | 5 |
| T-061 | Usage ledger, credit packs, SMS credit reservation (idempotent) | 5 |
| T-065 | White-label branding, domains, branded build pipeline | 6 |
| T-070 | Telephony adapter, inbox/chat | 7 |
| T-080 | AI orchestration layer + passenger assistant + operator copilot | 8 |
| T-085 | Shared marketplace | 9 |
| T-090 | Dedicated deployments, licensing (binary vs source vs IP) | 10 |
| T-100 | Staging/prod environments, backups + restore test, monitoring, runbooks | 11 |
