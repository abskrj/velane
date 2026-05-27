# Runeforge — Build Phases

This document tracks the phased implementation plan for Runeforge. Each phase delivers a working, shippable increment. Phases build on each other — complete Phase N before starting Phase N+1.

---

## Phase 1 — Core Runtime (complete)

**Goal:** A fully runnable local stack that can execute code snippets synchronously.

### Delivered

- Go control plane HTTP API (chi router)
- Postgres schema with auto-applied migrations
- API key authentication (prefix lookup + bcrypt, `invoke` / `manage` / `admin` scopes)
- Snippet CRUD — create, list, get, delete
- Versioned deployments — each publish creates an immutable numbered version
- `dev` and `prod` environments per snippet
- Sync invocation: `POST /v1/invoke/{tenant}/{snippet}` → blocks until result
- ProcessExecutor — HTTP bridge to per-language executor containers (Bun, Python)
- Executor runtimes — Bun (`runner.ts`) and Python (`runner.py`) as persistent HTTP servers
- Tenant isolation enforced on every read and invocation
- Full test suite: 28 unit tests (zero deps) + integration tests (require `TEST_DATABASE_URL`)
- `docker-compose.yml` for local development

### Stack

- Control plane: Go 1.22, chi, pgx/v5, zap
- Runtimes: Bun 1.1, Python 3.12 + FastAPI
- Storage: Postgres 16
- Infrastructure: Docker Compose (local), Kubernetes + Terraform (production path)

---

## Phase 2 — Async, Streaming & Warm Pool (complete)

**Goal:** Support long-running and streaming snippets at scale; eliminate cold-start penalty for hot snippets.

### Delivered

- **Async invocation** — `X-Invoke-Mode: async` returns `202 { invocation_id }` immediately; snippet runs in background
- **Async polling** — `GET /v1/invocations/{id}` already exists; Phase 2 adds webhook delivery on completion (`callback_url` in invoke body)
- **Streaming invocation** — `X-Invoke-Mode: stream` returns `text/event-stream`; snippet yields chunks via `yield` (Python) or async generator (Bun)
- **Redis job queue** — async jobs enqueued to Redis, worker pool dequeues and dispatches to executor
- **Warm pool manager** — K8s Deployment per language; tenant-configurable `min_instances` per snippet; slot claim/release via Redis atomic ops
- **Version pinning** — callers can specify `?version=v3` to invoke a pinned version instead of the active one

### New API surface

```
POST /v1/invoke/{tenant}/{snippet}
  X-Invoke-Mode: async    → 202 { invocation_id, status_url }
  X-Invoke-Mode: stream   → 200 text/event-stream
  Body: { ..., callback_url?: string }  (async only)
```

### New infrastructure

- Redis (added to docker-compose and Terraform)
- Worker service (Go) — pulls from Redis queue, dispatches to executor, updates invocation record, fires webhook
- `SnippetEnvironment.min_instances` — warm pool manager watches this and maintains ready slots

---

## Phase 3 — Staging, Canary & Secrets (complete)

**Goal:** Full three-environment promotion flow with safe traffic-splitting and secret injection.

### Delivered

- **Staging environment** — third env (`dev` → `staging` → `prod`); `POST /v1/snippets/{id}/versions/{num}/publish?env=staging`
- **Canary traffic splitting** — `POST /v1/snippets/{id}/canary` sets `{ version_id, percent }` on the prod environment; Traffic Router sends X% to canary version, 100-X% to stable
- **Rollback** — re-publish any archived version to instantly swap active version
- **Secrets manager** — `POST /v1/secrets` to store encrypted key-value pairs; injected as env vars at invocation time; never returned in API responses
- **Egress policy engine** — per-tenant IP/CIDR blocklist enforced via iptables inside executor net namespace; default blocks `169.254.0.0/16`, RFC1918 ranges, and configurable domain sinkhole

### New API surface

```
POST   /v1/secrets                    → create secret (manage scope)
GET    /v1/secrets                    → list secret names (never values)
DELETE /v1/secrets/{id}              → delete secret

POST   /v1/snippets/{id}/canary       → set canary { version_id, percent }
DELETE /v1/snippets/{id}/canary       → remove canary (full traffic to active)

GET    /v1/tenants/{slug}/egress      → get egress policy
PUT    /v1/tenants/{slug}/egress      → update egress policy (admin scope)
```

### New data model

```sql
secrets (id, tenant_id, snippet_id nullable, name, value_encrypted, environments[])
snippet_environments.canary_version_id
snippet_environments.canary_pct
```

---

## Phase 4 — Developer Surfaces (complete)

**Goal:** Give engineers three ways to write and deploy snippets: Web IDE, CLI, and Git push-to-deploy.

### Delivered

- **Web IDE** — Monaco editor in React; syntax highlighting for Bun/Python; inline error display; test-invoke panel; version history sidebar; publish button with env selector
- **CLI tool** (`runeforge` binary, distributed via npm and Homebrew)
  ```
  runeforge login               # authenticate, store key in system keychain
  runeforge snippets list
  runeforge snippets push <file>  # create/update draft, optionally publish
  runeforge invoke <slug> [--env prod] [--input '{}']
  runeforge logs <slug>
  ```
- **Git webhook integration** — connect a GitHub/GitLab repo; push to `main` → deploy to `staging`; push a tag (`v*`) → deploy to `prod`; PR branch → deploy to `dev` (preview env)

### New services

- `web-ide/` — Vite + React SPA, deployed to `app.runeforge.io`
- `cli/` — Go CLI (cobra), distributed as single binary
- Webhook receiver endpoint on control plane: `POST /v1/webhooks/git`

---

## Phase 5 — Observability (complete)

**Goal:** Give engineers full visibility into every invocation — logs, metrics, and replay.

### Delivered

- **Invocation schema upgrades** — added `input_ref`, `output_ref`, `stderr_ref`, and `cpu_ms` to invocation model/migrations
- **Log query API** — `GET /v1/logs/snippets/{id}` implemented with filters (`limit`, `env`, `status`, `start_time`, `end_time`)
- **Metrics API** — `GET /v1/metrics/snippets/{id}?window=1h|24h|7d` implemented with aggregates (`count`, `avg`, `p50`, `p95`, `p99`) + time series
- **Replay API** — `POST /v1/invocations/{id}/replay` implemented with manage-scope enforcement + tenant-level replay opt-in
- **Replay privacy toggle** — tenant-level `replay_enabled` gate enforced before replay
- **Post-invocation observability hooks** — scheduler + worker now emit completion events into pluggable observability pipeline interfaces
- **Multi-tenant namespace provider** — added pluggable `TenantProvider` interface with `SharedTenantProvider` and `NamespacedTenantProvider`
- **Infra placeholders wired** — ClickHouse and MinIO services + related env vars added to local compose for Phase 5 runtime plumbing

### New infrastructure

- ClickHouse (added to docker-compose and Terraform)
- MinIO / S3 bucket (log and replay payload storage)
- `ObservabilityWorker` — async goroutine pool that ships log lines and metrics rows after each invocation

### Notes

- Current implementation uses Postgres-backed query paths for logs and metrics APIs, with external-store interfaces in place for phased rollout.
- ClickHouse/MinIO are wired in local infrastructure and can be enabled incrementally behind the observability interfaces.

---

## Phase 6 — MCP Server (Cursor / Claude Code Integration) (complete)

**Goal:** Let engineers connect Cursor, Claude Code, or any MCP-compatible AI agent directly to Runeforge to generate and deploy snippets without leaving their IDE.

### Delivered

- **MCP server service** — standalone Go service at `services/mcp-server`
- **Two transports implemented**:
  - HTTP/SSE hosted at `/mcp` (zero install — add URL to IDE config)
  - stdio mode via `cmd/stdio` (for IDEs that only support stdio)
- **JSON-RPC methods implemented**:
  - `initialize`
  - `ping`
  - `tools/list`
  - `tools/call`
- **10 MCP tools exposed** and wired to control-plane APIs:

| Tool              | Scope needed |
| ----------------- | ------------ |
| `list_snippets`   | invoke       |
| `get_snippet`     | invoke       |
| `create_snippet`  | manage       |
| `update_draft`    | manage       |
| `publish_snippet` | manage       |
| `invoke_snippet`  | invoke       |
| `get_logs`        | invoke       |
| `list_secrets`    | manage       |
| `set_secret`      | manage       |
| `get_metrics`     | invoke       |

- **Auth passthrough** — Bearer token forwarded to control-plane for scope/tenant enforcement
- **Docker Compose integration** — `mcp-server` service added, listening on `:8090`, configured with `CONTROL_PLANE_URL=http://control-plane:8080`
- **MCP usage docs** — root README includes compose startup and JSON-RPC usage examples

### Developer setup (after this phase ships)

```json
// .cursor/mcp.json or ~/.claude/mcp.json
{
  "mcpServers": {
    "runeforge": {
      "url": "https://api.runeforge.io/mcp",
      "headers": { "Authorization": "Bearer rf_xxxx" }
    }
  }
}
```

### New services

- `services/mcp-server/` — Go service, thin wrapper over control plane API

---

## Phase 7 — Embeddable Dashboard (complete)

**Goal:** Let any org embed a white-label snippet browser directly into their own portal via a single `<iframe>` tag.

### Scope

- **Embed token API** — `POST /v1/embed/tokens` issues short-lived, read-only tokens scoped to a tenant (optionally to specific snippet IDs)
- **Embed app** — React SPA served from `embed.runeforge.io`:
  - Snippet list with search, language filter, env filter
  - Snippet detail: code viewer (Monaco, read-only), version sidebar, env status badges, recent invocations summary, p95 latency badge
  - Environment switcher (dev / staging / prod)
- **White-label theming** — theme via URL params (`?theme=dark&accent=6366f1`) or persisted branding config per tenant (`logo_url`, `accent_color`, `font_family`)
- **iframe security** — embed subdomain uses `Content-Security-Policy: frame-ancestors *`; main app keeps `frame-ancestors 'none'`

### Integration (one line for orgs)

```html
<iframe
  src="https://embed.runeforge.io?token=et_xxxx"
  width="100%"
  height="700"
  frameborder="0"
/>
```

### New services

- `services/embed-dashboard/` — Vite + React, deployed to `embed.runeforge.io`

---

## Phase 8 — Tenant Admin Dashboard & White-Label

**Goal:** Give each tenant org a self-serve admin dashboard to manage their Runeforge account, configure white-label branding for the embedded dashboard, and govern their engineers' access.

### Scope

- **Tenant admin portal** — separate React app at `admin.runeforge.io`; accessible to users with the `admin` API scope
- **White-label branding config** — org admins set logo URL, accent colour, font family, and custom domain from a UI; stored per-tenant in DB; fetched by the embed app on load; URL params remain as overrides
- **Custom domain for embed** — org configures `snippets.acme.com` → points to `embed.runeforge.io` via CNAME; TLS via Let's Encrypt / cert-manager
- **API key management UI** — create, revoke, and scope API keys; view last-used timestamps; copy key on creation (never shown again)
- **Team member management** — invite engineers by email; assign `invoke / manage / admin` roles; revoke access
- **Usage dashboard** — invocation counts, GB-seconds consumed, error rates — per snippet and per time window; powered by Phase 5 ClickHouse metrics
- **Egress policy editor** — visual UI to add/remove blocked CIDRs and domains instead of raw JSON via API

### New data model

```sql
-- Branding config (extends existing egress_policy pattern on tenants)
ALTER TABLE tenants ADD COLUMN branding JSONB NOT NULL DEFAULT '{}';
-- { logo_url, accent_color, font_family, custom_domain, hide_branding }

-- Team members / user accounts
CREATE TABLE users (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT,           -- null if OAuth-only
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tenant_members (
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'manage' CHECK (role IN ('invoke', 'manage', 'admin')),
    invited_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, user_id)
);
```

### New API surface

```
GET  /v1/tenants/{slug}/branding              → get branding config
PUT  /v1/tenants/{slug}/branding              → update branding (admin scope)

GET  /v1/tenants/{slug}/members               → list team members
POST /v1/tenants/{slug}/members/invite        → invite by email
DELETE /v1/tenants/{slug}/members/{userID}    → revoke access

GET  /v1/tenants/{slug}/usage                 → usage summary (from ClickHouse)
```

### New services

- `apps/admin/` — Vite + React, deployed to `admin.runeforge.io`
- Reuses `packages/ui` components from Phase 4 (MetricsBadge, API key management UI)

### Relationship to Phase 7 embed

Phase 7 builds the embed app and accepts branding config as URL params. Phase 8 adds the admin UI where orgs configure that branding through a proper form — the embed app simply fetches it from `GET /v1/tenants/{slug}/branding` on load instead of relying on URL params alone. No changes to the embed app itself.

---

## Phase 9 — Hardening & Advanced Features

**Goal:** Production-grade security hardening, full schema-driven API docs, and enterprise auth.

### Scope

- **Firecracker executor plugin** — pluggable `Executor` interface implementation using AWS Firecracker microVMs; VM-boundary isolation; snapshot/restore for sub-50ms warm starts; requires KVM (bare metal or metal EC2 instances)
- **OpenAPI spec generation** — at publish time, extract Zod / Pydantic schemas and emit a full OpenAPI 3.1 spec for the snippet's invoke endpoint; expose at `GET /v1/snippets/{id}/openapi.json`
- **JWT auth** — RS256 JWTs as an alternative to API keys; short-lived (15min) + refresh tokens; intended for Web IDE sessions and user-facing callers; also enables Phase 8 team member login without API keys
- **Seccomp profiles** — production-grade syscall allowlist for the ProcessExecutor; block `ptrace`, `mount`, `clone(CLONE_NEWUSER)`, `perf_event_open`, etc.
- **Audit log** — append-only log of all management actions (publish, secret create, egress change, member invite) per tenant; queryable by admin

---

## Summary

| Phase | Theme          | Key deliverable                                                             |
| ----- | -------------- | --------------------------------------------------------------------------- |
| 1     | Core Runtime   | Sync invocation, API keys, Postgres, docker-compose                         |
| 2     | Scale          | Async + streaming, Redis queue, warm pool                                   |
| 3     | Safety         | Staging, canary, secrets, egress policy                                     |
| 4     | DX             | Web IDE (Monaco + test runner + log streaming), CLI, git push-to-deploy     |
| 5     | Visibility     | Logs, metrics, replay, multi-tenant K8s namespaces                          |
| 6     | AI Integration | MCP server (Cursor / Claude Code integration)                               |
| 7     | Embedding      | iframe embed dashboard, embed tokens, read-only snippet viewer              |
| 8     | Tenant Admin   | Admin portal, white-label branding config, team management, usage dashboard |
| 9     | Hardening      | Firecracker, OpenAPI gen, JWT, seccomp profiles, audit log                  |
