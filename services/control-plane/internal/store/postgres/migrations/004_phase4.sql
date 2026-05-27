-- Git webhook integrations per snippet
CREATE TABLE IF NOT EXISTS git_integrations (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    tenant_id   TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    snippet_id  TEXT NOT NULL REFERENCES snippets(id) ON DELETE CASCADE,
    provider    TEXT NOT NULL CHECK (provider IN ('github', 'gitlab')),
    repo_url    TEXT NOT NULL,
    secret      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (snippet_id)
);
CREATE INDEX IF NOT EXISTS idx_git_integrations_tenant_id ON git_integrations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_git_integrations_snippet_id ON git_integrations(snippet_id);
