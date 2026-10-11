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
| fleet.read (drivers, vehicles) | ✓ | ✓ | ✓ | | |
| fleet.manage (+ document approval) | ✓ | ✓ | | | |
| pricing.manage (zones, categories, fares) | ✓ | ✓ | | | |
| booking.read | ✓ | ✓ | ✓ | ✓ | ✓ |
| dispatch.manage (phone bookings, assign, cancel, board) | ✓ | ✓ | ✓ | | |
| reports.read | ✓ | ✓ | | ✓ | |

`driver` and `passenger` are membership roles with **no** staff permissions. Their endpoints
(`/driver/*`, `/my/*`, `/quotes`) are guarded by `RoleRoute` and always resolve the caller's own
driver/customer record from the session — a driver can only act on their own offers and trips,
a passenger only on their own bookings (others return 404).

Extra rules:
- Only an `owner` can invite another `owner` and complete onboarding.
- Suspended/closed tenants: reads allowed, writes return `403 tenant_inactive`.
- Closing a tenant revokes sessions of staff who belong to no other tenant (platform staff excluded).

## Planned (not built)
Phone verification for passengers (K-009), corporate
account roles, API-key principals with scoped permissions, AI tool principals (act as the requesting user, never above).
