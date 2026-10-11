-- Uploaded files (driver documents, business verification documents). Bytes live outside the
-- database (local volume now, object storage later); this table is the authority on ownership.

CREATE TABLE files (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    uuid NOT NULL REFERENCES tenants(id),
    uploaded_by  uuid NOT NULL REFERENCES users(id),
    content_type text NOT NULL CHECK (content_type IN ('image/jpeg', 'image/png', 'image/webp', 'application/pdf')),
    size_bytes   bigint NOT NULL CHECK (size_bytes BETWEEN 1 AND 10485760),
    sha256       bytea NOT NULL,
    storage_key  text NOT NULL UNIQUE,   -- "<tenant_id>/<file_id>": keys are tenant-prefixed by construction
    file_name    text NOT NULL CHECK (length(file_name) BETWEEN 1 AND 200),
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX files_tenant_idx ON files(tenant_id);
ALTER TABLE files ENABLE ROW LEVEL SECURITY;
ALTER TABLE files FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON files USING (app_platform_scope() OR tenant_id = app_tenant_id());

ALTER TABLE driver_documents ADD COLUMN file_id uuid REFERENCES files(id);

-- Business verification documents reviewed by platform staff (D-021).
CREATE TABLE tenant_documents (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id),
    doc_type    text NOT NULL CHECK (doc_type IN ('business_registration', 'owner_id', 'tax_registration', 'other')),
    file_id     uuid NOT NULL REFERENCES files(id),
    note        text CHECK (length(note) <= 500),
    uploaded_by uuid NOT NULL REFERENCES users(id),
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tenant_documents_tenant_idx ON tenant_documents(tenant_id);
ALTER TABLE tenant_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_documents FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_documents USING (app_platform_scope() OR tenant_id = app_tenant_id());
