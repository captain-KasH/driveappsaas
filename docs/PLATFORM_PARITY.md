# Platform parity — operator console + platform (super-admin) console for Taxios

Researched **2026-10-12** (all sources accessed that day; tags `[Pn]` resolve in §D). Companion to `UBER_PARITY.md`
(rider/driver apps) and `COMPETITOR_FEATURE_MATRIX.md` (vendor matrix, `[O1]`-style IDs there). Question answered here:
*what does an Uber-grade OPERATIONS layer look like, how do mature white-label vendors structure multi-tenant
management, and what exactly must Taxios build.* Pattern parity only — no Uber names, marks or layouts are copied.

Legend: `(UV)` = not verified from a primary source (search snippet, third-party listing, or common practice);
`(old)` = primary source ≥ 3 years old; `(3P)` = third party. Status marks in §C5 follow `FEATURE_STATUS.md`
(✅ verified · 🟡 partial · ⬜ missing · ⛔ blocked externally). Terms: **platform** = Taxios super-admin (`/admin`),
**tenant/operator** = a taxi/tuk business (`/console`), **knob** = a runtime setting the apps read.

Hard constraints carried from `CLAUDE.md`: every tenant table has `tenant_id` + FORCE RLS + policy on `app_tenant_id()`;
platform scope only inside `PlatformRoute`; money in integer minor units; append-only money/history tables; secrets
hashed/encrypted, never displayed; unknown tenant → 404; missing permission → 403.

---

## A. Uber's own operations layer (benchmark)

Uber exposes three operations surfaces publicly and describes the internal ones in its engineering blog. Taxios' tenant
= an Uber *fleet partner* **plus** an Uber *city-ops team* **plus** an Uber *Central coordinator*, all in one console.

| # | Uber surface / concept | What it does (primary source) | Taxios translation | Src |
|---|---|---|---|---|
| A1 | **Fleet Hub / Supplier Portal** (fleet partners) | Web tool, free: add/edit/remove drivers; register vehicles and assign drivers to any approved vehicle (unassigned drivers pick an available vehicle before a shift); upload documents to complete fleet registration; "Only fleet partners can add vehicles and all the documents needed to go online"; earnings at driver **and** vehicle level; Live Map of drivers; balance + transactions; downloadable earnings reports; in-dashboard Help. Payout schedule not published. | Operator console Drivers/Vehicles/Earnings sections already map 1:1; add **vehicle-level** earnings, driver↔vehicle assignment with "pick a free vehicle at shift start", and a per-fleet balance ledger (cash: driver owes commission). | [P1] |
| A2 | **Uber Central** (dispatcher books for guests) | Coordinator enters rider phone + pickup/drop-off, vehicle type, driver notes; rider gets **SMS with details + web tracking link, no app required**; schedule up to 30 days ahead; recurring/"request again"; shared ride up to 5 drop-offs; in-progress/upcoming trips in one place; monthly statements + travel reports; coordinator sets rider **language** for SMS. Rider never charged — the business pays. | Phone-booking screen in the dispatch board = Central: phone-first customer lookup, SMS tracking link on the tenant's web domain, language per customer (si/ta/en), corporate account billing (B2B statement) later. | [P2] |
| A3 | **Uber for Business** billing model | Central trips bill the organisation; statements and spend visibility per org. | `corporate_accounts` with monthly statement per tenant (v1.1); dispatcher can tag a booking to an account. | [P2] |
| A4 | **Driver deactivation policy** (public, per city) | Reason categories: background check, **expired documents**, safety (failed Real-Time ID Check, unsafe driving, impairment, harassment, unapproved vehicle), fraud (fake/manipulated trips, duplicate accounts, falsified docs, cancellation abuse), **ratings below a city minimum (average of last 500)**, discrimination, general policy violations. Process: notify when access is at risk, manual review, in-app Review Center with evidence upload, temporary hold vs permanent loss, steps to regain access. Abusive-customer reports excluded. | Tenant-configurable **driver standing** engine: warning → temporary suspension → deactivation with reason codes, appeal note, audit; rating floor per tenant (rolling window N trips); document expiry auto-suspends via `driver_compliance` (exists). | [P3] |
| A5 | **Real-Time ID Check** | Periodic selfie match vs profile photo at login and between trips; not shown while moving; human or software review; biometric data deleted within 24 h; failure → immediate temporary suspension pending checks. | v1.1 "identity check" = driver takes a selfie at go-online; dispatcher-reviewed (no biometrics vendor); PDPA review before any automated matching. | [P4], [P5] |
| A6 | **H3 grid / surge zones** | Hexagonal hierarchical grid; demand vs supply measured per hexagon drives surge; neighbourhood-level analysis where point-level is too costly; stable cells independent of hand-drawn zones. | Operator zones stay PostGIS polygons (stands, airports, city limits). H3 only for heatmap/surge (later, T-143); surge off by default (LK operators prefer fixed/night tariffs). | [P6] |
| A7 | **Dispatch (DISCO) / batching / ETA ranking** | See `UBER_PARITY.md` §3: batched matching minimising average wait, road-ETA not distance, in-flight drivers considered. | Dispatch policy knob per tenant: `nearest` (today) → `eta` (T-013b) → `batch` (v1.1); plus `stand_fifo` and `broadcast` which Uber doesn't need but LK stands do. | UBER_PARITY [S6][S7] |
| A8 | **Experimentation Platform (XP)** | >1,000 concurrent experiments across apps; flags + treatment monitoring; 2024 engine moved evaluation client-side (100× faster). | Taxios needs **feature flags with per-tenant overrides** (T-125), not A/B at this scale; evaluate server-side, cache per tenant for 60 s. | [P7], [P8] |
| A9 | **COTA (support ticket assistant)** (old) | In-house support platform, hundreds of thousands of tickets/day; ML suggests contact type + reply template to agents; v1 cut English resolution time >10 %. | Support inbox per tenant with **contact-type taxonomy + canned replies in si/ta/en**; link each ticket to a booking/driver/customer; AI suggestions later (M12). | [P9](old) |
| A10 | **Project RADAR (fraud)** | Rules-based early detection with humans in the loop: AI drafts a blocking rule for a narrow marketplace slice, analyst approves; focus on payment fraud. | Platform-level fraud rules (duplicate accounts, GPS jump > X km/min, cancel-abuse ratio, self-booking) → **flag, don't auto-ban**; reviewer approves in the platform console. | [P10] |
| A11 | **Fleet-level live map + driver/vehicle performance** | Live Map; earnings per driver/vehicle; (2016 UberFLEET: performance dashboards, daily/weekly trip stats) (old). | Operator live map (all drivers + trips), performance per driver (acceptance, cancel, rating, on-time) and per vehicle (utilisation, revenue). | [P1], [P11](old) |
| A12 | **Public status / reliability** | No public status page cited here; reliability tiers in UBER_PARITY [S32]. | Platform health page + per-tenant provider health (SMS/push/maps). | — |

**Benchmark summary.** Uber's ops layer is (1) a fleet portal, (2) a phone/guest booking console with SMS links,
(3) a policy-driven driver standing/deactivation engine with appeal, (4) grid-based pricing/analytics, (5) rules+human
fraud review, (6) flag/experiment infrastructure, (7) a ticket system with suggested replies. None of these require ML
to reach parity *in function* at single-operator scale; they require good data models, audit, and configurable policy.

---

## B. How white-label vendors structure multi-tenant management

Verification: most vendor claims are marketing/manual pages (`Adv` in the matrix). TaxiCaller manual pages returned
HTTP 403 on fetch; their content is quoted from search snippets `(UV)`. "Platform-level" = set by the SaaS owner;
"tenant-level" = set by the operator.

| Capability | Onde | Yelowsoft | Jugnoo | TaxiCaller | iCabbi | Autocab (Ghost) | Cabstartup / clone vendors | Generic SaaS admin pattern | Taxios level |
|---|---|---|---|---|---|---|---|---|---|
| Tenant / multi-company model | One "company" per partner; branded apps under a **separate store author account per partner** [O8] | Multi-brand/multi-service "in deployments" (3P, UV) | White-label per licence (own deployment) [P12] | "multi-company profile in your account" to manage several fleets (UV) [P13] | **Sites** = extra fleet inside one system, overflow between sites [P14] | Per-operator install; iGo marketplace for demand sharing | Per-licence source code; sub-admins (UV) | Tenant row + status + plan + owner | Platform: tenant lifecycle; tenant: sub-fleets later (`fleets` table) |
| Operator/dispatcher accounts | "Operators" tab in My hub; access types; **limited access** = create/view/change orders but not other operators'/passengers' orders; suspend/delete operator [P15] | Role panels (UV) | Admin + dispatcher dashboards [P12] | Users + user permits (UV) [P16] | Mobile dispatch app with permissions [P17] | "permission-controlled modules" [P18] | Role-based staff access (UV) | RBAC, least privilege | Tenant: roles exist (owner/admin/dispatcher/finance/viewer); add `fleet_manager`, `support` |
| Dispatch policy config | Queue algorithm, manual mode, 100 km panel radius [O5] | Auto/manual, drag-drop allocation, request broadcasting (plan-gated) [Y2] | Nearest-driver auto [J2] | Winner = **closest car or first in zone queue**; priority jobs block lower-priority within radius/zone; advanced zone queues (UV) [P19] | Zones each with own **dispatch rules**, profiles **scheduled by day/time** or started manually, real-time profile changes, zone blocking [P14] | Intelligent auto-dispatch, late-job warnings, "The Wall" (regular-customer priority + direct drivers to demand), driver bidding for pre-bookings, dual queue zones [P18], [P20] | Configurable radius/timeouts (UV) | — | **Tenant + zone**: policy, radius, offer TTL, retries, schedule profiles |
| Zones | Service types + areas [O3] | Geofenced zones (UV) | "identify high demand areas" [P12] | Assignment zones + tariff zones, **GeoJSON import/export** (UV) [P16] | Zones + Sites [P14] | Zones central to coverage [P20] | Geofences (UV) | — | Tenant (`service_zones` exists); add zone kinds (service/stand/surcharge), GeoJSON import |
| Fare rules / tariffs | Dynamic pricing in FAQ [O4] | Per-plan | Surge in dashboards [J2] | Tariffs + tariff zones (UV) | Configurable | Configurable | Per city/ride-type rates (UV) | — | Tenant + category (`fare_rules` versioned exists); add time-of-day/night, airport flat, surcharges |
| Driver docs / compliance | — | Compliance reporting (3P) | — | Driver permits + vehicle permits, approval flow, alerts, suspensions (UV) [P16] | Driver Records: licences, insurance, performance stats [P14] | — | Doc upload + approval (UV) | — | Tenant (`driver_documents` + `driver_compliance` exist); add vehicle docs, expiry jobs, standing |
| Commission / settlement | Driver wallet; Company API changes driver balances [O7] | Per-trip commission deducted; settlement adjustments [Y2] | Earnings summary [J2] | — | Driver Pay wallet, scheduled/instant payouts [I1] | Accounts module (reviewed as weak) [A3] | Commission per city/type/tier, payout schedule, manual payouts (UV) [P21] | Ledger + statements | Tenant: commission rules; append-only driver ledger; cash-in |
| Reports / BI | Performance reports [O1] | Wallet/transaction reports [Y2] | Revenue analysis [P12] | Reports with Report IDs pullable via API [T3] | **Data Suite: 290+ parameters, scheduled delivery** [P14] | Analytics product [P18] | Basic | Per-tenant dashboards + platform roll-up | Tenant reports + platform usage/revenue |
| Automations / hooks | — | — | — | — | **Hooks**: auto tasks after events (e.g. thank-you SMS) [P14] | — | — | Webhooks | Tenant: notification templates per event; outbound webhooks (T-192) |
| Branded apps / store accounts | Onde uploads; separate author account per partner [O8] | Paid add-on [Y1] | Included [J1] | — | Fleet's own store listing [I2] | White-label passenger app [P18] | Source handed over | — | Platform: build request queue; tenant: app identity + own store credentials (D-022) |
| Plans / entitlements / usage | Plan + revenue share [O1] | Trip buckets, extra-trip fee, languages per plan, broadcasting plan-gated [Y1] | Ride buckets [J1] | Per activated vehicle [T1] | — | — | Licence | Boolean features, numeric limits, config entitlements; usage meters per window; behaviour at limit predefined [P22], [P23] | **Platform**: plans, entitlements, flags; tenant sees usage |
| Impersonation / support access | — (not published) | — | — | — | — | — | — | Elevated role, time-boxed, both actor and target logged, persistent UI banner, cannot impersonate higher-privilege users [P24] | Platform: "login as" with reason + TTL + audit (T-020) |
| Audit log | — | — | — | — | — | — | — | Who/when/what/why, before/after, append-only, exportable; alert on role/API-key/auth changes [P24] | Both levels (`audit_events` exists, append-only) |
| API / webhooks | REST+WS, Operator/Company/Wallet APIs, sandbox [O6] | — | Open APIs + webhooks [J3] | API keys, partner pricing/booking push [T3] | Gated partner API [I4] | — | — | Scoped keys, signed webhooks | Tenant: API keys + webhooks (T-191/192) |
| Status / SLA | 99.98 % uptime claim [O4] | — | — | — | — | — | — | Status page | Platform health page |

**Patterns that recur across every mature vendor:** (1) zones carry their own dispatch rules and queues; (2) dispatch
profiles switch by schedule (day/night, event); (3) dispatcher accounts have scoped visibility; (4) driver records bundle
documents + performance; (5) settlement is a wallet/ledger with adjustments; (6) reports are schedulable; (7) branded
apps live under the operator's own store account; (8) plans gate features and meter trips/vehicles. **What none of them
publish** (and where Taxios can lead): audited impersonation, signed webhook docs, per-tenant feature flags, a public
entitlement model, and tenant-isolation tests as a security claim.

---

## C. Target multi-tenant platform for Taxios

### C1. Platform (super-admin) console — `/admin`

Served only by `PlatformRoute` handlers; roles `super_admin` (all) and `support` (read + time-boxed support access).
MFA (T-006) required for both before launch.

| Area | Capabilities (target) | Key rules | Today |
|---|---|---|---|
| Tenant directory | Search/filter by status, plan, country, created, last activity; health badges (drivers online, bookings 24 h, overdue invoice, failed providers). | Cursor pagination (T-106). | 🟡 directory + filters live |
| Tenant lifecycle & verification | Queue of `pending_verification` with documents viewer, approve/reject with reason, request more info; status actions `activate / suspend (reason, until) / close (grace, export)`; reactivation; isolation mode; owner contact. | Self-registered tenants must upload documents (exists). Suspension blocks writes (exists) and must also push "service paused" to apps. | ✅ queue/approve/reject; ⬜ suspend-until, close-with-export |
| Plans, subscriptions, entitlements | Plan catalogue as data; assign/change plan; trial dates; **entitlement matrix** = boolean features + numeric limits (drivers, vehicles, staff, bookings/month, SMS/month) + config entitlements (data retention, languages); manual overrides per tenant with expiry; invoice list, mark paid (bank transfer), grace/overdue policy. | Entitlement check is server-side in one hook used by every tenant route; behaviour at limit predefined (block / warn / allow-overage). | ⬜ (M10) |
| Feature flags | Global flag registry with per-tenant override and % rollout; read through the same entitlement hook; audited changes. | Cache per tenant 60 s; flags never bypass permissions. | ⬜ (T-125) |
| Usage & revenue | Per-tenant meters (bookings, completed trips, SMS, push, API calls, storage) by month; MRR/ARR, overdue, churn; CSV export. | Meters written idempotently from the outbox (T-030/T-061). | ⬜ |
| Global audit | Cross-tenant `audit_events` search (actor, tenant, action, target, time), export; alerts on role changes, API-key creation, auth settings, impersonation. | Append-only (exists). | ⬜ UI (table exists) |
| Support access ("login as") | Start session with **tenant, reason code, free text, TTL ≤ 60 min**, optional owner consent request; every action logged with `actor_user_id` + `on_behalf_of_tenant`; persistent banner in console; cannot act as `owner`-only actions (billing, delete) unless explicitly granted; session list + revoke. | Platform scope never leaks: support session is a normal tenant-scoped transaction with `app.support_session_id` set; RLS unchanged. | ⬜ (T-020) |
| Announcements | Platform → tenant banners/in-app notices (maintenance, policy), scheduled, per-plan/per-tenant targeting, read receipts. | | ⬜ |
| Support inbox (platform) | Tickets raised by tenants (billing, build requests, abuse); link to tenant; SLA timers; canned replies. | | ⬜ |
| Health / status | Service health (DB, worker lag, outbox backlog, provider adapters), per-tenant provider failures, public status page. | `/healthz` exists. | 🟡 |
| App-build requests | Queue per tenant: app identity (name, bundle/package id, icon, splash), store credential status (uploaded/validated/expired, **never displayed**), build + submit runs with logs, store review status, version history. | D-022: tenant's own store accounts; creds encrypted at rest. | ⬜ (T-173/174) |
| Domains | Subdomain allocation `{slug}.taxios.lk`; custom-domain requests with TXT verification + TLS status; one domain per tenant for booking site/tracking links. | A tenant can't claim another's domain. | ⬜ (T-171) |
| Abuse / fraud controls | Platform rules (duplicate phone/email across tenants, GPS jump, cancel-abuse, self-booking, promo abuse) raise **flags** reviewed here; actions: warn tenant, freeze driver/customer, suspend tenant. Rate-limit overrides. | Rules flag, humans decide (A10). | ⬜ |
| Data requests (PDPA) | Queue of export/erasure requests (from tenant console or passenger app); produce JSON/CSV bundle; erasure = anonymise PII, keep ledger/audit rows; deadline timer; evidence stored. | Sri Lanka PDPA No. 9 of 2022: rights of access, rectification, erasure, objection, withdrawal via written request (ss. 13–16); breach notice to DPA within 72 h (draft regs); statutory response window not confirmed in this pass `(UV)` — treat as 30 days until verified. | ⬜ (T-133) |
| Platform settings | Default knobs (§C3 platform scope), provider credentials (SMS/email/push/maps) with per-tenant override allowed by plan, legal texts versions. | Secrets env/encrypted; never in logs. | 🟡 (env only) |

### C2. Operator console — `/console` (Uber-grade live ops)

| Area | Capabilities (target) | Dispatch/money rules | Today |
|---|---|---|---|
| Overview | Live KPIs (online drivers, searching/assigned/in-progress, unassigned > N min, cash owed, docs expiring 7 d), today vs yesterday, alerts feed. | | 🟡 overview |
| **Live map** | All online drivers (status colour, category icon, stale-location flag), active trips with route lines, open bookings as pins, zone overlays, filters by category/zone/status, click → driver/booking card, follow mode. WebSocket feed (T-031), 5 s polling fallback. | Location older than `driver.location_stale_s` greys the pin. | ⬜ (dispatch tab lists only) |
| **Dispatch board** | Columns: unassigned / offered / assigned / en-route / arrived / in-progress / scheduled; **alerts**: unassigned > N min, late pickup (ETA > scheduled), driver stationary > N min on trip, route deviation, no-show candidate; manual assign (eligible driver list ranked by ETA, stale-location warning), reassign (with reason, frees previous driver), bulk assign/cancel, dispatcher notes, search by phone/plate/ref, keyboard-first. | Lock order booking → offer → driver; manual offer TTL knob; auto-reassign on decline/expiry when `dispatch.auto_reassign` on. | 🟡 queue + manual offer; alerts/reassign/bulk ⬜ (T-146) |
| Bookings on behalf (phone) | Phone-first customer lookup/create, address search + map pin, category + quote, now/scheduled/recurring, notes, corporate account tag, language; **SMS tracking link** on tenant domain; "request again" from history; WhatsApp intake later. | Idempotency-Key per submit (exists). | 🟡 (no SMS link, no recurring) |
| Zones, categories, fare rules | Zones with kinds (service area, stand/queue, airport, surcharge) + GeoJSON import/export; categories (tuk, mini, car, van) with seats/luggage/icon; versioned fare rules per category: base/distance/time/minimum/wait/booking fee/night window/scheduled fee/airport flat/zone-to-zone flat/cancellation + no-show fee/rounding/tax; upfront-fare deviation tolerance; surge (opt-in, capped, audited) later; **dispatch profiles by schedule** (day/night/event). | Money integer minor units; rule versions append-only (exist). | ✅ zones/categories/fares; ⬜ zone kinds, night/airport/flat, profiles, cancellation fees (T-141), surge (T-143) |
| Promotions | Promo codes (amount/%, cap, window, first-ride, per-customer limit, category/zone scope), referral credits later; usage report. | Validated server-side; discount is a ledger line. | ⬜ (T-142) |
| Customers (CRM) | Search, profile, trips, spend, saved places, notes/tags, language, blocked flag + reason, corporate accounts (members, monthly statement), PDPA export/erasure request, CSV import/export. | | 🟡 list + CSV export |
| Drivers | Onboarding pipeline (invited → documents → review → approved), document viewer with expiry + reminders (30/7/1 d) and auto-suspend, vehicle assignment (fixed or "pick free vehicle at shift start"), availability/shift log, performance (acceptance %, cancel %, rating window, on-time %, trips/day), **standing**: warnings → temporary suspension (until) → deactivation with reason codes + appeal note, rating floor, identity-check photo (v1.1), earnings + commission owed, messages. | Eligibility = `driver_compliance` view only. | ✅ profile/docs/approval/suspend; ⬜ pipeline UI, reminders, shifts, performance, standing engine, earnings (T-130/132/134) |
| Vehicles | Docs (registration, insurance, revenue licence, emission) with expiry; inspection + service dates; maintenance state (in service / workshop / retired) with history; capacity + accessibility; utilisation and revenue per vehicle. | Expired vehicle doc blocks dispatch like driver docs. | ✅ basics; ⬜ docs/maintenance (T-131) |
| Earnings, commission, settlement (cash) | Commission rules (% or flat per category/zone/time, driver-specific override); per-trip ledger lines (fare, commission, discount, adjustment); **driver balance** = cash collected − commission owed; cash-in at office (receipt no.), wallet top-up later; daily/weekly driver statements (PDF/SMS summary); operator settlement report; adjustments with reason + audit. | Ledger append-only; balance derived, never stored as the source of truth; no float. | ⬜ (T-132/144/162) |
| Reports / BI | Trips, revenue, cash, cancellations, driver/vehicle performance, zone heatmap, hourly demand, scheduled email/CSV; API pull by report ID (v1.1). | `reports.read`. | 🟡 summary + CSV |
| Branding, app identity, website | Logo/colour/tagline/links (exists); app name, icon, splash, store links; booking website template + SEO + legal pages; custom domain request; tracking-page theme. | | 🟡 brand ✅, rest ⬜ (T-170…174) |
| Notification templates | Per event (booking confirmed, driver assigned, arriving, arrived, receipt, cancelled, doc expiring, statement) × channel (SMS/push/email/in-app) × language (si/ta/en); placeholders; test send; delivery log + cost. | Outbox-driven, deduped (T-190). | ⬜ |
| Team & roles | Owner/admin/dispatcher/finance/viewer + `fleet_manager` (sub-fleet scope) + `support_agent`; invitations; per-member sessions; dispatcher scope (zone/fleet) like Onde's limited access. | Last owner can't be removed (exists). | ✅ core; ⬜ scoped roles |
| Settings | PIN policy (off/night/always), cancellation policy (free window, fee, no-show wait), wait-time grace + rate, dispatch policy + offer timeout + max radius + retries + auto-reassign + stand FIFO, location staleness, quote TTL, scheduled lead time, upfront-fare deviation, languages + default, phone country, SOS numbers, driver-visible contact policy (masked/hotline/direct), data retention. | Stored in `tenant_settings`, validated ranges, audited; apps read via `/v1/public/tenants/{slug}` + authenticated settings endpoint. | ⬜ (constants hard-coded in `booking.go`) |
| Integrations | API keys (scopes, rotation, last used), webhooks (events, secret, retries, dead-letter, replay, logs), provider health, SMS/email sender identity. | Keys stored as SHA-256 (as sessions). | ⬜ (T-191…193) |
| Support inbox | Tickets from passengers (app) and drivers, linked to booking; contact types; canned replies si/ta/en; SLA; escalate to platform. | | ⬜ |
| Subscription & billing | Plan, usage vs limits, invoices, payment instructions, add-ons. | Read-only from entitlements. | ⬜ (T-185) |

### C3. Per-tenant configuration knobs (runtime)

Scope: **P** platform default (overridable per tenant unless marked P-only) · **T** tenant · **Z** zone · **C** category.
Used by: **PA** passenger app · **DA** driver app · **DS** dispatch worker/board · **FA** fare engine · **NT** notifications.
Today every dispatch value is a Go constant (`booking.go`: OfferTTL 30 s, ManualOfferTTL 90 s, DispatchRadiusM 8000,
SearchTimeout 10 min, ScheduledLeadTime 15 min, QuoteTTL 10 min; location freshness 2 min) — they become T-scope rows.

| Knob | Scope | Default | Used by |
|---|---|---|---|
| `brand.*` (logo, colour, text colour, tagline, support phone/email, terms/privacy URLs) | T | — | PA DA web |
| `app.identity` (name, bundle/package id, icon, splash, store URLs) | T | slug-derived | build pipeline, PA DA |
| `locale.languages` / `locale.default` | P→T | en,si,ta / en | PA DA NT console |
| `locale.phone_country` | P→T | LK | PA DA console |
| `locale.currency`, `locale.timezone` | T | LKR / Asia/Colombo | FA PA DA reports |
| `categories[]` (code, names si/ta/en, icon, seats, luggage, sort) | C | tuk, mini, car, van | PA DA DS |
| `zone.kind` (service / stand / airport / surcharge), `zone.polygon` | Z | service | DS FA PA |
| `dispatch.policy` (nearest / eta / stand_fifo / broadcast / batch) | T→Z | nearest | DS |
| `dispatch.profile_schedule[]` (policy + radius + TTL by weekday/time) | T→Z | none | DS |
| `dispatch.offer_ttl_s` / `dispatch.manual_offer_ttl_s` | T | 30 / 90 | DS DA |
| `dispatch.max_radius_m` (+ expand steps) | T→Z | 8000 | DS |
| `dispatch.search_timeout_s` | T | 600 | DS PA |
| `dispatch.max_offers_per_booking` | T | 8 | DS |
| `dispatch.auto_reassign` (on decline/expiry/abort) | T | true | DS |
| `dispatch.scheduled_lead_s` | T | 900 | DS PA |
| `dispatch.block_mutual_one_star` | T | false | DS |
| `driver.location_stale_s` / `driver.ping_idle_s` / `driver.ping_trip_s` | P→T | 120 / 10 / 5 | DS DA |
| `driver.rating_floor` + `driver.rating_window_trips` | T | off / 100 | standing engine, DA |
| `driver.doc_expiry_reminders_d` | T | 30,7,1 | NT DA |
| `driver.identity_check` (off / go_online / daily) | T | off | DA console |
| `driver.contact_policy` (masked / hotline / direct) | T | hotline | PA DA |
| `driver.destination_filter_per_day` | T | 0 | DA DS |
| `booking.quote_ttl_s` | T | 600 | FA PA |
| `booking.upfront_deviation_pct` / `_min` | T | 20 / 10 | FA DA PA |
| `booking.allow_scheduled_days_ahead` | T | 30 | PA console |
| `booking.pin_policy` (off / night / always) + `night_window` | T | night, 21:00–05:00 | PA DA |
| `booking.cancel_free_window_s`, `booking.cancel_fee_minor`, `booking.no_show_wait_s`, `booking.no_show_fee_minor` | T→C | 120 / 0 / 300 / 0 | FA PA DA |
| `booking.wait_grace_s`, `fare.wait_rate_minor_per_min` | T→C | 180 / tariff | FA DA |
| `fare.rules` (versioned: base, per km, per min, minimum, booking fee, night %, scheduled fee, airport flat, zone flat, rounding, tax) | C(+Z) | — | FA |
| `fare.surge_enabled` / `fare.surge_cap` | T→Z | false / 1.5 | FA PA |
| `fare.commission` (% or flat per C/Z/time) | T→C | 10 % | ledger DA |
| `cash.enabled`, `card.enabled` | T | true / false | PA DA |
| `promo.enabled` | T (entitlement) | false | PA FA |
| `safety.sos_numbers` | P→T | 119, 1990 | PA DA |
| `safety.share_trip_domain` | T | {slug}.taxios.lk | PA NT |
| `notify.templates[event][channel][lang]` | P→T | platform defaults | NT |
| `notify.sms_sender_id` | T (plan) | TAXIOS | NT |
| `passenger.auth` (phone_otp / email) | P→T | email until OTP ⛔ | PA |
| `passenger.saved_places_max` | P | 10 | PA |
| `entitlements.*` (feature booleans, limits) | P per plan → T override | plan | all servers |
| `flags.*` | P → T override | off | all |
| `support.tenant_contact` (hotline, hours) | T | — | PA DA |
| `data.retention_days` (locations, events) | P per plan → T | 365 | jobs |
| `api.rate_limit` | P per plan | 600/min | API |

### C4. Data model additions (all tenant tables: `tenant_id` + FORCE RLS + `app_tenant_id()` policy + isolation test)

Existing: `tenants, users, sessions, user_tokens, tenant_memberships, tenant_invitations, tenant_documents, files,
service_zones, vehicle_categories, fare_rules, fare_rule_versions, drivers, driver_documents, vehicles, customers,
booking_quotes, bookings, booking_events, dispatch_offers, trips, location_updates, payments, ratings, audit_events`.

| Table (new) | Scope | Key columns | Notes |
|---|---|---|---|
| `tenant_settings` | T | tenant_id, key, value jsonb, updated_by, updated_at; PK (tenant_id,key) | §C3 knobs; audited; validated server-side |
| `tenant_domains` | T | tenant_id, hostname UNIQUE, kind (subdomain/custom), txt_token, verified_at, tls_status | hostname unique across platform |
| `app_identities` | T | tenant_id, variant (passenger/driver), display_name, ios_bundle_id, android_package, icon_file_id, splash_file_id, store_urls jsonb | one per variant |
| `store_credentials` | T | tenant_id, store (apple/google), ciphertext bytea, key_version, validated_at, expires_at, status | **never returned by API**; AES-GCM with platform KMS key |
| `app_builds` | T (+P view) | tenant_id, app_identity_id, version, build_number, status (requested/building/submitted/review/live/failed), log_file_id, requested_by, finished_at | platform queue |
| `dispatch_profiles` | T | tenant_id, zone_id null, name, policy, radius_m, offer_ttl_s, schedule jsonb, active | schedule by weekday/time |
| `zone_kinds` → column on `service_zones` | T | kind, stand_fifo bool, surcharge_minor | prefer column over table |
| `driver_standing_events` | T | tenant_id, driver_id, kind (warning/suspension/deactivation/reinstatement), reason_code, note, until, created_by | append-only; drives `driver_compliance` |
| `driver_shifts` | T | tenant_id, driver_id, vehicle_id, started_at, ended_at, online_s, trips | derived from status changes |
| `driver_identity_checks` | T | tenant_id, driver_id, photo_file_id, result, reviewed_by, created_at | photo deleted after review (retention knob) |
| `vehicle_documents` | T | tenant_id, vehicle_id, kind, file_id, expires_at, status | mirrors `driver_documents` |
| `vehicle_maintenance` | T | tenant_id, vehicle_id, state, inspection_due, service_due, note, created_at | history append |
| `commission_rules` | T | tenant_id, category_id null, zone_id null, driver_id null, kind (pct/flat), value, valid_from, valid_to | most specific wins |
| `ledger_entries` | T | tenant_id, driver_id, booking_id null, kind (fare_cash/commission/discount/adjustment/cash_in/payout), amount_minor, currency, ref, created_by | append-only; balance = SUM |
| `cash_settlements` | T | tenant_id, driver_id, amount_minor, receipt_no, received_by, created_at | writes a `ledger_entries` row |
| `corporate_accounts`, `corporate_members` | T | name, billing contact, monthly statement; member customer_ids | B2B (v1.1) |
| `promo_codes`, `promo_redemptions` | T | code, kind, value, cap, window, limits; booking_id, customer_id | redemption unique per (promo, booking) |
| `customer_notes` / `customer_flags` | T | customer_id, note / flag, reason, created_by | blocked customers excluded from booking |
| `support_tickets`, `support_messages` | T (+P for platform tickets) | subject, contact_type, status, booking_id/driver_id/customer_id, assignee; messages | SLA timers |
| `notification_templates` | T (P defaults with tenant_id NULL readable) | event, channel, lang, subject, body, version | P rows via platform policy |
| `notifications` | T | channel, recipient, template, status, provider_ref, cost_minor, dedupe_key UNIQUE | outbox-driven |
| `api_keys` | T | key_hash, scopes, last_used_at, revoked_at | hash only |
| `webhook_endpoints`, `webhook_deliveries` | T | url, secret_hash, events; attempt, status, response, next_retry | dead-letter + replay |
| `data_requests` | T | subject (customer/driver), kind (export/erasure), status, due_at, bundle_file_id, completed_by | PDPA |
| `fraud_flags` | T (+P view) | subject type/id, rule, score, evidence jsonb, status, reviewed_by | rules flag, humans decide |
| `outbox_events` | T | aggregate, type, payload, seq, delivered_at | T-030; feeds WS, NT, meters |
| **Platform-scoped** (no tenant_id; readable only under `app_platform_scope()`): `plans`, `plan_entitlements` (plan_id, key, kind bool/limit/config, value), `subscriptions` (tenant_id FK, plan_id, status, trial_ends, current_period), `invoices` + `invoice_payments` (manual, append-only), `tenant_entitlement_overrides` (tenant_id, key, value, expires_at), `feature_flags` + `feature_flag_overrides`, `usage_meters` (tenant_id, meter, period, qty; idempotent via event id), `support_sessions` (actor_user_id, tenant_id, reason_code, note, expires_at, ended_at), `announcements` + `announcement_reads`, `platform_settings`, `provider_credentials` (ciphertext). Tenant-side reads of their own subscription/usage go through explicit endpoints, not RLS on these tables. |

### C5. Priority

| Item | Prio | Status today | Task |
|---|---|---|---|
| Tenant lifecycle + verification queue | MVP | ✅ | T-120…T-124 |
| `tenant_settings` + settings page replacing Go constants (dispatch/PIN/cancel/wait/languages) | MVP | ⬜ | new T-129 |
| Plans + subscriptions + entitlement hook + manual invoices + tenant billing page | MVP | ⬜ | T-180…T-185 |
| Feature flags with per-tenant override | MVP | ⬜ | T-125 |
| Support access ("login as") with reason/TTL/audit + banner | MVP | ⬜ | T-020 |
| Global audit UI + alerts on sensitive changes | MVP | ⬜ (table ✅) | T-124 |
| App identity + store credentials (encrypted) + build request queue | MVP | ⬜ | T-173/T-174 |
| Live map (WebSocket) + dispatch alerts/reassign/bulk/notes | MVP | 🟡 | T-031, T-146 |
| Phone booking with SMS tracking link (public tokenised page) | MVP | 🟡 (no link ⛔ SMS) | T-041, T-171 |
| Commission rules + append-only ledger + cash-in + driver statements | MVP | ⬜ | T-132/T-144 |
| Driver standing engine (warn/suspend/deactivate + appeal) + doc expiry reminders | MVP | 🟡 (suspend ✅) | T-130 |
| Vehicle docs + maintenance | v1.1 | ⬜ | T-131 |
| Zone kinds + stand FIFO + dispatch schedule profiles | v1.1 | ⬜ | new T-147 |
| Night/airport/zone-flat tariffs + cancellation/no-show fees | MVP (night/flat), v1.1 (fees) | ⬜ | T-141 |
| Notification templates si/ta/en + delivery log | MVP (needs ⛔ SMS/email creds) | ⬜ | T-190, T-040, T-041 |
| Team scoped roles (fleet_manager, dispatcher zone scope, support_agent) | v1.1 | 🟡 | T-122 follow-up |
| Customers CRM (notes/flags/corporate accounts/import) | v1.1 | 🟡 | T-127, T-133 |
| Promotions | v1.1 | ⬜ | T-142 |
| Usage meters + platform revenue reporting | v1.1 | ⬜ | T-061, T-186 |
| Announcements, platform + tenant support inbox | v1.1 | ⬜ | new T-188 |
| API keys + webhooks + provider health | v1.1 | ⬜ | T-191…T-193 |
| Custom domains + TLS | v1.1 | ⬜ | T-171 |
| PDPA data export/erasure queue | v1.1 (before 1k passengers) | ⬜ | T-133 |
| Fraud flags (rules + review) | v1.1 | ⬜ | new T-226 |
| Driver identity check (selfie, human review) | later | ⬜ | — |
| Surge + H3 heatmap | later | ⬜ | T-143 |
| Sub-fleets (Sites) + cross-tenant overflow | later | ⬜ | T-085 |
| Health/status page, metrics | MVP (internal), v1.1 (public) | 🟡 | T-223 |

---

## D. Sources (accessed 2026-10-12)

| ID | URL | Note |
|---|---|---|
| P1 | https://www.uber.com/us/en/earn/fleet-management/ | Fleet Hub: drivers, vehicles, docs, earnings per driver/vehicle, Live Map, balance, reports, Help |
| P2 | https://www.uber.com/us/en/business/products/central/ | Uber Central: phone-number booking, SMS tracking link, 30-day scheduling, recurring, 5-stop shared, statements, language |
| P3 | https://www.uber.com/us/en/drive/driver-app/deactivation-review/ | Deactivation policy: reason categories, Review Center, ratings window, notice |
| P4 | https://www.uber.com/en-GB/newsroom/uber-launches-real-time-id-check-for-drivers-in-the-uk | Real-Time ID Check UK: human/software review, biometrics deleted in 24 h, temporary suspension on failure |
| P5 | https://privacyinternational.org/sites/default/files/2021-12/Uber response to Privacy International - 19 Nov 2021.pdf | (3P-hosted Uber letter) check cadence: login + between trips, not while moving |
| P6 | https://www.uber.com/blog/h3/ | H3 for surge, supply/demand analysis, stable cells |
| P7 | https://www.uber.com/blog/xp/ | Experimentation Platform (2018, old) |
| P8 | https://www.uber.com/blog/making-ubers-experiment-evaluation-engine-100x-faster | Client-side flag evaluation (2024) |
| P9 | https://www.uber.com/blog/cota/ · https://www.uber.com/gb/en/blog/cota-v2/ | COTA support assistant (2018, old) |
| P10 | https://www.uber.com/en-DK/blog/project-radar-intelligent-early-fraud-detection | Project RADAR: rules + human analysts (2022) |
| P11 | https://www.medianama.com/2016/12/223-uber-fleet-launch-india/ | UberFLEET 2016 (3P, old) |
| P12 | https://jungleworks.com/download/Jugnoo%20Pitch%20Deck%20%281%29_compressed.pdf | Jugnoo white-label deck: admin dashboard, auto dispatch, demand areas, revenue analysis |
| P13 | https://taxicaller.com/manuals/admin-panel/2/en/topic/your-companies | TaxiCaller multi-company profile (UV — fetch 403, search snippet only) |
| P14 | https://icabbi.com/platform/dispatch/ | iCabbi: zones with own rules, scheduled profiles, Sites, Hooks, Driver Records, Data Suite 290+ params |
| P15 | https://onde.app/learn/onde-system-setup-tips/lessons/3 | Onde operators tab, access types, limited dispatch access |
| P16 | https://www.manula.com/manuals/taxicaller-nordic-ab/admin-panel/3/en/topic/settings | TaxiCaller settings sections, zones GeoJSON, permits (UV — fetch 403, snippet) |
| P17 | https://icabbi.com/?p=16254496 | iCabbi Pulse mobile dispatch: live drivers, queues, zone blocking, messaging (UV snippet) |
| P18 | https://www.autocab.com/solution/booking-and-dispatch | Autocab Ghost: intelligent auto-dispatch, configurable dashboard, late-job warnings, The Wall (UV snippet) |
| P19 | https://www.taxicaller.com/manuals/admin-panel/2/en/topic/advanced-auto-dispatch | TaxiCaller winner = closest or zone-queue first; priority blocking (UV — fetch 403, snippet) |
| P20 | https://apps.apple.com/app/id1290768019 · https://au.autocab.com/?p=3654 | Autocab Driver Companion: pre-booking bidding, dual queue zones (UV snippet) |
| P21 | https://www.marketplace.wrteam.in/docs/etaxi/docs/features · https://www.ongraph.com/driver-commission-management-system-for-taxi-apps/ | Clone-vendor commission/payout feature lists (3P, UV) |
| P22 | https://www.stigg.io/blog-posts/the-engineers-guide-to-entitlements-what-they-are-and-why-your-current-setup-isnt-enough | Entitlement kinds: boolean, numeric limit, config; single source of truth (vendor) |
| P23 | https://docs.schematichq.com/what-is-schematic | Usage-based entitlements, behaviour at limit, attribution (vendor) |
| P24 | https://www.skills.sh/peterbamuhigire/skills-web-dev/saas-admin-backoffice-tooling · https://help.octopuspro.com/user-impersonation-login-as | Impersonation guardrails (elevated role, TTL, dual attribution, banner), audit fields, alerting (UV practitioner guides) |
| P25 | https://dpa.gov.lk/newsappeals.php · https://dpa.gov.lk/newsinquiry.php | Sri Lanka PDPA No. 9 of 2022: rights via written request (ss. 13–16), fee s. 17(6), draft 72 h breach notice; response deadline not confirmed (UV) |
| P26 | https://tech.nearme.jp/en/blog/2021-05-08-multitenant-dispatch/ | Multi-tenant dispatch: one dispatch system per operating company (3P) |

Not found in this pass (UV, left out of the tables): Uber "Vehicle Solutions" as a product name; any vendor (Onde,
Yelowsoft, Jugnoo, iCabbi, Autocab, Cabstartup, Mobisoft, Grepix/Appicial, SpotnRides) publishing a super-admin /
impersonation / feature-flag model; Uber city-ops dashboards as a primary source.
