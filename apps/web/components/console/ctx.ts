export type Tenant = {
  id: string; slug: string; name: string; businessType: string; countryCode: string; status: string; createdAt: string;
  legalName: string | null; registrationNo: string | null; taxNo: string | null; contactEmail: string | null; contactPhone: string | null;
  billingEmail: string | null; address: string | null; timezone: string; currency: string; language: string;
  statusReason: string | null; submittedAt: string | null; verifiedAt: string | null;
};
export type ChecklistStep = { key: string; done: boolean; required: boolean; missing?: string[] };
export type Checklist = { status: string; verified: boolean; canSubmit: boolean; steps: ChecklistStep[] };
export type Ctx = { tenantId: string; tenant: Tenant; role: string; reload: () => Promise<void> };

// Mirrors services/api/internal/tenancy/permissions.go for showing/hiding UI only — the API decides.
const all = ["tenant.read", "tenant.manage", "staff.read", "staff.manage", "audit.read", "fleet.read", "fleet.manage", "pricing.manage", "booking.read", "dispatch.manage", "reports.read"];
const perms: Record<string, string[]> = {
  owner: all,
  admin: all,
  dispatcher: ["tenant.read", "staff.read", "fleet.read", "booking.read", "dispatch.manage"],
  finance: ["tenant.read", "audit.read", "booking.read", "reports.read"],
  viewer: ["tenant.read", "booking.read"],
};
export const can = (role: string, perm: string) => (perms[role] ?? []).includes(perm);

export type Category = { id: string; code: string; name: string; seats: number; active: boolean };
export type Zone = { id: string; name: string; active: boolean; area: GeoJSON.Geometry };
export type Vehicle = { id: string; categoryId: string; plate: string; make: string; model: string; colour: string; status: string };
export type Driver = {
  id: string; fullName: string; phone: string; email: string; status: string; availability: string; vehicleId: string | null;
  hasAccount: boolean; lat: number | null; lng: number | null; locationAt: string | null; locationFresh: boolean; complianceIssues: string[];
};

/** Rupee text → minor units without float drift ("120.5" → 12050). */
export function toMinor(v: string): number {
  const [r, c = ""] = v.trim().split(".");
  return Number(r || 0) * 100 + Number((c + "00").slice(0, 2));
}
export const fromMinor = (m: number) => (m / 100).toFixed(2);

export const ISSUE_TEXT: Record<string, string> = {
  driver_not_approved: "not approved",
  no_vehicle: "no vehicle",
  vehicle_not_approved: "vehicle not approved",
  documents_missing_or_expired: "documents missing or expired",
};
