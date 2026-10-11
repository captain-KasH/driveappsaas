-- M4 self-service onboarding (D-021): email verification + password reset tokens, business profile,
-- a pending_verification state and the platform verification decision.

ALTER TABLE users ADD COLUMN email_verified_at timestamptz;
-- Every existing account came through a trusted path (admin bootstrap or an invitation sent to that address).
UPDATE users SET email_verified_at = created_at;

CREATE TABLE user_tokens (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id),
    purpose    text NOT NULL CHECK (purpose IN ('verify_email', 'reset_password')),
    token_hash bytea NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX user_tokens_user_idx ON user_tokens(user_id);
ALTER TABLE user_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_tokens FORCE ROW LEVEL SECURITY;
-- Tokens are looked up by hash before any identity is known, so that path runs in platform scope.
CREATE POLICY user_tokens_access ON user_tokens USING (app_platform_scope() OR user_id = app_user_id());

ALTER TABLE tenants DROP CONSTRAINT tenants_status_check;
ALTER TABLE tenants ADD CONSTRAINT tenants_status_check
    CHECK (status IN ('onboarding', 'pending_verification', 'active', 'suspended', 'closed'));

ALTER TABLE tenants
    ADD COLUMN legal_name        text CHECK (length(legal_name) BETWEEN 1 AND 200),
    ADD COLUMN registration_no   text CHECK (length(registration_no) BETWEEN 1 AND 64),
    ADD COLUMN tax_no            text CHECK (length(tax_no) BETWEEN 1 AND 64),
    ADD COLUMN contact_email     citext,
    ADD COLUMN contact_phone     text,
    ADD COLUMN billing_email     citext,
    ADD COLUMN address           text CHECK (length(address) <= 500),
    ADD COLUMN language          text NOT NULL DEFAULT 'en' CHECK (language IN ('en', 'si', 'ta')),
    ADD COLUMN created_by        uuid REFERENCES users(id),
    ADD COLUMN submitted_at      timestamptz,
    ADD COLUMN verified_at       timestamptz,
    ADD COLUMN verified_by       uuid REFERENCES users(id),
    ADD COLUMN verification_note text;
-- Tenants that exist today were all created by platform staff: trusted, hence verified.
UPDATE tenants SET verified_at = created_at;
CREATE INDEX tenants_status_idx ON tenants(status);
