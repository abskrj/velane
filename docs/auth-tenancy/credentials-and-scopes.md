---
title: Credentials and Scopes
description: Understand authentication credentials, scopes, and active org context.
sidebar_position: 1
---

# Credentials, Scopes, and Active Org Context

Velane supports three credential types. Use the right one for each surface.

## Credential types

- `Session JWT`
  - Used by the admin portal login flow
  - Best for human dashboard usage
- `API key` (`vl_...`)
  - Best for automation, scripts, CLI, and server-to-server calls
- `Embed token` (`et_...`)
  - For embedded experiences
  - Intended for embed flows, not full admin access

## Scope model

Velane permissions are scope-based:

- `invoke`: read/invoke level actions
- `manage`: create/update operational resources
- `admin`: destructive/team/admin actions

As a rule:

- reads generally require lower privilege
- writes require `manage`
- sensitive tenant administration requires `admin`

## Tenant context

Velane is multi-tenant. Many API requests are resolved in tenant context.

Tenant resolution is server-side:
- API key requests resolve tenant directly from the key.
- Session requests resolve tenant from your active org membership.

## Practical defaults

- Human admins: use session login in the portal
- CI/automation: use `vl_` API keys with minimum required scope
- Embeds: use `et_` tokens and keep TTL short

## Security basics

- Do not share API keys in client-side code
- Rotate keys periodically
- Keep embed tokens short-lived
- Prefer least privilege by default

## Related docs

- [Auth and Request Flow](./auth-and-request-flow.md)
- [Security Non-Negotiables](../security/non-negotiables.md)
