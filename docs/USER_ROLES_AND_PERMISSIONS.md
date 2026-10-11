# User roles and permissions

Identity is global (`users`); authority is per tenant (`tenant_memberships.role`) or platform-wide
(`users.platform_role`). Checks are server-side in `tenancy.TenantRoute` / `tenancy.PlatformRoute`.
Source: `services/api/internal/tenancy/permissions.go`.

## Platform roles
| Role | Can |
|---|---|
| `super_admin` | create tenants, change tenant status (reason required), list tenants |
| `support` | list tenants (read-only). Tenant private data access needs a reason-coded, time-limited grant (T-020, not built) |

Platform roles grant **no** access to tenant workspace routes. A super admin is not a tenant member.

## Tenant roles
| Permission | owner | admin | dispatcher | finance | viewer |
|---|---|---|---|---|---|
| tenant.read | ✓ | ✓ | ✓ | ✓ | ✓ |
| tenant.manage (profile) | ✓ | ✓ | | | |
| staff.read | ✓ | ✓ | ✓ | | |
| staff.manage (invite) | ✓ | ✓ | | | |
| audit.read | ✓ | ✓ | | ✓ | |

Extra rules:
- Only an `owner` can invite another `owner` and complete onboarding.
- Suspended/closed tenants: reads allowed, writes return `403 tenant_inactive`.
- Closing a tenant revokes sessions of staff who belong to no other tenant (platform staff excluded).

## Planned (not built)
Passenger and driver identities (tenant-scoped profiles linked to a user, phone-verified), corporate
account roles, API-key principals with scoped permissions, AI tool principals (act as the requesting user, never above).
