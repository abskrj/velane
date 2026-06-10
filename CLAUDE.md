# Velane — Claude Code Guide

## Working style

Before writing any code, always discuss the problem first. Ask clarifying questions until the requirements, constraints, and approach are clear. Only start implementing once the user has confirmed the plan. If a task is ambiguous, propose options and let the user choose rather than making assumptions.

## Repo layout

```
velane/
├── services/
│   ├── control-plane/      # Go 1.26 API server (chi, pgx/v5, zap)
│   ├── executor-runtime/   # Bun + Python sandboxed HTTP runners
│   ├── mcp-server/         # MCP protocol server for Cursor / Claude Code
│   └── cli/                # Cobra CLI (velane login / push / invoke)
├── apps/
│   ├── admin/              # Vite + React admin portal
│   └── embed-dashboard/    # Vite + React embeddable viewer
└── platform-libraries/     # Canonical source for built-in libs (bun/ + python/)
```

Each Go service has its own `go.mod`. Module paths:
- `github.com/abskrj/velane/services/control-plane`
- `github.com/abskrj/velane/services/mcp-server`
- `github.com/abskrj/velane/services/cli`

## Essential commands

```bash
# Build + start everything
make up           # docker compose up --build -d

# Stop + wipe volumes
make down         # docker compose down -v

# Compile control-plane locally (requires Go 1.26+)
make build        # runs copy-platform-libs first, then go build ./...

# Run Go tests (control-plane)
cd services/control-plane && go test ./...

# TypeScript type-check (admin)
cd apps/admin && npx tsc --noEmit

# Add a new platform library
# 1. Add files under platform-libraries/{bun|python}/<slug>/
# 2. Run: make copy-platform-libs
# 3. Rebuild: make build  (or make up)
```

## Go conventions (control-plane)

- **Never commit** unless explicitly asked. The user's standing instruction is: do not commit unless told.
- Always run `go build ./...` from inside `services/control-plane/` — not from the repo root.
- Handler pattern: each handler file exposes a `*Store` interface for its DB needs (e.g. `BrandingStore`). Use the narrowest interface, not `*postgres.Store`, except in `TenantsHandler` which needs direct store access.
- **Tenant isolation** — every slug-based handler must verify the authenticated tenant matches the URL slug:
  ```go
  authTenant := middleware.TenantFromContext(r.Context())
  if authTenant == nil || authTenant.ID != tenant.ID {
      writeError(w, http.StatusForbidden, "access denied")
      return
  }
  ```
  This is already in: egress, branding, usage, members handlers. Do NOT skip it for new slug-based endpoints.
- **Scope middleware** — all authenticated routes must use `middleware.RequireScope(scope, log)`. Minimum scopes: `invoke` for reads, `manage` for writes, `admin` for destructive/team actions.
- **Error helpers** — use `writeError(w, status, msg)` and `writeJSON(w, status, v)` from `handlers/helpers.go`. Never write raw JSON inline.
- **API key prefix** is `vl_`. Embed token prefix is `et_`. Do not change these.
- Migrations live in `internal/store/postgres/migrations/`. Number them sequentially (`011_`, `012_`, …). They run automatically on startup.
- Platform libraries are embedded via `//go:embed all:files` in `internal/platformlibs/loader.go`. The `files/` directory is gitignored (except `.gitkeep`). Always run `make copy-platform-libs` before building locally.

## Auth model

| Credential | Where stored | Validated by |
|---|---|---|
| Session JWT (RS256) | `localStorage.sessionToken` | `SessionAuth` middleware → `JWTProvider.ValidateSession` |
| API key (`vl_…`) | `localStorage.apiKey` | `Auth` middleware → `ValidateAPIKey` |
| Embed token (`et_…`) | `localStorage.apiKey` | `AuthEmbed` middleware or `Auth` middleware (synthetic key) |

JWT access tokens expire in 15 min; refresh tokens last 7 days. `ValidateSession` checks issuer — do not remove that check.

Embed tokens get a synthetic API key with scopes `["invoke", "manage"]` — they must NOT have `admin` scope.

## Frontend conventions (apps/admin)

- API calls go through `src/lib/api.ts`. The `request()` function handles 401s automatically:
  - `vl_` key → throws `"Invalid API key"`
  - expired JWT → clears localStorage and redirects to `/login`
  - `et_` embed token → throws `"Unauthenticated"`
- Tenant slug comes from `localStorage.getItem('tenantSlug')`.
- Tailwind only — no CSS files, no inline `style=`. Monochrome palette (`gray-900` primary, `gray-50`/`gray-100` backgrounds).
- Primary buttons: `rounded-lg bg-gray-900 text-white hover:bg-gray-800`.
- Do not add `console.log` statements.
- Type-check before reporting done: `npx tsc --noEmit`.

## Nango rules (non-negotiable)

- **Nango is never exposed to the browser or host network.** No `ports:` mapping in docker-compose. Ever.
- All Nango traffic is proxied through the control plane:
  - API calls → Go `nango.Client` (internal HTTP, server-side only)
  - Logo/asset images → `GET /v1/nango-assets/*` proxy endpoint
  - OAuth callbacks → control plane receives and forwards
  - Connect UI → Nango's Connect UI is embedded via session tokens, not direct browser access
- `NANGO_SECRET_KEY` and `NANGO_PUBLIC_KEY` are control-plane config only — not Nango container env vars.
- Snippet code never calls Nango directly; all integration calls go through `/v1/proxy/{provider}/*`.

## Security rules (non-negotiable)

1. **Tenant isolation** on every slug-based endpoint (see pattern above).
2. **No admin scope for embed tokens** — synthetic key scopes must stay `["invoke", "manage"]`.
3. **JWT issuer check** in `ValidateSession` must remain.
4. **Slug validation** on `CreateTenant`: must match `^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`.
5. **Role validation** on `InviteMember`: only `invoke`, `manage`, `admin` are valid roles.
6. **RequireScope** on every authenticated route — no naked authenticated routes.

## Libraries design

Platform libraries are grouped by **integration** (e.g. Salesforce, Google Sheets, Google Docs). Each integration can have multiple capability slugs (e.g. `salesforce-cases`, `salesforce-contacts`). The Libraries UI shows integrations as the top level and lists their capabilities beneath each one.

**Platform library code must export a class**, not standalone functions. This lets users instantiate with config and extend behaviour:

```typescript
// Bun — index.ts
export class SalesforceCases {
  constructor(private config?: { instanceUrl?: string; accessToken?: string }) {}
  async createCase(fields: CaseFields): Promise<CreateCaseResult> { … }
  async getCase(id: string): Promise<CaseRecord> { … }
  async updateCase(id: string, fields: Partial<CaseFields>): Promise<void> { … }
  async deleteCase(id: string): Promise<void> { … }
}
export default SalesforceCases
```

```python
# Python — module.py
class SalesforceCases:
    def __init__(self, instance_url=None, access_token=None): …
    def create_case(self, fields: dict) -> dict: …
    def get_case(self, case_id: str) -> dict: …
    def update_case(self, case_id: str, fields: dict) -> None: …
    def delete_case(self, case_id: str) -> None: …
```

The `meta.json` should include an `integration` field for grouping:
```json
{ "name": "Cases", "integration": "Salesforce", "description": "…" }
```

## Adding a new platform library

1. Create `platform-libraries/{bun|python}/<slug>/` with:
   - `index.ts` (Bun) or `module.py` (Python) — export a class as default (see Libraries design above)
   - `meta.json` — `{"name": "…", "integration": "…", "description": "…"}`
   - `README.md` — markdown docs shown in the admin UI
2. Run `make copy-platform-libs` to sync into `services/control-plane/internal/platformlibs/files/`.
3. The library is auto-loaded at startup — no DB changes needed.
4. Import path: `@velane/<slug>` (Bun) or `from velane import <slug_with_underscores>` (Python).

## Docker / local dev notes

- First run: `make down` first if you have an old `postgres_data` volume with stale credentials.
- Bootstrap env vars (`BOOTSTRAP_EMAIL`, `BOOTSTRAP_PASSWORD`, `BOOTSTRAP_TENANT`) create the first admin user on startup — remove them after first boot in production.
- `ENCRYPTION_KEY` and `JWT_PRIVATE_KEY` must be stable in production. Ephemeral (default) values mean secrets can't be decrypted and all JWTs invalidate on restart.
- MCP server runs on `:8090`, admin portal on `:8092`, control-plane on `:8080`.

## Production deployment (OpenTofu)

Infrastructure lives under `infra/terraform/`. Use **OpenTofu** (`tofu`), not HashiCorp Terraform. Config files keep the `terraform.tfvars` name (OpenTofu convention); state files may be `terraform.tfstate`.

```bash
# 1. EKS cluster (once) — infra/terraform/aws-eks
cd infra/terraform/aws-eks
cp terraform.tfvars.example terraform.tfvars   # edit domain, region, etc.
tofu init && tofu apply

# 2. Velane workloads — infra/terraform/k8s
cd infra/terraform/k8s
cp terraform.tfvars.example terraform.tfvars   # edit images, secrets, OAuth, domain
tofu init && tofu apply

# Useful outputs
tofu output oauth_redirect_uris
tofu -chdir=../aws-eks output -raw acm_certificate_arn
```

- Pin container images to CI tags (`ghcr.io/abskrj/velane-<service>:sha-<short-sha>` or `:latest` after a main-branch build).
- `PUBLIC_BASE_URL` and OAuth client env vars are wired via `terraform.tfvars` → control-plane secret.
- EKS kubeconfig must be valid (`aws eks update-kubeconfig` or set `kubeconfig_context` in tfvars).
- Full walkthrough: `infra/terraform/k8s/README.md`.

## License

Dual-licensed: AGPL-3.0 (open source) + commercial license for proprietary use.
Contact: abskrj@icloud.com
