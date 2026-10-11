# Product requirements

Source of truth: the master build specification (Parts 4–10, 17). This file tracks requirements by
application with their delivery stage and status. Status: ✅ done (with test evidence) · 🟡 partial · ⬜ not started.

## Applications
| App | Key requirements | Stage | Status |
|---|---|---|---|
| Super-admin console | tenant provisioning/suspension/reactivation/closure, plans, modules, invoices, usage, providers, licences, flags, support, audit, dashboards | 1, 5, 10 | 🟡 tenant lifecycle API done; no UI |
| Tenant portal | profile/branding, zones, categories, fares, drivers, vehicles, staff, customers, bookings, payments, subscriptions, reports, API keys | 1–6 | 🟡 profile + staff invitations + audit API |
| Dispatcher console | pending/active bookings, driver freshness, eligibility, ETA, manual/auto assign, reassign, notes, escalation | 2–3 | ⬜ |
| Passenger app (RN + web) | phone verify, address search, server fare, immediate/scheduled booking, tracking, receipt, rating, support | 3 | ⬜ |
| Driver app (RN) | registration, documents, vehicle, availability, offers, accept/reject, trip lifecycle, cash confirm, earnings | 3 | ⬜ |
| White-label | branding, domains, branded builds via controlled release pipeline | 6 | ⬜ |
| Shared marketplace | operator discovery, eligible offers, single-operator fulfilment attribution, commission | 9 | ⬜ |

## Cross-cutting rules
- Never require location permission to browse or enter an address manually.
- Stale driver location is shown as stale, never as live.
- A driver cannot accept overlapping incompatible bookings, start an unauthorized trip, or change a fare.
- Fare estimates are server-generated and labelled estimates; final fare is computed server-side.
- Marketplace bookings are fulfilled by exactly one selected operator (decision D-010) — never duplicated.
- Super-admin does not browse tenant private data; support access is reason-coded, time-limited and audited (T-020).

## Commercial MVP acceptance (spec Part 17)
Operational journey steps 1–18 and SaaS monetization steps 1–11. Progress per step is tracked in
`PROJECT_STATUS.md`. Currently met: step 1 (admin creates tenant), step 2 (owner completes
onboarding — prerequisites added in T-012), step 18 for Stage 1 tables.
