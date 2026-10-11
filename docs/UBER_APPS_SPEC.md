# Uber rider + driver apps — screen-level spec for Taxios (pass 2)

Researched 2026-10-12. Complements `docs/UBER_PARITY.md` (pass 1; not edited). This is a **build spec**: for each
screen we describe layout, elements top-to-bottom, actions, data, gestures, transitions, empty/loading/error/offline
states and notifications. Everything is a *pattern*; the operator's brand (T-007) supplies all colour, logo, type and
copy. No Uber marks, colours, icons, sounds or exact layouts are to be reproduced.

Legend: `[Sn]` → §5 sources (all accessed 2026-10-12). **(UV)** = not confirmed by a primary source (observed
behaviour / third-party only). **(old)** = primary source ≥ 3 yrs old. **(US)**/**(UK)**/**(AU)**/**(IN)** = market the
source describes; Uber LK availability was not checked for any row. Layout vocabulary: **Map+Sheet** = full-bleed map
behind a bottom sheet with snap points *C* (collapsed: title + CTA), *H* (half), *F* (full); **FS** = full-screen
page; **Modal** = dialog/sheet over the current screen. Sheet snap heights and gestures are **(UV)** unless tagged.

Verified 2026 state of the apps (primary): rider home = quick-access icons (Ride, Hotels, Food, Reserve) + saved
locations + bottom nav Home / Services / Activity / Account; "Where to?" bar redesigned as one search; AI voice
booking announced [S1]. Driver app 2026: AI home summary, Trip Guide (pickup + drop-off + map preview + estimated
earnings in one view), Discover (promotions/reservations/airports/events), Tracker (earnings + Pro status + Quests),
voice assistant, trip-change clarity (new drop-off + fare change shown), upfront-fare protection, chat protections,
metrics fairness review, Points Pass [S2][S3][S4].

---

## 1. Rider app

### 1.1 Onboarding & account

| # | Screen | Layout | Elements top→bottom | Actions | Data / rules | States & notifications | Src |
|---|---|---|---|---|---|---|---|
| R-ON1 | Welcome | FS | Brand mark; one-line value prop; **phone number field with country code** (primary); "Continue"; social/email alternative links; legal line (Terms/Privacy) | Continue → OTP | Phone is the identity in most countries; email/social are alternatives | Offline: inline banner, keep input | [S5] |
| R-ON2 | OTP | FS | "Enter the 4–6 digit code sent to +94…"; code boxes (auto-read SMS on Android); resend timer; "didn't get a code?" help link | Auto-submit on last digit | Code never requested by support (anti-phish copy) | Wrong code: shake + inline error; resend after cooldown | [S5] |
| R-ON3 | Name (+ email) | FS | First name, last name, optional email; "first name is shown to drivers" helper | Next | First name + star rating + verified badge are what drivers see | — | [S5][S6] |
| R-ON4 | Consent | FS / Modal | Terms + privacy acceptance; notification permission prompt follows first booking, not at sign-up (UV) | Accept | — | Location permission requested on Home with rationale | [S5] |
| R-ON5 | Rider verification (US) | FS flow | Auto cross-check of name/phone vs third-party data; fallback upload gov ID + selfie; result = blue **verified badge** shown to drivers pre-accept and en route | Verify / Skip | Optional; anonymous payment methods may force ID | Badge lost if name/phone/email changes | [S6] |

### 1.2 Home → request

| # | Screen | Layout | Elements top→bottom | Actions | Data | Gestures / transitions | States / notifications | Src |
|---|---|---|---|---|---|---|---|---|
| R-H1 | Home | Map+Sheet (sheet ≈ H by default; map shows nearby vehicles as small category icons) | Status/promo banner (optional); **"Where to?" search bar** with a "Later/Reserve" time pill at its right; quick-access icon row (Ride, Reserve, …); saved places rows (Home, Work, +add); recent/suggested destinations (2–3 rows, map-pin icon); promo/feature cards carousel; **bottom nav: Home · Services · Activity · Account** | Tap search → R-S1; tap saved/recent → skips search, goes to R-P1 with pickup = current location; tap time pill → R-RS1 | Nearby supply from a lightweight "vehicles near me" feed (jittered positions, UV); recents from trip history | Sheet drag C↔H↔F; map pan allowed behind sheet; long press on map (UV) | Loading: skeleton rows, map still rendered; no location permission: sheet shows "turn on location" row and pickup falls back to typed search; offline: top banner, recents still tappable (cached, Lite lesson) | [S1][S7][S8] |
| R-S1 | Destination search | FS (keyboard up) | Back; two stacked fields — pickup (prefilled "Current location", editable) and drop-off (focused); connector dots between them; "+" add stop (multi-stop up to N); "Set location on map" row; "Search for a Home/Work" rows if unset; results list: saved → recent → live search hits, each row = icon, name, secondary address line, distance (UV) | Pick row → R-P1 (if pickup is current location, pickup confirm may be skipped when GPS accuracy is good); "Set on map" → R-S2 | One search now returns rides + food + items (not relevant to us) | Keyboard dismisses on pick; slide-in from bottom | Empty query: shows saved + recents; no results: "No results — set location on map"; offline: cached recents only, search disabled with hint | [S1][S9] |
| R-S2 | Pin-on-map | Map (full) + small bottom card | Back; fixed centre pin; address label above pin (reverse-geocoded, updates on pan end); card: resolved address + "Confirm pickup/destination" | Pan map under pin; confirm | Reverse geocode debounced ~300 ms after pan end (UV) | Pin lifts on drag start, drops on end (UV) | Geocode fail: show coordinates + "Use this location" | [S9] (old) |
| R-P1 | Products / price sheet | Map+Sheet (route drawn pickup→dropoff; map fits route with sheet inset; sheet H by default, F to compare) | Drag handle; "Choose a ride" title; product rows: category icon, name, seat count, **drop-off time or pickup ETA**, upfront **price**, optional badge (faster / cheaper / "recommended", UV); selected row highlighted with a one-line description; **payment row** (method icon + name + chevron) above the CTA; CTA "Choose <product>" (full width) | Select row; tap payment → R-PAY; scroll list; CTA → R-P2 or request | Upfront price = base + time + distance (expected route) + dynamic + tolls/fees; shown before request; re-quoted if destination/stops/route change materially | Sheet snaps C/H/F; list scrolls inside at F; CTA pinned footer | Loading: skeleton rows + grey route; no supply: row shows "No cars nearby" and CTA disabled with "Schedule instead"; quote expired (UV ~minutes): re-fetch silently | [S10][S11] |
| R-PAY | Payment & promo | Modal (half sheet) | Methods list (cash is first-class in cash markets); radio selection; "Add payment method"; "Add promo code" row; business/personal profile toggle (if set) | Pick → return to R-P1 | Cash: rider pays driver the in-app amount; no tip flow on cash | Promo invalid: inline error, stays open | [S12][S13] |
| R-P2 | Confirm pickup | Map (zoomed to pickup) + card | Draggable pin (map pans under fixed pin); pickup label; **suggested pickup spots** as dots the pin snaps to; venue/airport: list of designated zones instead of free pin; card: pickup address + optional "notes for driver" + CTA "Confirm pickup" | Pan to adjust (within radius, else re-quote); confirm → R-M1 | Suggested spots derived historically from where drivers actually stopped | Snap haptic (UV) | Pickup moved > X m: price re-quoted and CTA label updates | [S9][S14] (old) |

### 1.3 Matching → trip → end

| # | Screen / state | Layout | Elements top→bottom | Actions | Data | Transitions | States / notifications | Src |
|---|---|---|---|---|---|---|---|---|
| R-M1 | Searching | Map+Sheet (C) | Pulsing ring at pickup; sheet: "Finding your ride" + progress bar; requested product + price; "Cancel" (plain) | Cancel = free before match | Matching is batched over a few seconds and minimises *average* wait, not nearest-first; avoids 1★ pairs | Enters after confirm; exits to R-E1 on `accepted`, or to R-NF | No driver: "No drivers available — try again / schedule / other product"; app backgrounded: push "Driver found" on match | [S15][S16] |
| R-E1 | Driver en route | Map+Sheet (C/H) | Map: driver icon gliding along road, route to pickup, ETA bubble; sheet top: **"Meeting at <pickup> · N min"** (or "Arriving in N min"); driver card: photo, first name, rating, vehicle colour/make/model, **plate in large type**, verified/compliments chips; action row: **Call** (masked), **Message**, **Share trip**, **Safety (shield)**; "Add PIN"/PIN digits if enabled; secondary: change pickup note, cancel (with fee warning when past window) | Cancel → R-CX; call/message; share → OS share sheet with live link | Driver location refresh every few seconds; ETA re-computed server-side | On driver `arriving` (≤ ~0.2 mi / 320 m) sheet copy changes to "Driver is nearby" | Push: "Driver assigned"; "Arriving soon"; iOS Live Activity starts here (driver name/photo, plate + vehicle photo, time-to-pickup, progress bar whose last 20 % = final 2 min); rider cancel after 2 min from match (Economy) → fee | [S17][S18][S19][S20] |
| R-A1 | Driver arrived (+PIN) | Map+Sheet (H) | "Your driver has arrived" headline; wait-timer text ("free wait ends in m:ss", then "wait fee applies"); **PIN** (4 digits, large) with "Tell the driver this PIN" copy if enabled; plate/vehicle card; call/message/share/safety; cancel (no-show fee warning) | Tell PIN verbally; driver enters it; trip starts only on match | Wait fee starts 2 min after arrival (5 min Black; 2 min at airports); not charged if a cancel/no-show fee is charged; grace & no-show windows both start at GPS-determined arrival | Push "Your driver has arrived" + Live Activity switches to plate-first layout | Driver can cancel as no-show after ≥ 5–7 min (Economy) → fee to rider | [S19][S20][S21][S22] |
| R-T1 | On trip | Map+Sheet (C) | Map: car = you, route to destination, remaining ETA; sheet: **"Arriving at h:mm"** + destination line; actions: **Add/Change destination** (re-priced; driver sees new drop-off + fare change), **Share trip**, **Safety shield** (emergency, report, audio record), split fare (card markets), driver card collapsed | Change destination → search → re-quote → confirm | Final price keeps upfront unless destination/stops/route change materially, or wait/tolls/cleaning added; "upfront fare protection" means a *nearby* lower destination still pays original fare | Live Activity shows drop-off time | RideCheck: long unexpected stop / crash / off-route → "Are you OK?" prompt to both parties | [S2][S10][S23] |
| R-X1 | Trip end: rate & pay | FS (slides up over map) | Fare total + method (cash: "Pay driver Rs X in cash"); 5-star control; < 5★ → "what went wrong" chips; 5★ → **compliments** chips; tip (card markets); "Submit"; link to receipt | Submit → Home; skip allowed | Ratings are anonymous; driver average = last 500 rated trips; 1★ pairs never re-matched | Receipt push/email after completion | Cash markets: receipt flags "paid in cash"; cleaning fee can be added post-trip after review (max USD 225 US) | [S24][S25][S26] |
| R-CX | Cancel confirmation | Modal | Reason chips (driver not moving, wrong pickup, …); fee line ("Rs X cancellation fee" or "free"); "Cancel ride" (destructive) / "Keep ride" | — | Fee = max(city minimum, time+distance from match to pickup); waived if driver is late vs ETA/not moving (old UK) | — | Dispute later via Help → "Review my cancellation fee" | [S22][S27] |
| R-NF | No driver found | Sheet | Message; buttons: Try again (same product), Try another product, Schedule, Call operator (our addition) | — | — | — | — | (UV) |

### 1.4 Activity, receipts, Reserve

| # | Screen | Layout | Elements | Actions / rules | Src |
|---|---|---|---|---|---|
| R-AC1 | Activity | FS (tab) | Segments Upcoming / Past; rows: map thumbnail, destination, date-time, price, status badge (completed / cancelled / scheduled) | Tap → R-AC2; upcoming rows show "Edit / Cancel" | [S1][S7] |
| R-AC2 | Trip detail | FS | Static map with route; date; driver (photo, name, plate); fare breakdown (base, time, distance, dynamic, fees, wait, tolls, promo, total, method); buttons **Get help**, **Resend receipt**, **Download PDF/invoice**, **Rebook/Return trip**, lost item | Help opens trip-scoped help | (UV) |
| R-AC3 | Receipt (email/in-app) | Doc | Operator header, trip id, times, route, breakdown, payment, support link | — | (UV) |
| R-RS1 | Reserve (schedule) | Modal → sheet | Date + time pickers (**up to 90 days** ahead; suggested ≥ 2 h); airport mode: airline + flight number → pickup time suggestion; product list with Reserve pricing; CTA "Reserve" | Price locked (no surge) but changes on stops/destination/route/toll changes; driver wait included (≈ 5 min); ride confirmed only when driver details arrive (driver assigned ahead of time); cancellation fees higher than on-demand (AU/IN: free > 1 h before) | [S28][S29] |
| R-RS2 | Upcoming reservation card | Home banner + Activity | "Pickup tomorrow 05:30 · BIA" with driver details when assigned; edit/cancel | Rider notified when driver assigned and when en route | [S28] |

### 1.5 Account, safety, help, accessibility, profiles, Lite

| # | Screen | Layout | Elements / behaviour | Src |
|---|---|---|---|---|
| R-AC | Account | FS (tab) | Header: photo, name, rating; rows: Settings, Messages, Saved places, Safety, Family & teens, Business profile, Privacy, Legal, Help, Sign out, Delete account | [S1] (UV rows) |
| R-SET | Settings | FS | Personal info (name/phone/email), Safety preferences (PIN verification: always / scheduled day+time / off; Share-trip reminders: every trip / after an hour / when leaving venue; trusted contacts up to 5; audio recording), Accessibility, Notifications, Language, Appearance | [S30][S31] |
| R-SP | Saved places | FS | Home, Work, favourites (name + address + icon); add via search or map pin | [S7] |
| R-SF | Safety toolkit | Modal (from shield on every trip screen) | Rows: **Emergency** (calls local number; shows current address + trip details to read out, shares with services where supported), **Share my trip**, **Record audio** (encrypted on device; attachable to a report), **Safety centre**, **Report a safety issue** (on-trip discreet report, follow-up ≤ 24 h), **Check your ride** (plate/make/photo), Driver profile (rating, trip count, compliments) | [S30][S31] |
| R-HELP | Help | FS | Trip-scoped list (pick trip) → issue tree: lost item (driver call-back with rider number; return fee passed to driver, e.g. EUR 5 / GBP 25 if delivered), fare review (route deviation, wrong charge, cleaning fee dispute), cancellation fee review, safety report, lost receipt; chat/ticket with support | [S27][S32][S26] |
| R-A11Y | Accessibility | FS | Screen-reader labels; deaf/HoH driver matched → call disabled, text only, rider told driver is deaf/HoH and prompted for destination; WAV/Assist products in some cities | [S33] (old) |
| R-FAM | Family & teens | FS | Family profile with shared payment; invite member/teen (13–17) by contact; teen must complete safety onboarding; teen trips: always-on PIN, RideCheck (more sensitive), live tracking auto-shared to guardian, destination locked (driver cannot change), guardian notified on every request, guardian can contact driver; guardian may book for teen; highly-rated drivers only | [S34] (UK) |
| R-BIZ | Business profile | Toggle on R-PAY | Personal/Business badge; separate payment + expense code/memo; receipts to work email; weekly/monthly reports. Uber **Central** = coordinator web dashboard books for guests by name + phone + pickup/drop-off + vehicle + notes; guest gets SMS with driver/vehicle and a web tracking page, no app needed; scheduled up to 30 days; org billed monthly | [S35][S36] (old) |
| R-LITE | Uber Lite | Separate Android app | < 5 MB, 4 taps with little typing, maps hidden by default, city top places cached offline, landmark pickups when GPS fails, cash, emergency + share kept | [S37] (old/IN) |

---

## 2. Driver app

### 2.1 Onboarding, documents, vehicle

| # | Screen | Layout | Elements / behaviour | Src |
|---|---|---|---|---|
| D-ON1 | Sign-up | FS web/app | Phone + OTP, name, email, city; choose "drive" product; requirements checklist (licence, registration, insurance; profile photo; background check; vehicle inspection where required) | [S38] |
| D-DOC | Documents | FS (Account → Documents) | List: one row per required document with status chip **Pending / Approved / Rejected (reason) / Expiring / Expired**; tap row → "Take photo" (camera overlay with corner guides) or upload; tips: originals only, all four corners, not blurry/dark. Expired or rejected blocks going online until re-approved | [S38] |
| D-VEH | Vehicles | FS | Vehicle list (plate, make/model/year, colour, status); add vehicle → docs per vehicle; fleet-managed drivers pick an approved fleet vehicle before a shift if none is assigned | [S39] |
| D-FLEET | Fleet partner (Uber Fleet / Fleet Hub) | Web + app | Live map of drivers' locations and routes; call/message driver; add/remove/assign drivers to approved vehicles; metrics (trips, online hours, earnings) + pay statements; recruiting marketplace; 24/7 support. Only the fleet can add vehicles/vehicle docs; drivers may bring own vehicle | [S39][S40] |

### 2.2 Home, status, demand, preferences

| # | Screen | Layout | Elements top→bottom | Actions | Data / rules | States | Src |
|---|---|---|---|---|---|---|---|
| D-H1 | Home (offline) | Map (full) + top strip + bottom bar | Top-left: menu/profile; top-centre: **earnings pill** (today's net, tappable, hideable); top-right: Discover/inbox icon; map with **demand heat shading** + surge/Boost zones with amounts; **AI home summary card** ("busy near X until 19:00", upcoming events); bottom: big **GO** button; preferences (clipboard icon) → D-PREF; "sparkle" icon → voice assistant | GO → online; drag map | Heat map: zones shaded light→dark by surge level; surge shown as multiplier/add-on/included-in-fare by city | GPS off: blocking banner "Location needed"; docs expiring: yellow banner; blocked: red card with reason + "Fix" | [S2][S4][S41][S42] |
| D-H2 | Home (online) | Map + status strip | Status strip at bottom: "You're online · finding trips" with a **pause/hand button** (stop new requests), **Go offline**; queue position when in an airport FIFO zone ("#12 in queue, ~N drivers ahead"); Tracker chip (earnings + Pro + Quest progress) | Offline; open Trip Radar list; open preferences | Airport queue: FIFO by zone-entry time, not proximity; lose place by going offline, leaving zone, declining several in a row or cancelling; rider cancel keeps place; empty queue → nearest drivers outside zone | Waiting: subtle pulse on own marker; weak signal: "poor connection" chip | [S43][S44] |
| D-PREF | Trip preferences | Modal sheet | Toggles per product/trip type (rides, deliveries, long trips, Reserve); **Destination Mode** (set destination; Flexible vs Faster route; **2 uses/day**, 3 for Platinum/Diamond; a use starts on an Exclusive request; decline cap 2 (Faster) / 3 (Flexible) per session); Women Rider Preference; **Rider Rating Preference** (minimum rider rating); Area Preferences (Platinum+, 2 h/day) | Save | Exclusive declines still count toward acceptance rate | — | [S45][S46][S41] |
| D-RADAR | Trip Radar | Card/list over map | When off-trip: list/map of nearby open trips shared with several drivers; each row = pickup distance/ETA, earnings, trip length, destination area; tap **Match** on one or many; while > 10 mph only one request + route overview is shown | Match (raise hand); dismiss | Match not first-tap: ranked by ETA etc.; no acceptance-rate impact; cancelling after match counts; counts toward promotions; individual requests still arrive labelled **Exclusive** | — | [S47][S48] |

### 2.3 Offer → pickup → trip → end

| # | Screen / state | Layout | Elements top→bottom | Actions | Rules | Transitions / notifications | Src |
|---|---|---|---|---|---|---|---|
| D-OF1 | Trip request (offer) card | Full-width card over map (map shows pickup + route preview) | Top: product/trip type + tags (**Exclusive**, Long trip + est. time, Reserve, surge/Boost+ amount, verified rider badge); **upfront earnings** in largest type; pickup: ETA + distance ("4 min · 1.2 km"); trip: time + distance; pickup address, **destination address/area** (Gold+ in some markets); rider first name + **star rating**; **countdown** ring/bar (~15 s); big **Accept** (tap anywhere on card accepts in some markets); "No thanks"/× top-left; loud tone + vibration + flash (deaf mode) | Accept → D-NAV1; decline/ignore → expires, goes to next driver | Decline/expiry of an Exclusive request lowers acceptance rate (rolling last 100 exclusive requests; shown after first 10); low acceptance → warnings/temporary inability to go online, never permanent deactivation; cancellation rate = cancels ÷ accepts, same window; city max → alert, temporary block, deactivation if continued; "good" < 5 % (AU Pro: ≤ 3 %) | Back-to-back: next offer arrives while finishing current trip (route in second colour); can pause with hand button | [S49][S50][S51][S52][S53] |
| D-NAV1 | Navigate to pickup | Map (turn-by-turn) + top instruction banner + bottom trip card (C) | Next manoeuvre + distance; ETA to pickup; bottom card: rider first name + rating, pickup address, **Call / Message** (masked number), **Trip Guide** (pickup + drop-off + map preview + est. earnings in one view), "Arrived" when within geofence; overflow: cancel, report, safety, switch nav app | Nav choice: built-in Uber Navigation or hand-off to Google Maps/Waze (Account → App Settings → Navigation); voice: "message my rider I'm 3 minutes away" (US) | Rider sees `arriving` at ≤ ~0.2 mi | Auto "arriving" push to rider; arrival determined by GPS | [S2][S49][S54][S55] |
| D-WAIT | Waiting at pickup | Bottom card (H) | "Waiting for <name>" + **timer** counting from arrival; free-wait countdown then "wait fee applies" (2 min; 1 min on extra stops soon); **PIN entry** (4 boxes) if rider has PIN on; Call/Message with quick replies ("I'm here", "Where are you?"); **Cancel** → reasons; **no-show cancel** unlocks after threshold (≈ 5–7 min Economy / 10 Comfort / 15 premium) and pays the fee | Enter PIN → "Start trip" enabled | Wrong PIN: inline error, can't start; cancel with "rider no-show" after threshold does not count against cancellation rate | — | [S19][S21][S22][S30] |
| D-TRIP | On trip | Map nav + bottom card (C) | ETA/arrival time; destination; rider name; stops list (if multi-stop) with "Arrived at stop"; **trip change banner** ("Rider changed destination → Rs X new fare"); safety shield; **slide to complete** at destination (swipe, not tap) | Swipe start (at pickup) / swipe complete; add stop on rider's request | Upfront fare protection: nearby lower destination keeps original fare; chat protection: abusive rider message → driver may cancel without stats impact | Next offer may arrive (back-to-back) | [S2][S49] |
| D-END | Trip end | FS card | Earnings for trip (fare, surge/boost, tips pending); **rate rider** 1–5 (1★ auto-blocks future matches) + feedback tags; "Report a problem" (mess/cleaning fee within 3 days with photos + receipt; max USD 225 US); cash markets: **"Collect Rs X in cash"** screen with amount and "Collected" confirmation; then back to online map (or next trip card) | Done | 2-way ratings; driver can review "Satisfaction/Cancellation rate" fairness adjustments | Receipt to rider | [S2][S26][S41] |

### 2.4 Earnings, promotions, Pro, payouts

| # | Screen | Layout | Elements / behaviour | Src |
|---|---|---|---|---|
| D-ER1 | Earnings (tab) | FS | Header: **balance** + "Cash out"; period selector (day / week, Mon–Sun) with bar chart; totals: trips, online hours, **net earnings**; list: itemised trips (time, route, earnings) → D-ER2; sections: Promotions, Tips, Consecutive-trips, Adjustments, Expenses (tolls); "Earnings details"; weekly statement every Monday (first page: total earned + cash-outs; then trip earnings / promotions / tips; weekly service fee section; negative service fee possible when promotions exceed fees) | [S56][S57] |
| D-ER2 | Trip earnings detail | FS | Rider price → your earnings; lines: base/time/distance, surge or Boost+, wait time, tolls, tips (added later), cancellation/no-show fee, **service fee** (= rider price − driver earnings, excluding tips/tolls/certain taxes); map thumbnail; "Get help with this trip" | [S57] |
| D-PROMO | Promotions / Opportunities | FS lists | **Quest**: N trips in window → bonus (e.g. 40 trips/week → +$40; 30 trips/3 days → +$100), progress tracker at top of map, swipe to see trips remaining; **Boost+**: sign up for a session (closes 1 h before), extra per trip starting in blue zones in the window; **Consecutive trips**: start a trip in zone during window, complete 2 more without declining/going offline → bonus; rider cancels don't break streak; counts toward Quest; surge: zone amounts on map | [S58][S59] |
| D-PRO | Uber Pro | FS | Tiers Blue (default) / Gold / Platinum / Diamond; points per trip (1 off-peak, 5 at peak hours; none for cancels); 3-month program periods, status carries one period; thresholds per city shown in-app; quality gates: rating ≥ 4.85, cancellation ≤ 3 % (AU), acceptance minimum per city; rewards by tier (fuel/EV discounts, trip duration + direction shown on offers from Gold, extra destination uses + priority airport rematch + status protection at Platinum, priority support at Diamond); **Points Pass** protects status during a break | [S41][S50][S60] |
| D-PAY | Cash out | Modal | Instant Pay to debit card up to 6×/day (US, fee); weekly auto-deposit otherwise. **n/a for Taxios cash-first** — see §4 | [S61] |
| D-TRACK | Tracker | Chip on map → sheet | Earnings today, Pro status, active Quests in one view | [S2] |

### 2.5 Account, support, accessibility

| # | Screen | Elements / behaviour | Src |
|---|---|---|---|
| D-ACC | Account | Profile photo + name + rating + Pro tier; Documents (with expiry chips), Vehicles, Payment, App settings (Navigation app choice, sounds, night mode, deaf/HoH mode), Privacy, Tax info, Help, Sign out | [S38][S54] |
| D-STATS | Ratings & metrics | Star rating (last 500 rated trips), compliments + feedback themes, acceptance rate, cancellation rate with counts used, "review this metric" (fairness), city minimum warnings and quality-course prompts | [S50][S51] |
| D-ALERT | Document expiry | Status chip "Expiring" on the document row + home banner; expired → cannot go online until resubmitted and approved. (Proactive push reminder: UV) | [S38] |
| D-SUP | Help | Trip-scoped issues (fare adjustment, rider made a mess, rider no-show, unsafe rider → "Trip Issues and Adjustments"), account issues, 24/7 safety line, Report Hub lessons | [S26][S49][S2] |
| D-SAFE | Safety toolkit (driver) | Emergency button with location share; Follow My Ride; **Record My Ride** (selfie cam + mic, encrypted; recordings reviewable by driver for eligible incidents); speed-limit alerts; anonymous calls; GPS tracking of every trip; crash report + claim; Women Rider Preference; Rider Rating Preference; 1★ auto-block; (India 2026) Ambulance Assistance, trip-verification PIN | [S41][S62] |
| D-A11Y | Deaf/HoH mode | Flashing light for requests; rider call disabled, text only; rider prompted for destination and told driver is deaf/HoH | [S33] (old) |
| D-VOICE | Voice assistant | Off-trip: sparkle icon → mic: "Where is it busy?", "How much did I make this week?", "How long is the airport queue?"; on-trip: mic on map → "Message my rider that I'm 3 minutes away", "Call my rider" (confirm first); report road conditions. US, English, rolling out Sep 2026 | [S4] |

---

## 3. Shared platform behaviours

| Topic | Uber behaviour (as observed in the apps) | Taxios rule | Src |
|---|---|---|---|
| Realtime | Persistent push stream to both apps (seq numbers, resend on reconnect); Live Activity updates sent by a backend "out-of-app" service via APNs with server-side debouncing/throttling of state changes | T-031 WebSocket + outbox; a server-side "notification policy" decides which booking state changes fan out to push/Live Activity (debounce ≤ 1/5 s) | [S20], pass 1 S5 |
| Location cadence | Trip GPS ≈ every 4 s; arrival/arriving geofence ≈ 0.2 mi; GPS determines wait-timer start | Driver: 4–5 s on trip, 10–15 s idle; arrival geofence 100 m (operator knob); wait timer = server time of `arrived` event, not client clock | [S19][S21], pass 1 S3 |
| Upfront price + adjustments | Locked at request; changes on stops/destination/material route or duration change/un-factored toll; wait fee per minute after grace (2 min; Black 5; airports 2; 1 min on extra-stop trips soon); cleaning fee post-trip with evidence (≤ USD 225); upfront protection when a nearby new destination lowers the fare; driver sees new drop-off and fare delta | Quote locked; re-price on destination/stop change (both parties see delta); GPS-final only if deviation > operator threshold; wait charge per minute after grace (operator knob); cleaning fee = operator-raised adjustment with photo, audited | [S2][S10][S21][S26] |
| Cancellation (rider) | Free before match; fee if > 2 min after match (Economy) / 5 min premium; fee = max(city min, time+distance since match); waived when driver late/not moving (old) | Operator knobs: free window (default 2 min), flat fee, waiver when driver ETA slips > N min; cash → fee becomes rider balance owed, collected on next trip or waived | [S22][S27] |
| Cancellation (driver) / no-show | May cancel after waiting ≥ 5–7 min (Economy) → no-show fee; redispatch on driver cancel; wait fee not charged on top of a cancel fee; no-show cancels excluded from driver cancellation rate | Same; dispatcher can override; "rider no-show" reason requires `arrived` + threshold elapsed | [S22][S51] |
| Ratings maths & thresholds | Driver rating = mean of last 500 rated trips; city minimum → alert → course → deactivation; cancellation rate = cancels ÷ accepts (rolling 100); acceptance rate = accepted exclusive ÷ exclusive received (rolling 100, after 10); 1★ pairs blocked; rider deactivation threshold undisclosed | Store per-trip ratings both ways; compute rolling-500 mean + rolling-100 accept/cancel rates nightly; operator-set floors trigger dispatcher alerts, never auto-deactivation; 1★ exclusion in dispatch SQL | [S50][S51][S52] |
| Fraud signals | GPS spoofing (fake-location apps), inflated distance/time, accepting with no intent to complete, dummy accounts, false fees; ML feasibility model on velocity/acceleration/elevation; human review before deactivation; warnings first | Server-side plausibility: speed > 150 km/h between pings, altitude jumps, mock-location flag on Android (`isMock`), distance ratio GPS vs routed > 1.5 → flag for operator review; never auto-block | [S63][S64] |
| ETA display | Pickup ETA in minutes on product rows and en-route sheet; on-trip shows arrival clock time; Live Activity progress bar last 20 % = last 2 min | Minutes for pickup, clock time for drop-off; round up; never show "0 min" — use "Arriving" | [S20] |
| Live Activities / ongoing notification | iOS Lock Screen + Dynamic Island: driver name/photo, plate + vehicle photo, time-to-pickup, progress bar, drop-off time; push-driven; reduced pickup cancellations ~2 % | v1.1 via EAS dev build: iOS ActivityKit + Android foreground notification with the same 4 fields | [S20] |
| Deep links | Universal links for trip share page, promo codes, Help articles; guest riders get an SMS web-tracking page (Central) | `https://<tenant-domain>/t/<token>` share page (public, tokenised), `taxios://booking/<id>` in-app | [S36] |
| App size / perf | Rider iOS ≈ 400 MB install, Driver iOS ≈ 408 MB (US); Lite < 5 MB; Lite target: usable on sub-3G, screen transitions < 300 ms | Budget: ≤ 60 MB per app, cold start ≤ 2 s on a 2019 Android, booking payloads ≤ 10 KB, cache last state for offline open | [S37][S65] |

---

## 4. Delta — Taxios today → Uber parity

Source of truth for "today": `apps/mobile/src/passenger/PassengerApp.tsx` and `apps/mobile/src/driver/DriverApp.tsx`
(read 2026-10-12). Both apps are **form-first (no map)**, poll every 4 s, foreground-only, no push.
Priority: **MVP** (before first operator live) · **v1.1** (first 3 months) · **later** · **n/a-LK** (cash-first: tips,
split fare, Instant Pay are n/a; drivers owe commission to the operator instead). "Knob" = per-tenant setting.

### 4.1 Rider

| Screen | Taxios today | Missing for parity | Priority | Operator knobs |
|---|---|---|---|---|
| R-ON1–4 Sign-up | Operator code → name, phone, email, **password** form; sign-in by email+password | Phone-first + **OTP** (SMS), no password; name only; terms consent line; email optional | **MVP** | OTP sender id; terms/privacy URLs (have) |
| R-ON5 Verification | — | Optional ID/selfie badge | later | on/off |
| R-H1 Home | Single card: pickup/drop-off fields, category chips, when chips, history list | **Map-first home** with nearby vehicles, "Where to?" bar, saved places, recents, bottom nav (Home/Activity/Account) | **MVP** | Show nearby vehicles on/off; jitter radius |
| R-S1 Search | `PlaceSearch` component per field | Two-field full-screen search with recents/saved, "set on map", add stop | **MVP** (stops v1.1) | Search provider; landmark POI set |
| R-S2 Pin on map | "Use my current location" only | Fixed-pin pan map + reverse geocode | **MVP** | — |
| R-P1 Products | Chips `name · seats`; "See fare" button → breakdown | Product **rows with upfront price + ETA per category** fetched in one call (quote all categories), route on map, payment row, pinned CTA | **MVP** | Categories per operator (have); ETA display on/off; fixed vs estimated label |
| R-PAY Payment/promo | Cash hard-coded | Payment row (cash default, card later), promo code field | v1.1 (promo T-142) | Methods enabled; promo rules |
| R-P2 Confirm pickup | — | Pin-adjust + driver note (note exists at booking) | **MVP** | Pickup-move re-quote radius |
| R-M1 Searching | Status text "Finding you a driver…" + cancel | Pulse on map, progress, product/price summary, "call operator" fallback, no-driver sheet | **MVP** | Operator hotline (have `supportPhone`); search timeout |
| R-E1 En route | Driver name, plate, vehicle text, "location is live" flag | **Live driver marker + route + ETA**, driver photo + rating, **call/message**, **share trip**, safety shield, cancel-fee warning | **MVP** (chat v1.1) | Call mode: direct / operator hotline / masked (later); share page domain |
| R-A1 Arrived + PIN | Status text only | Arrived banner + push, wait timer, **PIN display**, no-show warning | **MVP** | PIN: off / all / night-only (hours); free wait minutes; wait fee/min |
| R-T1 On trip | Status text | Route + arrival clock, change destination (re-quote), share, safety | **MVP** (change dest v1.1) | Deviation threshold %; allow destination change |
| R-X1 End / rating | Receipt + 1–5 chips + comment | Fare-first end screen, "pay driver Rs X cash" copy, < 5★ reason chips, 5★ compliments, SMS receipt option; tips n/a | **MVP** | Chip sets (si/ta/en); SMS receipts on/off |
| R-CX Cancel | Alert with confirm | Reason chips + fee line | v1.1 (T-141) | Free window, fee, waiver rule |
| R-AC1/2 Activity | Last 10 bookings in a list; detail = same Active view | Upcoming/Past segments, map thumbnail, breakdown, resend receipt, rebook, help entry | **MVP** (thumbnails later) | — |
| R-RS1 Reserve | Chips now/+30m/+1h/+2h/tomorrow; server starts search 15 min before | Date/time picker ≤ N days, airport mode (BIA flight), driver-assigned-ahead notifications, Reserve cancel rules | **MVP** (picker) / v1.1 (flight) | Max days ahead; lead time; assign-ahead minutes; Reserve cancel fee |
| R-AC Account/settings | Sign out, legal links | Profile edit, saved places, safety prefs, language (si/ta/en), notifications, delete account | **MVP** | Languages enabled |
| R-SF Safety | — | Shield → Emergency (119/1990 + dispatcher alert), share trip, report issue; audio recording later | **MVP** | Emergency numbers; dispatcher alert channel |
| R-HELP Help | tel/mailto links | Trip-scoped help: lost item (driver call-back), fare review, cancellation review, safety report → operator console queue | v1.1 | Support hours; lost-item fee |
| R-A11Y | `accessibilityRole` on rows | Labels on markers/CTAs, 44 pt targets, Dynamic Type, deaf/HoH driver handling | v1.1 | — |
| R-FAM Family/Teen | — | Family profile, teen accounts with always-on PIN + guardian tracking | later | Min age; allowed |
| R-BIZ Business / Central | Web console books on behalf (dispatcher) | Business profile toggle in app; guest SMS tracking link for console bookings | v1.1 (guest SMS) / later (profile) | Corporate accounts |
| R-LITE | n/a | Keep our app light instead: cached recents, landmark pickups, small payloads | **MVP** (principles) | — |
| Notifications | None (polling) | Push: assigned / arriving / arrived / started / completed / receipt; Live Activity v1.1 | **MVP** (T-153) | Templates per language |

### 4.2 Driver

| Screen | Taxios today | Missing for parity | Priority | Operator knobs |
|---|---|---|---|---|
| D-ON1 Sign-up | Email + password from operator invite; forgot-password web link | Phone-first + OTP sign-in; self-apply flow (D-021 pattern for drivers) with checklist | v1.1 | Self-apply on/off; required doc types |
| D-DOC Documents | Inline card: list with status/review note, add doc (type, number, expiry, camera/photo) | Dedicated screen; **Expiring** state + home banner + push reminder N days before; camera corner guides; per-vehicle docs | **MVP** (expiring/banner, T-130) | Doc types + validity; reminder days |
| D-VEH Vehicles | Plate shown; assignment by operator | Vehicle list/pick-before-shift for fleet vehicles | v1.1 | Driver may add own vehicle on/off |
| D-FLEET | Operator web console (zones, drivers, vehicles, dispatch) | Live fleet map, call/message driver, driver metrics (trips/hours/earnings), pay statements, roster — our console = Uber Fleet Hub | **MVP** (live map + metrics in console) | — |
| D-H1/H2 Home | Card: name, availability badge, plate, GO button, GPS status text | **Map-first home**, earnings pill, status strip with Go offline + pause, demand shading (later), blocking banners for GPS/docs | **MVP** (map, pill, strip) / later (heat) | Heat map on/off |
| D-PREF Preferences | — | Destination mode (limited uses), trip-type toggles, rider-rating floor | v1.1 | Uses/day; allowed filters |
| D-RADAR Trip Radar | — | Open-trip list to claim (stand-style) | later | On/off (conflicts with dispatcher) |
| D-OF1 Offer card | Countdown badge (s), pickup addr + distance, drop-off addr, passenger first name, estimate, Accept/Decline | Full-width card with **earnings in large type**, pickup ETA, trip time/distance, destination area, rider rating, tags (scheduled/long trip), countdown ring, **sound + vibration + flash**, tap-to-accept; delivered-ack; consequences copy (acceptance rate) | **MVP** | Offer TTL (have); show rider rating; sound on/off; acceptance-rate tracking on/off |
| D-NAV1 Navigate | External Google Maps links per leg; status buttons | In-app map with route + ETA, nav app chooser (Google/Waze/in-app later), rider call/message, **auto "arrived"** by geofence with manual fallback, "arriving" push to rider | **MVP** (map, chooser, call) / v1.1 (auto-arrive) | Nav app default; geofence m |
| D-WAIT Waiting | "I've arrived" then "Passenger on board — start trip" buttons | **Wait timer**, free-wait → fee state, **PIN entry**, quick-reply messages, **no-show cancel** unlocked after threshold (fee), reasons | **MVP** (timer, PIN, no-show) | Free wait min; wait fee/min; no-show minutes + fee; PIN policy |
| D-TRIP On trip | Buttons + "End trip early" with reasons | **Swipe to start / complete**, route + ETA, stops, trip-change banner with fare delta, back-to-back offer | **MVP** (swipe) / v1.1 (changes, back-to-back) | Allow back-to-back |
| D-END End + cash | CashCard: final fare lines, amount received, discrepancy note, confirm | **Rate rider** (stars + tags; 1★ block), trip earnings + **commission owed** line, report problem (mess/no-show), then return to online map | **MVP** (rating, commission line) | Commission %/flat per category; rating tags |
| D-ER1/2 Earnings | — | Day/week totals, trip list, per-trip detail (fare, wait, cancel fees, commission), **"owed to operator" balance + settle history** (replaces cash-out), weekly statement | **MVP** (T-132/T-144) | Settlement cadence; statement day |
| D-PROMO Promotions | — | Quest-style trip-count bonus; peak-hour bonus | later | Operator-defined, audited |
| D-PRO Tiers | — | Driver standing: rating, acceptance, cancellation with explanations and fairness review by dispatcher | v1.1 (metrics) / later (tiers) | Floors; review workflow |
| D-PAY Cash out | — | n/a: cash-first; commission settlement instead | n/a-LK | — |
| D-ACC Account | Switch business, sign out | Profile, ratings/metrics page, app settings (nav, sounds, language), help | **MVP** (basic) | — |
| D-SUP Help | — | Trip-scoped issues → operator console queue; call operator | v1.1 | Support channel |
| D-SAFE Safety | — | Emergency button + dispatcher alert; share trip; speed alerts later; recording later (PDPA review) | **MVP** (SOS) | Emergency numbers |
| D-A11Y Deaf/HoH | — | Flash alerts, text-only rider contact | v1.1 | — |
| D-VOICE | — | Not needed; consider large-button + si/ta voice prompts for low literacy | later | — |
| Location | Foreground 5 s / 10 m, 10 s min send, 30 s heartbeat | **Background location** (dev build), batching + replay on reconnect, mock-location flag | **MVP** (T-152b) | Idle/trip cadence |
| Airport queue | — | FIFO zone queue (BIA) with position | later | Zone polygon; FIFO on/off |
| Notifications | None | Push for offers (high priority + sound), trip changes, doc expiry, settlement | **MVP** (offers) | — |

### 4.3 Twelve largest changes (summary for the owner)
1. Both apps become **map-first** (map + bottom sheet on every live state) instead of forms.
2. Rider sign-up moves to **phone + OTP**, no password; name only.
3. Product selection = **rows with upfront price + ETA per category**, route on map, pinned CTA.
4. Rider live-trip screens: **driver marker, call/share/safety**, arrived banner with **PIN** and wait timer.
5. Driver offer card: **earnings-first**, pickup ETA/trip length/destination/rider rating, sound + vibration, countdown ring.
6. Driver pickup flow: in-app map, nav-app chooser, **auto-arrived geofence**, **wait timer + no-show cancel with fee**, PIN entry.
7. **Swipe** to start/complete; rate rider; commission-owed line on the cash screen.
8. **Earnings tab** with day/week, trip detail and an **"owed to operator" settlement ledger** (replaces cash-out).
9. **Push notifications** both sides (offers, assigned/arriving/arrived/receipt), Live Activity in v1.1.
10. **Background location** + offline replay of driver actions (needs EAS dev build, not Expo Go).
11. Activity/receipts, saved places, account/settings, safety toolkit (SOS → 119/1990 + dispatcher) in the rider app.
12. Operator knobs for all of the above: PIN policy, wait/no-show/cancel fees, categories + commission, nav default, languages.

---

## 5. Sources (accessed 2026-10-12)

| ID | Source | URL |
|---|---|---|
| S1 | Uber newsroom — GO-GET 2026: one app for everything (2026-04-29): home icons, bottom nav, One Search, voice bookings, Reserve | https://www.uber.com/us/en/newsroom/go-get-2026/ |
| S2 | Uber newsroom — Only on Uber 2026 (2026-09-24): Trip Guide, Discover, Tracker, trip-change clarity, upfront fare protection, chat protections, metrics fairness, Points Pass, earlier wait fees | https://www.uber.com/us/en/newsroom/only-on-uber-2026/ |
| S3 | (3P) The Rideshare Guy — Only on Uber 2026 explained | https://therideshareguy.com/only-on-uber-2026-new-features-for-drivers-explained/ |
| S4 | Uber blog — Voice in the Driver app (2026-09-17) | https://www.uber.com/us/en/blog/voice-in-the-driver-app/ |
| S5 | Uber Help — I did not receive the SMS verification when signing up; (3P) sign-up walkthroughs (UV order) | https://help.uber.com/h/588008d4-2b46-4952-b42c-f12ffcb04278 |
| S6 | Uber — Rider verification (US) | https://www.uber.com/us/en/safety/rider-verification |
| S7 | Uber newsroom — "We're redesigning the Uber app just for you" (2023-02-22): Services tab, Activity hub, saved places after "Where to?" | https://www.uber.com/us/en/newsroom/were-redesigning-the-uber-app-just-for-you/ |
| S8 | Uber newsroom — New rider app (2016) (old) | https://www.uber.com/newsroom/newriderapp |
| S9 | Uber developer — Guest Rides: pickup refinement / zones; (3P, old) TechCrunch suggested pickup points | https://developer.uber.com/docs/guest-rides/guest-ride-api-build-guide/pickup-refinement ; https://techcrunch.com/2015/07/08/uber-suggested-pickup-points |
| S10 | Uber — Upfront pricing (AU) / UK upfront pricing | https://www.uber.com/au/en/ride/how-it-works/upfront-pricing ; https://www.uber.com/gb/en/blog/uk-upfrontpricing |
| S11 | Uber newsroom — Only on Uber (2022): upfront fares + destination for drivers | https://www.uber.com/us/en/newsroom/only-on-uber/ |
| S12 | Uber LK — Cash in Colombo (2016) (old, LK) | https://www.uber.com/en-LK/blog/cashincmb |
| S13 | Uber — Authorisation holds (LK) | https://uber.com/en-LK/blog/auth-hold |
| S14 | Uber developer — Guest Rides dispatch & cancellation (states: processing/accepted/arriving ≤ 0.2 mi/in_progress/completed; fees) | https://developer.uber.com/docs/guest-rides/guest-ride-api-build-guide/dispatch-and-cancellation |
| S15 | Uber — Marketplace: Matching (batching, 1★ blocking) | https://www.uber.com/us/en/marketplace/matching/ |
| S16 | Uber Eng — Real-time push platform (RAMEN) | https://www.uber.com/blog/real-time-push-platform/ |
| S17 | Uber — Safety toolkit (AU newsroom) | https://www.uber.com/en-AU/newsroom/safetytoolkitau/ |
| S18 | Uber — Verify your ride (PIN) (AU) | https://www.uber.com/en-AU/blog/verify-your-ride/ |
| S19 | Uber Help — Wait time fees (2-min grace; Black 5; airports 2; not stacked with cancel fee; GPS arrival) | https://help.uber.com/h/22b6db22-0f86-4765-ac72-da549a2302d7 |
| S20 | Uber Eng — Live Activity on iOS (2024-07-25): per-state content, OOA push service, throttling, results | https://www.uber.com/blog/live-activity-on-ios/ |
| S21 | Uber Help — Review cancellation fee (rider/driver cancel policy; receipt check) | https://help.uber.com/h/6bec690f-ee35-40ba-96ee-c38a8ae796e0 |
| S22 | Uber UK blog — Changes to wait times (old); product thresholds via Guest Rides doc [S14] and help (2 min rider; 5–7 / 10 / 15 min driver) | https://www.uber.com/en-GB/blog/changes-to-wait-times |
| S23 | Uber IN newsroom — RideCheck & safety features | https://www.uber.com/en-IN/newsroom/new-safety-features |
| S24 | Uber — How to improve your rating (US) | https://www.uber.com/us/en/blog/how-to-improve-your-uber-rating/ |
| S25 | (3P, old) TechCrunch — Compliments (2016) | https://techcrunch.com/2016/11/21/uber-debuts-compliments-letting-riders-give-drivers-specific-positive-feedback/ |
| S26 | Uber Help — Rider made a mess (cleaning fee ≤ USD 225, 3 days, evidence) | https://help.uber.com/h/6a60bda7-d8a6-46b7-bd95-2cdf62c7e987 |
| S27 | Uber Help — Lost item return fee (EU); Uber PL blog — contact driver feature; UK £25 if delivered | https://help.uber.com/h/633c684e-1d4b-426a-bd4f-6bee488dac6a ; https://www.uber.com/pl/en/blog/feature-contact-with-driver |
| S28 | Uber — Reserve: how it works (90 days, locked rate, footnotes) ; AU scheduled rides (free cancel > 1 h) | https://www.uber.com/ride/how-it-works/reserve/ ; https://www.uber.com/au/en/ride/how-it-works/scheduled-rides/ |
| S29 | Uber IE blog — Schedule rides with Uber Reserve (5-min wait included, > 2 h) | https://www.uber.com/en-IE/blog/schedule-rides-in-advance-with-uber-reserve/ |
| S30 | Uber — Rider safety features (AU): Verify Your Ride, Share My Trip, RideCheck, on-trip reporting, driver profile | https://www.uber.com/en-AU/ride/safety/ |
| S31 | (3P) Time Out / Yahoo — scheduled PIN & share-trip automation (2024–25) | https://www.timeout.com/usa/news/uber-rolls-out-a-handful-of-new-features-to-help-riders-feel-safe-041024 |
| S32 | Uber (SA/GB) — Information about taking trips: request shows destination + fare, Long trip label, slide to start/complete, nav setting path, Help → Trip Issues | https://www.uber.com/sa/en/drive/basics/how-to-take-trips |
| S33 | Uber blog — Features for deaf/HoH partners (2015) (old); GAAD 2024 | https://www.uber.com/blog/exciting-app-features-for-deaf-and-hard-of-hearing-partners/ ; https://www.uber.com/us/en/newsroom/globalaccessibilityawarenessday2024 |
| S34 | Uber — Teen accounts (UK) | https://www.uber.com/gb/en/ride/teens/ |
| S35 | Uber newsroom — Business profiles (2015–16) (old) | https://www.uber.com/newsroom/business-profiles-3 |
| S36 | Uber for Business — Central (guest bookings, SMS tracking, 30-day scheduling) | https://www.uber.com/us/en/business/central/ |
| S37 | Uber Eng — Engineering Uber Lite; Uber Lite page (MX) | https://eng.uber.com/blog/engineering-uber-lite/ ; https://www.uber.com/mx/en/u/uber-lite-app |
| S38 | Uber Help — Setting up your driver account (Account → Documents, statuses, rejection reasons, expiring) | https://help.uber.com/h/01116d00-abb3-491c-bdbe-eb273d705f44 |
| S39 | Uber — Fleet management / Fleet Hub (DE) | https://www.uber.com/de/en/earn/fleet-management/ |
| S40 | Apple App Store — Uber Fleet (v1.4, 229.5 MB; live map, call/message, roster, metrics, pay statements) | https://apps.apple.com/us/app/-/id6502092547 |
| S41 | Uber — Driver safety (US): Women Rider Preference, Rider Rating Preference, PIN, Follow/Record My Ride, speed alerts, 1★ block | https://www.uber.com/us/en/drive/safety/ |
| S42 | Uber — How surge works (driver, AU) | https://www.uber.com/au/en/drive/driver-app/how-surge-works/ |
| S43 | Uber — Airport driver information pages (FIFO queue rules) | https://www.uber.com/global/en/r/airports/ktw/driver-information/ |
| S44 | Uber blog — Sea-Tac fraud notices (2018) (old) | https://www.uber.com/us/en/blog/sea-tac-fraud/ |
| S45 | Uber blog — Improved Destination Mode (2026-05-12): Flexible/Faster, 2 uses/day, decline caps | https://www.uber.com/us/en/blog/destination-mode/ |
| S46 | Uber — Driver destinations basics | https://www.uber.com/drive/basics/driver-destinations |
| S47 | Uber blog — All-new Trip Radar (2022-08-12) | https://www.uber.com/us/en/blog/all-new-trip-radar/ |
| S48 | Uber CA blog — Trip Radar (Match, Exclusive label, > 10 mph single request) | https://www.uber.com/ca/en/blog/trip-radar-canada |
| S49 | Uber CA blog — More flexibility for drivers (CA request card fields) | https://www.uber.com/blog/california/more-flexibility-for-drivers-in-california |
| S50 | Uber blog — Understanding acceptance and cancellation rates (2025-07-15) | https://www.uber.com/mx/en/blog/understanding-acceptance-and-cancellation-rates/ |
| S51 | Uber — Deactivation policy (PR, 2020): rating = last 500, city minimum, cancellation/acceptance handling, fraud | https://www.uber.com/pr/en/legal/deactivation-policy/ |
| S52 | Uber blog — Driver deactivation policy | https://www.uber.com/blog/driver-deactivation-policy |
| S53 | Uber — Back-to-back trips (delivery pages; rides UV) | https://www.uber.com/drive/delivery/basics/back-to-back-trips/ |
| S54 | Uber GB — Taking trips (nav app setting) ; (3P) TechCrunch in-app navigation | https://www.uber.com/gb/en/drive/basics/how-to-take-trips ; https://techcrunch.com/?p=1464602 |
| S55 | Uber Eng — DeepETA | https://www.uber.com/blog/deepeta-how-uber-predicts-arrival-times/ |
| S56 | Uber blog — Driver statements (weekly) | https://www.uber.com/us/en/blog/driver-statements/ |
| S57 | Uber — Service fee explained | https://uber.com/us/en/drive/driver-app/service-fee |
| S58 | Uber — Promotions (Quest, Boost+) | https://www.uber.com/us/en/drive/promotions/ |
| S59 | Uber blog — Consecutive trips | https://www.uber.com/blog/consecutive-trips-earnings |
| S60 | Uber — Uber Pro (AU rewards page; KE program terms) | https://www.uber.com/au/en/drive/rewards/ ; https://www.uber.com/ke/en/legal/uber-pro-program-terms/ |
| S61 | Uber — Instant Pay | https://www.uber.com/us/en/drive/driver-app/instant-pay/ |
| S62 | (3P) Business Standard — Uber India safety features 2026-06-30 (Record My Ride, Ambulance Assistance, PIN) — page 403 on fetch, search snippet only (UV) | https://www.business-standard.com/companies/news/ride-hailing-platform-uber-launches-new-safety-features-for-riders-drivers-126063001156_1.html |
| S63 | Uber — Fraud activities (GB) | https://www.uber.com/gb/en/drive/driver-app/fraud-activities |
| S64 | (3P) Justia — Uber patent: location-spoofing detection | https://patents.justia.com/patent/12113807 |
| S65 | Uber Eng — iOS app size; Apple App Store — Uber Driver (407.6 MB US) | https://eng.uber.com/how-uber-deals-with-large-ios-app-size/ ; https://apps.apple.com/app/uber/id1131342792 |

### Not verified in this pass (UV)
- Sheet snap heights, exact element order inside sheets, haptics, marker interpolation — no primary design source; derived from observed behaviour and 2023/2026 newsroom screenshots. Mobbin/pttrns were not accessible (auth).
- Offer countdown (≈ 15 s) is from a regional Uber page; current US value not confirmed. Tap-anywhere-to-accept varies by market.
- Back-to-back for *rides* (only delivery pages found); proactive document-expiry push; rider deactivation threshold; Reserve reservation-fee amount; driver assign-ahead window for Reserve.
- Uber LK availability of PIN, Reserve, Live Activities, Trip Radar, Pro, voice, teen accounts — not checked.
