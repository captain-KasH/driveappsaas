# Implementation roadmap (master prompt §34 milestones)

Status: ✅ done (evidence in PROJECT_STATUS/FEATURE_STATUS) · 🟡 partial · ⬜ todo · ⛔ blocked externally.
Priority: P0 = needed for the first commercial launch · P1 = v1.1 · P2 = later.
Old task IDs (T-000…T-100) are kept where the work exists; new work gets IDs per milestone (T-1xx = M1, …).

## Build order (D-025, re-sequenced 2026-10-12 for Uber parity — D-027)
M1 → M4 → **M4b settings layer (T-129)** → **M6 realtime + geo (T-030, T-031, T-013b)** → **M11 push/SMS (T-190, T-153, T-041)** →
**M7 app rewrite (T-155…T-159)** → **M5 ledger (T-132/T-144)** → M9 → M10 → M8 → M14 (launch) → M12 → M13. Reason: the owner's
first goal is the SaaS owner → tenant admin → white-label app chain end to end, and the apps must match Uber's flows
(`UBER_APPS_SPEC.md` §4) which need settings, realtime, maps and push underneath; billing before online payments because v1
invoicing is manual (D-023). Parity gap analysis: `PLATFORM_PARITY.md` §C5 (platform/operator), `UBER_APPS_SPEC.md` §4 (apps),
`SRI_LANKA_MARKET.md` §8 (local decisions D-LK-1…14).

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
| T-112 | `client: mobile` login → 180-day session; sliding expiry on use (web 30 d); `GET/DELETE /v1/auth/sessions`, `POST /v1/auth/logout-all`; home page "Where you're signed in" | P0 | | ✅ `TestSessionManagement` (no refresh tokens: opaque sessions + sliding expiry cover the need, D-002) |
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
| T-128 | i18n foundation: UI strings externalised; en/si/ta switch (Noto Sans Sinhala/Tamil bundled); per-tenant default language (by district); phone validation per country; `Rs 1,250` formatting | **P0** | both apps + SMS templates in si/ta | ⬜ |
| T-129 | **Tenant settings layer** (D-028): `tenant_settings` + validation + audit + one `settings.For(tenant)` hook; replace `booking.go` constants (offer TTLs, radius, search timeout, lead time, quote TTL, location staleness) + new knobs (dispatch policy, PIN policy + night window, cancel free window/fee, no-show wait/fee, wait grace, upfront deviation %, commission, languages, contact policy, SOS numbers, nav default); public subset on `/v1/public/tenants/{slug}`; console Settings page (`PLATFORM_PARITY.md` §C3) | **P0** | test: two tenants with different offer TTLs dispatch differently; no business constant left in Go | ⬜ |
| T-020 | Audited support access ("login as"): reason code, TTL ≤ 60 min, dual attribution, persistent banner, every access in `support_sessions` + audit | P0 | | ⬜ |
| T-188 | Announcements + platform/tenant support inbox (tickets, SLA timers) | P1 | | ⬜ |

## M5 — Fleet and customer operations
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-016 | Drivers, vehicles, documents, approval | P0 | | ✅ |
| T-130 | Driver document scans attached via `fileId` (uploader/driver/fleet staff only; cross-tenant and other-driver reads → 404); expiry alerts job ⬜ | P0 | cross-tenant read → 404 | 🟡 uploads ✅ `TestFileUploadsAndAccess`; alerts ⬜ |
| T-131 | Vehicle lifecycle: insurance/registration docs, inspection + service dates, maintenance state, capacity, accessibility | P1 | | ⬜ |
| T-132 | **Commission-credit ledger** (D-029): `commission_rules` (% / flat daily / tiered; category/zone/driver scope), append-only `ledger_entries`, balance = SUM, negative floor stops offers, cash-in (staff, bank import by reference), daily/weekly statements, driver-app earnings tab + "owed to operator" | P0 | ledger lines append-only; mutation test on the floor | ⬜ |
| T-135 | Driver standing engine: rolling-500 rating, rolling-100 acceptance/cancellation (nightly), floors → dispatcher alerts, warn → suspend(until) → deactivate with reason + appeal; `driver_standing_events` feeds `driver_compliance`; doc-expiry reminders 30/7/1 d (with T-130) | P0 | never auto-deactivates | ⬜ |
| T-136 | Rider ratings by drivers + mutual-1★ exclusion in dispatch (knob) | P0 | | ⬜ |
| T-137 | Driver register export for NTC/WP audits (NIC, licence class, CR, revenue licence, insurance, WP permit) + vehicle docs (merges into T-131) | P1 | | ⬜ |
| T-133 | Passenger saved places, preferences, consent, data export + deletion request | P1 | | ⬜ |
| T-134 | Shifts / working hours | P2 | | ⬜ |

## M6 — Booking, fares, dispatch
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-012…T-017 | Zones, categories, fares, quotes, bookings, dispatch, trips, cash, receipts, reports | P0 | | ✅ |
| T-030 | Transactional outbox (`outbox_events`, same tx) + in-binary worker, stable IDs, tenant context, consumer dedupe | P0 | retried event delivered once | ⬜ |
| T-031 | Realtime hub: authenticated WebSocket per tenant, per-booking seq numbers, resend from last-seen seq, 4 s heartbeat / 7 s reconnect, latest-state-only; Redis fan-out when multi-instance; polling stays as fallback; offer **delivered-ack** so unseen offers aren't declines | P0 | reconnect test replays missed events | ⬜ |
| T-013b | **Geo adapter** (D-031): tenant POI table with si/ta/en aliases + trigram search; self-hosted Nominatim/Photon + OSRM/Valhalla on the LK extract (search, route, matrix, map-match for fare distance); Google Places/Routes behind a per-tenant paid flag; learned per-zone/hour speed table | **P0** | quote distance = routed; ETA from matrix | ⬜ |
| T-148 | Quote **all categories in one call** with pickup ETA each; upfront fare lock + deviation rule (knob) + re-price on destination/stop change with both parties shown the delta | P0 | `TestUpfrontFare` | ⬜ |
| T-149 | Arrived/wait/PIN/no-show server rules: geofence auto-`arrived` (knob) with manual fallback, wait timer from server `arrived` time, PIN (off/night/always) checked before `in_progress`, `rider_no_show` only after `arrived` + threshold, fees per T-141 | P0 | state tests | ⬜ |
| T-140 | Booking extras: passenger count, luggage, multi-stop, flight number, corporate reference, accessibility | P1 | | ⬜ |
| T-141 | Cancellation / no-show / waiting fees as fare-rule fields (free window, flat fee, waiver when driver ETA slips; cash → owed balance on next trip; wait grace + per-minute) + night band 22:00–05:00 + peak multiplier + holiday calendar + airport/zone flat (meter-shaped rules, D-LK-2) | **P0** (night/flat/wait), P1 (fees) | | ⬜ |
| T-142 | Promo codes + discounts (validated server-side, usage limits) | P1 | | ⬜ |
| T-143 | Surge / demand pricing (opt-in per tenant, capped, audited) | P2 | | ⬜ |
| T-144 | Commission configuration (feeds T-132) | P0 | | ⬜ |
| T-145 | Recurring bookings | P2 | | ⬜ |
| T-146 | Dispatch board (Fleet-Hub grade): live map of drivers/trips over T-031, alerts (unassigned > N min, late, stationary, route deviation), reassign with reason, bulk actions, dispatcher notes, filters/search; phone booking with SMS tracking link on the tenant's domain | **P0** | | ⬜ |
| T-147 | Zone kinds (service / stand / airport / surcharge), stand-FIFO queue dispatch (tuk societies, BIA), GeoJSON import, scheduled dispatch profiles (day/night/event) | P1 | | ⬜ |
| T-150b | Hire products: hourly/day packages (hours + km caps, extra-km), intercity one-way/return per-km, corporate accounts with monthly statement | P1 | | ⬜ |

## M7 — Passenger and driver applications (Expo, D-024)
| ID | Task | P | Acceptance | Status |
|---|---|---|---|---|
| T-150 | `apps/mobile`: ONE Expo SDK 57 codebase; `APP_VARIANT=passenger\|driver` picks the app, optional `TAXIOS_TENANT` pins a per-operator build (D-022); runtime brand from `/v1/public/tenants/{slug}`; session in the keychain | P0 | both variants run in Expo Go on the iOS simulator against prod | 🟡 runs in Expo Go (2026-10-11); EAS build profiles ⬜ (T-174) |
| T-151 | Passenger app: operator code (or pinned), join/sign in (mobile session), address search + current location, vehicle, now/+30m/+1h/+2h/tomorrow, quote, idempotent booking, live status, driver card, cancel, history, receipt, rating, support/legal links. Missing: map + live driver pin, OTP ⛔, push, account deletion, date picker | P0 | journey steps on device | ✅ journey run on iOS simulators 2026-10-11 (PROJECT_STATUS "device run"); Android pending |
| T-152 | Driver app: sign in, business picker, compliance status, online/offline, foreground GPS (10 s), offer countdown + accept/decline, trip steps, Google Maps hand-off, end-early reasons, cash confirmation, documents with camera/photo upload. Missing: background location (T-152b, needs dev build), earnings (T-132), push | P0 | | ✅ same run (offer → trip → GPS final fare → cash); background GPS ⬜ T-152b |
| T-153 | Push notifications (Expo push → FCM/APNs) through T-190: offers (high priority + sound), assigned/arriving/arrived/started/receipt, doc expiry, settlement; Live Activity / ongoing notification v1.1 | P0 | | ⬜ |
| T-154 | Mobile tests: permissions, deep links, session expiry, backgrounding, GPS failure, offline replay; device matrix Samsung A05 / Honor X7b / Redmi 9A (D-LK-10) | P1 | | ⬜ |
| T-155 | **Passenger app rewrite, map-first** (`UBER_APPS_SPEC.md` §1, §4.1): map + bottom sheet shell (snap points), "Where to?" home with saved/recent places + nearby vehicles, two-field search with POIs + set-on-map pin, product rows with upfront price + ETA (T-148), payment row (cash explicit), searching sheet (free cancel, call operator, no-driver), en-route with live driver marker + route + ETA + call/share/safety, arrived banner + PIN + wait timer, on-trip, end screen ("pay driver Rs X cash", rating chips, SMS receipt), Activity, Reserve date picker, account/settings (si/ta/en), safety toolkit (119/1990 + dispatcher alert), public share page | **P0** | journey on both simulators via `make dev-tunnel` + screenshots in PROJECT_STATUS | ⬜ |
| T-156 | **Driver app rewrite, map-first** (§2, §4.2): map home + earnings pill + status strip, documents screen with expiring state, earnings-first offer card (fare, pickup ETA, trip length, destination, rider rating, countdown ring, sound/vibration, delivered-ack), in-app map + nav chooser, wait timer + PIN entry + no-show cancel, swipe to start/complete, cash screen with commission line, rate rider, earnings tab (T-132), SOS, account | **P0** | same | ⬜ |
| T-157 | Phone + OTP sign-up/sign-in for passengers (D-030) once T-041 lands; name-only profile; terms consent | P0 | ⛔ SMS creds | ⬜ |
| T-158 | EAS **dev build** (not Expo Go): background location (replaces T-152b), batching + replay, mock-location flag, Live Activity, push tokens; minSdk 28, APK ≤ 30 MB, cold start ≤ 3 s | P0 | | ⬜ |
| T-159 | Offline tolerance: cached recents/saved places + last booking state; idempotent driver actions queued and replayed in order | P0 | airplane-mode test | ⬜ |

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
| T-007 | Branding: logo (tenant's own image file), colour (+ computed text colour), tagline, support phone/email, website/terms/privacy links; validated; audited; `brand` on the public tenant endpoint + public logo route; booking page themed; console Brand card | P0 | booking page + apps render it | ✅ `TestBranding` (typography/languages → T-128) |
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
| T-041 | SMS/OTP provider adapter: local gateway (Notify.lk / Text.lk / Dialog eSMS) with per-tenant registered sender ID + platform fallback, si/ta ≤ 70-char segments, promo blackout 20:00–08:00, Twilio cross-border fallback (⛔ account) | P0 | | ⬜ |
| T-194 | PDPA readiness (by 2027-01-01): DPO fields, consent copy si/ta, data export/erasure request queue with deadlines, breach runbook, DPIA on location data, retention knobs | P1 | | ⬜ |
| T-226 | Fraud flags: rule-based (speed > 150 km/h between pings, altitude jumps, mock location, GPS/routed ratio > 1.5, duplicate accounts, cancel abuse) → human review in consoles; never auto-ban | P1 | | ⬜ |
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
**T-129 settings layer** (unblocks every per-operator knob), then **T-030 → T-031** (outbox + realtime), **T-013b** (geo adapter + POIs), **T-148/T-149** (upfront fare, arrived/PIN/no-show rules), **T-190 + T-153** (push), then the **T-155/T-156 map-first app rewrite** tested on both simulators against the server dev stack, then **T-132/T-135** (ledger, standing). Each with tests, committed and deployed separately.
