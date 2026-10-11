-- Stage 2–3 core mobility: zones, categories, versioned fares, fleet, bookings, dispatch, trips, cash.
CREATE EXTENSION IF NOT EXISTS postgis;

ALTER TABLE tenants
    ADD COLUMN timezone text NOT NULL DEFAULT 'Asia/Colombo',
    ADD COLUMN currency char(3) NOT NULL DEFAULT 'LKR';

-- Drivers and passengers are tenant members too, with no staff permissions.
ALTER TABLE tenant_memberships DROP CONSTRAINT tenant_memberships_role_check;
ALTER TABLE tenant_memberships ADD CONSTRAINT tenant_memberships_role_check
    CHECK (role IN ('owner', 'admin', 'dispatcher', 'finance', 'viewer', 'driver', 'passenger'));
ALTER TABLE tenant_invitations DROP CONSTRAINT tenant_invitations_role_check;
ALTER TABLE tenant_invitations ADD CONSTRAINT tenant_invitations_role_check
    CHECK (role IN ('owner', 'admin', 'dispatcher', 'finance', 'viewer', 'driver'));

CREATE TABLE service_zones (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  uuid NOT NULL REFERENCES tenants(id),
    name       text NOT NULL CHECK (length(name) BETWEEN 1 AND 120),
    area       geometry(MultiPolygon, 4326) NOT NULL CHECK (ST_IsValid(area)),
    active     boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, name)
);
CREATE INDEX service_zones_area_idx ON service_zones USING gist (area);

CREATE TABLE vehicle_categories (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  uuid NOT NULL REFERENCES tenants(id),
    code       text NOT NULL CHECK (code ~ '^[a-z0-9_]{2,32}$'),
    name       text NOT NULL CHECK (length(name) BETWEEN 1 AND 80),
    seats      int NOT NULL CHECK (seats BETWEEN 1 AND 60),
    active     boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, code)
);

CREATE TABLE fare_rules (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id),
    zone_id     uuid NOT NULL REFERENCES service_zones(id),
    category_id uuid NOT NULL REFERENCES vehicle_categories(id),
    current_version int NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, zone_id, category_id)
);

-- Immutable: a booking references the exact version it was priced with.
CREATE TABLE fare_rule_versions (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id             uuid NOT NULL REFERENCES tenants(id),
    fare_rule_id          uuid NOT NULL REFERENCES fare_rules(id),
    version               int NOT NULL CHECK (version > 0),
    currency              char(3) NOT NULL,
    base_minor            bigint NOT NULL CHECK (base_minor >= 0),
    per_km_minor          bigint NOT NULL CHECK (per_km_minor >= 0),
    per_minute_minor      bigint NOT NULL CHECK (per_minute_minor >= 0),
    minimum_minor         bigint NOT NULL CHECK (minimum_minor >= 0),
    booking_fee_minor     bigint NOT NULL DEFAULT 0 CHECK (booking_fee_minor >= 0),
    scheduled_fee_minor   bigint NOT NULL DEFAULT 0 CHECK (scheduled_fee_minor >= 0),
    waiting_per_min_minor bigint NOT NULL DEFAULT 0 CHECK (waiting_per_min_minor >= 0),
    free_waiting_minutes  int NOT NULL DEFAULT 0 CHECK (free_waiting_minutes >= 0),
    night_surcharge_bps   int NOT NULL DEFAULT 0 CHECK (night_surcharge_bps BETWEEN 0 AND 10000),
    night_start           time,
    night_end             time,
    tax_bps               int NOT NULL DEFAULT 0 CHECK (tax_bps BETWEEN 0 AND 10000),
    round_to_minor        bigint NOT NULL DEFAULT 100 CHECK (round_to_minor > 0),
    created_by            uuid NOT NULL REFERENCES users(id),
    created_at            timestamptz NOT NULL DEFAULT now(),
    UNIQUE (fare_rule_id, version),
    CHECK ((night_start IS NULL) = (night_end IS NULL))
);

CREATE TABLE vehicles (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id),
    category_id uuid NOT NULL REFERENCES vehicle_categories(id),
    plate       text NOT NULL CHECK (length(plate) BETWEEN 2 AND 20),
    make        text NOT NULL DEFAULT '',
    model       text NOT NULL DEFAULT '',
    colour      text NOT NULL DEFAULT '',
    status      text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'suspended')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, plate)
);

CREATE TABLE drivers (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id),
    user_id      uuid REFERENCES users(id),
    full_name    text NOT NULL CHECK (length(full_name) BETWEEN 1 AND 200),
    phone        text NOT NULL CHECK (phone ~ '^\+[1-9][0-9]{7,14}$'),
    email        citext NOT NULL,
    status       text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'suspended')),
    availability text NOT NULL DEFAULT 'offline' CHECK (availability IN ('offline', 'available', 'busy')),
    vehicle_id   uuid REFERENCES vehicles(id),
    location     geometry(Point, 4326),
    location_at  timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, phone),
    UNIQUE (tenant_id, user_id)
);
CREATE INDEX drivers_location_idx ON drivers USING gist (location);

CREATE TABLE driver_documents (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id),
    driver_id   uuid NOT NULL REFERENCES drivers(id),
    doc_type    text NOT NULL CHECK (doc_type IN ('driving_licence', 'national_id', 'vehicle_insurance', 'revenue_licence', 'police_report')),
    number      text NOT NULL CHECK (length(number) BETWEEN 1 AND 64),
    expires_on  date,
    status      text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by uuid REFERENCES users(id),
    reviewed_at timestamptz,
    review_note text,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX driver_documents_driver_idx ON driver_documents(driver_id);

ALTER TABLE tenant_invitations ADD COLUMN driver_id uuid REFERENCES drivers(id);

CREATE TABLE customers (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  uuid NOT NULL REFERENCES tenants(id),
    user_id    uuid REFERENCES users(id),
    full_name  text NOT NULL CHECK (length(full_name) BETWEEN 1 AND 200),
    phone      text NOT NULL CHECK (phone ~ '^\+[1-9][0-9]{7,14}$'),
    phone_verified_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, phone),
    UNIQUE (tenant_id, user_id)
);

CREATE TABLE booking_quotes (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        uuid NOT NULL REFERENCES tenants(id),
    customer_id      uuid NOT NULL REFERENCES customers(id),
    zone_id          uuid NOT NULL REFERENCES service_zones(id),
    category_id      uuid NOT NULL REFERENCES vehicle_categories(id),
    fare_rule_version_id uuid NOT NULL REFERENCES fare_rule_versions(id),
    pickup           geometry(Point, 4326) NOT NULL,
    pickup_address   text NOT NULL,
    dropoff          geometry(Point, 4326) NOT NULL,
    dropoff_address  text NOT NULL,
    scheduled_for    timestamptz,
    est_distance_m   int NOT NULL,
    est_duration_s   int NOT NULL,
    estimate_method  text NOT NULL,
    currency         char(3) NOT NULL,
    est_total_minor  bigint NOT NULL,
    breakdown        jsonb NOT NULL,
    expires_at       timestamptz NOT NULL,
    created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE bookings (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        uuid NOT NULL REFERENCES tenants(id),
    reference        text NOT NULL,
    customer_id      uuid NOT NULL REFERENCES customers(id),
    quote_id         uuid NOT NULL REFERENCES booking_quotes(id),
    created_by       uuid REFERENCES users(id),
    channel          text NOT NULL CHECK (channel IN ('passenger_app', 'dispatcher')),
    idempotency_key  text NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 128),
    status           text NOT NULL CHECK (status IN ('scheduled', 'searching', 'assigned', 'completed', 'cancelled', 'expired')),
    payment_method   text NOT NULL CHECK (payment_method IN ('cash')),
    scheduled_for    timestamptz,
    search_started_at timestamptz,
    driver_id        uuid REFERENCES drivers(id),
    vehicle_id       uuid REFERENCES vehicles(id),
    assigned_at      timestamptz,
    cancelled_by     uuid REFERENCES users(id),
    cancel_reason    text,
    passenger_note   text,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, idempotency_key),
    UNIQUE (tenant_id, reference),
    UNIQUE (quote_id)
);
CREATE INDEX bookings_tenant_status_idx ON bookings(tenant_id, status, created_at DESC);
CREATE INDEX bookings_customer_idx ON bookings(customer_id, created_at DESC);

CREATE TABLE booking_events (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id     uuid NOT NULL REFERENCES tenants(id),
    booking_id    uuid NOT NULL REFERENCES bookings(id),
    event         text NOT NULL,
    from_status   text,
    to_status     text,
    actor_user_id uuid REFERENCES users(id),
    reason        text,
    data          jsonb NOT NULL DEFAULT '{}',
    occurred_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX booking_events_booking_idx ON booking_events(booking_id, id);
CREATE TRIGGER booking_events_no_update BEFORE UPDATE OR DELETE ON booking_events
    FOR EACH ROW EXECUTE FUNCTION audit_events_immutable();

CREATE TABLE dispatch_offers (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id),
    booking_id   uuid NOT NULL REFERENCES bookings(id),
    driver_id    uuid NOT NULL REFERENCES drivers(id),
    source       text NOT NULL CHECK (source IN ('auto', 'dispatcher')),
    distance_m   int,
    status       text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'expired', 'withdrawn')),
    offered_at   timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL,
    responded_at timestamptz
);
-- One live offer per booking and per driver: the exclusive-assignment guarantee.
CREATE UNIQUE INDEX dispatch_offers_one_pending_per_booking ON dispatch_offers(booking_id) WHERE status = 'pending';
CREATE UNIQUE INDEX dispatch_offers_one_pending_per_driver ON dispatch_offers(driver_id) WHERE status = 'pending';
CREATE INDEX dispatch_offers_expiry_idx ON dispatch_offers(expires_at) WHERE status = 'pending';

CREATE TABLE trips (
    booking_id      uuid PRIMARY KEY REFERENCES bookings(id),
    tenant_id       uuid NOT NULL REFERENCES tenants(id),
    driver_id       uuid NOT NULL REFERENCES drivers(id),
    vehicle_id      uuid NOT NULL REFERENCES vehicles(id),
    status          text NOT NULL CHECK (status IN ('driver_assigned', 'en_route', 'arrived', 'in_progress', 'completed', 'aborted')),
    en_route_at     timestamptz,
    arrived_at      timestamptz,
    started_at      timestamptz,
    completed_at    timestamptz,
    aborted_at      timestamptz,
    abort_reason    text,
    distance_m      int,
    duration_s      int,
    waiting_s       int,
    distance_source text,
    final_total_minor bigint,
    currency        char(3),
    final_breakdown jsonb,
    created_at      timestamptz NOT NULL DEFAULT now()
);
-- A driver can hold at most one live trip.
CREATE UNIQUE INDEX trips_one_live_per_driver ON trips(driver_id) WHERE status NOT IN ('completed', 'aborted');

CREATE TABLE location_updates (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   uuid NOT NULL REFERENCES tenants(id),
    driver_id   uuid NOT NULL REFERENCES drivers(id),
    booking_id  uuid REFERENCES bookings(id),
    location    geometry(Point, 4326) NOT NULL,
    recorded_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX location_updates_trip_idx ON location_updates(booking_id, recorded_at) WHERE booking_id IS NOT NULL;

-- Append-only money records. Corrections are new rows with kind = 'adjustment'.
CREATE TABLE payments (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id),
    booking_id   uuid NOT NULL REFERENCES bookings(id),
    kind         text NOT NULL CHECK (kind IN ('cash_collected', 'adjustment')),
    amount_minor bigint NOT NULL,
    currency     char(3) NOT NULL,
    recorded_by  uuid NOT NULL REFERENCES users(id),
    note         text,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX payments_one_cash_per_booking ON payments(booking_id) WHERE kind = 'cash_collected';
CREATE TRIGGER payments_no_update BEFORE UPDATE OR DELETE ON payments
    FOR EACH ROW EXECUTE FUNCTION audit_events_immutable();

CREATE TABLE ratings (
    booking_id  uuid PRIMARY KEY REFERENCES bookings(id),
    tenant_id   uuid NOT NULL REFERENCES tenants(id),
    stars       int NOT NULL CHECK (stars BETWEEN 1 AND 5),
    comment     text CHECK (length(comment) <= 1000),
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Same isolation policy on every tenant-owned table.
DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY['service_zones','vehicle_categories','fare_rules','fare_rule_versions','vehicles',
    'drivers','driver_documents','customers','booking_quotes','bookings','booking_events','dispatch_offers',
    'trips','location_updates','payments','ratings'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format('CREATE POLICY tenant_isolation ON %I USING (app_platform_scope() OR tenant_id = app_tenant_id())
                    WITH CHECK (app_platform_scope() OR tenant_id = app_tenant_id())', t);
  END LOOP;
END $$;

CREATE TRIGGER fare_rule_versions_immutable BEFORE UPDATE OR DELETE ON fare_rule_versions
    FOR EACH ROW EXECUTE FUNCTION audit_events_immutable();

-- Single source of truth for whether a driver may be dispatched (availability and
-- location freshness are checked separately at dispatch time). security_invoker keeps RLS in force.
CREATE VIEW driver_compliance WITH (security_invoker = true) AS
SELECT d.id AS driver_id, d.tenant_id,
       array_remove(ARRAY[
         CASE WHEN d.status <> 'approved' THEN 'driver_not_approved' END,
         CASE WHEN v.id IS NULL THEN 'no_vehicle' WHEN v.status <> 'approved' THEN 'vehicle_not_approved' END,
         CASE WHEN EXISTS (
           SELECT 1 FROM unnest(ARRAY['driving_licence', 'national_id', 'vehicle_insurance']) AS req(doc_type)
           WHERE NOT EXISTS (
             SELECT 1 FROM driver_documents dd
             WHERE dd.driver_id = d.id AND dd.doc_type = req.doc_type AND dd.status = 'approved'
               AND (dd.expires_on IS NULL OR dd.expires_on >= (now() AT TIME ZONE t.timezone)::date)))
           THEN 'documents_missing_or_expired' END
       ], NULL) AS issues
FROM drivers d
JOIN tenants t ON t.id = d.tenant_id
LEFT JOIN vehicles v ON v.id = d.vehicle_id;
