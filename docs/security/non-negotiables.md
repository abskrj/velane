---
title: Security Non-Negotiables
description: Core security requirements and invariants in Velane.
sidebar_position: 1
---

# Security Non-Negotiables

This page summarizes the core security guarantees Velane is designed to keep.

## Tenant isolation first

Every tenant-scoped action must respect tenant boundaries. Cross-tenant access is denied.

## No admin access for embed tokens

Embed tokens (`et_...`) are for embed use cases and must not gain admin privilege.

## Session validation must stay strict

Session token verification checks issuer and signature. Do not weaken these checks.

## Scope checks are mandatory

Authenticated routes should enforce minimum required scopes:

- `invoke` for read/invoke actions
- `manage` for write operations
- `admin` for sensitive administration

## Integration boundary

Integration credentials stay server-side. Snippet code should call Velane integration pathways, not raw credential-bearing endpoints.

## Key management

- keep signing/encryption keys stable in production
- rotate operational keys safely
- avoid exposing secrets in logs or client apps
