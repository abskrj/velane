---
title: Embed Mode vs Embed Dashboard
description: Compare Velane's two embed experiences and when to use each.
sidebar_position: 2
---

# Embed Mode vs Embed Dashboard

Velane supports two embed experiences that look similar but serve different goals.

## Quick comparison

| Experience | Primary use | Typical audience | Token |
|---|---|---|---|
| Admin embed mode | Restricted admin-style experience | Internal power users | `et_...` |
| Embed dashboard | Read-only iframe viewer | External stakeholders/customers | `et_...` |

## Admin embed mode

- Entry is through the admin app flow
- Useful when you still want parts of the management-oriented interface
- Some sensitive sections remain hidden

## Embed dashboard

- Standalone read-only dashboard for embedding
- Best for product/customer-facing visibility views
- Focuses on viewing snippet state, not platform administration

## Which one should you choose?

- choose admin embed mode for internal operational viewers
- choose embed dashboard for customer-facing or low-risk viewing access

## Security guidance

- use short TTL embed tokens
- rotate and revoke tokens when access changes
- prefer least privilege and minimal exposure

## Related docs

- [Integrations Overview](./overview.md)
- [Credentials and Scopes](../auth-tenancy/credentials-and-scopes.md)
- [Security Non-Negotiables](../security/non-negotiables.md)
