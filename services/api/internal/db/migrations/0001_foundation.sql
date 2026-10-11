-- Stage 1 foundation: identity, tenants, memberships, invitations, sessions, audit.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

-- Helpers read the per-transaction context set by the API (see internal/db/tx.go).
CREATE FUNCTION app_tenant_id() RETURNS uuid LANGUAGE sql STABLE AS
$$ SELECT nullif(current_setting('app.tenant_id', true), '')::uuid $$;
CREATE FUNCTION app_user_id() RETURNS uuid LANGUAGE sql STABLE AS
$$ SELECT nullif(current_setting('app.user_id', true), '')::uuid $$;
CREATE FUNCTION app_platform_scope() RETURNS boolean LANGUAGE sql STABLE AS
$$ SELECT coalesce(current_setting('app.platform', true), '') = 'on' $$;

CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         citext NOT NULL UNIQUE,
    full_name     text NOT NULL CHECK (length(full_name) BETWEEN 1 AND 200),
    password_hash text,
    platform_role text CHECK (platform_role IN ('super_admin', 'support')),
    disabled_at   timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id),
    token_hash   bytea NOT NULL UNIQUE,
    created_at   timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL,
    revoked_at   timestamptz,
    user_agent   text,
    ip           inet
);
CREATE INDEX sessions_user_idx ON sessions(user_id);

CREATE TABLE tenants (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug          text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$'),
    name          text NOT NULL CHECK (length(name) BETWEEN 1 AND 200),
    business_type text NOT NULL CHECK (business_type IN (
        'taxi', 'three_wheeler', 'independent_driver', 'airport_transfer', 'chauffeur',
        'tourism', 'hospitality', 'corporate', 'shuttle', 'startup', 'enterprise_fleet')),
    country_code  char(2) NOT NULL DEFAULT 'LK',
    status        text NOT NULL DEFAULT 'onboarding' CHECK (status IN ('onboarding', 'active', 'suspended', 'closed')),
    isolation     text NOT NULL DEFAULT 'shared' CHECK (isolation IN ('shared', 'dedicated_db', 'dedicated_deployment')),
    status_reason text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tenant_memberships (
    tenant_id  uuid NOT NULL REFERENCES tenants(id),
    user_id    uuid NOT NULL REFERENCES users(id),
    role       text NOT NULL CHECK (role IN ('owner', 'admin', 'dispatcher', 'finance', 'viewer')),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, user_id)
);
CREATE INDEX tenant_memberships_user_idx ON tenant_memberships(user_id);

CREATE TABLE tenant_invitations (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id),
    email       citext NOT NULL,
    role        text NOT NULL CHECK (role IN ('owner', 'admin', 'dispatcher', 'finance', 'viewer')),
    token_hash  bytea NOT NULL UNIQUE,
    invited_by  uuid NOT NULL REFERENCES users(id),
    expires_at  timestamptz NOT NULL,
    accepted_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tenant_invitations_tenant_idx ON tenant_invitations(tenant_id);

-- Append-only. tenant_id NULL = platform-level event.
CREATE TABLE audit_events (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    occurred_at   timestamptz NOT NULL DEFAULT now(),
    tenant_id     uuid REFERENCES tenants(id),
    actor_user_id uuid REFERENCES users(id),
    action        text NOT NULL,
    target_type   text NOT NULL,
    target_id     text NOT NULL,
    reason        text,
    request_id    text,
    metadata      jsonb NOT NULL DEFAULT '{}'
);
CREATE INDEX audit_events_tenant_idx ON audit_events(tenant_id, occurred_at DESC);

CREATE FUNCTION audit_events_immutable() RETURNS trigger LANGUAGE plpgsql AS
$$ BEGIN RAISE EXCEPTION 'audit_events is append-only'; END $$;
CREATE TRIGGER audit_events_no_update BEFORE UPDATE OR DELETE ON audit_events
    FOR EACH ROW EXECUTE FUNCTION audit_events_immutable();

-- Row-level security. FORCE so the table owner is bound too; only superusers bypass.
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenants FORCE ROW LEVEL SECURITY;
CREATE POLICY tenants_access ON tenants USING (
    app_platform_scope()
    OR id = app_tenant_id()
    OR id IN (SELECT tenant_id FROM tenant_memberships WHERE user_id = app_user_id())
);

ALTER TABLE tenant_memberships ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_memberships FORCE ROW LEVEL SECURITY;
CREATE POLICY memberships_access ON tenant_memberships USING (
    app_platform_scope() OR tenant_id = app_tenant_id() OR user_id = app_user_id()
) WITH CHECK (app_platform_scope() OR tenant_id = app_tenant_id());

ALTER TABLE tenant_invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_invitations FORCE ROW LEVEL SECURITY;
CREATE POLICY invitations_access ON tenant_invitations USING (
    app_platform_scope() OR tenant_id = app_tenant_id()
);

ALTER TABLE audit_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_events FORCE ROW LEVEL SECURITY;
CREATE POLICY audit_read ON audit_events FOR SELECT USING (
    app_platform_scope() OR (tenant_id IS NOT NULL AND tenant_id = app_tenant_id())
);
CREATE POLICY audit_insert ON audit_events FOR INSERT WITH CHECK (
    app_platform_scope() OR (tenant_id IS NOT NULL AND tenant_id = app_tenant_id())
);
