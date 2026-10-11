# API contracts — v1

Base path `/v1`. JSON only; unknown request fields are rejected (`400 invalid_json`), bodies ≤ 1 MiB.
Auth: `Authorization: Bearer txs_…` (opaque session token, 30-day TTL).
Every response carries `X-Request-ID` (client may send one ≤ 64 chars).

Error shape (all errors):
```json
{"error": {"code": "slug_taken", "message": "slug \"x\" is already in use", "requestId": "a1b2…"}}
```
Common codes: `unauthenticated` 401 · `forbidden` 403 · `not_found` 404 · `invalid_json` 400 · `internal` 500.

OpenAPI generation: T-011 (not yet generated — this file is the contract until then).
Rate limits: not yet enforced (T-005). Audit column: ✓ = writes an audit event in the same transaction.

| Method & path | Auth / permission | Scope | Request | Success | Business errors | Audit |
|---|---|---|---|---|---|---|
| `GET /healthz` | none | — | — | 200 `{status}` | 503 `db_unavailable` | |
| `POST /v1/auth/login` | none | — | `{email,password}` | 200 `{token,userId,platformRole}` | 401 `invalid_credentials` (same for unknown email) | |
| `POST /v1/auth/logout` | session | — | — | 204 | | |
| `POST /v1/auth/accept-invite` | invitation token | platform (by token hash) | `{token,fullName,password}` | 200 `{token,tenantId,role}` | 404 `invalid_invitation`, 422 `weak_password`/`name_required`, 401 `invalid_credentials` (existing account, wrong password) | ✓ |
| `GET /v1/me` | session | user | — | 200 `{user, memberships[]}` | | |
| `POST /v1/platform/tenants` | `super_admin` | platform | `{slug,name,businessType,countryCode?,ownerEmail}` | 201 `{tenant, ownerInvitation{email,role,token,expiresAt}}` | 409 `slug_taken`, 422 `invalid_tenant`/`invalid_email` | ✓ |
| `GET /v1/platform/tenants` | `super_admin`,`support` | platform | — | 200 `{tenants[]}` (≤200, newest first) | | |
| `POST /v1/platform/tenants/{id}/status` | `super_admin` | platform | `{status,reason}` | 200 `{tenant}` | 422 `reason_required`, 409 `invalid_transition`, 404 | ✓ |
| `GET /v1/tenants/{id}` | member · `tenant.read` | tenant | — | 200 `{tenant,myRole}` | 404 non-member | |
| `PATCH /v1/tenants/{id}` | `tenant.manage` | tenant, write | `{name}` | 200 `{tenant}` | 422 `invalid_name`, 403 `tenant_inactive` | ✓ |
| `POST /v1/tenants/{id}/onboarding/complete` | owner | tenant, write | — | 200 `{tenant}` | 409 `invalid_transition` | ✓ |
| `GET /v1/tenants/{id}/members` | `staff.read` | tenant | — | 200 `{members[]}` | | |
| `POST /v1/tenants/{id}/invitations` | `staff.manage` (owner for role=owner) | tenant, write | `{email,role}` | 201 `{invitation}` | 422 `invalid_email`/`invalid_role` | ✓ |
| `GET /v1/tenants/{id}/audit-events` | `audit.read` | tenant | — | 200 `{events[]}` (latest 100) | | |

Idempotency: tenant creation is naturally idempotent on `slug` (409 on replay). A general
`Idempotency-Key` mechanism lands with bookings (T-015).

Invitation tokens are returned once in the API response until email delivery exists (T-040); only
their hash is stored.
