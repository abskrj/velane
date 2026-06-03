---
title: Auth and Request Flow
description: Visual guide to authentication, tenant resolution, and authorization flow.
sidebar_position: 2
---

# Auth and Request Flow

This page shows how requests are authenticated and authorized before Velane executes work.

## High-level flow

```mermaid
flowchart TD
  A[Client request] --> B{Credential type}
  B -->|Session JWT| C[Session validation]
  B -->|API key vl_*| D[API key validation]
  B -->|Embed token et_*| E[Embed token validation]
  C --> F[Tenant context resolution]
  D --> F
  E --> F
  F --> G[Scope check invoke/manage/admin]
  G --> H[Handler logic]
  H --> I[Response]
```

## What this means for users

- Authentication answers: who is calling?
- Scope checks answer: what can they do?
- Tenant context answers: where can they do it?

All three are needed for safe multi-tenant behavior.

## Practical examples

- Dashboard user: session login, tenant selected in UI, role-based access
- CI automation: API key with least required scope
- Embedded usage: embed token with reduced privileges

## Related docs

- [Credentials and Scopes](./credentials-and-scopes.md)
- [Security Non-Negotiables](../security/non-negotiables.md)
- [Request Lifecycle](../invoke/request-lifecycle.md)
