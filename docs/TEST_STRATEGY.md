# Test strategy

## Layers
| Layer | Where | Runs against |
|---|---|---|
| Domain unit tests (fares, state machines, dispatch ranking) | `internal/<module>/*_test.go` | pure Go |
| API + isolation integration | `internal/server/server_test.go` | real Postgres as `taxios_app` (RLS live) |
| Concurrency (double assignment, idempotency) | integration, parallel goroutines | real Postgres |
| Provider adapters | recorded sandbox fixtures + failure injection | stubbed HTTP |
| E2E (web) / mobile UI | Playwright / Detox or Maestro — when apps exist | staging |
| Load | k6 — booking + location update traffic | staging |
| Backup restore | scripted restore + smoke test | staging |

Run everything: `make test-remote` (rsyncs the working tree to the packline server and runs `deploy/test.sh`: fresh PostGIS 16-3.4 on tmpfs + Go 1.26 container). `deploy.sh` runs the same suite before every deploy. `make test` runs locally only where Postgres has PostGIS. Integration tests **fail** (not skip) without a database.
Every new test must be seen failing once (mutate the guarded code or policy) before it counts as evidence.

## Mandatory scenarios (spec Part 15)
| # | Scenario | Status |
|---|---|---|
| 1 | Tenant cannot access another tenant's bookings | ✅ `TestMVPJourney` step 18, `TestMobilityTablesRLS`, `TestTenantIsolation*` |
| 2 | Two drivers can't both get the same exclusive assignment | ✅ `TestExclusiveAssignmentConcurrency` (8 parallel accepts → exactly 1) + unique partial indexes |
| 3 | Retried booking request doesn't duplicate | ✅ `TestBookingIdempotency` (5 concurrent retries → 1 booking), phone-booking retry in `TestQuoteAndBookingRules` |
| 4 | Fake payment-success redirect can't mark paid | ⬜ T-050 |
| 5 | Replayed verified webhook has no duplicate effect | ⬜ T-050 |
| 6 | Unpaid paid module can't activate | ⬜ T-060 |
| 7 | Retried SMS job doesn't double-deduct credits | ⬜ T-061 |
| 8 | Expired entitlement blocks new usage | ⬜ T-060 |
| 9 | AI outage doesn't break booking | ⬜ T-080 |
| 10 | AI can't read another tenant's data | ⬜ T-080 |
| 11 | Driver with invalid/expired docs not assignable | ✅ `TestExpiredDocumentsBlockDispatch` (mutation-checked: removing the compliance predicate fails it) |
| 12 | Completed trip can't be reopened | ✅ `TestCompletedTripIsFinal` |
| 13 | Source licence ≠ IP ownership | ⬜ T-090 |
| 14 | Provider outage surfaced, not faked success | ⬜ T-050 |
