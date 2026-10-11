# Decisions

| ID | Date | Decision | Why |
|---|---|---|---|
| D-001 | 2026-10-11 | Go modular monolith, stdlib `net/http` (1.22+ patterns), `pgx/v5`, no ORM or router lib | Spec default; fewest moving parts; explicit SQL keeps RLS and money logic visible |
| D-002 | 2026-10-11 | Opaque session tokens (SHA-256 stored), not JWT | Instant revocation (logout, tenant closure) without a denylist |
| D-003 | 2026-10-11 | Isolation via FORCE RLS + transaction-local `app.*` settings + non-bypass `taxios_app` role | Defence in depth: a missing WHERE clause can't leak another tenant |
| D-004 | 2026-10-11 | Users are global; authority via per-tenant membership | One person can staff several operators; passenger/driver profiles will be tenant-scoped records linked to a user |
| D-005 | 2026-10-11 | Forward-only embedded SQL migrations with a 40-line runner | No migration-tool dependency; rollback by compatible releases + backups |
| D-006 | 2026-10-11 | Platform scope only in `PlatformRoute` and invite acceptance | Keeps cross-tenant reads to an auditable, greppable surface |
| D-007 | 2026-10-11 | Static role→permission map in code | No tenant needs custom roles yet; move to tables when one does |
| D-008 | 2026-10-11 | Background work = Postgres outbox + in-process worker; Redis only for real-time | Durable and transactional without new infrastructure |
| D-009 | 2026-10-11 | Product name **Taxios** (spec working title MobilityOS) | Owner decision |
| D-010 | 2026-10-11 | Marketplace bookings fulfilled by exactly one selected operator | Spec forbids duplicate bookings without a rule; allocation can come later as an explicit policy |
| D-011 | 2026-10-11 | Unknown tenant and non-member both return 404 | Don't reveal which tenant IDs exist |
