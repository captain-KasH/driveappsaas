# Project audit (master prompt §3)

_Audited 2026-10-11 against the working tree at commit `c616826` (+ this docs change). Re-audit at the start of each milestone._

## Repository facts
| Item | Finding |
|---|---|
| Backend | Go 1.26.5 modular monolith, `services/api` (4,826 LOC). Deps: `pgx/v5 5.11.0`, `x/crypto 0.58.0` only. stdlib `net/http` routing. |
| Database | PostgreSQL 16 + PostGIS 3.4. Two forward-only migrations (`0001_foundation`, `0002_mobility`), FORCE RLS on every tenant table, non-bypass app role `taxios_app`. |
| Web | `apps/web` Next.js 16.3.6 / React 19.3.0 / TypeScript 5.9.3 / Leaflet 1.9.4 (2,141 LOC). Static export embedded into the Go binary. No ESLint config; `tsc --noEmit` available but not run in CI. |
| Mobile | **None.** `apps/mobile` does not exist. |
| Realtime / jobs / cache | No WebSocket, no Redis, no job queue. One in-process dispatch worker (5 s tick). Clients poll. |
| API surface | 53 routes under `/v1` (16 identity/tenancy/platform + 37 fleet/booking). No OpenAPI document. |
| Tests | 18 test functions (2 unit tables in `fare`, 16 integration in `server` against real PostGIS). Run on the packline server via `make test-remote` / `deploy/test.sh`; `deploy.sh` runs them before every deploy. `.github/workflows/ci.yml` exists but Actions is **not used** (minutes cap). No lint/typecheck step; no E2E; no mobile tests. |
| Deployment | Docker Compose stack `/data/apps/taxios` on the packline server, behind packline's nginx (vhost `taxios.packline.lk`), Cloudflare proxied. Live and healthy. No staging. No DB backups for Taxios. No monitoring/metrics/error tracking beyond JSON logs + `/healthz`. |
| Config | Env vars: `DATABASE_URL`, `MIGRATE_DATABASE_URL`, `ADDR`, `TAXIOS_TRUST_PROXY`, `TAXIOS_ALLOW_RLS_BYPASS`, `TAXIOS_ADMIN_PASSWORD` (bootstrap only), `TAXIOS_TEST_DATABASE_URL`. Only `deploy/.env.example` exists (DB passwords); no root `.env.example`, no startup validation beyond `mustEnv`. |
| Secrets | None in git (checked). Tokens stored as SHA-256; bcrypt passwords; `deploy/.env` chmod 600 on the server. |
| Code hygiene | 5 marked placeholders (`ponytail:` comments): in-process worker (single replica), straight-line route estimate, static role map, OSM/Nominatim maps, localStorage token. No dead code found; no duplicated logic found. Build is green. |
| Git | Clean tree, 10 commits on `main`, no uncommitted user work. |

## Classification (what is real)
### Verified working (integration-tested on real PostGIS, deployed)
- Identity: email+password login, opaque sessions, logout, `/me` (`TestAuthSessions`).
- Tenancy: platform admin creates tenant → owner invitation → accept → onboarding complete (requires ≥1 zone, category, fare rule); staff roles owner/admin/dispatcher/finance/viewer; suspend/reactivate/close with reason; append-only audit (`TestOnboardingJourney`, `TestSuspensionAndClosure`, `TestAuditIsAppendOnly`).
- Isolation: API 404 for non-members; RLS on every table; mutation-checked (`TestTenantIsolationAPI/RLS`, `TestMobilityTablesRLS`).
- Fleet: zones (PostGIS polygons), vehicle categories, immutable versioned fare rules, vehicles, drivers, document records + review, approval gated on documents, `driver_compliance` view (`TestMVPJourney`, `TestExpiredDocumentsBlockDispatch`).
- Fare engine: integer minor units; base/distance/time/waiting/minimum/night/booking fee/scheduled fee/tax/rounding (9 table cases).
- Booking: quotes (zone, category, schedule window), idempotent create (`Idempotency-Key`, 5 concurrent → 1), passenger and dispatcher phone bookings, cancel rules.
- Dispatch: nearest eligible driver, 30 s exclusive offers, reject/expiry → next, 10 min search timeout, scheduled activation, manual offer (90 s), concurrency-safe (8 parallel accepts → 1).
- Trip: `driver_assigned → en_route → arrived → in_progress → completed/aborted`; final fare from GPS track with the quote's rule version; completed is final; cash record (once); receipt; rating; operator views, dispatch board, summary report.
- Web UI (deployed): landing, login, invite accept, role hub, thin platform admin, operator console (overview/setup/drivers/vehicles/bookings/dispatch/team), driver web app, passenger booking page. **Anonymous pages verified in the browser; signed-in flows not yet walked (K-016).**

### Partially implemented
- Platform admin console: create/list/status only. No verification queue, plans, subscriptions, add-ons, usage, revenue, flags, support access, announcements, metrics.
- Operator dashboard: no customers, promotions, branding, integrations, subscription, exports, settings.
- Driver documents: numbers + expiry only; **no file upload**.
- Rate limiting: per-IP at the edge only; no per-account limit.
- Reports: one summary report with definitions; no exports, no BI.
- Maps: OSM tiles + Nominatim from the browser (low-volume policy); route estimate is a straight-line heuristic.

### UI-only / mocked / hardcoded
- None found. Every button on the deployed site calls a real endpoint. (The spec's "no fabricated metrics" rule has been followed.)

### Missing
- Self-service tenant registration, email verification, password reset, MFA, OTP, refresh tokens, session management UI.
- Tenant business profile (legal name, registration, tax, billing contact, business type, country/locale/language), verification documents, draft/pending-verification states, resumable onboarding checklist, CSV import.
- Branding (logo, colours, typography, app identity), domains, booking-site templates, publishing workflow, white-label app build pipeline.
- Mobile apps (passenger, driver), push notifications, background location.
- Realtime (WebSocket), durable jobs/outbox, Redis.
- Online payments, ledger, refunds, reconciliation, driver earnings/commissions/settlements.
- Plans, subscriptions, trials, invoices, entitlements, add-on catalogue, usage metering, limits.
- Communications (email, SMS, WhatsApp, in-app, voice), templates.
- API keys, webhooks, integration settings, OpenAPI.
- Multi-stop, recurring, passenger count/luggage, flight info, corporate references, promo codes, surge, cancellation/no-show fees, tolls.
- Customers CRM, saved places, loyalty/referrals, account export/deletion.
- AI (voice receptionist, copilot, BI). Licensing/deployment registration. Marketplace.
- Backups, restore test, monitoring/metrics/error tracking, staging, security headers/CSP, dependency scanning, E2E tests.

### Blocked (external)
- Email delivery provider (needed for signup verification, invitations, receipts) — no account yet.
- SMS/OTP provider — no account. Payment providers PayHere / Payable — no merchant/sandbox accounts.
- Legal review: PDPA 2022, transport licensing, call recording, licence agreements.
- Per-tenant store apps need each tenant's Apple/Google developer accounts (Apple guideline 4.2.6; D-022).

### Security / reliability risks
| Risk | Severity | Mitigation task |
|---|---|---|
| Web session token in localStorage, no CSP | medium (no third-party scripts yet) | T-221 |
| No per-account login limit | medium | T-005 |
| Single API replica + in-process worker (no advisory lock) | low now | T-030 |
| No DB backups for Taxios | **high** before real tenants | T-220 |
| No monitoring/alerting | medium | T-223 |
| OSM/Nominatim usage policy at scale | low now | T-013b |
| Invitation/verification tokens returned in API responses until email exists | medium | T-040 |

## Document mapping (prompt §3 → repo)
| Prompt name | Repo file |
|---|---|
| PROJECT_AUDIT | `PROJECT_AUDIT.md` (this) |
| PRODUCT_REQUIREMENTS | `PRODUCT_REQUIREMENTS.md` (+ `PRODUCT_VISION.md`) |
| SYSTEM_ARCHITECTURE | `ARCHITECTURE.md` |
| DATA_MODEL | `DATABASE_DESIGN.md` |
| API_CONTRACTS | `API_CONTRACTS.md` |
| SECURITY_MODEL | `SECURITY_AND_PRIVACY.md` (+ `USER_ROLES_AND_PERMISSIONS.md`) |
| COMPETITOR_FEATURE_MATRIX | `COMPETITOR_FEATURE_MATRIX.md` |
| FEATURE_STATUS | `FEATURE_STATUS.md` (per prompt section) — `PROJECT_STATUS.md` keeps the verification log |
| IMPLEMENTATION_ROADMAP | `IMPLEMENTATION_PLAN.md` (milestones M1–M14) |
| DEPLOYMENT | `DEPLOYMENT_RUNBOOK.md` |
| BILLING_AND_ENTITLEMENTS, INTEGRATIONS, DECISIONS | same names |
| TESTING_STRATEGY | `TEST_STRATEGY.md` |
| KNOWN_LIMITATIONS | `KNOWN_ISSUES.md` |
| (master prompt) | `MASTER_PROMPT.md` |

## Baseline test run (before any Milestone-1 change)
See the "Baseline" line in `PROJECT_STATUS.md` (filled from `make test-remote` on 2026-10-11).
