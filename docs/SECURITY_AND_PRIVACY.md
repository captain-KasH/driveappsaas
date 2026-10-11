# Security and privacy

## In place (verified by tests in `services/api/internal/server/server_test.go`)
| Control | Evidence |
|---|---|
| Tenant isolation at API (non-member → 404, no existence leak) | `TestTenantIsolationAPI` |
| Tenant isolation at DB (FORCE RLS, transaction-local scope, non-bypass app role) | `TestTenantIsolationRLS` (mutation-checked: fails when a policy is opened) |
| Append-only audit (trigger + revoked grants) | `TestAuditIsAppendOnly` |
| Opaque session tokens stored as SHA-256, revocable, expiring | `TestAuthSessions` |
| bcrypt passwords, 10–72 chars, uniform error + timing for unknown email | `TestAuthSessions` |
| Single-use, expiring, hashed invitations | `TestOnboardingJourney` |
| Suspended tenants read-only; closed tenants sign out sole-tenant staff | `TestSuspensionAndClosure` |
| Reason required for tenant status changes, audited | `TestSuspensionAndClosure`, `TestOnboardingJourney` |
| Strict JSON decoding (unknown fields rejected, 1 MiB cap); 500s leak no detail | `TestInvalidTenantInput` |
| RLS on every Phase 1 table incl. the `driver_compliance` view (security_invoker) | `TestMobilityTablesRLS` |
| Passengers/drivers can only reach their own bookings, offers and trips (404 otherwise) | `TestMVPJourney` step 18, `TestExclusiveAssignmentConcurrency` |
| Driver phone never shown to passengers; passenger phone never shown to drivers | `TestMVPJourney` (no `phone` key) |
| Edge rate limits: auth + sign-up 10/min/IP, API 20/s/IP (packline nginx) | nginx config (packline `edb4fb3b`); no automated test |
| Client IP from `X-Real-IP` only when `TAXIOS_TRUST_PROXY=1` (set only behind our nginx) | code: `identity.RemoteIP` |
| `serve` refuses a DB role that bypasses RLS | manual smoke 2026-10-11: superuser URL → "role bypasses row-level security" (no automated test) |

## Open (tracked in IMPLEMENTATION_PLAN / KNOWN_ISSUES)
Per-account login rate limiting (T-005; per-IP exists at the edge), MFA for platform roles (T-006), reason-coded time-limited
support access (T-020), secrets encryption for provider credentials (T-050), dependency scanning in CI (T-004),
security headers/CORS once web clients exist, session listing/revocation UI.

## Privacy
- Collect the minimum: passenger phone is needed for verification; location only when granted, never required to type an address.
- Driver/passenger contact details are masked through the platform where telephony allows.
- Retention policies per data class (location history, recordings, chats) are defined before those features ship.
- Sri Lanka Personal Data Protection Act No. 9 of 2022 applies — legal review required before launch (KNOWN_ISSUES K-003).
