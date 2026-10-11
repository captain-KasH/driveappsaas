# Known issues and external dependencies

| ID | Issue | Impact | Next step |
|---|---|---|---|
| K-001 | No PostGIS on the local Postgres 16 (only pgcrypto, citext) | Service-zone polygons and nearest-driver queries need it | T-010: `brew install postgis` (check it targets postgresql@16) or use docker-compose |
| K-002 | Docker not installed locally | `docker-compose.yml` and CI parity unverified locally | Install Docker/OrbStack or rely on CI |
| K-003 | Legal review not done: PDPA 2022, transport licensing per vehicle type, CBSL rules for any stored value, call recording consent | Blocks launch, not development | Owner to engage counsel |
| K-004 | No payment/SMS merchant or sandbox accounts | Stage 4/7 integrations can't be verified | Owner to apply for PayHere / Payable / SMS sandbox |
| K-005 | Invitation tokens are returned in the API response (no email yet) | Admin must hand the link to the owner manually | T-040 email delivery |
| K-006 | No login rate limiting | Credential stuffing risk once exposed | T-005 before any public deployment |
| K-007 | CI workflow written but not yet run (no commits pushed before this one) | Unverified CI | Check first GitHub Actions run |
| K-008 | Platform tenant list is capped at 200, no pagination | Fine until >200 tenants | Cursor pagination when needed |
