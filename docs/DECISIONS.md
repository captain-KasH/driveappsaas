# Decisions

| ID | Date | Decision | Why |
|---|---|---|---|
| D-001 | 2026-10-11 | Go modular monolith, stdlib `net/http` (1.22+ patterns), `pgx/v5`, no ORM or router lib | Spec default; fewest moving parts; explicit SQL keeps RLS and money logic visible |
| D-002 | 2026-10-11 | Opaque session tokens (SHA-256 stored), not JWT | Instant revocation (logout, tenant closure) without a denylist |
| D-003 | 2026-10-11 | Isolation via FORCE RLS + transaction-local `app.*` settings + non-bypass `taxios_app` role | Defence in depth: a missing WHERE clause can't leak another tenant |
| D-004 | 2026-10-11 | Users are global; authority via per-tenant membership | One person can staff several operators; passenger/driver profiles will be tenant-scoped records linked to a user |
| D-005 | 2026-10-11 | Forward-only embedded SQL migrations with a 40-line runner | No migration-tool dependency; rollback by compatible releases + backups |
| D-006 | 2026-10-11 | Platform scope only in `PlatformRoute`, invite acceptance, and (since 0003) account-token flows + `POST /v1/businesses` (first membership) | Keeps cross-tenant reads to an auditable, greppable surface |
| D-007 | 2026-10-11 | Static role→permission map in code | No tenant needs custom roles yet; move to tables when one does |
| D-008 | 2026-10-11 | Background work = Postgres outbox + in-process worker; Redis only for real-time | Durable and transactional without new infrastructure |
| D-009 | 2026-10-11 | Product name **Taxios** (spec working title MobilityOS) | Owner decision |
| D-010 | 2026-10-11 | Marketplace bookings fulfilled by exactly one selected operator | Spec forbids duplicate bookings without a rule; allocation can come later as an explicit policy |
| D-011 | 2026-10-11 | Unknown tenant and non-member both return 404 | Don't reveal which tenant IDs exist |
| D-012 | 2026-10-11 | Host on the packline server as a separate Docker stack (`/data/apps/taxios`, own PostGIS), served at taxios.packline.lk via packline's nginx | Owner decision; no new server cost. Taxios never shares packline's DB |
| D-013 | 2026-10-11 | Tests run on the server in Docker (`make test-remote`), not GitHub Actions or the dev Mac | Actions minutes capped; dev Mac lacks PostGIS/Docker |
| D-014 | 2026-10-11 | Drivers and passengers are tenant members (roles `driver`, `passenger`) with zero staff permissions; their routes use `RoleRoute` | One membership/RLS model for every actor |
| D-015 | 2026-10-11 | Driver eligibility lives in one SQL view `driver_compliance` (security_invoker); dispatch, approval and go-online all read it | No drift between UI and dispatch rules |
| D-016 | 2026-10-11 | Fare locked at quote time: a booking keeps the fare rule version of its quote; final fare recomputes with that version using the GPS track (fallback: estimate distance) | Price changes never alter accepted bookings |
| D-017 | 2026-10-11 | Exclusive assignment = booking row lock + unique partial indexes (one pending offer per booking/driver, one live trip per driver) | Correct under concurrency even if app checks regress |
| D-018 | 2026-10-11 | Cash only until a payment provider's API is verified; online methods rejected with `payment_method_unavailable` | Spec forbids fabricated provider integrations |
| D-019 | 2026-10-11 | Dispatch worker runs in-process every 5 s; single API replica until it takes a Postgres advisory lock | Simple; scale-out is a known upgrade |
| D-020 | 2026-10-11 | Master prompt v2 (`MASTER_PROMPT.md`) supersedes the MobilityOS spec; 14 milestones replace Stages 0–11; existing doc names kept with a mapping in PROJECT_AUDIT.md | Owner's new brief; renaming files would break links and history for no gain |
| D-021 | 2026-10-11 | Tenants self-register (verified email → draft business → documents → `pending_verification`) and go `active` only after a super admin approves in the platform console; super admins can still create tenants directly | Owner decision: self-service growth with a verification gate (transport businesses, legal exposure) |
| D-022 | 2026-10-11 | White-label apps = **separate store app per tenant from day one**, built from one Expo codebase with a build-time tenant pin + native assets, submitted under the **tenant's own** Apple/Google developer accounts (Apple App Review 4.2.6 forbids template apps under the provider's account); credentials stored encrypted, never displayed. The shared Taxios app with runtime branding remains for tenants without store accounts | Owner decision + Apple rule. Runtime branding is a prerequisite of the per-tenant build anyway |
| D-023 | 2026-10-11 | SaaS billing v1 = real plans/trials/invoices with **manual** payment recording (bank transfer / payment link, super admin marks paid, grace + overdue policy). Automatic recurring billing only after a provider's recurring API is verified with real credentials | No LK provider's recurring support verified; prompt forbids assuming it |
| D-024 | 2026-10-11 | Mobile = React Native with Expo (EAS Build/Submit, Expo push, OTA for JS only) | Owner decision; matches the owner's existing toolchain; per-tenant builds are EAS profiles |
| D-025 | 2026-10-11 | Build order M1 → M4 → M7 → M9 → M10 → M8 → M11 → M14 → M12 → M13 | Owner's first goal is SaaS owner → tenant admin → white-label app end to end; manual invoicing means billing doesn't wait for payments |
