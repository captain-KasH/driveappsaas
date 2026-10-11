# AI features

Status: none built (Stage 8). AI is optional — core booking and dispatch never depend on it.

## Orchestration layer (planned, T-080)
One `ai` module: provider config + fallback, per-tenant entitlement check, usage/cost ledger
(idempotent), timeouts, rate limits, structured outputs, versioned prompts, tool registry,
evaluation datasets, human escalation, audit of consequential tool calls.

## Tool contract
Every tool declares: name + version, authorized roles, tenant requirement, input schema, validation,
side effects, confirmation requirement, rate limit, audit requirement, failure behaviour.
Tools execute **as the requesting principal** through the same `db.InTx` scope and permission
checks as the HTTP API. No arbitrary SQL, files, cross-tenant reads, price changes or payment initiation.

## Modules
1. **Passenger assistant** — booking, clarification, status, cancellation eligibility, support, FAQs; Sinhala/Tamil/English (quality validated before release).
2. **Operator copilot** — answers from deterministic analytics queries; shows period, definitions, freshness.
3. **Dispatch assistant** — recommends from verified eligibility/ETA/freshness; deterministic engine stays authoritative.
4. **Support assistant** — classify, retrieve approved KB, summarize, suggest; escalate safety/disputes/low confidence.
5. **Growth assistant** — drafts only; sending requires authorization and consent.
6. **Operations intelligence** — summaries, anomalies, expiry reminders, forecasts (labelled as predictions).
7. **Voice booking** — discloses it is automated, fares/refs only from backend tools, read-back + explicit confirmation before booking.

The model never invents availability, fares, ETAs, payment success or booking references.
