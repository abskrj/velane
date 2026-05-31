CREATE TABLE IF NOT EXISTS connections (
    id                  TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    tenant_id           TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider            TEXT NOT NULL,
    nango_connection_id TEXT NOT NULL,
    display_name        TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_connections_tenant ON connections(tenant_id);

DROP TABLE IF EXISTS library_versions;
DROP TABLE IF EXISTS libraries;
