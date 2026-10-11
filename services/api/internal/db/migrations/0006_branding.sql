-- Runtime branding (T-007, D-022): what the booking page and the apps render for a tenant.
-- Native app identity (bundle id, icon, store listing) is a build-time concern (T-173) and is not here.
ALTER TABLE tenants
    ADD COLUMN logo_file_id  uuid REFERENCES files(id),
    ADD COLUMN brand_color   text NOT NULL DEFAULT '#0f766e' CHECK (brand_color ~ '^#[0-9a-f]{6}$'),
    ADD COLUMN tagline       text CHECK (length(tagline) <= 160),
    ADD COLUMN support_phone text,
    ADD COLUMN support_email citext,
    ADD COLUMN website_url   text CHECK (website_url ~ '^https?://' AND length(website_url) <= 300),
    ADD COLUMN terms_url     text CHECK (terms_url ~ '^https?://' AND length(terms_url) <= 300),
    ADD COLUMN privacy_url   text CHECK (privacy_url ~ '^https?://' AND length(privacy_url) <= 300);
