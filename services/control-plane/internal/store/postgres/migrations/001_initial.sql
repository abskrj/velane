CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS tenants (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key_hash TEXT NOT NULL,
    key_prefix TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT '{}',
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS snippets (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    language TEXT NOT NULL CHECK (language IN ('bun', 'python')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by TEXT NOT NULL DEFAULT '',
    UNIQUE (tenant_id, slug)
);

CREATE TABLE IF NOT EXISTS snippet_versions (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    snippet_id TEXT NOT NULL REFERENCES snippets(id) ON DELETE CASCADE,
    version_number INT NOT NULL,
    code TEXT NOT NULL DEFAULT '',
    input_schema TEXT NOT NULL DEFAULT '{}',
    output_schema TEXT NOT NULL DEFAULT '{}',
    timeout_ms INT NOT NULL DEFAULT 30000,
    max_memory_mb INT NOT NULL DEFAULT 128,
    max_cpu_percent INT NOT NULL DEFAULT 100,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by TEXT NOT NULL DEFAULT '',
    UNIQUE (snippet_id, version_number)
);

CREATE TABLE IF NOT EXISTS snippet_environments (
    snippet_id TEXT NOT NULL REFERENCES snippets(id) ON DELETE CASCADE,
    env TEXT NOT NULL CHECK (env IN ('dev', 'prod')),
    active_version_id TEXT REFERENCES snippet_versions(id),
    min_instances INT NOT NULL DEFAULT 0,
    PRIMARY KEY (snippet_id, env)
);

CREATE TABLE IF NOT EXISTS invocations (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    snippet_id TEXT NOT NULL REFERENCES snippets(id),
    version_id TEXT NOT NULL REFERENCES snippet_versions(id),
    environment TEXT NOT NULL,
    tenant_id TEXT NOT NULL REFERENCES tenants(id),
    status TEXT NOT NULL DEFAULT 'pending',
    input_payload TEXT NOT NULL DEFAULT '{}',
    output TEXT,
    error TEXT,
    stderr TEXT,
    duration_ms INT,
    peak_memory_mb INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_invocations_snippet_id ON invocations(snippet_id);
CREATE INDEX IF NOT EXISTS idx_invocations_tenant_id ON invocations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invocations_created_at ON invocations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_snippet_versions_snippet_id ON snippet_versions(snippet_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON api_keys(key_prefix);
