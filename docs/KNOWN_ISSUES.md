# Known issues and external dependencies

| ID | Issue | Impact | Next step |
|---|---|---|---|
| K-001 | No PostGIS on the dev Mac's Postgres 16 | Local `make test` fails at migration 0002 | Resolved by testing on the server (`make test-remote`); local install optional |
| K-002 | Docker not installed locally | `docker-compose.yml` and CI parity unverified locally | Install Docker/OrbStack or rely on CI |
| K-003 | Legal review not done: PDPA 2022, transport licensing per vehicle type, CBSL rules for any stored value, call recording consent | Blocks launch, not development | Owner to engage counsel |
| K-004 | No payment/SMS merchant or sandbox accounts | Stage 4/7 integrations can't be verified | Owner to apply for PayHere / Payable / SMS sandbox |
| K-005 | Invitation tokens are returned in the API response (no email yet) | Admin must hand the link to the owner manually | T-040 email delivery |
| K-006 | Login rate limiting only per IP at the edge (packline nginx `auth` zone, shared with packline) | Distributed credential stuffing per account not limited | T-005: per-email limit in-app |
| K-007 | `.github/workflows/ci.yml` exists but Actions is not used (minutes cap) | None while `deploy.sh` runs the suite | Delete or keep for when Actions is enabled |
| K-008 | Platform tenant list (200), bookings list (200) and passenger history (100) have no pagination | Fine for MVP volumes | Cursor pagination when needed |
| K-009 | Passenger phone numbers are not verified (no SMS provider yet) | A passenger could register someone else's number | T-041 OTP before public launch |
| K-010 | Route distance/ETA estimates are straight-line × 1.35 at 22 km/h; dispatch ranks by straight-line distance | Estimates can be off on winding routes; final fare uses the GPS track | T-013b routing provider |
| K-011 | Packline nginx bind-mounts `nginx.conf` read-only as a single file: a `git pull` doesn't reach the running container until nginx is recreated (packline deploys do this) | Vhost changes need a packline deploy or an nginx recreate | Documented in DEPLOYMENT_RUNBOOK |
| K-012 | No push/SMS notifications: passengers and drivers must poll (`/my/bookings/{id}`, `/driver/work`) | Higher latency/battery until T-031/T-042 | WebSocket + push in Stage 3 |
| K-013 | Driver documents have numbers and expiry but no file uploads yet | Reviewers can't see the document image | Object storage + signed URLs (Stage 3) |
| K-014 | Web maps use OpenStreetMap public tiles + Nominatim search/reverse geocoding from the browser | Their usage policies only allow low volume; fine for MVP, not for scale | Replace with the chosen maps provider (T-013b) |
| K-015 | Web session token is kept in localStorage; no Content-Security-Policy on the site yet | An XSS bug would expose the token | httpOnly cookie sessions + CSP before adding third-party scripts |
| K-016 | The UI journey (admin → owner → driver → passenger in the browser) has not been walked end to end; the API journey is fully tested | UI-only bugs may remain | Walk it once a platform admin exists; add Playwright E2E |
