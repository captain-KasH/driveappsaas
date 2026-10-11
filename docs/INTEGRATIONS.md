# Integrations

Rule: implement only capabilities verified against official provider documentation and, where
needed, a merchant/sandbox account. Nothing below is integrated yet. "Unverified" means no
capability may be coded until the research task records the source.

| Area | Candidate | Status | Task |
|---|---|---|---|
| Card / online payments (LK) | PayHere | unverified — research checkout, notify (checksum) callback, status API, refunds, recurring | T-050 |
| Card / online payments (LK) | Payment.lk (Payable) | unverified | T-050 |
| SMS / OTP (LK) | Dialog, Mobitel, Notify.lk, Text.lk, Twilio | unverified — sender ID rules, segment pricing, DLR | T-041 |
| Email | any SMTP / transactional API | not chosen | T-040 |
| Push | FCM / APNs (via Expo if RN app uses Expo) | not chosen | T-042 |
| Maps, geocoding, routing/ETA | Google Maps Platform, OSRM/Valhalla self-host, Mapbox | not chosen; cost vs. quality for LK addresses | T-013 |
| Telephony / IVR / voice AI | Twilio, Vonage, local SIP | unverified; recording consent varies by jurisdiction | T-070 |
| LLM providers | configurable | not chosen | T-080 |

## Adapter contract (to be built with the first provider)
Provider-neutral interfaces per capability (e.g. `payments.Gateway{CreateCheckout, Status, VerifyCallback, Refund}`),
one adapter per provider, a `provider_capabilities` row per verified capability, health checks,
encrypted tenant credentials, and webhook intake that stores the raw event, verifies the signature
per provider docs, dedupes by provider event ID, and processes asynchronously. Browser redirects are
never trusted as payment success.
