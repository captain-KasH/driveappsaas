# Project status

_Last updated: 2026-10-11_

## Phase
Master prompt v2 adopted 2026-10-11 (`MASTER_PROMPT.md`, D-020…D-025). Audit: `PROJECT_AUDIT.md`;
per-section status: `FEATURE_STATUS.md`; roadmap: `IMPLEMENTATION_PLAN.md` (M1–M14).
The ride engine is verified end to end (`TestMVPJourney`) and the web app is live at https://taxios.packline.lk.
**Milestone now: M4.** Done: T-110, T-120, T-121, T-123; T-124 partial. Also done: T-122, T-130 uploads, T-113 staff removal, T-126 customers + CSV exports. **Next task: T-112 (long-lived mobile sessions + sign-out-everywhere), then T-007 branding → M7 Expo apps.**

Baseline before M1 changes (2026-10-11, `make test-remote`, PostGIS 16-3.4 + Go 1.26 on the server):
`ok internal/fare 0.014s · ok internal/server 14.741s` (18 tests, all pass).

## Working (verified)
| Capability | Evidence |
|---|---|
| Platform admin, login/logout, sessions | `TestAuthSessions` |
| Tenant provisioning → owner invite → onboarding (requires zone + category + fare rule) | `TestOnboardingJourney` |
| Staff roles and permissions | `TestOnboardingJourney`, USER_ROLES_AND_PERMISSIONS.md |
| Zones (PostGIS), categories, immutable versioned fare rules | `TestMVPJourney`, `TestQuoteAndBookingRules` |
| Fare engine (integer minor units, minimum, night window across midnight, waiting, scheduled fee, tax, rounding) | `internal/fare` table tests (9 cases) |
| Drivers, vehicles, document review, approval gated on valid documents, go-online gated on compliance | `TestMVPJourney`, `TestExpiredDocumentsBlockDispatch` |
| Passenger registration on an operator's brand, quotes (outside-area / category / schedule rules) | `TestQuoteAndBookingRules` |
| Idempotent booking (passenger and dispatcher phone bookings) | `TestBookingIdempotency` (5 concurrent retries → 1 booking) |
| Nearest-eligible dispatch, exclusive offers, reject → next, expiry → next, search timeout, scheduled activation, manual assign | `TestRejectAndExpiryRedispatch`, `TestExclusiveAssignmentConcurrency`, `TestQuoteAndBookingRules` |
| Trip state machine; completed is final | `TestCompletedTripIsFinal` |
| Final fare from GPS track with the quote's fare version (estimate fallback) | `TestMVPJourney` step 13, `TestCompletedTripIsFinal` |
| Cash record (once, note required when ≠ fare), receipt, rating | `TestMVPJourney` steps 14–15 |
| Operator booking view (events, offers, payments), dispatch board, summary report with definitions | `TestMVPJourney` steps 16–17, `TestQuoteAndBookingRules` |
| Tenant isolation on every table (API + RLS) | `TestMVPJourney` step 18, `TestMobilityTablesRLS`, `TestTenantIsolation*` |
| Self-service sign-up → email verification → draft business → business details → checklist → submit → super admin reject/approve → public booking page (D-021) | `TestSelfServiceOnboarding` (mutation-checked: removing the email-verified guard fails it), `TestDraftBusinessLimit`, `TestAdminCreatedTenantIsTrusted` |
| Password reset (single-use 1 h token, revokes every session, proves email), manual email verification by platform staff (K-019) | `TestPasswordReset`, `TestManualEmailVerification` |

Mutation checks (2026-10-11): removing `cardinality(c.issues) = 0` from dispatch eligibility →
`TestExpiredDocumentsBlockDispatch` fails ("expired-document driver was offered"). Removing the
offer-state guard in accept → the concurrency test still passes because the driver-busy check
and unique partial indexes also block it (layered by design, D-017).

Verification run (2026-10-11): `make test-remote` → on the server, PostGIS 16-3.4 + Go 1.26:
```
--- PASS: TestCompute, TestEstimateRoute
--- PASS: TestMVPJourney (1.76s) · TestBookingIdempotency · TestExclusiveAssignmentConcurrency
--- PASS: TestExpiredDocumentsBlockDispatch · TestCompletedTripIsFinal · TestRejectAndExpiryRedispatch
--- PASS: TestQuoteAndBookingRules · TestMobilityTablesRLS · TestOnboardingJourney · TestTenantIsolationAPI
--- PASS: TestTenantIsolationRLS · TestAuditIsAppendOnly · TestSuspensionAndClosure · TestAuthSessions · TestInvalidTenantInput
ok  internal/fare · ok  internal/server 13.3s
```

## Deployment (T-018)
- Cloudflare: `A taxios.packline.lk → 164.68.97.219`, proxied; SSL mode Full (verified in dashboard).
- Server: `/data/apps/taxios`, first `deploy.sh` run applied migrations 0001+0002, API container healthy.
- packline nginx vhost committed (packline `edb4fb3b`), live after packline's 10:27 deploy recreated nginx (K-011).
- Public verification 2026-10-11 against https://taxios.packline.lk (commit `3aaa149`): `/healthz` → 200 `{"status":"ok"}`;
  unknown brand → 404 JSON; `/v1/me` without token → 401; 9 rapid bad logins → `401×6, 429×3` (edge limit);
  `deploy.sh` ran the full suite on the server before deploying (ok fare, ok server 11.9s); container healthy.
  packline.lk stayed 200 throughout.

## Website (T-019, 2026-10-11)
https://taxios.packline.lk serves the Next.js static site from the API binary (same origin as `/v1`):
`/` landing · `/login/` · `/invite/?token=` · `/home/` role hub · `/admin/` platform admin · `/console/?t=` operator
console (overview report, setup with map zones + fare rules + go-live, drivers + document review, vehicles,
bookings + timeline, live dispatch board with manual offers and phone bookings, team) · `/driver/?t=` driver web
app (online/offline, GPS sharing, offers with countdown, trip steps, navigation hand-off, cash) · `/book/?b=slug`
passenger booking (sign up/in, map + address search, quote, book, live driver, cancel, receipt, rating).
Verified live: landing, sign-in, unknown operator message, all assets 200, no unexpected console errors.
`/v1/*` unknown paths still return JSON 404. Not yet verified: the signed-in flows in the browser (K-016).

## Server validation (2026-10-11, after the alias incident K-017)
- DNS: taxios.packline.lk and packline.lk resolve to Cloudflare from 8.8.8.8, 1.1.1.1 and 9.9.9.9.
- packline.lk / www / `/healthz` → 200 with packline's own page; taxios.packline.lk `/`, `/login/`, `/healthz` → 200 Taxios.
- Shared network aliases: `api` → packline only, `taxios-api` → Taxios only; 30/30 packline `/v1/flash-sales/active` → 200.
- `deploy.sh` end to end: tests ok → `Deploy succeeded: cbdc510`; containers `taxios-api`, `taxios-db-1` (renamed `taxios-db` in D-026) healthy; nginx config valid.

## Not built yet
Mobile (React Native) apps; notifications/realtime; online payments; billing, add-ons, usage metering;
communications; AI; shared marketplace; white-label builds; enterprise licensing; backups for the
Taxios DB. See IMPLEMENTATION_PLAN.md.

## MVP journey progress (spec Part 17)
Operational: 1–18 ✅ via API (UI pending, T-019). Monetization: 1–11 ⬜ (Stage 5).

## Outstanding external dependencies
Legal review (K-003); payment and SMS sandbox accounts (K-004, K-009); routing/maps provider choice (K-010).

## M4 progress (2026-10-11)
Backend commit `12bdb24`, web commit `f92957e`: migration 0003, `/signup/ /verify/ /forgot/ /reset/ /start/` pages, console Setup
tab with business details + computed checklist + "Submit for review" / "Go live", `/admin/` verification queue with
search/filter, business detail, manual email verification. Suite: 23 tests green on the server before deploy.
Live verification 2026-10-11 after `Deploy succeeded: f92957e` (suite green on the server first): `/signup/ /verify/ /forgot/ /reset/ /start/ /admin/` → 200 and `/signup/` renders without console errors; `register` with a bad email → 422 `invalid_email`; `verify-email` bad token → 404 `invalid_token`; `forgot-password` unknown email → 200 `{sent}`; `POST /v1/businesses` unauthenticated → 401; packline.lk → 200. Signed-in walk-through still pending (K-016, needs the owner's super-admin login).

## Change log
- 2026-10-11 — T-122/T-130 files `aa15eb4` → `Deploy succeeded: aa15eb4`, volume `taxios_files` present, API healthy; anonymous hits on the new endpoints → 401/404. T-113 + T-126 customers/exports (next deploy).
- 2026-10-11 — D-026 docker naming: `Deploy succeeded: 5a15cc3`; server shows only `taxios-api`, `taxios-db` (+ `taxios_internal`, `taxios_db`, `taxios-api:latest`, `taxios-test_*`) beside `voiceai-*` and `packlink-*`.
- 2026-10-11 — M1 (audit, competitor matrix, web typecheck/build in the test stack) `4ae9940`, `f87592a`; M4 onboarding API `12bdb24` + web `f92957e`.
- 2026-10-11 — Phase 1 core mobility (T-010…T-017), server test runner, Docker deploy, DNS, nginx vhost. Commit `5c57015`+.
- 2026-10-11 — T-000…T-003 foundation. Commit `b46f605`.
