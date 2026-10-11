# Uber parity reference (rider app, driver app, platform) — for Taxios

Researched 2026-10-12. Purpose: make Taxios *work and feel* like Uber while every screen carries the
**operator's own brand**. We copy **patterns and behaviour, never pixels**: no Uber name, logo, wordmark,
colours, typefaces (Uber Move), icon set, illustration style, sounds or exact layouts. Brand comes from
T-007 (tenant logo/colour/computed text colour).

Legend — Source tags `[Sn]` resolve in §6 (all accessed 2026-10-12). `(UV)` = not verified from a primary
source in this pass (observed/common knowledge or third-party only). `(3P)` = third-party source.
`(old)` = primary source is ≥ 3 years old; current app may differ. Uber varies features by country; nothing
here is a claim about Uber in Sri Lanka unless tagged LK.

Priority keys used in §5: **MVP** (before the first operator goes live), **v1.1** (first 3 months live),
**later**, **n/a LK** (not applicable / not wanted in Sri Lanka).

---

## 1. Rider app — screen by screen

| # | Screen / state | Uber pattern (described) | Src |
|---|---|---|---|
| R1 | Home | Map-first full-bleed map centred on the rider's location with nearby cars drawn as small vehicle icons; one dominant search affordance ("Where to?") in a bottom card/sheet; below it saved places (Home/Work) and recent/suggested destinations as one-tap rows; product shortcuts (ride, reserve, other services) as a tile row. The 2016 redesign set the principle: the app opens on *one question* (destination) with frequent destinations as shortcuts. | [S20](old), (UV) for current tiles |
| R2 | Destination entry | Full-screen search: two stacked fields (pickup prefilled with "current location", drop-off focused), connector dots/line between them; results list mixes saved, recent and search hits with an icon per type; "Set location on map" row opens a fixed-centre-pin map you pan under; "+" adds stops (multi-stop). Keyboard dismisses when a result is chosen. | (UV); Lite POI-first entry [S11] |
| R3 | Product selection sheet | Route drawn on map (pickup → drop-off) above a bottom sheet listing products (vehicle icon, name, seats, pickup ETA or drop-off time, upfront price); one row preselected; sheet expands to compare all; "fastest/cheapest" style badges; primary CTA "Choose <product>" pinned to the sheet bottom. | [S4] upfront price, (UV) layout |
| R4 | Payment row + promo | A single tappable row above the CTA showing the current method (card / cash / wallet / business profile); tapping opens a method picker that also hosts "add promo code". Cash is a first-class method in cash markets (LK cash since 2016). | [S13](LK, old), [S14] |
| R5 | Confirm pickup | Map zooms to the pickup point; draggable pin (map pans under a fixed pin) with the street/landmark name shown; **suggested pickup spots** shown as dots the pin snaps to (historically from where drivers actually stopped); special venues (airports) use designated zones/pickup lists. CTA "Confirm pickup". | [S16](old), [S17] |
| R6 | Matching / searching | Pulsing ring at pickup on the map, sheet shows "Finding your ride" with progress + the requested product and price; cancel is free and prominent in this state. Matching is **batched** (system waits a few seconds to pool requests). | [S7], [S8] free pre-match cancel |
| R7 | Driver en route | Live car icon moving smoothly along the road toward the pickup; sheet top line = "Meeting at <spot> · N min"; driver card: photo, first name, rating, vehicle make/model/colour, **plate in large type**; actions: masked call + in-app chat, share trip status, safety shield (toolkit), cancel (fee warning if past the free window). | [S8], [S10], [S12] |
| R8 | Pickup arrival + PIN | "Your driver has arrived" (push + banner), wait-time clock starts after a grace period; optional **4-digit PIN** the rider tells the driver; driver cannot start the trip until it matches; rider gets an in-app confirmation. Can be "all rides" or "night only". Fallback guidance: check photo, plate, make/model. | [S9] |
| R9 | On trip | Route line to destination, ETA/arrival time in the sheet, live car = you; actions: change destination (re-priced), add stop, split fare, share trip, safety toolkit. Changing destination/stops/route changes the final price. | [S4], [S18] |
| R10 | Trip end | Receipt summary (fare, method, breakdown link); rating 1–5 stars, < 5 asks "what went wrong" chips; 5 stars offers **compliments** chips; optional tip (where enabled); then back to Home. Rating is anonymous; average = last 500 trips. | [S19](old), tip (UV per market) |
| R11 | Trip history / receipts | Activity list (upcoming + past), each with map thumbnail, date, price, status; detail page shows route, fare breakdown, driver, "get help", re-send/download receipt (email + PDF/business invoice). | (UV) |
| R12 | Scheduled rides (Reserve) | Pick date/time (Uber: up to 90 days ahead; older pages 30 days/≥ 2 h); fare **locked once a driver is assigned**; reservation fee; included wait window; rider cancel < 1 h before pickup is charged. Drivers must be online ahead of time. | [S21] (versions differ) |
| R13 | Account / settings | Profile, saved places, payment, family/business profiles, privacy, safety preferences (PIN, trusted contacts, RideCheck), accessibility, language, notifications, delete account. | [S9], (UV) |
| R14 | Safety | **Safety toolkit** behind a shield button on every trip screen: emergency call (local emergency number; app shows current address + map to read out), share trip with up to 5 trusted contacts (auto-prompt rules), safety centre, **RideCheck** (anomaly detection — long stop, possible crash, route deviation — pings both parties "Are you OK?"), **audio recording** (encrypted on-device, only Uber can decrypt and only if attached to a report; regional retention e.g. 7 days). Matching avoids pairs that rated each other 1 star. | [S10], [S12], [S7] |
| R15 | Accessibility | Screen-reader labels; driver "deaf/hard-of-hearing" mode: rider calling disabled, rider prompted to enter destination + told driver is deaf/HoH; driver request alerts via flashing light/vibration; WAV/Assist products in some cities. | [S22] |
| R16 | Notifications | Push for: driver assigned, arriving soon, arrived, trip started, receipt; iOS **Live Activity** on Lock Screen + Dynamic Island (ETA, status, vehicle + plate) since Feb 2023. Android equivalent: ongoing notification (UV). | [S23](old) |
| R17 | Offline / poor network | Uber Lite lessons: cache frequent/recent destinations; POI/landmark pickups instead of typed addresses; payloads < 1 MTU; usable on sub-3G with < 300 ms screen transitions. Main app: request state survives reconnects (server is source of truth; push stream resends unacked messages on reconnect). | [S11], [S5] |
| R18 | 2026 direction | GO-GET 2026: "one app for everything" (hotels, Travel Mode, Shop for Me) — **not relevant** to Taxios' single-operator scope. | [S24] |

### 1.1 Cancellation & waiting rules (rider side)
| Rule | Uber behaviour | Src |
|---|---|---|
| Before a driver is assigned | Free | [S8] |
| Rider cancels after accept | Fee if > 2 min after accept (Economy), > 5 min (Premium); fee = max(city minimum, time/distance from match to pickup) | [S8] |
| Driver waits / no-show | Driver may cancel with a no-show fee after waiting ~5–7 min (Economy), 10 (Comfort), 15 (Premium); when a cancel fee is charged no wait fee applies | [S8] |
| Fee waived | Driver > 5 min late vs original ETA, not progressing, or moving away | [S8](old UK) |
| Dispute | "Review cancellation fee" flow in Help | [S8] |

---

## 2. Driver app

| # | Area | Uber pattern (described) | Src |
|---|---|---|---|
| D1 | Go online/offline | Big single "Go" control on a map home; status strip ("You're online · finding trips"); earnings pill at the top (customisable/hideable); preferences (trip types, destination filter) reachable from the strip. | [S25](old 2018) |
| D2 | Demand / surge display | Heat-map overlay on the home map: zones shaded light→dark for low→high surge; surge shown on offers as multiplier, add-on amount or included in the upfront fare (varies by city); service-fee % unchanged during surge. | [S15] |
| D3 | Trip request card | Full-width card with countdown ring/bar (offer expires ~ seconds), loud tone + vibration (+ flash for deaf drivers); shows pickup distance/time, **upfront earnings**, trip length, **destination** area, rider rating, product, surge/boost tags. One tap to accept; decline/ignore lets it expire. **Trip Radar**: optional list of other nearby open trips to claim. | [S26], [S22] |
| D4 | Navigation | Built-in turn-by-turn or hand-off to Google Maps/Waze; returns to the app via overlay/bubble on Android (UV). | (UV) |
| D5 | Trip steps | Navigate to pickup → "Arrived" (geofenced; starts rider wait clock) → enter rider PIN if required → "Start trip" (swipe-to-confirm) → navigate to destination → "Complete trip" (swipe) → rate rider. Swipe gestures prevent accidental taps. | [S9], (UV) swipe |
| D6 | Waiting / no-show | Wait timer visible after arrival; cancel-for-no-show button unlocks after the threshold and pays the no-show fee. | [S8] |
| D7 | Ratings | Driver rates rider each trip; driver sees own average + compliments + feedback themes. | [S19] |
| D8 | Earnings tab | Balance → itemised trip list (fare, fees, tips, promotions), daily/weekly views, expenses; trackers for **Quests** (N trips or $X earnings in a window → bonus) and other opportunities with expiry; cash out from the same tab. | [S27](old), [S28](3P) |
| D9 | Cash-out | Instant Pay up to 6×/day to a debit card (US, fee per cash-out); LK: daily payments/cash-out at any time; LK cash-heavy — Uber penalises drivers who refuse card trips. Cash trips create a driver→platform balance owed (UV detail). | [S29], [S13] |
| D10 | Destination filter | Driver sets a destination; only trips heading that way are offered; limited uses per day (commonly 2; a use is consumed only when a trip is received). | [S30], (3P) limit |
| D11 | Onboarding / documents | Sign up → upload licence, vehicle registration, insurance, profile photo; status per document (pending/approved/expiring); app blocks going online until compliant. | (UV) — Taxios already has this (driver_compliance) |
| D12 | Safety | Same shield/toolkit: emergency button, share trip, speed alerts, RideCheck, audio recording where available. | [S10], [S12] |
| D13 | Support | In-app help by trip, report rider, lost items, fare adjustments. | (UV) |

---

## 3. Platform mechanics

| Topic | Uber approach | What Taxios should do | Src |
|---|---|---|---|
| Geospatial index | **H3** hexagonal hierarchical grid (res 0–15); `kRing` neighbours approximate circles for nearby-supply search; used for pricing/surge and dispatch analysis. Earlier DISCO sharded supply by cell IDs (S2). | PostGIS `ST_DWithin` + GiST is enough for one operator's fleet (hundreds of drivers). Adopt H3 (`h3-pg` or Go lib) only for surge zones/heatmaps (T-143). | [S1], [S6](3P,old) |
| Matching | **Batched**: wait a few seconds, pool riders + drivers, minimise *average* wait for all, not nearest-first; "closest ≠ quickest" (road ETA, rivers, overpasses); considers in-flight drivers finishing trips (DISCO); avoids 1-star pairs. | Keep sequential offer (one driver at a time, countdown) for MVP; rank candidates by **road ETA** not straight-line once T-013b lands; add a 2–5 s batch window + "finishing soon" drivers in v1.1; block pairs with mutual 1★. | [S7], [S6] |
| Offer protocol | Push offer to one driver; offer TTL ≈ seconds; ack needed to know it arrived (motivated the gRPC move). | Server-side offer expiry is authoritative (already). Record **delivered_at** ack per offer; don't count un-delivered offers as declines. | [S5], [S5b] |
| Realtime transport | **RAMEN**: persistent stream (SSE → gRPC bidi over QUIC/HTTP3); 1-byte heartbeat every 4 s, client reconnects after 7 s silence; per-message priority (high/med/low), TTL seconds–30 min, monotonically increasing **sequence numbers**, at-least-once with resend on reconnect from last-seen seq; only latest message per type sent; **Fireball** decides *when* to push (user action, app open, timers, bus events, geofence in/out). Before RAMEN, polling was ~80 % of gateway traffic. | T-031 WebSocket: per-user channel, seq numbers + "resume from seq" on reconnect, heartbeat 15–25 s (mobile radio cost), "latest-state-per-booking" snapshots rather than event replay; fall back to polling the booking endpoint every 5–10 s when the socket is down; push (T-153) for backgrounded apps. Outbox (T-030) feeds it. | [S5], [S5b] |
| Location updates | Trip GPS pings roughly every **4 s**; map-matching via HMM + Viterbi to snap noisy GPS to roads. | Driver: 4–5 s on trip / en-route, 10–15 s idle online, distance filter 10–20 m; batch + send on reconnect; snap to road server-side later (OSRM/Valhalla match) for fare distance. | [S3], [S31] |
| ETA / routing | ETAs drive fare, pickup estimate, matching; DeepETA (deep model on top of a routing engine, ms latency); DeepETT traffic forecasting. | Routing provider adapter (T-013b) — Google Routes or self-hosted OSRM with LK OSM; heuristic fallback; cache per (cell, cell, hour). Never claim live traffic we don't have. | [S2], [S3] |
| Upfront pricing | Estimate shown before request = base + time + distance (expected route + traffic) + dynamic charge + tolls/fees; **changes** if destination/stops change, route deviates materially, trip much longer, or pickup differs. Driver sees upfront earnings + destination. | Quote = upfront fare locked at booking with tolerance rules: re-price on destination/stop change; GPS-final when deviation > X % or > Y min (operator-configurable, audited). Taxios already computes GPS final fare — add the "keep upfront unless deviation" rule. | [S4], [S26] |
| Surge | Per-zone multiplier from local demand/supply, updated near-real-time; driver heat map; rider sees higher price before confirming. | Opt-in, capped, audited per tenant (T-143); many LK operators will want **fixed tariffs** — default off. | [S15] |
| Payments | Card **authorisation hold** at request (small or ≈ fare), converted to final charge at trip end; second hold if destination change exceeds it; trip declined if hold fails. Cash = pay driver the in-app amount. | MVP is cash: driver confirms cash collected (done). Card later via PayHere/Payable (T-160/161): pre-auth if provider supports it, else charge-after-trip with saved token; never fabricate provider APIs. | [S14], [S13] |
| Safety / fraud | RideCheck anomaly detection, PIN, audio, trusted contacts, emergency; 1★ pair blocking; auth hold as fraud check. | MVP: PIN (night-only default), share-trip link, SOS → call 119/operator, stop-too-long/route-deviation alert to dispatcher (we have a human dispatcher — use it). | [S9]–[S12] |
| Reliability | Two-region failover historically at 2× capacity; NSDI '26 "UFA" tiers services by criticality (2× → 1.3×, 99.97 %). Cell-based architecture: **no Uber primary source found**. | Single region is fine; priorities: idempotent booking (done), outbox + durable jobs (T-030), clients that recover from any state via `GET /bookings/{id}`. | [S32] |
| Offline-first clients | Lite: cached destinations, POI pickups, tiny payloads; stream resend on reconnect. | Cache last booking state + saved places; queue driver status transitions (arrived/start/end/cash) with idempotency keys and replay on reconnect. | [S11], [S5] |
| Notifications | Fireball chooses push channel/timing; Live Activities on iOS. | T-153 push; Live Activity / Android ongoing notification in v1.1 (needs dev build, not Expo Go). | [S5], [S23] |

---

## 4. Design-system patterns (describe, don't copy)

| Pattern | Uber-style behaviour | Taxios rule |
|---|---|---|
| Map + sheet layout | Map always behind; one bottom sheet carries the state's content + one primary CTA; map camera re-frames to fit relevant points (you + pickup, or route) with the sheet height as bottom padding. | Every rider/driver screen = `MapView` + `Sheet`; camera fit uses sheet height inset. |
| Sheet snap points | Collapsed (summary + CTA), half (list), full (details); drag handle; sheet height animates between states; content never hidden behind the CTA. | `@gorhom/bottom-sheet` (or Expo UI BottomSheet) with 2–3 snap points per state; CTA pinned in a footer. |
| Typography / spacing (Base) | Large, high-contrast type, few type sizes, generous whitespace, 4 px base spacing scale, near-monochrome UI with colour reserved for state/CTA. Base Web (open-source React) implements the language. Type specifics vary across 3P teardowns → treat as unverified. | System font (or tenant font later); our own scale (e.g. 12/14/16/20/28); 4-pt spacing; **tenant colour only on primary CTA, selected states, route line and own-vehicle marker**; neutral greys elsewhere so any brand colour works. |
| Motion — vehicle marker | Car glides between pings (interpolated position + rotated to bearing), not teleporting; route line drawn progressively; ETA text updates without layout jump. (UV — observed behaviour) | Interpolate between consecutive pings over the ping interval (linear on lat/lng + shortest-arc bearing), drop out-of-order pings by timestamp, snap > 100 m jumps. |
| Loading | Skeleton rows in product list; searching state = pulse at pickup; never a blank map. | Skeletons + pulse; optimistic UI only for local choices, never for money. |
| Empty / error | Inline, actionable ("No cars available — try again / schedule"); network banner at top; retry keeps the user's input. | One inline error component; offline banner; keep form state. |
| Accessibility | Screen-reader labels on map markers + actions, big touch targets, haptics on offer, flash alerts option, text alternatives to calls. | `accessibilityLabel` on every marker/CTA, 44×44 pt targets, Dynamic Type, Sinhala/Tamil fonts tested at large sizes. |

---

## 5. Parity matrix

| Feature | Uber behaviour | Taxios priority | Sri Lanka / cash-first notes |
|---|---|---|---|
| Map-first home with live nearby vehicles | Map + "Where to?" + shortcuts | **MVP** | Today the passenger app has no map (T-151 gap). Show vehicles as tuk / car icons per category; jitter positions slightly for driver privacy. |
| Saved places + recents | One-tap rows | **MVP** | T-133. Landmark-style labels ("near Kollupitiya junction"). |
| Destination search | Two fields, results list, set-on-map | **MVP** | LK addresses are landmark-based: search must include POIs (temples, junctions, bus stands, hospitals); allow free-text "notes for driver". |
| Set pickup on map (fixed pin) | Pan map under pin | **MVP** | Essential where geocoding is weak; reverse-geocode to nearest landmark. |
| Suggested pickup spots | Pin snaps to good spots | later | Needs history of where drivers actually stopped; collect it from trip data now. |
| Multi-stop | "+" add stop, re-priced | v1.1 | T-140. Common for errands by tuk. |
| Product list with upfront price + ETA | Sheet with price/ETA/seats | **MVP** | Categories = tuk, mini car, car, van; show seats + luggage. Prices in LKR integer rupees. |
| Payment row | Method row above CTA | **MVP** | Cash default; card/wallet later. Make "cash" explicit so drivers can't refuse. |
| Promo codes | In payment picker | v1.1 | T-142. |
| Batched matching | Few-second batch, min avg wait | v1.1 | MVP sequential offers fine at single-operator scale. |
| Road-ETA ranking | ETA not distance | v1.1 | T-013b; Colombo traffic makes straight-line misleading. |
| Searching state + free cancel | Pulse + cancel | **MVP** | Also offer "call the operator" — many LK operators still take phone bookings. |
| Driver card (photo, name, plate, vehicle, rating) | Plate prominent | **MVP** | Plate format "WP CAB-1234"; tuk colour helps identification. |
| Live driver on map + smooth animation | Interpolated marker | **MVP** | Depends on T-031 (or 5 s polling as interim). |
| Masked call | Proxy number | later | Needs telco proxy; MVP = in-app call button revealing operator hotline or driver number per tenant policy. |
| In-app chat | Text + quick replies | v1.1 | Quick replies in si/ta/en ("I'm at the gate"). |
| Share trip | Link with live map | **MVP** | Public tokenised tracking page on the tenant's web domain; WhatsApp share sheet. |
| Cancellation + no-show fees | Time windows, fee waivers | v1.1 | T-141. Cash market: fee can only be collected on the next trip → owed balance; default fees OFF. |
| Arrived + wait timer | Grace then paid wait | **MVP** | Waiting-time charge per minute exists in many LK tariffs. |
| Ride PIN | 4-digit, opt-in / night | **MVP** | Cheap, high trust signal; default "night only". |
| On-trip change destination | Re-price | v1.1 | Cash: final fare shown to both before payment. |
| Split fare | Per-rider card | n/a LK | Cash trips — not applicable. |
| Upfront price with deviation rules | Locked unless change | **MVP** | Fixed upfront fare is the main reason riders choose apps over meterless tuks; deviation threshold per operator. |
| Surge | Zone multiplier + heat map | later | Opt-in, capped; consider fixed "night tariff" instead (common in LK). |
| Receipt | In-app + email/PDF | **MVP** | Done (receipts); add SMS receipt option — email penetration is low. |
| Rating + compliments | Stars, chips | **MVP** | Rating done; add chips (si/ta/en). |
| Tips | In-app | n/a LK (cash) | Cash tips happen off-app; add later only with card. |
| Trip history | List + details | **MVP** | Done (history); add map thumbnail later. |
| Scheduled rides | Reserve, fare locked at assign | **MVP** | Already supports now/+30m…tomorrow; add date picker; airport runs (BIA) are the big use case. |
| Safety toolkit (SOS, share, trusted contacts) | Shield button | **MVP** | SOS calls 119 (police) / 1990 (Suwa Seriya ambulance) and alerts the operator's dispatcher. |
| RideCheck-style anomaly alerts | Auto "are you OK?" | v1.1 | Route deviation / long stop → dispatcher alert first (human in loop). |
| Audio recording | Encrypted, on-device | later | Legal review under LK PDPA 2022 first. |
| Accessibility (screen reader, deaf driver mode) | Flash alerts, chat-only | v1.1 | Also low-literacy: icons + numbers, voice prompts. |
| Push notifications | Assigned/arriving/arrived/receipt | **MVP** | T-153; SMS fallback for "driver arrived" (OTP gateway). |
| Live Activities / ongoing notification | Lock-screen ETA | v1.1 | Needs dev/EAS build. |
| Offline tolerance | Cached places, resend | **MVP** | Patchy 4G outside Colombo; idempotent retries for every driver action. |
| Languages | Many | **MVP** | Sinhala, Tamil, English (T-128); number formatting LKR "Rs 1,250". |
| Driver go online/offline | Big toggle + status strip | **MVP** | Done; add earnings pill. |
| Offer card with countdown | Earnings + destination shown | **MVP** | Done (countdown); show destination area + upfront fare + pickup ETA. |
| Trip Radar | List of open nearby trips | later | Could suit tuk stands; conflicts with dispatcher control — tenant option. |
| Navigation hand-off | Built-in or Google Maps/Waze | **MVP** | Done (Google Maps hand-off); add a "return to app" floating bubble on Android later. |
| Swipe to start/end | Prevents mis-taps | **MVP** | Cheap; also confirm cash amount on end. |
| Driver earnings tab | Daily/weekly, itemised | **MVP** | T-132/T-144; show commission owed to operator for cash trips. |
| Quests / incentives | Trip-count bonuses | later | Operator-configured; audit + money rules. |
| Cash-out | Instant Pay | n/a LK (MVP) | Cash-first: drivers *owe* commission; build a "settle with operator" ledger instead. |
| Destination filter | Limited uses/day | v1.1 | Popular with drivers going home; tenant limit. |
| Heat map for drivers | Zone shading | later | With H3 + surge. |
| Document onboarding | Upload, statuses, block | **MVP** | Done (driver_compliance); add expiry reminders (T-130). |
| Rider rating by driver | Each trip | **MVP** | Check T-152 covers it; feeds 1★ pair blocking. |
| 1★ pair blocking in matching | Never re-match | v1.1 | Simple SQL exclusion in dispatch. |
| Card pre-auth | Hold → capture | later | Depends on PayHere/Payable capabilities (T-160). |
| Corporate/business profile | Separate billing | later | B2B accounts matter for LK operators (hotels, offices). |
| Hotels / Travel Mode / delivery | Super-app | n/a | Out of scope. |

---

## 6. Sources (all accessed 2026-10-12)

| ID | Source | URL |
|---|---|---|
| S1 | Uber Eng — H3: Uber's Hexagonal Hierarchical Spatial Index (2018) | https://www.uber.com/blog/h3/ |
| S2 | Uber Eng — DeepETA: How Uber Predicts Arrival Times (2022) | https://www.uber.com/blog/deepeta-how-uber-predicts-arrival-times/ |
| S3 | Uber Eng — Scaling real-time traffic (DeepETT; GPS ≈ every 4 s) | https://www.uber.com/blog/scaling-real-time-traffic/ |
| S4 | Uber — Upfront pricing explained (AU) / What is upfront pricing (UK, 2023) / Know your fares (IN) | https://www.uber.com/au/en/ride/how-it-works/upfront-pricing ; https://www.uber.com/gb/en/blog/uk-upfrontpricing ; https://www.uber.com/en-IN/newsroom/kyf |
| S5 | Uber Eng — Uber's Real-Time Push Platform (RAMEN, Fireball) | https://www.uber.com/blog/real-time-push-platform/ |
| S5b | Uber Eng — Uber's Next Gen Push Platform on gRPC (2022) | https://eng.uber.com/blog/ubers-next-gen-push-platform-on-grpc/ |
| S6 | (3P, old) QCon 2015 summaries of DISCO dispatch | https://highscalability.com/how-uber-scales-their-real-time-market-platform/ |
| S7 | Uber — Marketplace: Matching (batching, 1★ blocking) | https://www.uber.com/us/en/marketplace/matching/ |
| S8 | Uber Help — Review cancellation fee; Uber developer — dispatch & cancellation; UK wait-time blog | https://help.uber.com/h/6bec690f-ee35-40ba-96ee-c38a8ae796e0 ; https://developer.uber.com/docs/guest-rides/guest-ride-api-build-guide/dispatch-and-cancellation ; https://www.uber.com/en-GB/blog/changes-to-wait-times |
| S9 | Uber — Verify your ride (PIN) | https://www.uber.com/en-AU/blog/verify-your-ride/ |
| S10 | Uber — Safety Toolkit (AU newsroom; TW blog) | https://www.uber.com/en-AU/newsroom/safetytoolkitau/ ; https://www.uber.com/en-TW/blog/safety-toolkit/ |
| S11 | Uber Eng — Engineering Uber Lite | https://eng.uber.com/blog/engineering-uber-lite/ |
| S12 | Uber — Audio recording for riders; India newsroom (RideCheck) | https://www.uber.com/jp/en/ride/safety/audio-recording/ ; https://www.uber.com/bh/en/blog/audio-recording ; https://www.uber.com/en-IN/newsroom/new-safety-features |
| S13 | Uber LK — Cash in Colombo (2016); UberTUK launch; LK daily payments | https://www.uber.com/en-LK/blog/cashincmb ; https://www.uber.com/en-LK/blog/launching-ubertuk-sri_lanka-your-tuk-tuk ; https://www.themorning.lk/articles/4P8KH5kWo8SOIB07D7yh |
| S14 | Uber — Authorisation holds (GB, NZ, LK) | https://www.uber.com/en-GB/blog/what-is-an-authorization-hold ; https://uber.com/en-LK/blog/auth-hold |
| S15 | Uber — How surge pricing works (driver) | https://www.uber.com/au/en/drive/driver-app/how-surge-works/ |
| S16 | (3P, old) TechCrunch — Suggested pickup points (2015) | https://techcrunch.com/2015/07/08/uber-suggested-pickup-points |
| S17 | Uber developer — Pickup refinement / Zones | https://developer.uber.com/docs/guest-rides/guest-ride-api-build-guide/pickup-refinement |
| S18 | Uber Help — Splitting a fare; Split Fare (GB) | https://help.uber.com/h/2ccba301-152e-4747-b207-e4281a1a2ba5 ; https://www.uber.com/gb/en/ride/how-it-works/split-fare |
| S19 | Uber — How to improve your rating; (3P) TechCrunch Compliments (2016) | https://www.uber.com/us/en/blog/how-to-improve-your-uber-rating/ ; https://techcrunch.com/2016/11/21/uber-debuts-compliments-letting-riders-give-drivers-specific-positive-feedback/ |
| S20 | Uber newsroom — New rider app (2016) | https://www.uber.com/newsroom/newriderapp |
| S21 | Uber — Uber Reserve FAQs (AU) / Reserve (CZ) | https://www.uber.com/en-AU/blog/faqs-about-uber-reserve/ ; https://www.uber.com/en-CZ/blog/reserve |
| S22 | Uber — Features for deaf/hard-of-hearing partners; GAAD 2024 | https://www.uber.com/blog/exciting-app-features-for-deaf-and-hard-of-hearing-partners/ ; https://www.uber.com/us/en/newsroom/globalaccessibilityawarenessday2024 |
| S23 | (3P) 9to5Mac / Engadget — Uber Live Activities | https://9to5mac.com/2022/12/09/uber-and-uber-eats-live-activities/ ; https://www.engadget.com/uber-ride-tracker-iphone-lock-screen-060908235.html |
| S24 | Uber newsroom — GO-GET 2026 (2026-04-29) | https://www.uber.com/us/en/newsroom/go-get-2026/ |
| S25 | (3P, old) Axios / The Points Guy — 2018 driver app redesign | https://www.axios.com/2018/04/10/uber-overhauls-its-1523354239 ; https://thepointsguy.com/news/ubers-redesigned-app-shows-drivers-where-they-can-get-peak-fares |
| S26 | Uber newsroom — Only on Uber (upfront fares + destination for drivers, Trip Radar, 2022) | https://www.uber.com/us/en/newsroom/only-on-uber/ |
| S27 | Uber newsroom — Refining the earnings experience (2016) | https://www.uber.com/gb/en/newsroom/refining-the-earnings-experience-2/ |
| S28 | (3P) The Rideshare Guy — Quests | https://therideshareguy.com/the-best-strategies-and-hacks-for-uber-quest/ |
| S29 | Uber — Instant Pay | https://www.uber.com/us/en/drive/driver-app/instant-pay/ |
| S30 | Uber newsroom — Driver destinations; (3P) Rideshare Guy destination filter | https://www.uber.com/us/en/newsroom/driver-destinations-2/ ; https://therideshareguy.com/understanding-the-destination-filter/ |
| S31 | Uber Eng — Mapping accuracy with CatchME (HMM map matching, 2019) | https://www.uber.com/blog/mapping-accuracy-with-catchme/ |
| S32 | NSDI '26 — Uber's Failover Architecture (UFA) | https://www.usenix.org/conference/nsdi26/presentation/bansal |

### Not verified in this pass
- Exact current (2026) rider-app layout, tile set and sheet snap heights — no primary design source found; layout rows tagged (UV) are observed patterns.
- Uber "cell-based architecture" — no Uber primary source found; only third-party articles.
- Base design system bottom-sheet component / type scale — only third-party teardowns; Base Web is the open React implementation.
- Smooth car-marker interpolation — observed behaviour, no Uber engineering post found.
- Driver destination-filter daily limit (2) — third-party only; Uber does not state it in the cited post.
- Whether Uber LK offers PIN, audio recording, Live Activities or tipping — country availability not checked.
