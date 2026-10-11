// Business types accepted by the API (tenants.business_type CHECK) with their labels.
export const BUSINESS_TYPES: [string, string][] = [
  ["taxi", "Taxi company"], ["three_wheeler", "Three-wheeler fleet"], ["independent_driver", "Independent driver"],
  ["airport_transfer", "Airport transfers"], ["chauffeur", "Chauffeur / premium"], ["tourism", "Tourism transport"],
  ["hospitality", "Hotel transport desk"], ["corporate", "Corporate / staff transport"], ["shuttle", "Scheduled shuttle"],
  ["startup", "New transport brand"], ["enterprise_fleet", "Large fleet"],
];
export const businessTypeLabel = (v: string) => BUSINESS_TYPES.find(([k]) => k === v)?.[1] ?? v.replaceAll("_", " ");
