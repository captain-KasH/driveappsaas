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
Rate limits: enforced at the edge (see end of file); none in-app yet. Audit column: ✓ = writes an audit event in the same transaction.

| Method & path | Auth / permission | Scope | Request | Success | Business errors | Audit |
|---|---|---|---|---|---|---|
| `GET /healthz` | none | — | — | 200 `{status}` | 503 `db_unavailable` | |
| `POST /v1/auth/login` | none | — | `{email,password}` | 200 `{token,userId,platformRole}` | 401 `invalid_credentials` (same for unknown email) | |
| `POST /v1/auth/logout` | session | — | — | 204 | | |
| `POST /v1/auth/accept-invite` | invitation token | platform (by token hash) | `{token,fullName,password}` | 200 `{token,tenantId,role}` | 404 `invalid_invitation`, 422 `weak_password`/`name_required`, 401 `invalid_credentials` (existing account, wrong password) | ✓ |
| `GET /v1/me` | session | user | — | 200 `{user{…,emailVerified}, memberships[]}` | | |
| `POST /v1/auth/register` | none | platform (new user) | `{email,fullName,password}` | 201 `{token,userId,emailVerified:false}` (+`verifyToken` only with `TAXIOS_DEV_TOKENS=1`) | 409 `email_taken`, 422 `invalid_email`/`name_required`/`weak_password` | ✓ |
| `POST /v1/auth/verify-email` | none | platform (by token hash) | `{token}` | 200 `{emailVerified:true}` | 404 `invalid_token` (used/expired, 24 h) | ✓ |
| `POST /v1/auth/resend-verification` | session | user | — | 200 `{sent}` | 409 `already_verified` | |
| `POST /v1/auth/forgot-password` | none | platform | `{email}` | 200 `{sent}` always (no enumeration) | | |
| `POST /v1/auth/reset-password` | none | platform (by token hash) | `{token,password}` | 200 `{token,userId}` — all other sessions revoked; email counts as verified | 404 `invalid_token` (1 h), 422 `weak_password` (checked before the token is spent) | ✓ |
| `POST /v1/businesses` | session, email verified | platform (first membership) | `{slug,name,businessType,countryCode?}` | 201 `{tenant}` (status `onboarding`, caller = owner) | 403 `email_unverified`, 409 `slug_taken`/`too_many_drafts` (3), 422 `invalid_tenant` | ✓ |
| `POST /v1/platform/users/verify-email` | `super_admin` | platform | `{email}` | 200 `{emailVerified:true}` | 404 `user_not_found` | ✓ (K-019 stop-gap) |
| `POST /v1/platform/tenants/{id}/verification` | `super_admin` | platform | `{decision: approve\|reject, note?}` (note ≥3 chars to reject) | 200 `{tenant}` — approve: `verifiedAt` set, `pending_verification → active`; reject: back to `onboarding` with `statusReason` | 422 `invalid_decision`/`reason_required`, 409 `invalid_transition` (closed), 404 | ✓ |
| `POST /v1/tenants/{id}/files` | any member except passenger | tenant, write | multipart field `file` (JPEG/PNG/WebP/PDF by sniffing, ≤ 10 MB) | 201 `{file{id,contentType,sizeBytes,fileName,createdAt}}` | 413 `file_too_large`, 415 `unsupported_file_type`, 400 `invalid_upload`, 422 `empty_file` | ✓ |
| `GET /v1/tenants/{id}/files/{fileId}` | uploader, the driver it belongs to, `fleet.read` or `tenant.manage` | tenant | — | 200 bytes (`Content-Type`, inline, no-store) | 404 otherwise (incl. other tenants) | |
| `GET /v1/platform/files/{fileId}` | `super_admin`,`support` | platform | — | 200 bytes | 404 | ✓ `file.viewed_by_platform` |
| `GET /v1/tenants/{id}/documents` · `POST …/documents` | `tenant.read` · `tenant.manage` | tenant | `{docType: business_registration\|owner_id\|tax_registration\|other, fileId, note?}` | 200 `{documents[]}` · 201 `{documentId}` | 422 `invalid_file`/`invalid_document` | ✓ |
| `GET /v1/platform/tenants/{id}` | `super_admin`,`support` | platform | — | 200 `{tenant, documents[], owners[{email,fullName,emailVerified}]}` | 404 | |
| `GET /v1/tenants/{id}/onboarding` | `tenant.read` | tenant | — | 200 `{status,verified,canSubmit,steps[{key,done,required,missing?}]}` — computed from rows (profile, zone, category, fare, vehicle, driver, staff, verification, live) | | |
| `POST /v1/platform/tenants` | `super_admin` | platform | `{slug,name,businessType,countryCode?,ownerEmail}` | 201 `{tenant, ownerInvitation{email,role,token,expiresAt}}` | 409 `slug_taken`, 422 `invalid_tenant`/`invalid_email` | ✓ |
| `GET /v1/platform/tenants` | `super_admin`,`support` | platform | `?status=&q=` (q matches name/slug/legal name) | 200 `{tenants[]}` (≤200, pending verification first, then newest) | | |
| `POST /v1/platform/tenants/{id}/status` | `super_admin` | platform | `{status,reason}` | 200 `{tenant}` | 422 `reason_required`, 409 `invalid_transition`, 404 | ✓ |
| `GET /v1/tenants/{id}` | member · `tenant.read` | tenant | — | 200 `{tenant,myRole}` | 404 non-member | |
| `PATCH /v1/tenants/{id}` | `tenant.manage` | tenant, write | any of `{name,legalName,registrationNo,taxNo,contactEmail,contactPhone,billingEmail,address,timezone,currency,language}` (`""` clears an optional field) | 200 `{tenant}` | 422 `invalid_<field>`/`invalid_input`, 403 `tenant_inactive` | ✓ |
| `POST /v1/tenants/{id}/onboarding/complete` | owner | tenant, write | — | 200 `{tenant,checklist}` — verified tenant → `active`; self-registered → `pending_verification` (D-021) | 409 `onboarding_incomplete` (lists missing steps) / `invalid_transition` | ✓ |
| `DELETE /v1/tenants/{id}/members/{userId}` | `staff.manage` (owner to remove an owner) | tenant, write | — | 200 `{removed}` | 409 `cannot_remove_self`/`last_owner`/`not_staff`, 403, 404 | ✓ |
| `GET /v1/tenants/{id}/customers` | `booking.read` | tenant | `?q=` name/phone | 200 `{customers[{…,bookings,completed,cancelled,spentMinor,lastBookingAt}]}` (≤500) | | |
| `GET /v1/tenants/{id}/exports/customers.csv` · `…/exports/bookings.csv` | `reports.read` | tenant | bookings: `?status=&from=YYYY-MM-DD&to=` | 200 `text/csv` attachment (money as decimals, UTC RFC 3339) | 403 | |
| `GET /v1/tenants/{id}/members` | `staff.read` | tenant | — | 200 `{members[]}` | | |
| `POST /v1/tenants/{id}/invitations` | `staff.manage` (owner for role=owner) | tenant, write | `{email,role}` | 201 `{invitation}` | 422 `invalid_email`/`invalid_role` | ✓ |
| `GET /v1/tenants/{id}/audit-events` | `audit.read` | tenant | — | 200 `{events[]}` (latest 100) | | |

Idempotency: tenant creation is naturally idempotent on `slug` (409 on replay). A general
`Idempotency-Key` mechanism lands with bookings (T-015).

Invitation tokens are returned once in the API response until email delivery exists (T-040); only
their hash is stored.

## Phase 1 — core mobility (all under `/v1/tenants/{id}` unless noted)

Money is always integer **minor units** (`*Minor`, LKR cents). Every write below runs in one
transaction with the tenant's RLS scope; booking state changes append to `booking_events`.

### Configuration (owner/admin: `pricing.manage`; read: `tenant.read`)
| Method & path | Request | Success | Errors |
|---|---|---|---|
| `POST /zones` | `{name, area: GeoJSON Polygon/MultiPolygon (lng,lat)}` | 201 `{zone}` | 409 `zone_exists`, 422 `invalid_area` |
| `GET /zones` | — | `{zones[]}` | |
| `POST /vehicle-categories` | `{code, name, seats}` | 201 `{category}` | 409 `category_exists`, 422 `invalid_category` |
| `GET /vehicle-categories` | — | `{categories[]}` (active) | |
| `POST /fare-rules` | `{zoneId, categoryId, baseMinor, perKmMinor, perMinuteMinor, minimumMinor, bookingFeeMinor?, scheduledFeeMinor?, waitingPerMinMinor?, freeWaitingMinutes?, nightSurchargeBps?, nightStart?, nightEnd? ("HH:MM"), taxBps?, roundToMinor? (default 100)}` | 201 `{fareRule}` = next immutable version | 422 `invalid_fare_rule` |
| `GET /fare-rules` | — | current version per (zone, category) | |

### Fleet (`fleet.manage` writes, `fleet.read` reads)
| Method & path | Request | Success | Errors |
|---|---|---|---|
| `POST /vehicles` | `{categoryId, plate, make, model, colour}` | 201 `{vehicle}` (pending) | 409 `plate_exists` |
| `POST /vehicles/{vid}/status` | `{status: approved\|suspended, reason?}` | 200 | |
| `POST /drivers` | `{fullName, phone (E.164 or 07…), email}` | 201 `{driver, invitation}` (invite token shown once) | 409 `driver_exists`, 422 `invalid_phone` |
| `GET /drivers`, `GET /drivers/{did}` | — | driver incl. `complianceIssues[]`, `locationFresh`, documents, vehicle | |
| `POST /drivers/{did}/documents` | `{docType, number, expiresOn?}` | 201 (pending) | 422 `invalid_document` |
| `POST /drivers/{did}/documents/{docId}/review` | `{decision: approved\|rejected, note?}` | 200 | |
| `PUT /drivers/{did}/vehicle` | `{vehicleId}` | 200 | 422 `invalid_assignment` |
| `POST /drivers/{did}/status` | `{status: approved\|suspended, reason?}` | 200 | 409 `documents_incomplete` (needs valid driving_licence, national_id, vehicle_insurance) |

### Driver app (role `driver`)
| Method & path | Request | Success | Errors |
|---|---|---|---|
| `GET /driver/me` | — | own driver record, documents, vehicle | |
| `POST /driver/documents` | as above | 201 | |
| `POST /driver/availability` | `{available}` | `{availability}` | 409 `not_eligible` (lists issues), 409 `on_trip` |
| `POST /driver/location` | `{lat, lng, recordedAt?}` (≤10 min old buffered points keep their time) | 200 | 422 `invalid_location` |
| `GET /driver/work` | — | `{offer \| null, activeTrip \| null}` | |
| `POST /driver/offers/{oid}/accept` | — | 200 `{booking}` | 409 `offer_unavailable`, 409 `not_eligible`, 404 not your offer |
| `POST /driver/offers/{oid}/reject` | — | 200; next driver is offered | 409 `offer_unavailable` |
| `POST /driver/trips/{bid}/{en-route\|arrived\|start\|complete\|abort}` | abort: `{reason}` | 200 `{booking}` | 409 `invalid_transition`, 422 `reason_required` |
| `POST /driver/trips/{bid}/cash` | `{amountMinor, note?}` | 201 `{matchesFare}` | 409 `already_recorded`/`invalid_state`, 422 `note_required` when amount ≠ fare |

### Passenger app (role `passenger`)
| Method & path | Request | Success | Errors |
|---|---|---|---|
| `GET /v1/public/tenants/{slug}` | — | `{tenant{id,name}, categories[]}` (active tenants only) | 404 |
| `POST /v1/public/tenants/{slug}/passengers` | `{fullName, email, phone, password}` | 201 `{token, tenantId, customerId}` | 409 `already_registered`/`phone_taken`, 401 existing email + wrong password |
| `POST /quotes` | `{pickup{lat,lng,address}, dropoff{…}, categoryId, scheduledFor?}` | 201 `{quote}` — `isEstimate: true`, breakdown, fare rule version, 10-min expiry | 422 `outside_service_area`, `category_unavailable`, `invalid_schedule` (20 min–30 days ahead) |
| `POST /my/bookings` | header `Idempotency-Key` (8–128); `{quoteId, paymentMethod: "cash", note?}` | 201 new · 200 replay of the same key | 422 `idempotency_key_required`, `quote_expired`, `payment_method_unavailable`; 409 `idempotency_key_reused`, `quote_already_used` |
| `GET /my/bookings`, `GET /my/bookings/{bid}` | — | booking; driver first name, vehicle, live location + freshness while the trip is live (no phone) | 404 not yours |
| `POST /my/bookings/{bid}/cancel` | `{reason?}` | 200 | 409 `invalid_transition` once in progress/completed |
| `GET /my/bookings/{bid}/receipt` | — | final server fare breakdown, distance source, payment status | 409 `no_receipt` before completion |
| `POST /my/bookings/{bid}/rating` | `{stars 1-5, comment?}` | 201 | 409 `already_rated`/`not_completed` |

### Operator / dispatcher console
| Method & path | Permission | Request | Success | Errors |
|---|---|---|---|---|
| `POST /bookings` | `dispatch.manage` | `Idempotency-Key`; `{customer{fullName, phone}, pickup, dropoff, categoryId, scheduledFor?, note?}` | 201 / 200 replay | as quotes |
| `GET /bookings?status=` | `booking.read` | — | latest 200 | |
| `GET /bookings/{bid}` | `booking.read` | — | full view incl. customer phone, events, offers, payments | |
| `POST /bookings/{bid}/assign` | `dispatch.manage` | `{driverId}` | 200 (offer to that driver, 90 s; `driverLocationFresh`) | 409 `invalid_state`, `driver_not_eligible` |
| `POST /bookings/{bid}/cancel` | `dispatch.manage` | `{reason}` | 200 | 409 `invalid_transition` |
| `GET /dispatch/board` | `dispatch.manage` | — | open bookings + non-offline drivers with freshness, compliance, pending offers | |
| `GET /reports/summary?from&to` | `reports.read` | YYYY-MM-DD in tenant TZ (default last 7 days) | metrics + definitions + period + freshness | 422 `invalid_date`/`invalid_range` |

Edge limits (packline nginx): `/v1/auth/*` and passenger sign-up share a per-IP limit of 10 req/min (burst 5) → 429; other routes 20 req/s (burst 40).
