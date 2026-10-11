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

Run everything: `make test`. Integration tests **fail** (not skip) without a database.
Every new test must be seen failing once (mutate the guarded code or policy) before it counts as evidence.

## Mandatory scenarios (spec Part 15)
| # | Scenario | Status |
|---|---|---|
| 1 | Tenant cannot access another tenant's bookings | 🟡 proven for Stage 1 tables; extend to bookings (T-015) |
| 2 | Two drivers can't both get the same exclusive assignment | ⬜ T-016 |
| 3 | Retried booking request doesn't duplicate | ⬜ T-015 |
| 4 | Fake payment-success redirect can't mark paid | ⬜ T-050 |
| 5 | Replayed verified webhook has no duplicate effect | ⬜ T-050 |
| 6 | Unpaid paid module can't activate | ⬜ T-060 |
| 7 | Retried SMS job doesn't double-deduct credits | ⬜ T-061 |
| 8 | Expired entitlement blocks new usage | ⬜ T-060 |
| 9 | AI outage doesn't break booking | ⬜ T-080 |
| 10 | AI can't read another tenant's data | ⬜ T-080 |
| 11 | Driver with invalid/expired docs not assignable | ⬜ T-016 |
| 12 | Completed trip can't be reopened | ⬜ T-017 |
| 13 | Source licence ≠ IP ownership | ⬜ T-090 |
| 14 | Provider outage surfaced, not faked success | ⬜ T-050 |
