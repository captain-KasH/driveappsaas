# Sri Lanka market & local integration landscape — Taxios

Research track for fitting an Uber-grade UX to Sri Lankan operators (tuk/mini/car/van, cash-first).
All sources accessed **2026-10-12**. Legend: **V** = verified against a primary/official page or
reputable local press; **UV** = unverified (secondary/aggregator/anecdote, or primary page could not
be fetched). Money is LKR unless stated. `[Sn]` points at the sources table (§9).

**TL;DR — what most changes our design**
1. Cash is the default and stays so; commission on cash trips is collected from a **driver prepaid
   credit** (PickMe-style daily/per-trip deduction topped up via eZ Cash/bank), not from the rider.
2. Regulated Western-Province tuk meter tariff (**Rs 110 first km, Rs 90/km after, Mar 2026**) is the
   anchor riders compare app fares against; the fare engine must express "first-km + per-km" + night 15 %.
3. The NTC (Amendment) Act No. 8 of 2025 brings tuks and ride-hailing under the NTC; an **app-mandate
   for all three-wheelers** is being drafted — operators will need exportable driver/vehicle registers.
4. PDPA core obligations (Parts I & III: lawful processing, DPO, breach notice, DPIA, cross-border)
   start **1 Jan 2027**; data residency questions apply to any non-LK hosting.
5. SMS OTP is cheap locally (Rs 0.40–0.90) but needs a registered 11-char sender ID per operator and
   a Business Registration; Twilio is ~Rs 120/SMS and one-way. WhatsApp auth is ~Rs 0.6–2.
6. OSM `name` tags are romanised and `name:si`/`name:ta` are sparse; landmark/junction search needs a
   tenant-curated POI layer, not just Nominatim/Google.
7. Android ≈ 79 % with a long tail of Android 11–13 low-end (Honor/ZTE/Samsung A0x) phones and
   ~20 Mbps median mobile — the apps need a ≤ 30 MB, offline-tolerant, Sinhala/Tamil-shaped UI.
8. PIN-to-start exists at Uber LK (2020) but not visibly at PickMe; make it an opt-in tenant policy.

---

## 1. Incumbents: apps & features (as of 2026)

| Platform | Categories / products | Booking modes | Payments | Safety / trust | Languages | Driver model & tooling | Status |
|---|---|---|---|---|---|---|---|
| **PickMe** (Digital Mobility Solutions Lanka PLC, listed) [S1–S9] | Rider app: Tuk, Nano/Flex, Mini, Car/Sedan, Minivan, Van; Bike; Truck (goods); Flash (parcel), Food, Market | On-demand, advance booking, "book for the day" (day hire), promo codes, airport pickups | Cash, card; real-time cashless settlement to drivers via HNB (2017); PhonePe UPI QR for Indian tourists (2024); **LankaQR via Dialog Pay** on Rides/Flash/Trucks (Jul 2026) | KYC-verified drivers, live tracking with plate; ETA/trip sharing; in-app emergency (SOS) button + 24/7 safety hotline; no public ride-PIN (UV) | EN, SI, TA (+ Nepali in store listing) | Tuk commission **12 %**, falls to **7 %** above Rs 8,000/day (Mar 2024, V); drivers pay a **daily "subscription"** via eZ Cash app or 20,000 outlets (2019, V); "PickMe Credit" top-up via People's Bank documented Mar 2025 (PDF unreadable, UV); driver app: trip scanner (radar), directional travel, advanced matching, multi-category per vehicle, real-time withdrawal, 1–4 GB data/month; peak charge 15–25 %, night 15 % after 22:00 on tuks (2016, UV today); corporate accounts with credit period + auto log sheets; BIA designated bay (2023) | 165 k driver/delivery earners monthly, 1.9 M consumers (Jul 2026); FY25/26 revenue Rs 8.7 bn, GTV Rs 84.5 bn |
| **Uber Sri Lanka** (Uber Lanka (Pvt) Ltd since Jul 2026) [S10–S16] | Auto/Tuk, Zip, Moto (bike), Go/X cars, XL, Intercity, Tuk Rentals (hourly: Rs 399/1 h or 10 km, 2020), CNG, Teens (Jul 2025), **Black** (Oct 2026), Reserve (schedule) | On-demand, Reserve, hourly | Cash + card (UberTUK launch), Uber One | **PIN verification**, RideCheck, on-trip safety helpline (2020); emergency button; plate-mismatch reporting; "cancel if unsafe" | EN, SI (city pages in Sinhala) | 25 % commission reported by protesting drivers (2017, UV today); penalties after 5 rejections (2017) | 5 districts / 40 cities (2024); Kegalle Apr 2026, Ratnapura Jun 2026 |
| **Kangaroo Cabs** (2588588) [S17–S18] | Tuk, Budget mini (A/C), Sedan, Semi-luxury, Minivan/Van | App, web, phone hotline; now or advance | Cash, card, promo | Driver photo/plate, in-app chat, SMS updates, e-mail receipt | EN | **Fixed Rs 300/day** driver fee instead of commission (2017, UV today); BIA pre-paid counter | Legacy call-centre fleet with app |
| **HelaGo** (Helakuru super-app; Jan 2026) [S19–S20] | Tuk/car (not itemised) | **Driver fare offers (bidding)**, rider picks; instant + "Auto Booking" | Cash + digital | not described | SI-first super-app | Driver onboarding via David Pieris Group dealer network; 300+ drivers Colombo 01–07 pilot | Western Province rollout from 6 Jan 2026 |
| **YOGO** (DPMC E-Drive, 2023) [S21] | Electric tuk | Street hail, app, hotline | Flat **Rs 65/km** (2023) | — | — | Operator-owned EV fleet | Niche |
| **Casons, Budget Taxi, meter-taxi fleets** | Car/van by phone | Hotline dispatch, flat/negotiated quotes, airport & outstation hires | Cash; corporate invoicing | — | SI/EN voice | Dispatcher radios/calls driver; driver calls rider to confirm "near X" | No verified 2025 feature data (UV) |

**Observed expectations & pain points (riders / drivers)** [S4, S22–S25]

| Rider expectation / pain | Evidence | Taxios implication |
|---|---|---|
| Upfront fare that holds; final fare > quote is the top complaint (driver doesn't end trip, route extended) | Trustpilot/LinkedIn accounts 2025–26 (UV, anecdotal) | Fixed quote by default; route-deviation detection; rider-visible trip end |
| Drivers prefer cash, sometimes demand more than app fare; tourists overcharged | Reddit/Trustpilot (UV); Ella/Galle conflicts (V) | Receipt shows agreed fare in large type; "report overcharge" flow; cash-only drivers flagged |
| Drivers call to ask destination before accepting, then cancel | Industry norm; not in press (UV) | Show destination + fare to driver at offer time (already in our offer model); cancel penalties |
| Cancellation after long wait; support unreachable | Trustpilot 2026 (UV) | SLA timers, auto-reassign, in-app support ticket |
| Change for cash: small notes scarce; fares read as "Rs 270" | Anecdotal (UV) | Round quotes to Rs 10; show "pay Rs 270" and optional tip |
| Driver side: commission with no say; fuel ≈ 40 % of fare; owner takes up to 50 % of earnings | RoW 2024 (V), ReadMe 2017 (V) | Transparent per-trip commission line; owner/driver split field on vehicle |
| Driver side: hostility from stand-based tuk drivers in tourist towns | RoW 2024, AsiaNews Jan 2026 (V) | Operator-scoped zones; stand/queue mode for existing tuk societies (§7) |

## 2. Payments in Sri Lanka

| Provider | Type | Methods | Fees / terms | Tokenisation / pre-auth / recurring | Payouts | Settlement | Notes |
|---|---|---|---|---|---|---|---|
| **PayHere** [S26–S28] | Aggregator IPG | Visa/MC/Amex, Genie, FriMi, eZ Cash, mCash, bank apps (UV list) | 2.9–3.9 % by plan (2019, UV); plugin listing cites 3.4 % + Rs 50 local / 3.9 % + Rs 50 intl (UV); no setup fee | **Preapproval API** (card-on-file), Recurring/Subscription API, Checkout API, mobile SDKs (V via SDK wrappers); no card hold/auth-capture publicly (UV) | None (collection only) | Daily payouts, ~T+2 (UV) | Needs business registration; best-documented LK gateway |
| **WebXPay** [S29] | IPG + instalments | Cards, XSPLIT instalments (min Rs 10 k; Rs 25 k with ComBank) | Package + installation fee; rates "published separately" | — | — | — | Rate card only on request |
| **Payable** (Sampath/ PayCorp family) | Bank IPG | Cards | ~4 % class (2019, UV) | — | — | — | No public 2025 data (UV) |
| Bank IPGs: Sampath PayCorp, Commercial Bank, HNB, Seylan [S29] | Direct acquiring | Cards | 3–4 % (2019, UV) | Bank-specific | — | T+1/T+2 (UV) | Direct contracts; needs company + turnover history |
| **Genie / Dialog Pay** (Dialog Finance) [S30–S31] | Wallet + gateway | Cards, LankaQR, payment links, Shopify checkout | Not published | — | — | — | Dialog Pay (2026) is what PickMe uses for LankaQR; API via Dialog business team (UV) |
| **FriMi** (Nations Trust) | Bank wallet | LankaQR, P2P | — | — | — | — | Consumer-side only for us |
| **eZ Cash** (Dialog) [S32] | Telco wallet | Agent cash-in at 20,000+ outlets, merchant payments | — | — | — | — | **Proven driver top-up channel** (PickMe 2019) |
| **mCash** (Mobitel) | Telco wallet | P2P, merchant QR; Rs 50 k/month cap (UV) | — | — | — | — | Secondary top-up channel |
| **LankaQR** (CBSL/LankaPay) [S33–S34] | National QR | Any LankaQR wallet/bank app | MDR cap 0.5 % (2020, V); merchant fee waived < Rs 5,000 from Apr 2026 (UV, blog) | n/a | n/a | Instant to merchant bank (via CEFTS) | Dynamic QR per trip is feasible; PickMe did it |
| **Koko** BNPL [S35] | 3-instalment BNPL | Via JustPay (bank account) | Merchant commission not public | — | — | — | Irrelevant for rides; maybe hires/corporate |
| **CEFTS / SLIPS / JustPay** (LankaPay) [S36] | Rails | CEFTS real-time 24/7 ≤ Rs 5 M; SLIPS 3 cycles/day, same-day if before 13:30; JustPay consumer debit | Bank-set; LankaPay caps (page not fetched, UV) | — | **Bank bulk-payout/H2H APIs** (e.g. StanChart API with PickMe: 24/7 collections, V) | — | Driver payouts = bank CEFTS file/API, not a gateway |

**Realistic commission collection for cash trips (operators ≤ 500 drivers)**

| Channel | How | Cost / friction | Verdict |
|---|---|---|---|
| Driver prepaid credit, deducted per trip | Operator sells credit; driver tops up by bank deposit/CEFTS to operator account (reference = driver code), eZ Cash/mCash merchant, LankaQR | Reconciliation by reference; eZ Cash merchant fee (UV) | **Default** (PickMe model; already matches our ledger) |
| Weekly/daily settlement at office | Driver pays cash at operator desk | Manual, leakage | Keep as "cash top-up by staff" with audit |
| Card trips net of commission | Gateway settles to operator; driver balance credited | Gateway 3–4 % | Card share small; pay driver via CEFTS batch |
| Fixed daily fee (Kangaroo) | Flat Rs 300/day class subscription | Simple | Offer as plan type per tenant |

## 3. Messaging (SMS / WhatsApp / push)

| Gateway | Price (LKR/SMS) | Sender ID (mask) | OTP / rules | Notes |
|---|---|---|---|---|
| **Dialog eSMS** [S37] | Business rate negotiated; consumer ref 0.50 | Alphanumeric ≤ 11 chars, max 5 masks/company, BR certificate, 3–7 business days | Supported; API via business team | Largest operator; API docs not public (UV) |
| **Mobitel eSMS** [S37] | Negotiated | As above | Some IDs restricted to OTP/OTT/transactional | — |
| **Hutch** | From 0.25 promo (UV) | — | — | Smallest network |
| **Notify.lk** [S38] | Pay&Go Rs 7,500 one-time then **0.60**; Lite Rs 690/mo (450 free, 1.10); Startup Rs 1,500 (0.90); SME Rs 3,600 (0.80); Pro 0.70; Heavy 0.60; Ultra Rs 18,000/mo (0.40) | Free IDs per plan (1–5); **no BR required**; form by e-mail | Transactional only in free quota; promo +Rs 0.25 tax; promo blackout 20:00–08:00 (TRCSL) | All networks, flat rate; WhatsApp OTP product |
| **Text.lk** [S39] | **0.64** flat | ≤ 11 chars; authorisation letter on letterhead; hours–3 days | Demo ID cannot send OTP; OAuth2 API; WhatsApp API | Cheapest advertised |
| **ShoutOUT** | Not public | — | OTP API | Marketing-oriented |
| **TextIt.biz** | Not found (UV) | — | — | — |
| **Twilio** [S37] | **≈ Rs 120 ($0.396)** | Pre-registration 3–4 weeks | No two-way SMS to LK; 10 MPS | Fallback only |
| **Infobip / Plivo / Sinch** | $0.137 / $0.216 / quote (UV) | LOA + 3–4 weeks | — | Enterprise fallback |
| **WhatsApp Business (Meta)** [S40–S41] | Auth template ≈ 0.60–2.00 (UV); utility similar; Meta re-priced LK on 1 Oct 2026 incl. "authentication-international" | Business verification; template approval | Twilio adds $0.005/msg; auth templates "not available" in some regions (check LK) | Good for receipts/ETA; not for first OTP |

Encoding: Sinhala/Tamil SMS are UCS-2 → **70 chars/segment** (160 GSM-7 for English). Push (FCM/APNs) is unrestricted; Expo push works in LK.

## 4. Maps & geo

**Google Maps Platform (list price per 1,000, 0–100 K tier; monthly free per SKU)** [S42, V]

| SKU | Free/mo | $/1 K | Use in Taxios |
|---|---|---|---|
| Dynamic Maps | 10 K | 7.00 | Rider map (prefer tiles/Mapbox/MapLibre) |
| 2D Map Tiles | 100 K | 0.60 | — |
| Autocomplete Requests | 10 K | 2.83 | Session usage free; Place Details Essentials 10 K free / 5.00 |
| Place Details Pro / Text Search Pro / Nearby Pro | 5 K | 17 / 32 / 32 | Avoid Pro fields |
| Geocoding | 10 K | 5.00 | Reverse-geocode pickup |
| Compute Routes Essentials / Pro / Enterprise | 10 K / 5 K / 1 K | 5 / 10 / 15 | Pro = traffic-aware ETA |
| Route Matrix Essentials / Pro | 10 K / 5 K | 5 / 10 | Dispatch ETA to N drivers (per element) |
| Roads Nearest Road / Route Traveled | 5 K | 10 | Snap-to-road |
| Navigation Request (Nav SDK) | 1 K | 25 | Driver turn-by-turn — expensive |
| Address Validation Pro | 5 K | 17 | Not useful in LK |

Volume discounts start at 20 % above 100 K/mo. Subscriptions: Starter $100 (50 K calls), Essentials $275, Pro $1,200.

| Alternative | Free tier (2026, UV secondary) | Price after | LK fit |
|---|---|---|---|
| **Mapbox** [S43] | 50 K web loads; 25 K mobile MAU; **100 K Directions/Matrix/Map-Matching req**; 2,500 Search Box sessions | $2/1 K routing → $1.20; $5/1 K loads; $11.50/1 K search sessions | Good routing economics; OSM-based data in LK |
| **HERE** [S43] | ~30 K transactions | ~$0.40–1/1 K | Decent road data; traffic weaker in LK (UV) |
| **OSM + self-hosted** [S44–S46] | Free | Server only | Geofabrik `asia/sri-lanka` extract; **OSRM** (`match` = map-matching), **Valhalla** (`trace_route`, time-aware but no live traffic), **GraphHopper**; Nominatim/Photon for geocoding |

**OSM quality & naming (V from OSM wiki/community):** `name` is conventionally **romanised English**; Sinhala/Tamil live in `name:si` / `name:ta` and are sparse and inconsistent; Nominatim returns `name` unless `accept-language` finds `name:en`. ~64 % of buildings mapped nationally (HDX AI estimate, 2025). Colombo roads/one-ways are usable for routing; outstation POIs thin. No live traffic from OSM — Colombo peak ETAs need Google Routes Pro or a learned per-zone/hour speed table from our own trip GPS.

**Addresses in practice:** "No. 12/3, Temple Road, Nugegoda" + "near X junction / opposite Y temple / Colombo 05". 5-digit postcodes exist (Colombo 00500 etc.) but are rarely spoken; Colombo "01–15" district numbers are. SL Post has **no What3words adoption** (V: nothing found). Junctions ("handiya"), temples, kovils, schools, bus halts and shops are the landmark vocabulary → a tenant-curated POI table beats any global geocoder.

## 5. Regulation & compliance

| Topic | Position (2026) | Source |
|---|---|---|
| **PDPA No. 9 of 2022** (amended No. 22 of 2025) | Part V (Authority) in force 17 Jul 2023; Parts VI, VIII–X 1 Dec 2023; Mar-2025 commencement repealed (Gazette 2427/34); **Gazette 2498/16 (22 Jul 2026): ss. 2–3, Part I, Part III in force 1 Jan 2027** = lawful/transparent processing, controller/processor contracts, **DPO designation, breach notification, DPIA, cross-border transfer rules**. Part II (data-subject rights) and Part VII (penalties ≤ Rs 10 M, doubling on repeat) await a further order. Cross-border: adequacy or consent/contract exceptions. Draft rules on breach notice/DPIA/rights published (titles only). | [S47–S49] V |
| **NTC (Amendment) Act No. 8 of 2025** | NTC now regulates "paratransport": three-wheelers, school vans, office transport; ride-hailing falls under the taxi segment. Oct 2025: no regulations yet, draft received; Feb 2026: four regulations in final stages, national registration drive for > 1 M tuks (operational stickers), **plan to require all three-wheelers to operate via apps**, standard fare structures, vehicle standards; coordination with Ministry of Digital Economy on a unified digital gateway. NTC hotline 1955 / WhatsApp 071 259 5555. | [S50–S51] V |
| **Western Province tuk tariff** | WP Road Passenger Transport Authority: **first km Rs 110, each further km Rs 90 (from 23 Mar 2026)**; previous Rs 100/90 (Jul 2024) with **+15 % 22:00–05:00**; union proposal Rs 130 first km (May 2026, pending). Meter compulsory under the 2002 WP statute / 2021 rules, widely unenforced; non-metered must display per-km charges. Registration/licensing of passenger tuks under 2002 regulations resumed 1 Oct 2025; permit on windscreen; quality-control report at registration. | [S52–S56] V |
| **Tax** | VAT **18 %** (from 1 Jan 2024); financial services VAT 20.5 % from 1 Jul 2026; **SSCL 2.5 %** on turnover above Rs 60 M/yr (Rs 15 M/quarter). VAT (Amendment) Act No. 4 of 2025: non-resident digital platforms must register and charge 18 % from 1 Jul 2026 above the same thresholds. Platform commission is a taxable service; treat commission + subscription as VAT-inclusive lines. | [S57–S58] V (platform treatment UV) |
| **Emergency numbers** | **119** police (misuse warning Dec 2025), **1990** Suwa Seriya ambulance (toll-free), **110** fire & rescue (some sources 111), 112 police mobile hotline, 1955 NTC transport complaints. | [S59–S60] V |
| **Driver / vehicle papers** | Driving licence with class (B1 three-wheeler; B car; C1/C van), NIC, vehicle registration (CR), **revenue licence**, insurance certificate valid for hire/passenger carriage, WP passenger service permit (where enforced), tuk society recommendation (Southern Province rule). Foreigners cannot be licensed for tuks (DMT 2025). Proposed min driver age 23 (union, UV). | [S61–S62] V/UV |
| **Common vehicle categories** | Tuk (Bajaj RE/TVS, 3 pax), Mini/Nano (Alto, Wagon R, Nano), Car (Prius, Axio, Vitz), Minivan (Every/Hijet), Van (Toyota KDH/HiAce, 6–14), Bike (Moto), EV tuks emerging. | industry norm, V via app category lists |

## 6. Culture & UX

| Area | Finding | Implication |
|---|---|---|
| Scripts & fonts | Android ships Noto Sans Sinhala/Tamil (OEM-dependent on old builds); iOS has Sinhala Sangam MN / Tamil Sangam MN. Forum history shows OEM shaping bugs (GSUB ligatures) on old Android. [S63] | Bundle **Noto Sans Sinhala + Noto Sans Tamil** (OFL) via expo-font; test ligatures (ක්‍ෂ, ශ්‍රී, கொ) on a Samsung A0x and a Honor; avoid italic/condensed |
| Trilingual UI | PickMe/Uber ship EN/SI/TA; PickMe added SI/TA mainly for **drivers**, many with low literacy. [S64] | Driver app: icon-first, large buttons, SI/TA default by tenant district; rider app: EN default with persistent language switch |
| Numbers / money | "Rs. 1,250.00" or "LKR 1,250"; integers in speech; coins Rs 1/2/5/10, notes 20–5,000. | Display `Rs 1,250` (no decimals) in apps; minor units internal; **quotes rounded to Rs 10** |
| Phones | +94, 10 digits domestic (07X XXX XXXX); 070 Mobitel, 071 Mobitel, 072/078 Hutch, 074/076/077 Dialog, 075 Airtel→Dialog. [S65] | Input mask `07X XXX XXXX`, store E.164, OTP to mobile only |
| Devices & network | Android 79 % / iOS 21 % (Jun 2026); Android versions spread 11–16 evenly; low-end (< $200) ≈ 84 % of South-Asia shipments, Samsung A05 / Honor X7b / ZTE Blade typical; median mobile 20 Mbps (early 2025) vs 65 Mbps (May 2026, UV); internet penetration 53.6 %. [S66–S68] | Targets: Android 9+ (minSdk 28), cold start < 3 s on 3 GB RAM, APK ≤ 30 MB, map tiles cached, API payloads < 20 KB, polling fallback when sockets drop |
| Cash etiquette | Riders expect exact change; drivers keep little change; tips uncommon but rounding-up common. | Show "Pay driver Rs 270 cash"; driver "mark paid, change given" step; optional "round up" tip |
| Shared tuks | Informal sharing exists on fixed routes (UV); apps don't offer pooled tuks. | Not v1 |
| Women's safety | IFC/PickMe 2020: 58 % of women feel safer with a woman driver; 90 % experienced harassment on public transport; PickMe recruits women drivers, no women-only product confirmed (UV). [S69] | Tenant flag "prefer woman driver" + share-trip + SOS to 119 in v1 |
| Time effects | Night surcharge 15 % after 22:00 (tuk meter + PickMe); peak charge 15–25 %; Uber busiest Fridays 17:00 and December; Poya (monthly full-moon) is a public holiday — alcohol sales banned, lower office demand, temple trips (UV behaviour); Avurudu 13–14 Apr exodus from Colombo; Vesak (May) road closures in Colombo. | Fare rules: time-band + calendar surcharges per tenant; holiday calendar seeded with Poya dates |

## 7. Business-model realities for operators

| Topic | Market practice | Source |
|---|---|---|
| Commission norms | PickMe tuk 12 % (7 % > Rs 8 k/day); Uber ~25 % (2017); Kangaroo Rs 300/day flat; HelaGo "open & fair" bidding (commission not public); YOGO owned fleet flat Rs 65/km | [S4, S17, S19, S21] |
| Collecting commission on cash trips | Prepaid driver credit/"subscription" deducted per trip, topped up via eZ Cash app/outlets, bank deposit, app; card trips settle to driver in real time (HNB) | [S32, S70] V |
| Driver economics | Fuel ≈ 40 % of fare; vehicle owner may take ~50 % of a non-owner driver's earnings; weekly fuel quotas matter (Uber+PickMe lobbied Mar 2026) | [S4, S17, S71] |
| Subscription willingness | Drivers resent per-ride commission; flat daily fees (Kangaroo) and bidding (HelaGo) are positioned against it | [S17, S19] |
| Stand-based dispatch | Tuk societies run stands with queue order; app drivers face hostility at tourist stands (Ella, Weligama, Sigiriya, Galle, Jaffna) | [S22, S4] V; mechanics UV |
| Corporate / hotel | PickMe corporate: credit period, paperless, route maps + automated log sheets (HNB Finance, NDB, Abans, Brandix…); hotels use desk-booked cars via phone fleets | [S72] V |
| Airport (BIA) | Official zone counters Rs 2,550–8,100 to Colombo (sources vary, UV); PickMe/Uber ≈ Rs 2,500–4,000; PickMe pays BIA Rs 700/taxi (UV); aviation ministry allowed PickMe/Uber at BIA Jun 2023 | [S73–S74] |
| Intercity / day hire | Car one-way Rs 70–90/km, return Rs 40–55/km + expressway tolls; day hire 8 h/100 km; KDH flat-roof ≈ Rs 25–27 k/day (100 km) + Rs 125/km extra, high-roof Rs 30–32 k + Rs 130/km; driver bata/accommodation included (UV, operator sites) | [S75] |

## 8. Design decisions for Taxios implied

| # | Decision | Detail |
|---|---|---|
| D-LK-1 | **Default categories per tenant seed** | `tuk` (3 pax), `mini` (Alto/Wagon R, 3–4), `car` (Prius/Axio, 4), `van` (KDH, 6–10), `bike` (moto, 1); optional `nano`, `minivan`, `luxury`, `ev_tuk`. Driver may hold several categories per vehicle (PickMe: Mini + Flex). |
| D-LK-2 | **Fare rule shape = meter shape** | Base covers first km (Rs 110 tuk), then per-km (Rs 90), per-minute waiting, **night band 22:00–05:00 +15 %**, peak multiplier 1.15–1.25 by zone, holiday calendar (Poya/Avurudu). Rider sees fixed upfront quote, rounded to Rs 10; "meter mode" (metered fare from GPS) selectable per tenant for tuk societies. |
| D-LK-3 | **Cash-first flow** | Cash is default and cannot be disabled; quote shown as "Pay driver Rs 270 cash"; driver confirms cash collected (+ change given); receipt by push + optional WhatsApp/SMS; disputes attach GPS trace. Card/LankaQR via PayHere Preapproval or tenant's Dialog Pay dynamic QR, **phase 2**. |
| D-LK-4 | **Commission-credit model** | Per-driver prepaid credit ledger (integer minor units, append-only); commission debited at trip completion for cash trips; **negative floor** per tenant (e.g. −Rs 2,000) after which offers stop; top-ups: staff cash (audited), bank transfer with driver reference (CSV/CEFTS import), eZ Cash/mCash/LankaQR webhook when available; plan types: % commission, flat daily fee, tiered (12 % → 7 % above daily threshold). Card-trip earnings credit the same ledger; payouts via bank CEFTS batch file. |
| D-LK-5 | **SMS fallback & OTP** | Primary OTP via local gateway (Notify.lk/Text.lk/Dialog eSMS) with **tenant-registered 11-char sender ID** (fallback platform ID "TAXIOS"); Twilio only as cross-border fallback; Sinhala/Tamil templates kept ≤ 70 chars/segment; no promo SMS 20:00–08:00; WhatsApp for receipts/ETA later. Push first, SMS when push unacked in 20 s for driver-arrived/trip events. |
| D-LK-6 | **Landmark search** | Tenant POI table (junctions, temples, kovils, schools, bus halts, hospitals, hotels) with SI/TA/EN aliases + trigram search, ranked above Nominatim/Photon (self-hosted, `name:si`/`name:ta` indexed) and Google Autocomplete (session-billed, Essentials fields only). "Near X" free-text note on every pickup; saved places; pickup pin snap via OSRM `match`. |
| D-LK-7 | **Routing & ETA** | OSRM/Valhalla self-hosted on Geofabrik LK extract for quotes and matrix; Google Routes Pro only when tenant enables "live traffic ETA"; learned per-zone/hour speed table from our own trips feeds ETA correction. |
| D-LK-8 | **PIN policy** | Ride PIN off by default; tenant can enable "all trips" or "night only (22:00–05:00)"; rider can opt in; PIN shown big, driver enters to start. |
| D-LK-9 | **Trilingual** | EN/SI/TA in both apps and SMS templates; tenant default language by district; driver app SI/TA-first with icon-led actions; bundle Noto Sans Sinhala/Tamil; currency `Rs 1,250`; phone `07X XXX XXXX`. |
| D-LK-10 | **Low-end performance targets** | minSdk 28, test matrix: Samsung A05 (Android 13/4 GB), Honor X7b, Redmi 9A (2 GB); cold start ≤ 3 s, TTI ≤ 5 s on 3G-class 2 Mbps; APK ≤ 30 MB; GPS heartbeat 5 s moving / 30 s idle; offline queue for status changes; map tiles cached for last 3 zones. |
| D-LK-11 | **Safety kit v1** | SOS → 119 dial + operator alert with live location; share trip link; "prefer woman driver" flag; plate shown large; driver photo required; 1990 ambulance quick-dial in SOS sheet. |
| D-LK-12 | **Compliance** | PDPA readiness by 1 Jan 2027: DPO field per tenant and platform, breach-notification runbook, DPIA on location data, data residency option (LK/SG region), consent copy in SI/TA; driver register export (NIC, licence class, CR, revenue licence expiry, insurance, WP permit) for NTC/WP audits; VAT 18 % + SSCL lines on operator invoices. |
| D-LK-13 | **Stand/queue mode** | Optional zone-level FIFO dispatch for tuk societies (stand geofence; first-in first-offered) alongside nearest-driver dispatch, to onboard existing stands rather than fight them. |
| D-LK-14 | **Hire products** | Hourly/day packages (hours + km caps, extra-km rate), intercity one-way vs return per-km, airport zone flat fares with toll flag, corporate accounts with monthly invoice + log sheet export. |

## 9. Sources (all accessed 2026-10-12)

| # | Source | URL | Status |
|---|---|---|---|
| S1 | PickMe rider app, App Store (categories, cash/card, advance & day booking, languages) | https://apps.apple.com/app/id1000163961 | V |
| S2 | PickMe Driver app, App Store (trip scanner, categories, withdrawal, data use) | https://apps.apple.com/ly/app/pickme-driver/id1520523477 | V |
| S3 | PickMe adds Sinhala & Tamil (Daily Mirror) | https://www.dailymirror.lk/ict/pickme-includes-sinhala-and-tamil-to-application/269-99899 | V |
| S4 | Rest of World, "Sri Lanka PickMe" (12 %/7 % commission, driver economics, BIA 2023) | https://restofworld.org/2024/sri-lanka-pickme/ | V |
| S5 | PickMe joins Dialog Pay / LankaQR (FT, 29 Jul 2026) | https://www.ft.lk/business/PickMe-joins-Dialog-Pay-merchant-ecosystem/34-795295 | V |
| S6 | PhonePe UPI at PickMe (IBS Intelligence) | https://ibsintelligence.com/ibsi-news/phonepe-pickme-enable-contactless-payments-for-indian-tourists-in-sri-lanka | V |
| S7 | PickMe FY25/26 results (The Morning) | https://www.themorning.lk/articles/CZKpQcsy7xp7GcFF8PVL | V |
| S8 | "PickMe Credit" (People's Bank PDF, Mar 2025) | https://www.peoplesbank.lk/roastoth/2025/03/PickMe-Credit.pdf | UV (binary not parsed) |
| S9 | Yamu, PickMe/Uber fare changes (peak/night, 2016) | http://www.yamu.lk/trending/pickme-and-uber-srilanka-what-the-fare-increases-mean-for-you/ | V (dated) |
| S10 | Uber LK newsroom: safety features (PIN, RideCheck, helpline) | https://www.uber.com/en-LK/newsroom/uber-announces-new-safety-features-in-sri-lanka | V |
| S11 | Uber Tuk Rentals (Aug 2020) | https://www.uber.com/en-LK/newsroom/uber-launches-tuk-rentals-in-sri-lanka-for-multi-hour-multi-stop-needs | V |
| S12 | How Sri Lanka Ubered in 2024 (FT) | https://www.ft.lk/motor/How-Sri-Lanka-Ubered-in-2024/55-772260 | V |
| S13 | Uber Black launch LK (FT, Oct 2026) | https://www.ft.lk/motor/Uber--brings-the--business-class-of-back-seats--with-Uber-Black/55-798287 | V |
| S14 | Uber expands to Kegalle (Daily Mirror, Apr 2026) | https://www.dailymirror.lk/print/business/Uber-expands-to-Kegalle/215-338572 | V |
| S15 | Uber for teens LK (FT, Jul 2025) | https://www.ft.lk/motor/Uber-for-teens-debuts-in-Sri-Lanka/55-779066 | V |
| S16 | UberTUK launch (cash + card) | https://www.dailymirror.lk/print/press_releases/UberTUK-rides-into-Colombo/335-154402 | V (dated) |
| S17 | ReadMe, "Silent voices against Uber" (25 % Uber, Rs 300/day Kangaroo, owner 50 %) | https://readme.lk/?p=83026 | V (2017) |
| S18 | Kangaroo Cabs app listing | https://apps.apple.com/app/id1166543666 | V |
| S19 | HelaGo launch (Ada Derana Biz, Jan 2026) | https://bizenglish.adaderana.lk/helago-launched-a-locally-built-platform-bringing-a-new-era-of-fair-ride-hailing-to-sri-lanka/ | V |
| S20 | HelaGo 1 M km milestone (Hiru) | https://hirunews.lk/english/business/473818/helago-crosses-1-mn-km-in-just-5-months-marking-major-milestone | V |
| S21 | DPMC E-Drive / YOGO (Rs 65/km) | https://www.dpg.lk/news-and-events/107/view | V |
| S22 | AsiaNews, tuk drivers vs apps (Jan 2026) | https://www.asianews.it/news-en/Tuk-tuk-drivers-against-paid-apps:-attacks-and-controversy-in-Sri-Lanka-64587.html | V |
| S23 | Trustpilot PickMe reviews | https://uk.trustpilot.com/review/pickme.lk | UV (anecdotal) |
| S24 | Global Press Journal, Jaffna drivers | https://globalpressjournal.com/asia/sri_lanka/ride-hailing-app-drives-off-unionized-taxis-northern-sri-lanka/ | V |
| S25 | Kimola PickMe Play-review analysis | https://kimola.com/reports/unlock-insights-with-the-pickme-app-user-feedback-analysis-google-play-en-154763 | UV |
| S26 | PayHere support FAQ | https://support.payhere.lk/general-questions | UV (403 on fetch) |
| S27 | PayHere Laravel wrapper (Preapproval/Recurring/Subscription APIs) | https://github.com/Visanduma/laravel-payhere | V (API names) |
| S28 | ReadMe, LK internet payment gateways 2019 (fees) | https://readme.lk/internet-payment-gateways-in-sri-lanka-2019/ | V (dated) |
| S29 | WebXPay pricing page (packages, XSPLIT) | https://orchardmothercare.webxpay.com/pricing | V |
| S30 | Genie Business app listing | https://apps.apple.com/app/id1613659355 | V |
| S31 | Dialog Genie Checkout (Shopify) | https://apps.shopify.com/dialog-genie-checkout | V |
| S32 | eZ Cash × PickMe driver subscription payments (Dialog, 2019) | https://dialog.lk/news/eZ-Cash-Partners-PickMe-to-Facilitate-Easy-Subscription-Fee-Payments-for-Drivers | V |
| S33 | CBSL LankaQR launch (MDR 0.5 %) | https://www.cbsl.gov.lk/en/news/Launch-of-the-National-LANKAQR-initiative | V |
| S34 | ACE blog, LK wallets (LankaQR fee waiver Apr 2026) | https://acemoneytransfer.com/blog/top-digital-wallet-apps-sri-lanka | UV |
| S35 | Koko × LankaPay JustPay (FT) | https://www.ft.lk/financial-services/Koko-and-LankaPay-form-strategic-partnership/42-773157 | V |
| S36 | LankaPay FAQ (CEFTS/SLIPS) | https://www.lankapay.net/en/faq | V |
| S37 | sent.dm Sri Lanka SMS pricing guide (Oct 2025) | https://sent.dm/en/resources/sms-pricing/sri-lanka-sms-pricing | UV (aggregator) |
| S38 | Notify.lk pricing | https://notify.lk/pricing | V |
| S39 | Text.lk FAQ | https://text.lk/faq | V |
| S40 | Meta WhatsApp Business pricing | https://developers.facebook.com/documentation/business-messaging/whatsapp/pricing | V (LK card not extracted) |
| S41 | go4whatsup LK WhatsApp pricing | https://go4whatsup.com/sri-lanka/whatsapp-business-api-pricing/ | UV |
| S42 | Google Maps Platform core pricing list | https://developers.google.com/maps/billing-and-pricing/pricing | V |
| S43 | Mapbox pricing 2026 (Woosmap/Storerocket); HERE pricing trackers | https://www.woosmap.com/blog/mapbox-pricing ; https://pricingsaas.com/companies/here | UV |
| S44 | OSM community: language for `name` in Sri Lanka | https://community.openstreetmap.org/t/what-language-should-we-use-for-the-main-tag-name-in-sri-lanka/103979 | V |
| S45 | OSM Help: Nominatim accept-language | https://help.openstreetmap.org/questions/69547/how-to-convert-reverse-goecode-result-to-english | V |
| S46 | HDX OSM buildings LK (64 % coverage) | https://geo.btaa.org/catalog/577cb335-bb51-4530-9f9d-36fec057167d | UV |
| S47 | FT, PDPA regime takes effect 1 Jan 2027 (Gazette 2498/16) | https://www.ft.lk/front-page/Data-protection-compliance-regime-takes-effect-on-1-Jan-2027/44-795776 | V |
| S48 | ReadMe Digital Policy Watch, PDPA timeline | https://readme.lk/digital-policy-watch/personal-data-protection-act-no-9-of-2022/ | V |
| S49 | ConsentStack LK PDPA summary (penalties, cross-border) | https://www.consentstack.io/regulations/lk-pdpa | UV |
| S50 | The Morning, NTC "no regulations yet" (Oct 2025) | https://www.themorning.lk/articles/CwXXpi75OIGIjlBjV2j0 | V |
| S51 | The Morning, NTC app-mandate plan (8 Feb 2026) | https://www.themorning.lk/articles/kYD7QwpLulbYWPYpYx4t | V |
| S52 | The Morning, tuk fare Rs 110/90 (23 Mar 2026) | https://www.themorning.lk/articles/WE3bogAM74EGohcMEMOO | V |
| S53 | Newsfirst, first-km fare Rs 110 | https://newsfirst.lk/2026/03/23/first-kilometer-fare-for-three-wheelers-raised-to-rs-110 | UV (429) |
| S54 | Daily Mirror, WP tuk fares regulated (Jul 2024, night +15 %) | https://dailymirror.lk/latest-news/Western-Province-three-wheeler-taxi-fares-regulated/342-287310 | V |
| S55 | WP Gazette 2258/25 (Dec 2021) three-wheeler & taxi rules | https://documents.gov.lk/view/egz/2021/12/2258-25_E.pdf | V (not parsed) |
| S56 | Newsfirst, WP registration resumes 1 Oct 2025 | https://www.newsfirst.lk/2025/09/29/from-oct-1-ticketless-travel-will-cost-you | V |
| S57 | Chambers Doing Business 2026 LK (VAT 18 %, SSCL) | https://gpg-pdf.chambers.com/Doing-Business-In-2026/1010/ | V |
| S58 | Fonoa, LK VAT on cross-border digital services | https://www.fonoa.com/resources/blog/sri-lanka-imposes-vat-cross-border-digital-services | V |
| S59 | Wikipedia, Telephone numbers in Sri Lanka (prefixes, emergency) | https://en.wikipedia.org/wiki/Telephone_numbers_in_Sri_Lanka | V |
| S60 | Ada Derana, 119 misuse warning (Dec 2025) | https://www.adaderana.lk/news/116150/police-warn-against-misuse-of-119-hotline-legal-action-against-offenders | V |
| S61 | Southern Province three-wheeler regs (Gazette 1777/39, 2012) | https://documents.gov.lk/view/extra-gazettes/2012/9/1777-39_E.pdf | V |
| S62 | Ada Derana, DMT no tuk licences for foreigners | https://www.adaderana.lk/news/114914/no-driving-licences-will-be-issued-to-foreigners-to-drive-three-wheelers-dmt-confirms- | V |
| S63 | Sony community, Sinhala rendering on Android | https://community.sony.ch/t5/z-series/can-t-read-sinhala-font/m-p/3361788 | UV (dated) |
| S64 | FT, PickMe in Sinhala and Tamil | https://ft.lk/Business/pickme-up-in-sinhala-and-tamil/34-508999 | V |
| S65 | JustCall LK phone format | https://www.justcall.io/blog/sri-lanka-phone-number-format.html | UV |
| S66 | StatCounter LK mobile OS & Android versions | https://gs.statcounter.com/os-market-share/mobile/sri-lanka | V |
| S67 | DataReportal Digital 2025 Sri Lanka | https://datareportal.com/reports/digital-2025-sri-lanka | V |
| S68 | IDC South Asia 2Q24 (LK Honor/ZTE/Samsung, price bands) | https://www.idc.com/resource-center/press-releases/south-asia-smartphone-market-grew-by-39-5-in-2q24-says-idc/ | V (dated) |
| S69 | IFC study, women riders LK (2020) | https://www.ifc.org/en/pressroom/2020/more-female-riders-could-boost-revenues-for-ride-hailing-in-sri-lanka-finds-new-ifc-study | V |
| S70 | FT, PickMe real-time payment transfers via HNB (2017) | https://www.ft.lk/Motor/PickMe-enables-real-time-payment-transfers-for-driver-affiliates/55-645418 | V |
| S71 | Daily Mirror, Uber/PickMe fuel quotas (Mar 2026) | https://www.dailymirror.lk/print/business-news/Uber-Pickme-urge-enhanced-fuel-quotas-to-sustain-mobility-and-delivery-services/273-336033 | V |
| S72 | PickMe corporate transport (HNB Finance) | https://pickme.lk/pickme-welcomes-hnb-finance-to-the-corporate-transport-partnership-fold | V |
| S73 | Newswire, PickMe parking slot at BIA (2023) | https://www.newswire.lk/2023/03/15/pickme-granted-designated-parking-slot-at-bia-for-affordable-travel/ | V |
| S74 | Island, BIA earns Rs 700 per PickMe taxi | https://island.lk/bia-to-earn-rs-700-per-taxi-from-pickme/ | UV |
| S75 | Lakpura taxi/KDH rates; Jungletide hire guide | https://www.lakpura.com/pages/taxi ; https://jungletide.com/hiring-drivers-in-sri-lanka/ | UV |
| S76 | StanChart API case study with PickMe (24/7 collections) | https://www.sc.com/en/news/corporate-investment-banking/api-solution-sri-lanka-pickme-case-study/ | V |

Open items to verify with vendors before build: PayHere 2026 rate card + Preapproval API docs; Dialog eSMS
business rate and mask SLA; Dialog Pay/LankaQR dynamic-QR API; LankaPay customer-charge caps; WP meter
enforcement status under the forthcoming NTC regulations; PickMe Credit top-up mechanics (S8).
