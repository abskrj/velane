CREATE TABLE IF NOT EXISTS integration_credential_profiles (
    id                        TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    tenant_id                 TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider                  TEXT NOT NULL,
    alias                     TEXT NOT NULL DEFAULT 'default',
    name                      TEXT NOT NULL,
    nango_provider_config_key TEXT NOT NULL UNIQUE,
    config_json               JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_default                BOOLEAN NOT NULL DEFAULT FALSE,
    created_by                TEXT NOT NULL DEFAULT '',
    deleted_at                TIMESTAMPTZ,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_integration_credential_profiles_tenant_provider_alias_active
    ON integration_credential_profiles (tenant_id, provider, alias)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_integration_credential_profiles_tenant_provider_default_active
    ON integration_credential_profiles (tenant_id, provider)
    WHERE is_default = TRUE AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_integration_credential_profiles_tenant_provider_active
    ON integration_credential_profiles (tenant_id, provider, created_at)
    WHERE deleted_at IS NULL;

ALTER TABLE connections
    ADD COLUMN IF NOT EXISTS provider_config_key TEXT,
    ADD COLUMN IF NOT EXISTS credential_profile_id TEXT REFERENCES integration_credential_profiles(id) ON DELETE SET NULL;

UPDATE connections
SET provider_config_key = provider
WHERE provider_config_key IS NULL OR provider_config_key = '';

ALTER TABLE connections
    ALTER COLUMN provider_config_key SET NOT NULL;
