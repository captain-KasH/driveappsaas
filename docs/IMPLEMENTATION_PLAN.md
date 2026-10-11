# Implementation roadmap (master prompt §34 milestones)

Status: ✅ done (evidence in PROJECT_STATUS/FEATURE_STATUS) · 🟡 partial · ⬜ todo · ⛔ blocked externally.
Priority: P0 = needed for the first commercial launch · P1 = v1.1 · P2 = later.
Old task IDs (T-000…T-100) are kept where the work exists; new work gets IDs per milestone (T-1xx = M1, …).

## Build order (D-025)
M1 → M4 → M7 → M9 → M10 → M8 → M11 → M14 (launch) → M12 → M13, with M5/M6 gaps picked up when a
milestone needs them. Reason: the owner's first goal is the SaaS owner → tenant admin → white-label app
chain end to end; billing before online payments because v1 invoicing is manual (D-023).

## M1 — Repository stabilization
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-000 | Initial audit, docs, CLAUDE.md | P0 | | ✅ |
| T-101 | Re-audit against master prompt v2: PROJECT_AUDIT, FEATURE_STATUS, roadmap, decisions | P0 | docs reflect the real tree; baseline suite green | ✅ 2026-10-11 |
| T-102 | Competitor feature matrix (Onde, Yelowsoft, TaxiCaller, iCabbi, Autocab, Jugnoo, Limo Anywhere, Gazoop, Cabica) with 40 dated sources; MVP vs long-term; pricing; Apple 4.2.6 | P0 | `COMPETITOR_FEATURE_MATRIX.md` | ✅ 2026-10-11 (vendor-advertised unless marked 3P/Official) |
| T-103 | Root `.env.example` lists every variable the binary reads; `mustEnv` exits 2 with `error: X is required` | P0 | `serve` without `DATABASE_URL` exits with a named error | ✅ (`main.go` `mustEnv`; `.env.example` rewritten) |
| T-104 | `deploy/test.sh` runs `go vet` + `gofmt -l` (already) and now `npm ci && tsc --noEmit && next build` in a `web` service (`SKIP_WEB=1` for API-only loops) | P0 | a type error fails `make test-remote` | ✅ `make test-remote` 2026-10-11: Go ok, `✓ Compiled successfully`, `Finished TypeScript`, 10 static pages |
| T-105 | Remove unused `.github/workflows/ci.yml` or make it opt-in | P2 | | ⬜ |

## M2 — Backend and database foundation
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-001 | httpx, health, migrations runner, logging | P0 | | ✅ |
| T-030 | Transactional outbox + durable jobs (retries, dedupe, advisory-locked worker) | P0 | retried job has one effect; worker safe with 2 replicas | ⬜ |
| T-011 | OpenAPI document generated from the route table, served at `/v1/openapi.json` | P1 | | ⬜ |
| T-106 | Cursor pagination on list endpoints (tenants, bookings, drivers, customers) | P1 | | ⬜ |

## M3 — Authentication and tenant isolation
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-002 / T-003 | Identity, tenancy, RLS, roles, invitations, audit | P0 | | ✅ |
| T-110 | Email verification + password reset (hashed single-use tokens; SMTP from env or logged; `TAXIOS_DEV_TOKENS=1` returns tokens in tests) | P0 | tests: unverified can't create a business; reset token single-use | ✅ `TestSelfServiceOnboarding`, `TestPasswordReset` (2026-10-11) |
| T-112 | Long-lived mobile sessions + refresh/rotation + "sign out everywhere" | P0 | | ⬜ |
| T-113 | `DELETE /members/{userID}`: staff only, owners only by owners, never the last owner or yourself; sessions revoked when they have no other business; audited | P0 | | ✅ `TestStaffRemoval` |
| T-005 | Per-account login limiting in-app | P1 | 11th attempt → 429 | 🟡 |
| T-006 | MFA (TOTP) for platform roles | P1 | | ⬜ |
| T-020 | Reason-coded, time-limited, audited support access | P1 | | ⬜ |

## M4 — Tenant onboarding and platform administration  ← **current**
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-120 | Tenant business profile: legal name, registration/tax no., contacts, billing, address, timezone/currency/language; status `onboarding → pending_verification → active → suspended/closed` (D-021); migration 0003 + PATCH + tests | P0 | state machine tested; existing tenants marked verified | ✅ `TestSelfServiceOnboarding`, `TestAdminCreatedTenantIsTrusted` |
| T-121 | Self-service: `POST /v1/auth/register`, `POST /v1/businesses` (verified email, max 3 drafts); web `/signup/`, `/verify/`, `/forgot/`, `/reset/`, `/start/` | P0 | test: register → verify → create business → appears in verification queue | ✅ mutation-checked (email-verified guard removed → test fails) |
| T-122 | Verification documents: `files` module (local volume `taxios_files`, tenant-prefixed keys, sniffed type, 10 MB cap), `tenant_documents`, checklist step `documents` required for self-registered tenants, platform detail + audited platform file reads | P0 | reviewer sees the uploaded certificate | ✅ `TestBusinessDocumentsRequiredForSelfService`, `TestFileUploadsAndAccess` |
| T-123 | Resumable onboarding checklist API (computed from rows) + console Setup tab (business details card, checklist, submit/go-live) + platform approve/reject | P0 | checklist reflects DB state; no client-side-only progress | ✅ `GET /onboarding`; UI deployed |
| T-124 | Platform console v2: tenant directory + search/filter ✅, verification queue ✅, tenant detail (profile ✅; plan/usage/audit ⬜), global audit log ⬜, announcements ⬜, manual email verification ✅ | P0 | live screens call real endpoints | 🟡 |
| T-125 | Platform feature flags (global + per-tenant overrides) read through the entitlement hook | P1 | | ⬜ |
| T-126 | Operator dashboard additions: customers tab (search, aggregates) ✅, CSV exports of customers + bookings (`reports.read`) ✅, settings/profile ✅ (T-123); branding tab (T-007) ⬜, subscription tab (M10) ⬜ | P0 | | 🟡 `TestCustomersAndExports` |
| T-127 | CSV import for drivers and vehicles with validation report + duplicate handling | P1 | | ⬜ |
| T-128 | i18n foundation: UI strings externalised; en/si/ta switch; per-tenant default language; phone validation per country | P1 | | ⬜ |

## M5 — Fleet and customer operations
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-016 | Drivers, vehicles, documents, approval | P0 | | ✅ |
| T-130 | Driver document scans attached via `fileId` (uploader/driver/fleet staff only; cross-tenant and other-driver reads → 404); expiry alerts job ⬜ | P0 | cross-tenant read → 404 | 🟡 uploads ✅ `TestFileUploadsAndAccess`; alerts ⬜ |
| T-131 | Vehicle lifecycle: insurance/registration docs, inspection + service dates, maintenance state, capacity, accessibility | P1 | | ⬜ |
| T-132 | Driver earnings + commission records per trip (configurable % / flat), driver statement | P0 | ledger lines append-only | ⬜ |
| T-133 | Passenger saved places, preferences, consent, data export + deletion request | P1 | | ⬜ |
| T-134 | Shifts / working hours | P2 | | ⬜ |

## M6 — Booking, fares, dispatch
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-012…T-017 | Zones, categories, fares, quotes, bookings, dispatch, trips, cash, receipts, reports | P0 | | ✅ |
| T-031 | Authenticated WebSocket (tenant-scoped) for dispatch board, driver offers, passenger tracking | P0 | | ⬜ |
| T-013b | Routing/ETA provider adapter (choice recorded in INTEGRATIONS) with heuristic fallback | P1 | | ⬜ |
| T-140 | Booking extras: passenger count, luggage, multi-stop, flight number, corporate reference, accessibility | P1 | | ⬜ |
| T-141 | Cancellation / no-show fees as fare-rule fields | P1 | | ⬜ |
| T-142 | Promo codes + discounts (validated server-side, usage limits) | P1 | | ⬜ |
| T-143 | Surge / demand pricing (opt-in per tenant, capped, audited) | P2 | | ⬜ |
| T-144 | Commission configuration (feeds T-132) | P0 | | ⬜ |
| T-145 | Recurring bookings | P2 | | ⬜ |
| T-146 | Dispatch alerts (unassigned > N min, late pickup), dispatcher notes, filters, search, reassignment | P1 | | ⬜ |

## M7 — Passenger and driver applications (Expo, D-024)
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-150 | `apps/mobile` Expo monorepo: `passenger`, `driver`, shared `packages/api` + `packages/brand`; runtime branding loaded from `/v1/public/tenants/{slug}/brand`; join by link/code | P0 | both apps build with EAS dev profile | ⬜ |
| T-151 | Passenger app: auth, OTP (⛔ provider), places, categories, quote, book/schedule, tracking, cancel, history, receipts, rating, settings, deletion | P0 | journey steps on device | ⬜ |
| T-152 | Driver app: auth, onboarding status, documents, online/offline, offers, trip steps, navigation hand-off, cash, earnings, background location (permitted, battery-aware) | P0 | | ⬜ |
| T-153 | Push notifications (Expo push → FCM/APNs) through the notification abstraction (T-190) | P0 | | ⬜ |
| T-154 | Mobile tests: permissions, deep links, session expiry, backgrounding, GPS failure, offline | P1 | | ⬜ |

## M8 — Payment operations
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-160 | Provider research (PayHere, Payable/Payments.lk): hosted checkout, webhook signing, refunds, recurring, settlement — documented with sources | P0 | INTEGRATIONS.md | ⬜ |
| T-161 | Payment abstraction: intents, hosted checkout, server-side status verification, signed webhook intake, idempotent events, states; sandbox run ⛔ until credentials | P0 | scenarios 4, 5, 14 | ⬜ |
| T-162 | Financial ledger (estimate, final fare, collection, refund, discount, fee, driver earning, adjustment) with invariants | P0 | | ⬜ |
| T-163 | Refund workflow (where provider supports) | P1 | | ⬜ |
| T-164 | Reconciliation + operator financial reports/exports | P1 | | ⬜ |

## M9 — White-label websites and branded experiences
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-007 | Branding: logo upload, colours, typography choice, contacts, legal links, languages; validated; audited; public `brand` endpoint | P0 | booking page + apps render it | ⬜ |
| T-170 | Branded booking website templates (2–3) + sections, SEO metadata, terms/privacy pages, consented analytics | P0 | | ⬜ |
| T-171 | Subdomain `{slug}.taxios.lk`-style routing + custom domain with ownership verification (TXT) + TLS automation | P1 | tenant can't claim another's domain | ⬜ |
| T-172 | Embeddable booking widget | P2 | | ⬜ |
| T-173 | App identity config per tenant (name, bundle/package id, icon, splash, store links) + "request branded build" workflow; platform-side approval | P0 | | ⬜ |
| T-174 | Per-tenant EAS build + submit pipeline using the tenant's own store credentials (encrypted at rest, never displayed) — D-022 | P0 | one tenant app built and submitted to TestFlight/internal track | ⬜ |
| T-085 | Shared marketplace foundations (operator listing, availability, routing to one operator) | P2 | | ⬜ |

## M10 — SaaS plans, subscriptions, add-ons
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-180 | Plan catalogue as data (free/trial/starter/growth/professional/enterprise; monthly/annual; limits; included add-ons) managed in the platform console | P0 | | ⬜ |
| T-181 | Subscriptions: trial → active → past_due → suspended → cancelled; upgrades/downgrades; history | P0 | state machine tested | ⬜ |
| T-182 | Entitlement service: server-side checks for features + limits (drivers, vehicles, staff, bookings/month); documented behaviour on expiry | P0 | scenarios 6, 8 | ⬜ |
| T-183 | Invoices + manual payment recording (bank transfer / link), grace period, overdue policy (D-023) | P0 | | ⬜ |
| T-184 | Add-on catalogue + purchase/activation sequence (BILLING_AND_ENTITLEMENTS) | P1 | | ⬜ |
| T-061 | Usage ledger (idempotent), quotas, overage, budget alerts | P1 | scenario 7 | ⬜ |
| T-185 | Tenant billing pages (plan, invoices, usage, add-ons) | P0 | | ⬜ |
| T-186 | Platform revenue + usage reporting | P1 | | ⬜ |

## M11 — Notifications and integrations
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-190 | Notification abstraction: channels, templates (branded, per language), outbox-driven, deduped, delivery status, metering | P0 | retried event → one message | ⬜ |
| T-040 | Email provider adapter (⛔ account) | P0 | | ⬜ |
| T-041 | SMS/OTP provider adapter (⛔ account) | P0 | | ⬜ |
| T-191 | API keys with scopes, rotation, rate limits | P1 | | ⬜ |
| T-192 | Webhooks: registration, signed delivery, retries, dead-letter, logs, replay | P1 | | ⬜ |
| T-193 | Tenant integration settings + provider health | P1 | | ⬜ |

## M12 — AI capabilities
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-200 | Voice provider research (LK/MY availability, numbers, languages, cost, recording law) | P2 | | ⬜ |
| T-201 | Tool-based booking agent (narrow tools, confirmation, fallback, metered) | P2 | scenarios 9, 10 | ⬜ |
| T-202 | Dispatcher copilot | P2 | | ⬜ |
| T-203 | BI layer | P2 | | ⬜ |

## M13 — Enterprise and self-hosted licensing
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-210 | Licence records, deployment registration, editions, expiry; no remote kill switch | P2 | scenario 13 | ⬜ |
| T-211 | Self-hosted distribution + docs | P2 | | ⬜ |

## M14 — Production hardening and launch readiness
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-220 | Nightly DB backups + restore test | **P0** | restore drill documented with result | ⬜ |
| T-221 | httpOnly cookie sessions, CSP, security headers, CORS policy, dependency/container scanning | P0 | | ⬜ |
| T-222 | Playwright E2E for the 12-step journey (§32) | P0 | | ⬜ |
| T-223 | Metrics, error tracking, alerts (payments, webhooks, jobs, quotas, backups) | P0 | | ⬜ |
| T-224 | Load test (booking + location traffic) | P1 | | ⬜ |
| T-225 | Launch docs: minimum infra, scaling triggers, incident response, DR | P0 | | ⬜ |

## Next actionable task
**T-103 + T-104** (small, unblock everything), then **T-120 → T-121 → T-122 → T-123 → T-124** (self-service onboarding + platform console), each with tests, committed and deployed separately.
