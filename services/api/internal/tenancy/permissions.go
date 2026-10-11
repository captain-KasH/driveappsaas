package tenancy

// Permission is checked server-side on every tenant-scoped endpoint.
type Permission string

const (
	PermTenantRead   Permission = "tenant.read"
	PermTenantManage Permission = "tenant.manage"
	PermStaffRead    Permission = "staff.read"
	PermStaffManage  Permission = "staff.manage"
	PermAuditRead    Permission = "audit.read"
	PermFleetRead    Permission = "fleet.read"
	PermFleetManage  Permission = "fleet.manage"   // drivers, vehicles, document approval
	PermPricing      Permission = "pricing.manage" // zones, categories, fare rules
	PermBookingRead  Permission = "booking.read"
	PermDispatch     Permission = "dispatch.manage" // create/assign/cancel bookings
	PermReportsRead  Permission = "reports.read"
)

// ponytail: static role->permission map; move to tables when tenants need custom roles.
var rolePermissions = map[string][]Permission{
	"owner":      allStaff,
	"admin":      allStaff,
	"dispatcher": {PermTenantRead, PermStaffRead, PermFleetRead, PermBookingRead, PermDispatch},
	"finance":    {PermTenantRead, PermAuditRead, PermBookingRead, PermReportsRead},
	"viewer":     {PermTenantRead, PermBookingRead},
	"driver":     {}, // driver app routes use RoleRoute
	"passenger":  {}, // passenger app routes use RoleRoute
}

var allStaff = []Permission{PermTenantRead, PermTenantManage, PermStaffRead, PermStaffManage, PermAuditRead,
	PermFleetRead, PermFleetManage, PermPricing, PermBookingRead, PermDispatch, PermReportsRead}

func RoleHas(role string, p Permission) bool {
	for _, have := range rolePermissions[role] {
		if have == p {
			return true
		}
	}
	return false
}

// validRole is for staff invitations; drivers are invited through the fleet module.
func validRole(role string) bool {
	_, ok := rolePermissions[role]
	return ok && role != "passenger"
}
