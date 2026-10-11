# Project status

_Last updated: 2026-10-11_

## Phase
**Phase 1 API complete.** The spec Part 17 operational journey (steps 1–18) runs end to end
through the API and is covered by `TestMVPJourney`. The API is deployed at
https://taxios.packline.lk (Docker stack on the packline server; see DEPLOYMENT_RUNBOOK.md).
**Next task: T-019** — operator/dispatcher web console + passenger web booking (Next.js), so the
journey is executable through a UI; then React Native passenger/driver apps.

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

## Not built yet
Mobile (React Native) apps; notifications/realtime; online payments; billing, add-ons, usage metering;
communications; AI; shared marketplace; white-label builds; enterprise licensing; backups for the
Taxios DB. See IMPLEMENTATION_PLAN.md.

## MVP journey progress (spec Part 17)
Operational: 1–18 ✅ via API (UI pending, T-019). Monetization: 1–11 ⬜ (Stage 5).

## Outstanding external dependencies
Legal review (K-003); payment and SMS sandbox accounts (K-004, K-009); routing/maps provider choice (K-010).

## Change log
- 2026-10-11 — Phase 1 core mobility (T-010…T-017), server test runner, Docker deploy, DNS, nginx vhost. Commit `5c57015`+.
- 2026-10-11 — T-000…T-003 foundation. Commit `b46f605`.
