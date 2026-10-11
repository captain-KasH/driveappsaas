package tenancy

// Permission is checked server-side on every tenant-scoped endpoint.
type Permission string

const (
	PermTenantRead   Permission = "tenant.read"
	PermTenantManage Permission = "tenant.manage"
	PermStaffRead    Permission = "staff.read"
	PermStaffManage  Permission = "staff.manage"
	PermAuditRead    Permission = "audit.read"
)

// ponytail: static role->permission map; move to tables when tenants need custom roles.
var rolePermissions = map[string][]Permission{
	"owner":      {PermTenantRead, PermTenantManage, PermStaffRead, PermStaffManage, PermAuditRead},
	"admin":      {PermTenantRead, PermTenantManage, PermStaffRead, PermStaffManage, PermAuditRead},
	"dispatcher": {PermTenantRead, PermStaffRead},
	"finance":    {PermTenantRead, PermAuditRead},
	"viewer":     {PermTenantRead},
}

func RoleHas(role string, p Permission) bool {
	for _, have := range rolePermissions[role] {
		if have == p {
			return true
		}
	}
	return false
}

func validRole(role string) bool { _, ok := rolePermissions[role]; return ok }
