# Competitor feature matrix — white-label ride-hailing / taxi dispatch SaaS

Research date: **2026-10-11** (every source below was accessed on that date unless noted). Scope: Onde, Yelowsoft,
TaxiCaller, iCabbi, Autocab, Jugnoo (Yelo / Cabstartup lineage), Limo Anywhere, plus Gazoop and Cabica as extra
credible white-label dispatch SaaS. Sri Lankan incumbents (PickMe, Uber LK) appear only as market context.

Column key. **Verif.** = `Adv` (vendor marketing/help page only), `Adv+3P` (vendor claim also echoed by a third-party
listing/review/press), `Official` (platform owner's primary document, e.g. Apple), `Press` (independent press/analyst).
Nothing here was verified hands-on; "verified capability" means "what the source states exactly", not "tested by us".
**Imp.** = must / should / nice. **Cx** = S/M/L estimated build complexity for Taxios. **Prio** = MVP / v1.1 / later.
**Taxios current status** per capability group is kept in `FEATURE_STATUS.md` (one row per prompt section) rather than
repeated here; the "Prio" column is what feeds `IMPLEMENTATION_PLAN.md`. Summary (2026-10-11): MVP table stakes that
Taxios already has = instant/scheduled booking, server fare estimate, auto + manual dispatch, phone-order entry, driver
documents, cash ledger + receipts, reports; MVP gaps = self-service signup, branding/white-label, mobile apps, push/SMS,
corporate accounts, CSV export, plans/billing.

## Sources (ID → URL; all accessed 2026-10-11)

| ID | URL | Note |
|---|---|---|
| O1 | https://onde.app/pricing | Onde pricing |
| O2 | https://onde.app/ | Onde home |
| O3 | https://onde.app/features | Onde features |
| O4 | https://onde.app/faq-page | Onde FAQ (dynamic pricing, gateways, uptime, 4-week launch, data ownership) |
| O5 | https://onde.app/operator-app | Onde operator app (queue algorithm, manual mode, PBX, Operator API) |
| O6 | https://onde.app/platform-api | Onde platform API (REST+WebSocket, sandbox, "webhook testing") |
| O7 | https://intercom.help/ondeeducate/en/articles/5463151-integration-with-onde-api | Onde API scopes + sandbox process |
| O8 | https://onde.app/blog/aso-seo/how-we-prepare-apps-for-submission-to-stores | Onde store submission ("separate author's account for each new business partner") |
| O9 | https://onde.app/glossary | Onde glossary (corporate account definition) |
| Y1 | https://www.yelowsoft.com/pricing/ | Yelowsoft pricing |
| Y2 | https://www.yelowsoft.com/features/ | Yelowsoft features |
| Y3 | https://www.yelowsoft.com/taxi-dispatch-software/ | Yelowsoft dispatch page |
| Y4 | https://www.capterra.com/p/169474/Yelowsoft/pricing/ | Capterra listing (3P; plan contents, review quotes) |
| T1 | https://www.taxicaller.com/en/faq | TaxiCaller FAQ (site returns 403 to fetchers; content taken from search index) |
| T2 | https://www.taxicaller.com/campaigns/en/quebec | TaxiCaller Québec campaign (CAD 29/vehicle/month, CAD 350 first year) |
| T3 | https://www.taxicaller.com/en/help/kb/what-is-the-taxicaller-api | TaxiCaller API KB |
| T4 | https://www.getapp.com/transportation-logistics-software/a/taxicaller/ | GetApp listing (3P) |
| I1 | https://www.icabbi.com/ | iCabbi home (Dispatch, Move AI, Driver Pay, Voice AI, Exchange, 99.999% SLA, per-driver licence) |
| I2 | https://www.icabbi.com/products/passenger-app | iCabbi passenger app features |
| I3 | https://www.icabbi.com/products/driver-app | iCabbi driver app features |
| I4 | https://icabbi.com/?p=16247250 | "What are iCabbi APIs?" (API keys via marketplace agreement) |
| A1 | https://www.autocab.com/ | Autocab home (Ghost, Phantom, Driver Companion, E-booker, iGo, surge, corporate portal) |
| A2 | https://apps.apple.com/app/id1290768019 | Autocab Driver Companion App Store listing |
| A3 | https://www.capterra.com/p/153213/Ghost-Cloud/ | Capterra Ghost Cloud (3P; 2.5/5 from 4 reviews) |
| J1 | https://www.jugnoo.io/pricing/ | Jugnoo pricing |
| J2 | https://www.jugnoo.io/taxi-dispatch-software/ | Jugnoo dispatch features |
| J3 | https://community.jungleworks.com/t/how-to-implement-webhooks/618 | Jungleworks forum: webhooks via open API docs |
| L1 | https://www.limoanywhere.com/pricing/ | Limo Anywhere pricing |
| L2 | https://www.limoanywhere.com/ | Limo Anywhere home |
| L3 | https://www.limoanywhere.com/?p=4489 | DriverAnywhere April 2021 notes (FlightStats on trip cards) |
| L4 | https://www.limoanywhere.com/?p=741 | ORES booking widgets announcement |
| G1 | https://www.gazoop.com/pricing | Gazoop pricing + add-ons |
| C1 | https://www.getapp.com/all-software/a/cabica/ | Cabica via GetApp (3P; no first-party pricing page found) |
| AP1 | https://developer.apple.com/app-store/review/guidelines/ | Apple App Review Guidelines 4.2.6 / 4.3(a) |
| AP2 | https://developer.apple.com/forums/thread/759119 | Apple dev forum: white-label apps from client accounts rejected under 4.3(a) |
| P1 | https://firstcapital.lk/wp-content/uploads/2025/12/PKME-Initition-Coverage-09.12.2025.pdf | First Capital PickMe initiation (65–70% share) |
| P2 | https://www.ft.lk/front-page/BIA-to-earn-Rs-700-per-taxi-from-PickMe/44-751362 | PickMe BIA airport deal |
| P3 | https://pickme.lk/pickme-welcomes-hnb-finance-to-the-corporate-transport-partnership-fold | PickMe Business Portal (950+ corporates) |
| U1 | https://www.uber.com/en-LK/blog/scheduledrides | Uber LK scheduled rides |
| U2 | https://www.uber.com/en-LK/newsroom/uber-launches-tuk-rentals-in-sri-lanka-for-multi-hour-multi-stop-needs | Uber Tuk Rentals |
| U3 | https://www.dailymirror.lk/print/business/Uber-expands-to-Kegalle/215-338572 | Uber Kegalle expansion (24 Apr 2026) |

## 1. Capability matrix

### 1.1 Passenger booking & journey

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Proposed Taxios differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Passenger booking | Instant + pre-order, multiple drop-offs, favourite addresses, in-app chat, share ride, SOS, cash/card/terminal, order for corporate clients | O3 | Adv | must | M | MVP | Phone-number-first login (LK norm), Sinhala/Tamil UI, cash default with card optional |
| Yelowsoft | Passenger booking | Instant + "Advance Trip Scheduling", "Daily or Weekly Ride Scheduler", SOS, in-app chat, "Web eBooker"; multi-stop NOT mentioned | Y2 | Adv | must | M | MVP | Multi-stop at MVP (Onde/iCabbi have it; Yelowsoft doesn't) |
| iCabbi | Passenger booking | Book/Pre-Book up to 7 days, Vias (driver side), fare estimates, vehicle selector with per-vehicle price, Track My Driver, Share Trip, Live Activities, chat, rate driver, trip history, Pair & Pay for street hails | I2, I3 | Adv | must | M | MVP | Upfront fixed quote with honest "estimate vs fixed" label; Live Activities later |
| Autocab | Passenger booking | White-label passenger app with live ETAs, E-booker web booking with travel-policy rules | A1 | Adv | must | M | MVP | — |
| Jugnoo | Passenger booking | "Book Rides Instantly", "Schedule Booking", "Driver Tracking", push+SMS status alerts, post-ride feedback | J2 | Adv | must | S | MVP | — |
| Limo Anywhere | Passenger booking | Online reservation engine, round trips, flight info at booking, Passenger Web App (PWA) | L2, L4 | Adv | should | M | v1.1 | Not a limo product; keep pre-book simple |
| TaxiCaller | Passenger booking | Booking via website, mobile apps, dispatchers, WhatsApp, corporate tool | T1 | Adv | must | M | MVP | WhatsApp booking intake is a cheap LK differentiator (v1.1) |

### 1.2 Driver app & lifecycle

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Driver app | Document registration, back-to-back rides, pre-order acceptance, in-app wallet + payouts, gamification, idle-time algorithm | O3 | Adv | must | M | MVP | Compliance view already drives eligibility; add expiry reminders (v1.1) |
| iCabbi | Driver app | Bid Maps, turn-by-turn, Vias, CarPlay/Android Auto, Driver Fatigue lock-outs, Driver Panic, NFC payments, earnings dashboard, in-app update centre; separate "Driver Docs" self-service onboarding portal | I3, I1 | Adv | must | L | v1.1 (fatigue/CarPlay later) | Fatigue rules are a regulator-friendly feature for LK tuk fleets — cheap to add as shift-hour caps |
| Autocab | Driver app | "Driver Companion" iOS/Android: available bookings, GPS nav, card/account payments, text messaging with office, automated pricing, pre-allocated bookings | A1, A2 | Adv+3P | must | M | MVP | — |
| Yelowsoft | Driver app | "Driver Document Management", "Driver Profile and Registration", "Multi-Driver Shift Assignment", ride queue (next trip auto-starts), preferred destination | Y2, Y4 | Adv+3P | must | M | MVP | — |
| Jugnoo | Driver app | Real-time requests, availability toggle, nav, start/end trip, earnings summary | J2 | Adv | must | S | MVP | — |
| Limo Anywhere | Driver app | "DriverAnywhere": accept/view assignments, status updates, GPS, FlightStats on trip cards | L3 | Adv | should | M | v1.1 | — |

### 1.3 Dispatch automation & manual

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Dispatch | "Intelligent queue algorithm" auto-assignment, "Manual Mode", phone-order entry, instant vs pre-order filters, driver workload heat map, 100 km panel radius, PBX integration | O5, O3 | Adv | must | M | MVP | Document the offer-chain rules openly (per-tenant config), not a black box |
| TaxiCaller | Dispatch | Closest-car, zone+queue FIFO, or broadcast-and-bid; route replay (routes, speeds, distances) | T1, T4 | Adv+3P | must | M | MVP (closest/FIFO), v1.1 (bidding) | Zone queue FIFO is what taxi stands / airport ranks in LK need |
| iCabbi | Dispatch | Built on Google Fleet Engine, "over 1000 real-time configurations", traffic-aware allocation, dispatch networking (overflow to partner fleets), Move AI grouped/linked trips (NEMT) | I1 | Adv | must | L | MVP (core), later (networking) | Cross-tenant overflow is a natural multi-tenant feature for Taxios (later) |
| Autocab | Dispatch | Ghost consolidates app/E-booker/corporate/phone; "Follow-on booking" for soon-to-clear drivers; "The Wall" prioritises regulars; demand management, no-show reduction | A1 | Adv | must | M | MVP (follow-on = back-to-back) | — |
| Yelowsoft | Dispatch | Auto + manual, "Manual Driver Assignment", drag-and-drop "Unassigned Trip Allocation", real-time driver map, "Request Broadcasting" (Professional plan) | Y2, Y4 | Adv+3P | must | M | MVP | — |
| Jugnoo | Dispatch | "Automated Dispatch" nearest driver; dispatcher dashboard manage bookings/drivers/customers | J2 | Adv | must | S | MVP | — |

### 1.4 Fleet / vehicle / shift

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Fleet | Service types (classic, SUV, van…), driver plans in "My hub" | O2, O3 | Adv | must | S | MVP | Tuk / bike / car / van categories with LK-specific capacity rules |
| Yelowsoft | Fleet | "Multi-Driver Shift Assignment" (vehicle shared across drivers), document management, unlimited vehicles | Y2, Y1 | Adv | should | M | v1.1 | Vehicle-shift model (one tuk, two drivers) is common in LK |
| TaxiCaller | Fleet | Billing is per *activated* vehicle per month (vehicle logged into by a driver); dispatchers free | T1 | Adv | should | S | MVP (vehicle entity) | — |
| iCabbi | Fleet | Driver Docs portal automates onboarding and record management; Driver Fatigue | I1, I3 | Adv | should | M | v1.1 | — |
| Cabica | Fleet | Document-expiry alerts in Starter tier | C1 | 3P | should | S | v1.1 | — |

### 1.5 Fare / surge / pricing rules

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Pricing | Tariffs per service type, "Fixed fare zones", dynamic pricing "based on factors like increased demand" | O3, O4 | Adv | must | M | MVP | Versioned fare rules with effective dates (append-only `fare_rule_versions`) is stronger than any vendor's visible offer; expose the rule that produced a quote |
| Yelowsoft | Pricing | "Custom Pricing Engine" (distance+time), "Zone-Wise Pricing", "Surge Pricing" auto in peak/zones, corporate custom pricing, airport custom pricing | Y2 | Adv | must | M | MVP (base), v1.1 (surge) | Surge with explicit caps + passenger disclosure (LK regulator optics) |
| Autocab | Pricing | Tariffs + surge pricing set up in advance; rules for busy periods | A1 | Adv | must | M | v1.1 | — |
| TaxiCaller | Pricing | "Fare rules" (FAQ); zone-based; no surge mentioned | T1 | Adv | must | M | MVP | — |
| Jugnoo | Pricing | "Surge Pricing" in dispatcher + admin dashboard | J2 | Adv | should | M | v1.1 | — |
| Cabica | Pricing | Zone builder with tariffs + surge (Professional tier) | C1 | 3P | should | M | v1.1 | — |

### 1.6 Payments / receipts / refunds / reconciliation

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Payments | "Over 20 different payment gateways" (Bepaid, Checkout.com, Stripe named), PCI DSS, 3DS, customer wallet, cash/terminal; min payment $0.10 footnote | O4, O2, O1 | Adv | must | L | MVP (cash + receipts), v1.1 (card) | LK rails first: cash, PayHere/Payments.lk/bank, LankaQR; refunds as ledger events, never float |
| iCabbi | Payments | Card, Apple Pay, Google Pay, cash, account; digital receipts; tipping; no-show & cancellation fees; NFC in driver app; "Driver Pay" wallet with scheduled/instant payouts | I2, I3, I1 | Adv | must | L | MVP (receipts, cancellation fee), v1.1 (wallet payouts) | Cancellation-fee policy as a tenant setting with audit trail |
| Autocab | Payments | Card/account payments in Driver Companion; online pay in E-booker; monthly consolidated corporate invoicing | A1, A2 | Adv+3P | must | M | MVP (cash/account), v1.1 (card) | — |
| Yelowsoft | Payments | Wallet top-up/recharge, wallet transaction reports, "Settlement Adjustment" credits/debits, QuickBooks add-on ($49) | Y2, Y1 | Adv | should | M | v1.1 | — |
| Gazoop | Payments | Stripe listed ("0.30% + 20¢" or "No Charge" by plan) | G1 | Adv | should | M | v1.1 | — |
| Limo Anywhere | Payments | "Limo Anywhere Pay" (US-only, rates not published) | L1, L2 | Adv | nice | — | later | — |
| TaxiCaller | Payments | Integrated payment processing with transaction fees (amount not published) | T4 | 3P | should | M | v1.1 | — |

### 1.7 Scheduled / recurring / airport / intercity / corporate / multi-stop

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Scheduled/corporate/multi-stop | Pre-order, multiple drop-offs, corporate clients ("Partner with corporate clients", corporate account = centralised billing/reporting/usage policies) | O3, O9 | Adv | must | M | MVP (pre-order, corporate account), v1.1 (multi-stop) | Corporate = tenant sub-account with monthly statement; already natural in Taxios tenancy model |
| Yelowsoft | Scheduled/recurring/airport/corporate | Advance scheduling, daily/weekly scheduler (recurring), "Airport Transfer Booking" with flight details, corporate accounts with auto invoicing | Y2 | Adv | must | M | MVP (scheduled, corporate), v1.1 (recurring, airport) | BIA airport flat fares + flight-number capture; intercity fixed routes (Colombo–Kandy etc.) |
| iCabbi | Scheduled/airport/corporate | Pre-book ≤7 days, flight-aware airport bookings, personal/business account switch, Business Solution portal for corporate/government | I2, I1 | Adv | must | M | MVP | — |
| Autocab | Corporate | Corporate accounts portal, spend management, consolidated monthly invoicing, E-booker travel-policy rules & budgets | A1 | Adv | must | M | MVP (account), v1.1 (policy/budget) | — |
| Jugnoo | Scheduled | "Schedule Booking"; corporate/multi-stop not mentioned | J2 | Adv | must | S | MVP | — |
| Limo Anywhere | Airport | Real-time flight tracking (FlightStats) on driver trip cards; round trips | L3, L2 | Adv | should | M | v1.1 | — |
| Gazoop | Airport | Flight Tracking & Alerts add-on 20¢/booking | G1 | Adv | should | M | v1.1 | — |

### 1.8 Ride pooling

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| iCabbi | Pooling | Move AI "grouped and linked trips" route optimisation (positioned for NEMT, not consumer pooling) | I1 | Adv | nice | L | later | Skip at MVP; none of the dispatch SaaS vendors ship consumer pooling as a headline feature |
| Jugnoo | Pooling | Separate "Carpool & Sharing" solution listed; not part of taxi dispatch page | J2 | Adv | nice | L | later | — |
| Yelowsoft / Onde / TaxiCaller / Autocab / Limo Anywhere | Pooling | Not mentioned on the pages fetched | Y2, O3, T1, A1, L2 | — | nice | L | later | — |

### 1.9 Websites & booking widgets

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Web booking | "Web app" browser booking, no install; SMS verification code | O3, O4 | Adv | must | M | MVP | Taxios already serves a web booking surface; add embeddable widget (v1.1) |
| Yelowsoft | Web booking | "Web eBooker"; "mobile-responsive website … will be extra" (price not published) | Y2, Y1 | Adv | must | M | MVP | — |
| Autocab | Web booking | E-booker (book, track, manage, pay online); online booking website + social media service | A1, A3 | Adv+3P | must | M | MVP | — |
| Limo Anywhere | Web booking | ORES customisable booking widgets; website design templates + CMS | L4, L2 | Adv | should | M | v1.1 | — |
| Gazoop | Web booking | Free websites for eligible customers; premium hosting $10–14/mo | G1 | Adv | should | S | v1.1 | — |
| Cabica | Web booking | Branded booking website in Professional tier | C1 | 3P | should | M | v1.1 | — |

### 1.10 Branded apps & custom domains

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Branded apps | Branded iOS+Android customer & driver apps included in every plan; Onde uploads to stores, "creates a separate author's account for each new business partner"; ~4 weeks launch | O1, O8, O4 | Adv | must | L | MVP (one shared binary or per-tenant build — see §4) | Per-tenant build pipeline from one RN codebase + tenant-owned developer accounts |
| Yelowsoft | Branded apps | White Label Customer App +$249 setup +$99/mo; White Label Driver App same; PWA Customer App +$149/+$49 | Y1 | Adv | must | L | MVP | Price white-label as a flat add-on like Yelowsoft, but LKR-denominated |
| Gazoop | Branded apps | Dedicated Customer App $750 per store per app, "needs app store approval"; whitelabel iOS/Android "subject to set-up charge" | G1 | Adv | must | L | MVP | — |
| iCabbi | Branded apps | "White-Labeled App" (fleet name, logo, brand); passengers download from fleet's own store listing | I2 | Adv | must | L | MVP | — |
| Autocab | Branded apps | White-label passenger app | A1 | Adv | must | L | MVP | — |
| Limo Anywhere | Branded apps | Passenger App (PWA) $199 + $99/mo (free on Black); Company Aliases (DBA) add-on | L1 | Adv | should | M | v1.1 | Custom domain per tenant is cheap on Taxios' nginx vhost model |
| Cabica | Branded apps | Branded passenger app in Professional tier | C1 | 3P | must | L | MVP | — |

### 1.11 CRM / loyalty / promos / referrals

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Promos | "Discount and referral programs", "Reward regular customers", in-app marketing campaigns in My hub | O3, O5, O1 | Adv | should | M | v1.1 | Promo codes with integer-minor-unit caps; no float discounts |
| Yelowsoft | Promos | "Promo Code Campaign", "Referral Program", "Reward" points; referral only on Professional+ | Y2, Y4 | Adv+3P | should | M | v1.1 | — |
| iCabbi | Promos | Promo codes (loyalty rewards, refunds, seasonal), Map Notification Bar offers; no referral named | I2 | Adv | should | M | v1.1 | — |
| Jugnoo | Promos | Promo codes; email/SMS/WhatsApp marketing via Hippo | J2 | Adv | should | M | v1.1 | — |
| Limo Anywhere | CRM | "Lead Quote Close" CRM (1% of bookings, $300/mo min) | L1, L2 | Adv | nice | M | later | — |
| Cabica | Promos | Promo codes + loyalty in Professional tier | C1 | 3P | should | M | v1.1 | — |

### 1.12 Driver commissions / earnings / settlements

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Settlement | Driver in-app wallet, "transparent earnings and easy payouts"; Company API can "change driver/s balance/s" | O3, O7 | Adv | must | M | MVP (ledger), v1.1 (payouts) | Append-only driver ledger with daily cash-settlement statements (cash-heavy LK) |
| iCabbi | Settlement | "Driver Pay": scheduled or instant payouts via automated wallet; earnings dashboard | I1, I3 | Adv | must | L | v1.1 | — |
| Yelowsoft | Settlement | Per-trip subscription deducts commission; "Settlement Adjustment" wallet credits/debits | Y2 | Adv | must | M | MVP | — |
| Jugnoo | Settlement | "Earnings Summary" in driver app; commission model not described | J2 | Adv | must | S | MVP | — |
| Autocab | Settlement | "maximising driver earnings" claim; accounts package called "basic" by a reviewer | A1, A3 | Adv+3P | must | M | MVP | — |

### 1.13 Reports / BI

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Reports | "Performance reports", "reports and advanced analytics" in My hub, business performance metrics | O1, O2, O3 | Adv | must | M | MVP | CSV export + per-tenant scheduled email of daily cash/trips |
| Yelowsoft | Reports | Trip, revenue, driver performance, rider engagement, wallet transaction reports; "Real-Time Demand Heatmap" | Y2 | Adv | must | M | MVP | — |
| TaxiCaller | Reports | Admin-panel reports each with a Report ID fetchable on a schedule via API | T3 | Adv | must | M | MVP (reports), v1.1 (API pull) | — |
| Autocab | Reports | "Analytics" module; "extensive management reports" | A1 | Adv | must | M | MVP | — |
| Jugnoo | Reports | "Smart Analytics": detailed reports, live driver maps, high-demand areas | J2 | Adv | must | M | MVP | — |
| iCabbi | Reports | Not described on fetched pages | I1 | — | must | M | MVP | — |

### 1.14 Notifications (SMS / email / chat / voice)

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Notifications | In-app chat (driver/customer), real-time driver notifications, SMS verification, PBX call-centre integration | O3, O5, O4 | Adv | must | M | MVP (push+SMS), v1.1 (chat) | LK SMS gateways (Dialog/Mobitel/Hutch aggregators) as pluggable providers |
| iCabbi | Voice | "Voice" IVR automating inbound/outbound calls, "Voice+" recognition, "Voice AI" NLP + Smart Suggestions; push, chat, Message Centre | I1, I2, I3 | Adv | should | L | later | IVR is the big-fleet differentiator; not MVP for app-first LK operators |
| Autocab | Voice | "Phantom" automated phone system (claims 50%+ calls automated day one), call recording/monitoring, text messaging office↔driver | A1 | Adv | should | L | later | — |
| Yelowsoft | Notifications | "Scheduled Texting", "Push Notification Manager", "Scheduled Emailing", "In-App Chat Messaging" | Y2 | Adv | must | M | MVP | — |
| Jugnoo | Notifications | Push + SMS trip alerts; WhatsApp marketing | J2 | Adv | must | S | MVP | — |
| Gazoop | Notifications | SMS priced per country (US 3¢), Android SMS Gateway app, call-centre tech per minute, call answering service per call | G1 | Adv | should | M | v1.1 | Usage-billed SMS per tenant |
| TaxiCaller | Voice | Caller ID + VoIP integration | T1 | Adv | nice | M | later | — |

### 1.15 APIs / webhooks / integrations

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | API | REST + WebSocket; Operator API (create/receive orders, summaries, track updates, search services/vehicles), Company API (driver balances), Wallet API (top-up methods); free sandbox by emailing support; "webhook testing" listed, no public webhook docs; SDKs "upcoming" | O6, O7, O5 | Adv | should | M | v1.1 | Public OpenAPI spec + signed webhooks from day one (none of these vendors publish webhook docs) |
| TaxiCaller | API | Partners check availability, get pricing, push confirmed bookings; report pull by Report ID; API keys | T3 | Adv | should | M | v1.1 | — |
| iCabbi | API | APIs + webhooks exist but keys issued only after marketplace partner agreement; Google Fleet Engine; "The Exchange" app/dispatch networking | I4, I1 | Adv | should | M | v1.1 | Open self-serve API keys vs iCabbi's gated model |
| Jugnoo | API | "open & free of cost developer APIs"; webhooks via Jungleworks open API docs | J2, J3 | Adv | should | M | v1.1 | — |
| Autocab | Integrations | iGo marketplace (Uber and other demand sources); no public API docs found | A1 | Adv | nice | L | later | — |
| Yelowsoft | Integrations | QuickBooks add-on; API not mentioned on features page | Y1, Y2 | Adv | nice | S | later | — |
| Limo Anywhere | Integrations | LA Net farm-in/out $0.25–0.75/trip; Partner Connectivity (GRIDD) per-trip fee | L1 | Adv | nice | L | later | — |

### 1.16 Tenant mgmt / licensing / subscriptions / usage billing

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Licensing | One-time fee + $99–259/mo + % revenue share of trip cost; unlimited drivers; Onde.Light no setup fee | O1 | Adv | must | M | MVP (plan + status), v1.1 (metering) | Per-tenant plan, suspended/closed enforcement already in `tenancy`; add usage counters |
| Yelowsoft | Licensing | Flat monthly with included trips + per-extra-trip fee; setup fee; 1/3/unlimited languages by plan | Y1 | Adv | must | M | MVP | Trip-bucket model is the most transparent for LK operators |
| TaxiCaller | Licensing | Per *activated* vehicle per month, resets monthly, prorated mid-month; >20 vehicles custom | T1, T2 | Adv | must | M | v1.1 | — |
| iCabbi | Licensing | Per-driver licence fee, monthly PAYG, 0–12-month terms; fleets ≥25 vehicles (US/CA) | I1 | Adv | must | M | v1.1 | — |
| Gazoop | Licensing | Per driver weekly/monthly/yearly; ≥1 active driver required; managed support tiers by ticket count | G1 | Adv | must | M | v1.1 | — |
| Limo Anywhere | Licensing | Month-to-month, upfront licence fee, billed in arrears, 30-day cancel notice, $25 reactivation | L1 | Adv | should | S | v1.1 | — |

### 1.17 Enterprise security / deployment / reliability

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Reliability | "2 million orders every day with a 99,98 % uptime record"; up to 9M orders/month; PCI DSS; 3DS; partner data remains partner's property | O4, O2 | Adv | must | M | MVP (status page), v1.1 (SLA) | Publish RLS/tenant-isolation tests as a security claim; status page |
| iCabbi | Reliability | "99.999% uptime SLA", Google Fleet Engine | I1 | Adv | must | L | later | — |
| Autocab | Hosting | Cloud-based (Ghost Cloud); permission-controlled module access | A1, A3 | Adv+3P | must | M | MVP (RBAC) | — |
| TaxiCaller | Compliance | Certified by Revenu Québec (fiscal compliance) | T1, T2 | Adv | nice | — | later | Local analogue: LK e-invoice/VAT readiness |

### 1.18 Accessibility / localization / international payments

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| iCabbi | Accessibility | Voice Commands, Text to Speech, font size, colour contrast, dark mode; FAQ claims WCAG compliance | I2 | Adv | should | M | v1.1 | WCAG AA on web console + Sinhala/Tamil fonts tested |
| Yelowsoft | Localization | Languages per plan: 1 / 3 / unlimited | Y1 | Adv | must | S | MVP (si/ta/en) | Trilingual at every plan, not tiered |
| Onde | Localization | 73 countries (2024); revenue share computed in local currency converted to USD; 20+ gateways | O4, O1 | Adv | should | M | v1.1 | LKR-native billing |
| Gazoop | Localization | SMS priced per country; Google Maps bring-your-own key | G1 | Adv | nice | S | v1.1 | — |

### 1.19 AI / fraud

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| iCabbi | AI | "AI embedded across the platform": Move AI routing, Voice AI NLP, Smart Suggestions, "AI-first" passenger app, Probability ETAs powered by Google | I1, I2 | Adv | nice | L | later | Rule-based fraud (GPS spoof, cancel abuse, duplicate accounts) before any "AI" |
| Onde | Fraud | 3DS checks, PCI DSS, no card data stored | O2 | Adv | should | S | v1.1 (with card rail) | — |
| Yelowsoft | AI | Demand heatmap only; no fraud/AI feature named | Y2 | Adv | nice | — | later | — |
| Jugnoo | AI | None named | J2 | — | nice | — | later | — |

### 1.20 Support tools / audit / compliance

| Competitor | Capability | Verified capability | Evidence | Verif. | Imp. | Cx | Prio | Differentiation |
|---|---|---|---|---|---|---|---|---|
| Onde | Support | 24/7 technical support, customer success team, free app updates, "Leave comments for operators and drivers", order event tracking | O1, O5 | Adv | must | S | MVP | Append-only `audit_events` + `booking_events` exposed as a per-booking timeline in console |
| TaxiCaller | Support | Free 24/7 chat/email/phone; route replay for disputes | T1 | Adv | must | S | MVP | — |
| Gazoop | Support | Paid managed support tiers ($69 / $299 / $1,199 per month by ticket count) | G1 | Adv | nice | S | later | — |
| Yelowsoft | Support | Capterra reviewer reported ride allocation, payments, driver visibility, reporting "malfunctioned or performed unpredictably"; 4.8/5 from 11 reviews | Y4 | 3P | — | — | — | Reliability is a sellable differentiator against low-cost vendors |
| Autocab | Support | Capterra 2.5/5 (4 reviews), support rated 1.0/5 | A3 | 3P | — | — | — | — |

## 2. Published pricing (all accessed 2026-10-11)

| Competitor | Model | Base / subscription | Per-booking / usage | Setup / one-time | White-label app fees | Source |
|---|---|---|---|---|---|---|
| Onde | One-time + monthly + revenue share | $99/mo (Starting up, Professional, Expert); $259/mo (Super App) | "% of trip cost" — exact % **not published**; "min payment $0.10" footnote | $4,500 / $8,500 / $23,000 / $6,500 | Branded iOS+Android driver & customer apps included; Onde.Light "no initial payment" (details not published) | O1 |
| Yelowsoft | Flat monthly + trip buckets | Starter $99 (500 trips) / Professional $270 (1,500) / Growth $750 (5,000) / Enterprise not published | $0.20 / $0.18 / $0.15 per extra trip | $999 / $1,499 / $1,999 | WL Customer App +$249 setup +$99/mo; WL Driver App +$249 +$99/mo; PWA +$149 +$49/mo; QuickBooks $49; website "extra" (not published) | Y1 (Capterra shows older $99/$250/$999 tiers — Y4) |
| TaxiCaller | Per activated vehicle | CAD 29 per vehicle/month (Québec page); aggregators list ~$28 | Payment-processing transaction fees — amount **not published** | "no setup costs"; Québec offer CAD 350 per vehicle covers year 1 | Rebranded passenger app "may incur a one-time development fee" — amount **not published** | T1, T2, T4 |
| iCabbi | Per-driver licence | **Not published** — per-driver monthly PAYG, tailored quotes, 0–12-month terms | — | Not published | Not published | I1 |
| Autocab | Custom | **Not published** ("competitive price"); one directory lists "$2 per vehicle per week" — unverified | — | Not published | Not published | A1, A3 |
| Jugnoo | Ride buckets + maintenance | Base price **not published**; Startup 24,000 rides, Growth 60,000 rides; Enterprise via sales | $0.10 / ₹6 (Startup), $0.075 / ₹4.5 (Growth) per extra ride | Annual maintenance from year 2: $200/mo or ₹12,500/mo | White-label apps included; 14-day trial; Growth adds complimentary SMS + gateways | J1 |
| Limo Anywhere | Monthly tiers by trip volume | Core $99–1,210 / Plus $249–3,399 / Black $549–5,999 per month (150–15,000 trips) | Overage $0.25–0.50/trip by plan; LA Net $0.25–0.75/trip; Shuttle $1 + 2.5%/ticket | $299 / $599 / $899 licence; $25 reactivation | Passenger PWA $199 + $99/mo (free on Black); DBA aliases $59 + $59/mo | L1 |
| Gazoop | Per driver | Standard $14 / Premium $18 / Ultimate $20 per driver/month (or $4/$5/$6 weekly) | Flight tracking 20¢, external bookings 20–30¢, managed Maps 15¢ per booking; SMS per country | None listed | Dedicated Customer App $750 per store per app; Ultimate whitelabel "subject to set-up charge" | G1 |
| Cabica | Per operator | Starter £30/mo (≤10 drivers, +£3/extra driver) / Professional £120/mo (≤50 drivers) / Enterprise custom | — | Not published | White-label included (scope varies by tier) — **third-party listing only** | C1 |

Takeaways: (a) nobody publishes a per-driver price below ~$14/driver/month except via custom quotes; (b) the two
transparent models are *trip buckets + overage* (Yelowsoft, Jugnoo, Limo Anywhere) and *per active vehicle/driver*
(TaxiCaller, Gazoop, iCabbi); (c) white-label apps are consistently a paid add-on ($150–$750 setup + ~$99/mo) unless
bundled into a $4.5k+ one-time fee (Onde); (d) revenue-share is used only by Onde.

## 3. Minimum viable commercial product vs long-term vision

**Day-one expectations (every competitor above offers these; an operator will not sign without them):**
1. Passenger app + web booking: instant and scheduled (pre-order) bookings, fare estimate before booking, live driver tracking, trip history, rating.
2. Driver app: online/offline, offer accept/decline, navigation hand-off, trip start/end, earnings summary, document upload.
3. Dispatch console: auto-assign (nearest, with a zone/FIFO option), manual assign/reassign, phone-order entry, live map, booking timeline.
4. Fleet: vehicle categories/service types, driver & vehicle records with document expiry, driver eligibility gate.
5. Fares: per-category distance+time tariff, fixed-fare zones (airport/intercity), cancellation/no-show fee setting; surge can wait but the data model must allow it.
6. Payments: cash with driver cash-collected ledger, receipts (email/SMS/PDF), corporate account "bill monthly".
7. Driver commissions/settlements: per-trip commission, driver balance ledger, daily/weekly statement.
8. Corporate accounts: business profile, monthly consolidated invoice, basic spend report.
9. Reports: trips, revenue, driver performance, cancellations; CSV export.
10. Notifications: push + SMS for booking lifecycle; SMS OTP login.
11. White-label: tenant branding (name, logo, colours) in apps and web; app listing under the tenant's own store accounts (see §4).
12. Tenant/licensing: plan, status (active/suspended/closed), 24/7-ish support channel, audit trail of admin actions.

**v1.1 (every serious competitor has it, operators ask within months):** surge/dynamic pricing with caps, promo codes + referral, multi-stop, recurring rides, airport flight-aware pickup, card payments via local gateway + refunds, driver wallet payouts, public API + signed webhooks, embeddable booking widget, custom domain per tenant, multi-language per tenant (si/ta/en), demand heatmap, usage-metered billing (trips/SMS).

**Long-term vision (differentiators only big vendors ship):** IVR/voice booking and Voice AI (iCabbi, Autocab), cross-fleet overflow networking (iCabbi Exchange, Autocab iGo, LA Net), consumer ride pooling, AI route grouping (NEMT), driver fatigue lock-outs, CarPlay/Android Auto, 99.999% SLA tier, super-app delivery vertical (Onde, Jugnoo), accounting integrations (QuickBooks), accessibility suite (voice commands, TTS), demand-generation services (ASO, install campaigns — Onde sells these as part of the one-time fee).

## 4. Apple App Store 4.2.6 "template app" rule

- Official text (Apple App Review Guidelines, 4.2.6 Minimum Functionality, AP1, accessed 2026-10-11): apps "created from a
  commercialized template or app generation service will be rejected unless they are submitted directly by the provider of
  the app's content. These services should not submit apps on behalf of their clients…". Alternative Apple explicitly
  allows: "a single binary to host all client content in an aggregated or 'picker' model". Guideline 4.3(a) Spam additionally
  forbids "multiple Bundle IDs of the same app".
- **Answer: yes.** Per 4.2.6 the client (the operator — "provider of the app's content") must submit under their own
  Apple Developer account; the SaaS vendor must not publish on the client's behalf from the vendor's account. An Apple
  dev-forum thread (AP2) shows that even client-account submissions get 4.3(a) rejections when binaries/metadata are near
  identical — so per-tenant customisation (branding, copy, screenshots, support URLs) matters as much as account ownership.
- How competitors describe it: **Onde** says it handles "the whole development cycle, testing, uploading the apps to Google
  Play and App Store", that it "creates a separate author's account for each new business partner", and that the client
  need not "read or understand any app store guidelines" (O8). It does not state whose Apple Developer Program membership
  that account belongs to — treat as *advertised, ownership unclear*. **Yelowsoft** publishes white-label app fees (Y1) but
  no page found describes the developer-account arrangement (searched 2026-10-11; not published). **Gazoop** states each
  dedicated customer app "needs app store approval" and is billed per store (G1). **iCabbi** says passengers download from
  the fleet's own store listing (I2).
- Implication for Taxios: ship one React Native codebase with a per-tenant build config; the operator enrols their own Apple
  Developer Program + Google Play Console, grants Taxios App Manager/admin access, and Taxios automates the upload (EAS or
  fastlane). Keep an Apple-sanctioned fallback: a single "Taxios" picker-model binary for tenants who cannot or will not
  enrol (Onde.Light-style free tier). Budget for the enrolment delay (D-U-N-S, 2FA Apple ID) in onboarding time — Onde's
  "around 4 weeks" (O4) is the market expectation.

## 5. Sri Lanka market context (PickMe, Uber LK) — accessed 2026-10-11

- PickMe: ~65–70% ride-hailing share per First Capital Research initiation (Dec 2025, P1 — analyst estimate, no formal
  share data); >100k drivers, ~6.4M trips/month per HNB Stockbrokers (Sep 2025, via search); Business Portal used by
  "over 950 corporates" with 24-h corporate line and credit period (P3, undated company claim); BIA airport rank deal
  (Rs. 700/taxi to BIA) and in-app airport pickup booking (P2, 2023); cash + cashless, live tracking, fare estimates,
  tuk/car/bike + food + logistics super-app. Surge pricing is in use (a travel directory claims rivals are 20–40% cheaper at
  peaks — anecdotal, unverified).
- Uber LK: scheduled rides 15 min–30 days ahead (U1; originally uberGO/uberX, dated), UberTUK, "Tuk Rentals" hourly
  multi-stop package (U2, pandemic era), expansion to Kegalle with cars/tuks/bikes on 24 Apr 2026 (U3). No LK-specific
  Uber for Business material found.
- Meaning for Taxios tenants (regional operators, hotel/airport fleets, corporate shuttles): the passenger baseline is set by
  PickMe — cash default, tuk-first categories, airport flat fares, corporate credit accounts, si/ta/en. A white-label
  operator competes on *dedicated supply + fixed corporate/airport pricing*, not on surge or super-app breadth.
