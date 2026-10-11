# TAXIOS — MASTER PROMPT FOR BUILDING A COMPLETE, PRODUCTION-READY, WHITE-LABEL RIDE-HAILING SAAS PLATFORM

_Saved verbatim from the owner on 2026-10-11 (version 2; supersedes the "MobilityOS" spec, Parts 1–18).
Owner's clarifications and decisions taken from it are recorded in `DECISIONS.md`; doc-name mapping in
`PROJECT_AUDIT.md`._

## 1. Your Role
Act as my principal software architect, senior full-stack engineer, mobile engineer, SaaS product strategist, DevOps engineer, security engineer, QA lead, and AI systems engineer.
Your mission is to transform my existing Taxios repository into a functional, commercially viable, multi-tenant ride-hailing SaaS platform that starts in Sri Lanka and can expand internationally.
This is an implementation task, not merely a planning, design, or documentation exercise.
I am building this product with AI-assisted development. Work directly inside the existing repository, understand the current implementation, make practical engineering decisions, implement real functionality, run tests, fix errors, and document how to operate the system.

Mandatory rules
1. Inspect the entire existing project before changing anything.
2. Do not assume that a feature exists because it is described in documentation or a previous conversation.
3. Preserve useful existing functionality. Do not rewrite working components without a technical reason.
4. Do not stop after creating a plan, architecture diagram, mockup, prototype, or backlog.
5. Implement functionality in small, testable milestones.
6. Do not mark a feature complete until its implementation and relevant tests have been verified.
7. Never substitute hardcoded data, fake API responses, local-only state, or mock screens for required production functionality.
8. Where a third-party service requires credentials, implement its integration boundary, configuration, error handling, and tests. Clearly mark the live connection as pending until actual credentials and a successful integration test are available.
9. Do not delete data, overwrite configuration, or perform destructive database migrations without first assessing the impact and preserving recoverability.
10. Maintain a clear record of completed work, remaining work, known limitations, and the next actionable task.
11. Use current, supported library versions and official documentation when verifying external integrations.
12. If something is uncertain, investigate it. If it cannot be verified, document the uncertainty instead of inventing facts.

## 2. Product Vision
Taxios is a white-label ride-hailing and mobility SaaS platform serving:
* Existing taxi operators that want to digitize their business.
* Entrepreneurs starting new taxi and mobility companies.
* Small and medium-sized fleets.
* Large taxi fleets and dispatch centres.
* Corporate transportation providers.
* Airport transfer and intercity operators.
* Rental and chauffeur businesses where appropriate.
* Enterprise customers requiring dedicated infrastructure.
* Technology partners and developers integrating mobility services.

Taxios must operate as a platform provider, not just as one taxi company's application.
The business model must support a central Taxios platform and many independently operated customer businesses, each with its own branding, configuration, staff, drivers, customers, plans, integrations, and operational data.

Required product experiences
1. Taxios platform administration: Manage the SaaS business, tenants, subscriptions, entitlements, add-ons, platform operations, and compliance.
2. Tenant/operator dashboard: Each taxi company manages its own business, fleet, drivers, bookings, customers, dispatchers, pricing, and reports.
3. Passenger applications: Branded apps and an optional shared Taxios marketplace app.
4. Driver applications: Branded driver apps and an optional shared driver application.
5. Public websites and booking pages: A website and online booking experience for each operator, including custom domains.
6. Dispatch centre: A real-time operational console for dispatchers and fleet managers.
7. Developer and integration layer: APIs, webhooks, integration configuration, and appropriate documentation.
8. Enterprise deployment options: Shared SaaS, dedicated tenant environments, self-hosted licensed deployments, and separately licensed source-code arrangements.

All these experiences must use a coherent backend, identity model, authorization system, and domain architecture.

## 3. Phase Zero — Audit the Existing Repository
Before implementation, inspect the repository thoroughly.
Determine:
* Languages, frameworks, packages, runtime versions, and dependency health.
* Web applications, mobile applications, backend services, and infrastructure.
* Existing API routes, domain logic, database schemas, migrations, authentication, and authorization.
* Existing tenant or organization concepts.
* Existing UI components, navigation, forms, validation, and responsive layouts.
* Existing integrations, background jobs, realtime connections, tests, and deployment configuration.
* Environment variables, secrets handling, logging, and error reporting.
* Broken builds, incomplete screens, TODOs, dead code, placeholder endpoints, and duplicated logic.
* Which features work end-to-end and which only appear implemented.
* Existing Git changes and user modifications that must be preserved.

Run the available tests and build commands before modifying the project, where safe.
Create and maintain these documents:
* `docs/PROJECT_AUDIT.md`
* `docs/PRODUCT_REQUIREMENTS.md`
* `docs/SYSTEM_ARCHITECTURE.md`
* `docs/DATA_MODEL.md`
* `docs/API_CONTRACTS.md`
* `docs/SECURITY_MODEL.md`
* `docs/COMPETITOR_FEATURE_MATRIX.md`
* `docs/FEATURE_STATUS.md`
* `docs/IMPLEMENTATION_ROADMAP.md`
* `docs/DEPLOYMENT.md`
* `docs/BILLING_AND_ENTITLEMENTS.md`
* `docs/INTEGRATIONS.md`
* `docs/TESTING_STRATEGY.md`
* `docs/DECISIONS.md`
* `docs/KNOWN_LIMITATIONS.md`

Adapt these paths if the repository has an established documentation structure.
The audit must clearly distinguish:
* Verified working functionality.
* Partially implemented functionality.
* UI-only functionality.
* Mocked or hardcoded functionality.
* Missing functionality.
* Blocked integrations.
* Security or reliability risks.

Then proceed with implementation. Do not wait for me to approve every routine technical decision. Ask questions only when an unresolved decision materially affects product behaviour, legal exposure, security, cost, or an irreversible architectural choice.

## 4. Competitor Research and Feature Completeness
Research current ride-hailing dispatch platforms, white-label taxi software, mobility SaaS products, and relevant adjacent products.
Include competitors such as Onde, Yelowsoft, TaxiCaller, iCabbi, and other credible providers discovered during research.
Use official product pages, documentation, API references, pricing pages, and reliable independent sources where available. Verify competitor claims and record the source and date. Distinguish advertised features from independently verified capabilities.
Create `docs/COMPETITOR_FEATURE_MATRIX.md` containing:
* Competitor and product.
* Feature category.
* Verified capability.
* Evidence/source.
* Taxios current status.
* Missing capability.
* Competitive importance.
* Estimated complexity.
* Recommended priority.
* Proposed Taxios differentiation.

Research at least these capability groups:
1. Passenger booking and journey management.
2. Driver applications and driver lifecycle management.
3. Dispatch automation and manual dispatch.
4. Fleet, vehicle, shift, and availability management.
5. Fare calculation, surge pricing, and pricing rules.
6. Payment collection, receipts, refunds, and reconciliation.
7. Scheduled, recurring, airport, intercity, corporate, and multi-stop trips.
8. Ride sharing or pooling, if commercially and operationally appropriate.
9. Taxi-company websites and online booking widgets.
10. Branded mobile apps and custom domains.
11. Customer CRM, loyalty, promotions, and referral systems.
12. Driver commissions, earnings, and settlements.
13. Reports, financial exports, and business intelligence.
14. Notifications, SMS, email, chat, and voice.
15. APIs, webhooks, integrations, and partner ecosystems.
16. Tenant management, licensing, subscriptions, and usage billing.
17. Enterprise security, deployment, and service reliability.
18. Accessibility, localization, and international payment support.
19. AI automation, operational intelligence, and fraud prevention.
20. Support tools, audit trails, incident management, and compliance.

Do not blindly copy competitors. Prioritize essential parity, customer demand, commercial feasibility, operational value, and defensible differentiation.
Identify a minimum viable commercial product separately from the long-term product vision. A feature that is not needed for the first launch must still be recorded if it matters to the roadmap.

## 5. Architecture and Engineering Principles
Use the existing technology stack if it is sound, maintainable, and suitable for the requirements.
If the current prototype lacks a coherent architecture, evaluate this default:
* Web applications: Next.js, React, and TypeScript.
* Passenger and driver mobile applications: React Native and TypeScript.
* Backend: Go modular monolith with well-defined domain modules.
* Primary database: PostgreSQL.
* Geographic queries and location-based operations: PostGIS.
* Caching, ephemeral coordination, and appropriate realtime support: Redis.
* Background work: Durable jobs with retries, failure handling, and idempotency.
* APIs: Versioned REST APIs with OpenAPI documentation.
* Live dispatch and trip updates: Authenticated WebSockets or an appropriate realtime transport.
* Deployment: Docker-based environments, with a practical CI/CD pipeline.
* Monitoring: Structured logs, metrics, tracing where appropriate, error reporting, and health checks.

Do not split the application into microservices prematurely. Start with clear domain boundaries and extract services only when justified by scale, team structure, isolation, or operational needs.
Use domain modules such as:
* Identity and access.
* Tenants and organizations.
* SaaS subscriptions and entitlements.
* Fleet and drivers.
* Passenger accounts.
* Booking and trip lifecycle.
* Dispatch and driver assignment.
* Fare engine.
* Payments and financial reconciliation.
* Communications.
* White-label configuration.
* Integrations.
* AI services.
* Analytics and reporting.
* Audit and compliance.

Define clear interfaces, dependency rules, transaction boundaries, and ownership of data.
Use appropriate database constraints, indexes, transactions, idempotency keys, pagination, validation, and consistent error formats.
Keep secrets out of source code. Provide `.env.example` with safe placeholders and validate required environment variables at startup.

## 6. Multi-Tenant SaaS Foundation
Tenant isolation is a critical requirement.
Implement a secure, explicit tenant model supporting multiple independent taxi businesses.
Each tenant must have:
* Tenant identifier and unique slug.
* Legal and display names.
* Business registration and tax information where applicable.
* Contact and billing information.
* Country, currency, timezone, locale, and preferred language.
* Branding, logo, colours, and app identity.
* Business configuration and operational settings.
* Subscription, entitlements, limits, and enabled add-ons.
* Domain and app configuration.
* Payment and communication provider settings.
* Tenant lifecycle status.
* Audit history.

Implement secure tenant context resolution for authenticated users, API requests, jobs, WebSocket connections, webhooks, and public booking pages.
Never trust a client-provided tenant identifier without authorization and membership validation.
Use tenant-aware database access and enforce isolation with appropriate application checks, database constraints, and PostgreSQL row-level security where suitable.
Test that one tenant cannot access another tenant's:
* Customers.
* Drivers.
* Vehicles.
* Bookings.
* Trip locations.
* Payment records.
* Reports.
* Uploaded files.
* Integrations.
* Staff and permissions.
* AI conversation history.
* Usage and billing information.

Ensure background jobs and administrative tools cannot accidentally bypass tenant isolation.
Provide a controlled, audited platform-support access mechanism where authorized personnel can assist a tenant without silently bypassing permissions.

## 7. Identity, Authentication, and Permissions
Implement:
* Email and password authentication.
* Secure password reset.
* Email verification.
* Session and refresh-token lifecycle management.
* Optional or policy-enforced multi-factor authentication.
* Rate limiting and brute-force protection.
* Secure logout and session revocation.
* Social login where commercially justified.
* Phone verification and OTP through configurable providers.
* Passenger, driver, dispatcher, fleet manager, accountant, tenant administrator, and platform administrator roles.
* Role-based access control with granular permissions.
* Tenant membership and invitations.
* Staff onboarding and offboarding.
* Permission-aware navigation and API authorization.
* Login and security audit events.

Passenger, driver, tenant staff, platform staff, and integration identities must have appropriate separate authorization boundaries.
Do not rely on hiding UI controls as a security mechanism. Every protected operation must be authorized on the server.

## 8. Tenant Registration and Onboarding
Build a complete self-service onboarding flow.
A new operator must be able to:
1. Register an account.
2. Verify email or phone as required.
3. Create a taxi business.
4. Choose a business type and operating country.
5. Configure currency, timezone, and language.
6. Provide business details and required verification documents.
7. Select a SaaS plan or trial, where offered.
8. Configure branding.
9. Add vehicles and drivers.
10. Configure service areas, pricing, and booking rules.
11. Configure payment and notification providers.
12. Generate a booking website.
13. Invite dispatchers and other staff.
14. Complete a guided setup checklist.
15. Publish the business when required checks are complete.

Support existing operators migrating from spreadsheets or another platform, including CSV imports with validation, error reports, and duplicate handling.
Implement draft, pending verification, active, suspended, and closed business states as appropriate.
Make onboarding resumable. Save progress securely and clearly show outstanding steps.
Do not allow businesses to bypass legally required verification or provider-specific approval requirements.

## 9. Platform Owner Dashboard
Create a separate Taxios administration console.
Include:
* Global tenant directory and search.
* Tenant onboarding and verification queue.
* Tenant status management.
* Plans, subscriptions, trials, and billing status.
* Add-on and entitlement management.
* Usage and overage reporting.
* SaaS revenue reporting.
* Payment failures and reconciliation exceptions.
* License issuance and deployment inventory.
* Integration and provider health.
* Global feature flags and controlled rollouts.
* Support tickets and operational incidents.
* Audit logs and security events.
* Platform-level usage, reliability, and performance metrics.
* Announcements and tenant notifications.
* Data retention and deletion workflows.
* Configuration for supported countries and providers.

Restrict sensitive actions by permission, require confirmation for destructive or financially sensitive actions, and record administrative changes in an immutable or appropriately tamper-resistant audit trail.
Do not expose unnecessary passenger or driver personal information to platform administrators.

## 10. Tenant Operator Dashboard
Each taxi operator must have an independent dashboard with:
* Business overview.
* Today's bookings and trips.
* Live fleet status.
* Available, busy, offline, and unavailable drivers.
* Booking and dispatch activity.
* Revenue and operational summaries.
* Driver and vehicle performance.
* Customer management.
* Service area and pricing settings.
* Promotions and loyalty configuration.
* Staff accounts and permissions.
* Website and branding configuration.
* Integration settings.
* Subscription and add-on management.
* Reports and exports.
* Support and notifications.

Create functional screens, validated forms, filters, pagination, loading states, empty states, and error states.
Use permission-aware layouts and responsive design. Do not create dashboards that display fabricated metrics or buttons that do nothing.

## 11. Fleet, Driver, and Vehicle Management
Implement full lifecycle management for drivers and vehicles.

Drivers
* Driver profiles and contact details.
* Verification and approval workflow.
* Licence and document expiry tracking.
* Vehicle assignment.
* Availability, shift, and working status.
* Trip history.
* Earnings and commission records.
* Performance metrics.
* Ratings and feedback.
* Notifications.
* Suspension and reactivation.
* Document access restrictions and retention rules.

Vehicles
* Vehicle details, category, capacity, and accessibility features.
* Registration and insurance documents.
* Inspection and service dates.
* Vehicle ownership or fleet assignment.
* Active, inactive, unavailable, and maintenance states.
* Vehicle-driver eligibility rules.
* Service and maintenance history.

Support driver onboarding, controlled document uploads, expiry alerts, and configurable approval requirements.

## 12. Passenger and Customer Management
Implement:
* Passenger profiles.
* Verified contact information.
* Saved pickup and destination places.
* Trip history and receipts.
* Booking preferences.
* Ratings and feedback.
* Support history.
* Consent and notification preferences.
* Loyalty, referral, and promotional eligibility.
* Account export and deletion workflows where applicable.

Minimize personal information collection and apply suitable retention policies.
Restrict access to live and historical location data. Do not retain detailed trip-location history indefinitely without a defined purpose and policy.

## 13. Booking and Trip Lifecycle
Implement a real booking engine.
Support:
* Immediate bookings.
* Scheduled bookings.
* Recurring bookings.
* Pickup and destination selection.
* Address search and map selection.
* Multi-stop trips.
* Passenger count and luggage.
* Vehicle categories.
* Accessibility requirements.
* Pickup instructions.
* Optional flight and airport-transfer information.
* Corporate booking references.
* Booking notes and permitted special requests.
* Fare estimates.
* Booking confirmation.
* Cancellation and no-show rules.
* Trip receipts and feedback.

Model booking and trip states explicitly. Define allowed state transitions and reject invalid transitions.
For example, define appropriate transitions between requested, quoted, confirmed, assigned, driver-arriving, arrived, in-progress, completed, cancelled, and disputed states. Adapt the state machine to the actual product requirements.
Handle duplicate requests, retries, simultaneous assignment attempts, cancellation races, delayed events, and network failures.
Persist important transitions and provide a traceable trip history.

## 14. Dispatch Console and Driver Assignment
Build a real-time dispatch interface.
Include:
* Booking queue.
* Live map of eligible drivers.
* Driver availability and trip status.
* Manual driver assignment.
* Automatic assignment.
* Scheduled dispatch.
* Driver acceptance and rejection.
* Assignment timeout and retry.
* Reassignment and cancellation.
* Dispatcher notes and activity history.
* Alerts for delayed pickups and unassigned bookings.
* Filters by service area, vehicle category, and status.
* Operational search and booking details.

Automatic assignment should consider configurable criteria such as distance, estimated arrival time, driver availability, vehicle suitability, service area, and tenant rules.
Prevent the same driver from accepting conflicting trips. Use server-side locking, transactions, or another reliable concurrency mechanism where appropriate.
Provide a safe manual override with authorization and audit history.
Separate driver GPS updates from business booking events. Handle stale locations, offline devices, GPS permission denial, and intermittent connectivity.

## 15. Fare Engine and Pricing
Build a configurable fare engine supporting:
* Base fare.
* Distance-based charges.
* Time-based charges.
* Minimum fare.
* Waiting charges.
* Cancellation fees.
* Tolls and additional charges.
* Vehicle-category pricing.
* Service-area and route-specific pricing.
* Scheduled-booking rules.
* Airport and intercity pricing.
* Corporate rates.
* Discounts and promo codes.
* Configurable commissions.
* Surge pricing or demand-based pricing, when enabled.
* Taxes and legally required fee disclosures.

Store the pricing-rule version and fare calculation details used for each quote and completed trip.
Clearly distinguish estimated fares from final fares. Define how waiting time, route changes, tolls, disputes, and adjustments affect the final charge.
Provide an explanation of fare calculations for authorized operators and support staff.
Test currency rounding, boundary conditions, invalid promotions, concurrent bookings, and pricing-rule changes.

## 16. Passenger Applications
Build a production-oriented passenger app using the existing mobile stack.
Include:
* Registration and login.
* Phone verification where required.
* Location permission and manual location selection.
* Address search and saved places.
* Vehicle category selection.
* Fare estimates.
* Booking and scheduled booking.
* Driver assignment status.
* Live driver tracking.
* Driver and vehicle details.
* Contact options with privacy controls.
* Cancellation.
* Payment method selection.
* Trip history and receipts.
* Ratings and feedback.
* Promotions and referrals.
* Push notifications.
* Support and complaints.
* Language selection.
* Account settings and deletion.

Include proper handling for location permission denial, network loss, expired sessions, app backgrounding, failed bookings, payment failures, and unavailable drivers.

## 17. Driver Applications
Build a dedicated driver app with:
* Secure authentication.
* Driver onboarding and verification status.
* Document uploads.
* Online/offline availability.
* Shift and availability management.
* Booking offers and acceptance.
* Navigation handoff or integrated navigation where supported.
* Pickup and destination details.
* Arrival, trip-start, and trip-completion workflows.
* Permitted waiting-time and trip adjustments.
* Passenger contact with privacy protections.
* Earnings and trip history.
* Commission breakdowns.
* Notifications and support.
* Profile and document-expiry alerts.

Implement background location tracking only where necessary and permitted, with explicit permissions, battery-conscious behaviour, and clear privacy disclosure.
Enforce server-side trip transitions and validate that the authenticated driver is authorized for the relevant trip.

## 18. Shared Marketplace and White-Label Products
Support two business models simultaneously.

Shared Taxios marketplace
Passengers can discover and book supported participating operators through a shared Taxios-branded experience.
Define how operators are listed, how service availability is determined, how bookings are routed, and which party is responsible for customer service and payment collection.
Do not assume a marketplace settlement or commission arrangement is supported by a payment provider until verified.

Separate branded operator products
Each tenant can have its own:
* Passenger app branding.
* Driver app branding.
* Logo, colours, typography, and splash screens.
* Business name and contact details.
* Booking website.
* Custom domain and subdomain.
* Booking widget.
* App links and deep links.
* Terms, privacy policy, and support links.
* Supported languages and currencies.
* Notifications and email templates.

Provide a tenant branding and publishing workflow.
Distinguish runtime-configurable branding from native app properties that require a separate app build, signing, store listing, or deployment. Do not claim that changing a dashboard setting automatically publishes a new branded app.
Document app store ownership, release management, signing credentials, custom-domain verification, and TLS requirements.

## 19. Websites and Online Booking
Build a configurable public website and booking experience for each operator.
Include:
* Homepage and company profile.
* Services and vehicle categories.
* Online booking.
* Scheduled booking.
* Contact and support information.
* Service areas.
* Pricing information where configured.
* Terms and privacy pages.
* Responsive mobile layout.
* Search-engine metadata.
* Analytics integration with appropriate consent controls.
* Custom domain support.
* Booking widget or embeddable integration.
* Branded email and booking confirmations.

Provide several maintainable templates and reusable sections rather than requiring a developer to build each website manually.
Implement domain ownership verification and prevent one tenant from claiming another tenant's domain.

## 20. Payments and Financial Operations
Keep ride payments and Taxios SaaS subscription billing as separate financial domains.

Ride payments
Design a payment abstraction layer for supported providers, with an initial Sri Lankan integration investigation for PayHere and Payment.lk.
Support provider capabilities where actually available:
* Payment initiation.
* Hosted checkout or supported SDK.
* Payment status verification.
* Signed webhook processing.
* Idempotent payment event handling.
* Success, failure, cancellation, and pending states.
* Refund workflows.
* Receipts.
* Reconciliation.
* Payment history.
* Exception handling.
* Operator financial reports.

Verify current provider documentation, onboarding requirements, currencies, fees, payment methods, recurring-payment support, refund support, webhook signatures, and settlement capabilities.
Do not assume that a provider supports recurring charges, marketplace payments, split settlement, driver payouts, or automated transfers.
Never treat a browser redirect as definitive proof of successful payment. Verify payment status server-side using the provider's supported method.
Do not store raw card details.

Financial records
Maintain auditable records distinguishing:
* Fare estimates.
* Final trip fares.
* Payment attempts.
* Successful collections.
* Refunds.
* Discounts.
* Platform or operator fees.
* Driver earnings.
* Adjustments.
* Settlement and reconciliation status.

Use a suitable ledger or double-entry accounting model if the business requirements justify it. Define invariants and reconciliation procedures.
Do not represent an internal balance as withdrawable money unless the underlying funds, legal model, and provider arrangements support that representation.
Do not implement a stored-value wallet, custody of customer funds, or money-transfer functionality without the necessary legal review and provider approvals.

## 21. Taxios SaaS Subscription Billing
Build a separate subscription and entitlement system.
Support:
* Free, trial, starter, growth, professional, and enterprise plans.
* Monthly and annual billing where supported.
* Tenant-specific plan configuration.
* Trial expiration.
* Upgrades and downgrades.
* Proration rules where supported.
* Renewal and cancellation.
* Payment failure and retry policies.
* Grace periods.
* Invoices and receipts.
* Subscription history.
* Plan limits.
* Add-on purchases.
* Usage billing.
* Manual enterprise contracts.
* Tax and invoice configuration appropriate to the billing jurisdiction.

Define the exact behaviour when a subscription becomes overdue, expires, or is cancelled. Preserve tenant data and access to permitted exports while restricting paid functionality according to documented policies.
Implement subscription state transitions and webhook idempotency.
Do not assume a ride-payment provider can handle Taxios's recurring SaaS billing until its capabilities are verified.

## 22. Add-On Marketplace and Usage Metering
Taxios must support selling additional capabilities to operators.
Design an extensible catalogue for add-ons such as:
* SMS packages.
* WhatsApp or other messaging integrations where officially supported.
* Email delivery.
* AI voice calls and receptionist services.
* AI usage credits.
* Additional staff accounts.
* Additional drivers or vehicles.
* Advanced analytics.
* Custom domains.
* White-label apps.
* CRM integrations.
* Accounting integrations.
* Premium dispatch automation.
* Advanced reporting.
* Additional storage.
* Enterprise support.
* API access and higher rate limits.
* Additional environments or deployment options.

Implement a central entitlement service that determines whether a tenant is permitted to use a capability.
Enforce entitlements on the server, not only in the user interface.
Implement usage events, usage aggregation, plan quotas, overage rules, billing periods, cost tracking, and usage reports.
For metered integrations, record units, timestamps, tenant, provider, operation type, and billing status. Prevent duplicate usage charges on retried operations.
Provide configurable usage limits and budget alerts. Support hard limits or soft limits according to the purchased plan and the nature of the service.
Never expose internal provider credentials to tenants.

## 23. AI Voice Agent and Receptionist
Develop AI capabilities as modular, optional, metered services.
Prioritize an AI voice receptionist that can serve a taxi operator's existing telephone number or supported voice channel.
Potential capabilities:
* Answer incoming calls.
* Identify the caller's language.
* Collect pickup, destination, time, passenger count, and other required booking information.
* Ask clarifying questions.
* Retrieve service availability and permitted fare estimates.
* Create a booking after validating required information.
* Read back important booking details for confirmation.
* Transfer to a human dispatcher when needed.
* Handle unsupported requests safely.
* Send booking confirmations through configured channels.
* Record appropriately limited conversation summaries.
* Categorize complaints and requests.

Investigate voice providers and their current Malaysian and Sri Lankan availability, telephone-number requirements, local language quality, latency, cost, recording policies, and data residency implications.
Use a tool-based architecture in which the AI can call narrowly scoped backend functions. The AI must never directly modify the database or bypass normal authorization, pricing, booking, or payment rules.
Require confirmation before committing ambiguous or financially consequential actions.
Implement consent and disclosure for AI interactions and call recording where required. Provide fallback behaviour when AI services are unavailable.
Meter voice usage and expose costs and limits to the tenant.

## 24. AI Dispatcher Copilot and Business Intelligence
Build an AI dispatcher assistant that can:
* Summarize operational activity.
* Identify unassigned or delayed bookings.
* Suggest eligible drivers for dispatchers to review.
* Explain operational alerts.
* Draft responses to customers.
* Summarize complaints.
* Identify common service problems.
* Produce natural-language summaries of authorized reports.

Add a business intelligence layer for:
* Booking volume and completion rates.
* Cancellation and no-show rates.
* Revenue and average fare.
* Driver utilization.
* Peak-hour and geographic demand.
* Customer retention.
* Promotional performance.
* Operator service quality.
* Subscription and add-on usage.

Consider demand forecasting, driver positioning recommendations, anomaly detection, and pricing recommendations after reliable data is available.
AI-generated recommendations must be distinguishable from verified facts. Do not automatically impose surge prices, suspend drivers, issue refunds, or take other consequential actions without the appropriate business rules and authorization.
Use tenant-scoped data access, strict query permissions, privacy controls, and cost limits. Never allow one tenant's data to appear in another tenant's AI context.

## 25. Additional Product Innovations
Research and prioritize differentiated features such as:
* Corporate transport accounts and centralized invoicing.
* Employee travel policies and approval workflows.
* Airport transfers and flight-aware pickup coordination where supported.
* Intercity and scheduled transport.
* Recurring school, staff, or business transportation.
* Multi-stop and group bookings.
* Accessible vehicle booking.
* Family or delegated booking profiles with appropriate consent.
* Driver incentives and retention tools.
* Customer loyalty and referral campaigns.
* Fraud and promotional-abuse detection.
* Dispatch workload recommendations.
* Natural-language reporting.
* Automated customer support.
* No-code booking website configuration.
* Open APIs and a partner integration marketplace.
* Tenant data export and migration tools.
* Multi-operator business groups.
* Enterprise service-level reporting.
* Fleet maintenance reminders.
* Optional rental, chauffeur, and other adjacent mobility modules.

For each innovation, document the user problem, business value, prerequisites, technical complexity, operating cost, risks, and recommended implementation phase.
Do not implement experimental features at the expense of reliable booking, dispatch, payments, and security.

## 26. Communications and Notifications
Create a provider abstraction for:
* Push notifications.
* Email.
* SMS.
* WhatsApp or other messaging channels where approved and supported.
* In-app notifications.
* Voice calls.

Support event-driven messages for booking confirmations, driver assignment, driver arrival, trip completion, receipts, cancellations, payment issues, subscription events, document expiry, and support updates.
Include templates, tenant branding, language selection, delivery status, retries, provider failure handling, and cost metering where applicable.
Do not send duplicate messages when an event is retried.
Store only the message content and delivery metadata necessary for the business purpose, subject to retention and privacy policies.

## 27. API, Webhooks, and Integrations
Create documented, versioned APIs for authorized integrations.
Support:
* API authentication and scoped permissions.
* API keys or OAuth where appropriate.
* Key rotation and revocation.
* Rate limiting and quotas.
* Idempotency keys for suitable write operations.
* Pagination and filtering.
* Consistent errors.
* Webhook registration and delivery.
* Signed webhook payloads.
* Retry policies and dead-letter handling.
* Delivery logs and replay controls.
* Versioning and deprecation policies.

Potential integrations include maps, geocoding, routing, payments, accounting, CRM, messaging, telephony, analytics, and fleet systems.
Create `docs/INTEGRATIONS.md` with configuration instructions, supported capabilities, known restrictions, required credentials, test procedures, and operating costs.
Use official provider documentation. Never invent endpoints, SDK methods, authentication flows, or integration capabilities.

## 28. Internationalization and Country Expansion
Build the platform with Sri Lanka as the initial market, without hardcoding Sri Lankan assumptions throughout the system.
Support configurable:
* Country and service area.
* Currency and monetary precision.
* Timezone and daylight-saving behaviour.
* Languages and locale.
* Phone-number validation.
* Address and map providers.
* Payment providers.
* Tax rules.
* Business registration fields.
* Invoice formats.
* Legal disclosures.
* Data retention.
* Driver documentation requirements.
* Supported vehicle categories.
* Pricing and fare rules.

Initial language priorities should be configurable and may include English, Sinhala, and Tamil.
Design for future Malaysian and other international tenants.
Do not assume that one payment provider, tax treatment, phone format, or legal requirement applies globally.
Use explicit feature and provider compatibility checks during tenant onboarding.

## 29. Licensing and Deployment Models
Taxios must support several commercial deployment models.

Model A: Shared hosted SaaS — Multiple tenants use a centrally operated, securely isolated platform.
Model B: Dedicated enterprise deployment — A customer receives a logically or physically dedicated environment with configurable infrastructure, monitoring, backups, support arrangements, and contractual service levels.
Model C: Self-hosted licensed deployment — A customer deploys a supported Taxios distribution in their own environment using a documented license and deployment configuration.
Model D: Source-code licensing — Offer source-code access under a separately defined commercial agreement and license terms. Do not treat a technical implementation as legal authorization to redistribute or sublicense third-party dependencies.

Implement the technical foundation for:
* License issuance and records.
* Deployment registration.
* Edition and feature entitlements.
* License expiry and renewal.
* Supported version tracking.
* Offline or restricted-connectivity verification where commercially necessary.
* Audit and support records.

Define failure behaviour when a license service is unavailable. Avoid remote kill switches that could unexpectedly interrupt critical customer operations.
Keep licensing enforcement separate from tenant identity and payment processing.
The actual license agreements, intellectual property terms, customer obligations, and jurisdiction-specific requirements must receive appropriate legal review.

## 30. Security, Privacy, and Compliance
Treat security as a foundational requirement.
Implement:
* Input validation and output encoding.
* Secure authentication and session handling.
* Authorization checks for every protected operation.
* Tenant isolation tests.
* CSRF protection where applicable.
* Appropriate CORS policies.
* SQL injection prevention.
* XSS prevention.
* Secure file uploads and access controls.
* Rate limiting.
* Secrets management.
* Encryption in transit.
* Appropriate encryption at rest.
* Secure password hashing.
* Audit logging.
* Dependency and container scanning.
* Backup and restore procedures.
* Administrative action controls.
* Security headers.
* Data retention and deletion policies.
* Incident response documentation.

For uploaded identity documents and sensitive passenger data, restrict access, limit retention, and avoid unnecessary duplication.
Assess relevant Sri Lankan privacy, payment, transport, and business requirements, together with requirements in each expansion market. Identify where specialist legal or compliance advice is needed rather than declaring the product compliant without evidence.
Never commit real credentials, production personal information, private keys, or payment secrets to Git.

## 31. Observability and Production Operations
Implement:
* Structured application logs.
* Request and correlation IDs.
* Tenant and trip correlation where safe.
* Application and dependency health checks.
* Error tracking.
* Metrics for latency, errors, throughput, and resource use.
* Background job monitoring.
* Database connection and performance monitoring.
* WebSocket connection monitoring.
* Payment and webhook failure alerts.
* AI provider usage and failure metrics.
* Tenant usage and quota alerts.
* Backup monitoring.
* Deployment and rollback procedures.

Ensure logs do not unnecessarily expose passwords, tokens, payment details, sensitive documents, or precise passenger location histories.
Document incident response, service recovery, disaster recovery, data restoration, and safe schema migration procedures.
Define initial service objectives and recovery targets as explicit, configurable business decisions rather than making unsupported availability guarantees.

## 32. Testing Requirements
Create a meaningful automated testing strategy.

Unit tests — Test domain rules, fare calculations, pricing, permissions, state transitions, subscription rules, usage metering, and validation.
Integration tests — Test APIs, database transactions, tenant isolation, authentication, payment callbacks, webhooks, background jobs, and external service adapters.
End-to-end tests — Cover complete journeys such as:
1. Register a tenant.
2. Configure a business.
3. Add a vehicle and driver.
4. Publish a booking page.
5. Create a passenger booking.
6. Assign a driver.
7. Accept and complete the trip.
8. Record a payment.
9. Generate a receipt.
10. Verify operator reports and trip history.
11. Purchase or change a SaaS plan.
12. Verify entitlements and usage limits.

Security tests — Verify that users cannot access unauthorized tenant records, manipulate fare totals, forge payment success, replay sensitive webhooks, bypass entitlements, or access another user's location or documents.
Reliability tests — Test retries, duplicate requests, interrupted network connections, failed jobs, unavailable providers, race conditions, and recoverable database failures.
Mobile tests — Validate permissions, deep links, session expiry, backgrounding, GPS failures, push notifications, and poor network conditions.
Run available linters, type checks, migrations, unit tests, integration tests, and production builds as the project evolves.
Never claim a test passed if it was not executed.

## 33. CI/CD and Deployment
Provide:
* Development environment setup.
* `.env.example`.
* Dockerfiles and Docker Compose where appropriate.
* Database migration and seed procedures.
* Safe test-data generation.
* Automated linting and type checking.
* Automated tests.
* Production build validation.
* Dependency scanning.
* Deployment instructions.
* Environment-specific configuration.
* Database backup and restoration.
* Rollback instructions.
* Migration safety procedures.
* Health checks.
* Monitoring setup.

Do not use real production data in development seeds.
Separate development, testing, staging, and production configurations.
Document the practical minimum infrastructure for an initial Sri Lankan launch, anticipated scaling bottlenecks, and the conditions under which infrastructure should be upgraded. Do not overengineer for hypothetical global scale before the core product works.

## 34. Required Implementation Milestones
Implement the project in this order unless the repository audit identifies a critical dependency requiring a different sequence.
- Milestone 1 — Repository stabilization: Audit the codebase, document the architecture, fix broken builds, establish environment configuration, and create a working baseline.
- Milestone 2 — Backend and database foundation: Implement core schema, migrations, validation, API conventions, error handling, logging, and testing infrastructure.
- Milestone 3 — Authentication and tenant isolation: Implement identity, tenant membership, roles, permissions, tenant-aware data access, and isolation tests.
- Milestone 4 — Tenant onboarding and platform administration: Implement self-service onboarding, the platform owner dashboard, tenant management, and operator dashboard foundations.
- Milestone 5 — Fleet and customer operations: Implement drivers, vehicles, documents, staff management, passenger profiles, and operational settings.
- Milestone 6 — Booking, fares, and dispatch: Implement booking lifecycle, fare estimates, pricing rules, manual dispatch, automatic assignment, realtime updates, and trip history.
- Milestone 7 — Passenger and driver applications: Connect the actual mobile applications to the backend and implement complete booking and trip workflows.
- Milestone 8 — Payment operations: Implement the payment abstraction, the first verified provider integration, payment state management, receipts, refunds where supported, and reconciliation.
- Milestone 9 — White-label websites and branded experiences: Implement branding, public booking pages, custom domains, shared marketplace foundations, and tenant publishing workflows.
- Milestone 10 — SaaS plans, subscriptions, and add-ons: Implement plan management, entitlements, subscription lifecycle, usage metering, add-on catalogues, and billing operations.
- Milestone 11 — Notifications and integrations: Implement supported communication channels, API documentation, webhooks, provider settings, and integration monitoring.
- Milestone 12 — AI capabilities: Add AI voice booking, dispatcher assistance, customer support, and operational analytics using controlled, testable service integrations.
- Milestone 13 — Enterprise and self-hosted licensing: Implement deployment configuration, license management, enterprise controls, and operational documentation.
- Milestone 14 — Production hardening and launch readiness: Complete security testing, end-to-end tests, performance checks, backup restoration, monitoring, deployment validation, and launch documentation.

Within every milestone, implement and verify smaller tasks rather than attempting a giant untestable change.

## 35. Definition of Done
A feature is complete only when:
* Its requirements are documented.
* The required backend logic exists.
* The database and migrations are implemented where necessary.
* Authorization and tenant isolation are enforced.
* The UI or mobile workflow is connected to real backend functionality.
* Loading, empty, validation, error, and retry states are handled.
* Relevant tests are written and executed.
* API documentation and configuration instructions are updated.
* Logging and operational handling are adequate.
* No critical placeholder or mock remains in the production path.
* Known limitations are documented.
* The relevant build and tests pass.

Where an external account, credential, provider approval, or commercial decision blocks completion, mark the integration as blocked and identify the exact next step.

## 36. Your Working Protocol
After the initial audit:
1. Summarize the actual repository state and the most important risks.
2. Present the prioritized implementation backlog.
3. Identify the first small, achievable milestone.
4. Begin implementing it immediately.
5. Inspect the relevant code before modifying it.
6. Make focused changes instead of rewriting unrelated modules.
7. Run tests and builds after meaningful changes.
8. Fix failures caused by your changes.
9. Update `docs/FEATURE_STATUS.md` and `docs/IMPLEMENTATION_ROADMAP.md`.
10. Continue with the next task that can be implemented safely.

If you reach a context limit, leave the repository in a coherent state and update the project documents with: what changed; files created or modified; commands and tests executed; test results; remaining issues; the exact next implementation task.
Do not repeatedly regenerate the plan instead of writing code.
If a task is too large, decompose it into smaller changes and continue from the highest-priority incomplete item.

## 37. Final Deliverables
The eventual product must include:
* A working Taxios platform administration console.
* Secure tenant onboarding and management.
* Functional operator dashboards.
* Fleet, driver, and customer management.
* Booking, fare calculation, and dispatch.
* Passenger and driver mobile applications.
* Public booking websites and white-label configuration.
* Payment integrations and financial reconciliation.
* SaaS subscriptions, entitlements, and add-on usage billing.
* Communication and integration services.
* AI capabilities introduced with suitable controls.
* Enterprise and self-hosted licensing foundations.
* Automated tests and CI/CD.
* Security, deployment, backup, and operational documentation.
* A verified competitor feature matrix.
* A transparent feature-status report.
* A prioritized roadmap for remaining work.

Do not claim that the entire product is production-ready merely because the screens render or the build succeeds. Production readiness requires successful validation of the actual workflows, security boundaries, integrations, deployment, and recovery procedures.

## 38. Start Now
Begin by auditing the existing Taxios repository.
Do not assume the current prototype is a complete platform. Establish what is actually implemented, research the most important competitive gaps, document the target architecture, and start building the highest-priority missing functionality.
Your objective is to deliver working software incrementally—not just produce an impressive specification.
Start with the audit, then continue directly into implementation. Do not stop after writing the documents.
