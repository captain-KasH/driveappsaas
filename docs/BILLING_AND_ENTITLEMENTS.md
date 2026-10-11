# Billing and entitlements

Status: not built (Stage 5). Design constraints recorded now so earlier stages don't contradict them.

## Separate accounting contexts
1. Platform subscription revenue · 2. Passenger fare transactions · 3. Marketplace fees ·
4. Operator-managed driver settlements · 5. Add-on subscriptions and usage · 6. Refunds and adjustments.
Taxios does **not** pool or custody operator money; operators settle through their own merchant
accounts unless legal and provider approval says otherwise (stored value needs CBSL review).

## Module activation sequence
authorize user → tenant eligibility → dependencies/conflicts → show real price + usage policy →
confirmation → subscription record → verified payment or authorized trial → entitlements →
secure credential collection → connection validation → operational only when prerequisites pass → audit.
A paid module is never active before payment/trial conditions are met.

## Usage ledger
`usage_events`: id, tenant_id, module_id, category, quantity + unit, occurred_at, provider_ref,
cost basis, billed amount (minor units), related operation, **idempotency key UNIQUE (tenant_id, key)**,
reconciliation status. Append-only; corrections are adjustment events. Supports quotas, prepaid packs,
overage, spending caps, low-credit alerts, exports. Retried jobs cannot double-deduct (enforced by the unique key, tested).

## Plans
Defined in the admin console as data (not code): price, interval, included quotas, module
compatibility. Core ride operations stay available per plan/contract even when an add-on expires.
