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
| K-017 | INCIDENT 2026-10-11: the Taxios compose service was named `api`; on packline's shared network that name became a DNS alias, so Docker round-robined ~60% of packline `/api` traffic to Taxios (404s) until packline switched to unique aliases (packline `d724634c`) | Resolved both sides | Taxios service renamed `taxios-api` (container `taxios-api`); rule in deploy/compose.yml: nothing generic on `edge` |
| K-018 | Apple App Review Guideline 4.2.6: template/white-label apps must be submitted under the client's own developer account | Every tenant wanting its own store app needs Apple ($99/yr) + Google ($25) developer accounts and must grant Taxios submit access | D-022; T-173/T-174 design around it; shared app as fallback |
| K-019 | No email provider account: verification, reset and invitation emails are logged, not sent (`comms.LogMailer`) | Self-signup in prod stalls at "check your email" until a super admin uses **Verify an email address manually** in `/admin/` (audited). SMTP works as soon as `TAXIOS_SMTP_*` are set in `deploy/.env` (stdlib STARTTLS/465, no provider SDK) | Owner to supply SMTP credentials (any provider) |
| K-020 | No object storage for uploads (documents, logos) | T-122/T-130/T-007 need it | Decide: local volume on the packline server vs S3-compatible bucket (packline already uses S3) |
