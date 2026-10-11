# Feature status (per master prompt section)

_Updated 2026-10-11. ✅ verified (test/evidence named) · 🟡 partial · ⬜ missing · ⛔ blocked externally.
Evidence conventions: test names live in `services/api/internal/server/*_test.go`; "live" = checked on https://taxios.packline.lk._

| § | Capability | Status | Evidence / gap | Next task |
|---|---|---|---|---|
| 6 | Tenant model: id, slug, names, status, timezone, currency, audit | 🟡 | + legal name, registration/tax, contacts, billing, address, language, verification (0003). Missing branding, plan, domains, provider settings | T-007, M10 |
| 6 | Tenant context resolution + RLS isolation (API, DB) | ✅ | `TestTenantIsolationAPI`, `TestTenantIsolationRLS`, `TestMobilityTablesRLS` | — |
| 6 | Isolation for jobs / WebSocket / webhooks / uploads / AI / billing | 🟡 | worker uses scoped `InTx`; others don't exist yet | T-030, T-031, T-130 |
| 6 | Audited platform-support access | ⬜ | | T-020 |
| 7 | Email+password, sessions, logout, revocation | ✅ | `TestAuthSessions` | — |
| 7 | Password reset, email verification | ✅ | `TestPasswordReset`, `TestSelfServiceOnboarding`; delivery needs SMTP creds (K-019), manual verification by staff meanwhile | T-040 |
| 7 | Refresh / long-lived mobile sessions | ⬜ | | T-112 |
| 7 | MFA | ⬜ | | T-006 |
| 7 | Rate limiting / brute force | 🟡 | per-IP at edge only | T-005 |
| 7 | Phone OTP | ⬜ | | T-041 ⛔ |
| 7 | Roles: owner/admin/dispatcher/finance/viewer/driver/passenger; platform super_admin/support | ✅ | `permissions.go`, `TestOnboardingJourney` | fleet manager role → T-122 |
| 7 | Invitations, staff onboarding/offboarding | ✅ | invite ✅; remove ✅ `TestStaffRemoval` | — |
| 8 | Self-service registration → business → verification → plan → checklist → publish | 🟡 | register/verify/business/documents/checklist/submit/approve ✅ (`TestSelfServiceOnboarding`, `TestBusinessDocumentsRequiredForSelfService`); plan step ⬜ (M10) | M10 |
| 8 | CSV import | ⬜ | | T-127 |
| 9 | Platform console | 🟡 | live `/admin/`: verification queue, search/filter, business detail, create, status, manual email verification. No plans/usage/flags/audit/announcements yet | T-124, T-125 |
| 10 | Operator dashboard | 🟡 | live `/console/`: overview, setup (details, documents, checklist), drivers, vehicles, bookings (+CSV), dispatch, customers (+CSV), team (invite/remove). Missing promotions, branding, integrations, subscription, support | T-007, M10 |
| 11 | Drivers: profile, approval, docs expiry, vehicle assignment, availability, suspension | ✅ | `TestMVPJourney`, `TestExpiredDocumentsBlockDispatch` | — |
| 11 | Driver document **files** ✅ (`TestFileUploadsAndAccess`); expiry alerts, earnings/commission, shifts, performance ⬜ | 🟡 | | T-130, T-132, T-134 |
| 11 | Vehicles: details, category, status, driver assignment | ✅ | `TestMVPJourney` | — |
| 11 | Vehicle docs, inspection/service dates, maintenance state/history, capacity/accessibility | ⬜ | | T-131 |
| 12 | Passenger profile, trip history, receipts, ratings | ✅ | `TestMVPJourney` steps 14–15 | — |
| 12 | Saved places, preferences, consent, loyalty/referrals, export/deletion | ⬜ | | T-133 |
| 13 | Immediate + scheduled bookings, categories, notes, estimates, confirmation, cancellation, receipts | ✅ | `TestQuoteAndBookingRules`, `TestMVPJourney` | — |
| 13 | Recurring, multi-stop, passenger count/luggage, accessibility, flight info, corporate ref, no-show rules, disputed state | ⬜ | | T-140, T-141, T-145 |
| 13 | Explicit state machine, duplicate/retry/race handling | ✅ | `TestCompletedTripIsFinal`, `TestBookingIdempotency`, `TestExclusiveAssignmentConcurrency` | — |
| 14 | Dispatch: queue, live map, manual + auto assign, scheduled, accept/reject, timeout/retry, cancel | ✅ | tests above; live `/console/` Dispatch tab | — |
| 14 | Realtime (no polling), alerts for delayed/unassigned, dispatcher notes, filters, search, reassignment | ⬜ | 5 s polling | T-031, T-146 |
| 15 | Base/distance/time/minimum/waiting/night/booking fee/scheduled/tax/rounding, versioned rules, estimate vs final, breakdown | ✅ | `fare_test.go`, D-016 | — |
| 15 | Cancellation fees, tolls, route/airport/intercity/corporate pricing, promo codes, commissions, surge | ⬜ | | T-141…T-144 |
| 16 | Passenger mobile app | ⬜ | web booking page exists (`/book/`) | T-150, T-151 |
| 17 | Driver mobile app | ⬜ | web driver app exists (`/driver/`) | T-150, T-152 |
| 18 | Branding, app identity, domains, booking widget, templates, publishing | ⬜ | | T-007, T-170…T-174 |
| 18 | Shared marketplace | ⬜ | | T-085 |
| 19 | Operator public website | 🟡 | booking page only, unbranded | T-171, T-172 |
| 20 | Ride payments: cash record | ✅ | `TestMVPJourney` step 14 (D-018) | — |
| 20 | Online payments, ledger, refunds, reconciliation, driver earnings | ⬜ / ⛔ | no provider account | T-160…T-164 |
| 21 | Plans, subscriptions, trials, invoices, grace/overdue policy | ⬜ | | T-180…T-183 |
| 22 | Add-on catalogue, entitlement service, usage metering, limits | ⬜ | | T-182, T-184, T-061 |
| 23 | AI voice receptionist | ⬜ | | T-200, T-201 |
| 24 | AI copilot + BI | ⬜ | summary report only | T-202, T-203 |
| 25 | Innovations register | ⬜ | | T-102 (matrix) |
| 26 | Notifications abstraction (push/email/SMS/in-app) | ⬜ / ⛔ | | T-040…T-042, T-190 |
| 27 | API keys, webhooks, OpenAPI | ⬜ | | T-011, T-191, T-192 |
| 28 | i18n: currency/timezone per tenant | 🟡 | tenant currency + timezone exist; UI English only; LK phone normalisation hardcoded | T-120, T-128 |
| 29 | Licensing / deployment registration | ⬜ | | T-210, T-211 |
| 30 | Security controls | 🟡 | see SECURITY_AND_PRIVACY.md table; missing CSP/headers, cookie sessions, uploads, scanning, backups | T-221, T-220 |
| 31 | Observability | 🟡 | JSON logs + request IDs + `/healthz`; no metrics/error tracking/alerts | T-223 |
| 32 | Tests | 🟡 | unit + integration + concurrency ✅; no E2E/security-suite/mobile/load | T-222, T-224 |
| 33 | CI/CD | 🟡 | server-run suite before deploy; no lint/typecheck/scanning/staging/backups | T-103, T-104, T-220 |
