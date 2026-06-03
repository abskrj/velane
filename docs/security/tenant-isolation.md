---
title: Tenant Isolation Model
description: How Velane enforces tenant boundaries across APIs and workflows.
sidebar_position: 2
---

# Tenant Isolation Model

Velane is multi-tenant by design. Isolation is a core platform behavior.

## What tenant isolation means

- tenant A cannot read or mutate tenant B resources
- tenant context is enforced on tenant-scoped APIs
- snippet invocation and management are bound to tenant ownership

## Where tenant context comes from

Tenant context is resolved from authenticated identity and request context (including tenant slug/header patterns used by the API surface).

## Safe usage guidance

- always pass the intended tenant context in direct API usage
- avoid broad admin credentials in shared automation
- test access boundaries in staging with separate tenants

## Common mistakes to avoid

- reusing keys across unrelated tenants
- assuming slug-only checks are enough in custom integrations
- skipping scope and tenant validation in new endpoints
