-- Drop the old one-connection-per-provider constraint.
ALTER TABLE connections DROP CONSTRAINT IF EXISTS connections_tenant_id_provider_key;

-- alias distinguishes multiple connections to the same provider (e.g. "prod" vs "sandbox").
-- Default 'default' keeps existing rows valid without a backfill.
ALTER TABLE connections ADD COLUMN IF NOT EXISTS alias TEXT NOT NULL DEFAULT 'default';

-- New unique key: one row per (tenant, provider, alias).
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'connections_tenant_provider_alias_key'
    ) THEN
        ALTER TABLE connections ADD CONSTRAINT connections_tenant_provider_alias_key
            UNIQUE (tenant_id, provider, alias);
    END IF;
END $$;

-- nango_connection_id is now nullable during the window between RecordConnection
-- (frontend) and webhook delivery (where the real Nango UUID arrives).
ALTER TABLE connections ALTER COLUMN nango_connection_id DROP NOT NULL;
