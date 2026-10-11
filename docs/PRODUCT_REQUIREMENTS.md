# Product requirements

Source of truth: `MASTER_PROMPT.md` (v2, §2 and §6–§33). Per-section status lives in `FEATURE_STATUS.md`; this file is the per-application view. This file tracks requirements by
application with their delivery stage and status. Status: ✅ done (with test evidence) · 🟡 partial · ⬜ not started.

## Applications
| App | Key requirements | Stage | Status |
|---|---|---|---|
| Platform (super-admin) console | tenant directory/search, **verification queue (D-021)**, status, plans/subscriptions/trials/invoices (manual payments, D-023), add-ons/entitlements, usage/revenue, feature flags, support access, audit, announcements, provider health, licences | M4, M10, M13 | 🟡 create/list/status UI only |
| Tenant (operator) dashboard | self-service signup + onboarding checklist, profile/branding, zones, categories, fares, drivers, vehicles, staff, customers, bookings, payments, subscription + add-ons, reports/exports, integrations, API keys | M4–M6, M9–M11 | 🟡 setup, fleet, bookings, dispatch, team built; no signup/customers/branding/billing |
| Dispatcher console | queue, live map, freshness, eligibility, manual/auto assign, phone bookings, cancel; realtime, alerts, notes, filters, reassignment | M6 | 🟡 polling board live; realtime/alerts/notes ⬜ |
| Passenger app (Expo + web) | phone verify, address search, server fare, immediate/scheduled booking, tracking, receipt, rating, support, push, language | M7 | 🟡 web booking page live; mobile ⬜ |
| Driver app (Expo + web) | onboarding, documents, availability, offers, trip lifecycle, navigation, cash, earnings, background location | M7 | 🟡 web driver app live; mobile ⬜ |
| White-label | branding (runtime), booking site templates, domains with ownership verification, **per-tenant store apps built from one codebase under the tenant's own store accounts (D-022)** | M9 | ⬜ |
| Shared marketplace | operator discovery, eligible offers, single-operator fulfilment attribution, commission | M9 (P2) | ⬜ |

## Cross-cutting rules
- Never require location permission to browse or enter an address manually.
- Stale driver location is shown as stale, never as live.
- A driver cannot accept overlapping incompatible bookings, start an unauthorized trip, or change a fare.
- Fare estimates are server-generated and labelled estimates; final fare is computed server-side.
- Marketplace bookings are fulfilled by exactly one selected operator (decision D-010) — never duplicated.
- Super-admin does not browse tenant private data; support access is reason-coded, time-limited and audited (T-020).

## Commercial MVP acceptance (master prompt §32 E2E journey)
1 register tenant · 2 configure business · 3 add vehicle + driver · 4 publish booking page · 5 passenger booking ·
6 assign driver · 7 accept + complete trip · 8 record payment · 9 receipt · 10 operator reports/history ·
11 purchase/change SaaS plan · 12 entitlements + usage limits.
Met via API today: 2, 3, 5–10 (`TestMVPJourney`). Not met: 1 (self-signup, T-121), 4 (branded page, T-170), 11–12 (M10).
