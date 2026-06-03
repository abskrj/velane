---
title: MCP Overview
description: Use Velane's first-class MCP server for agent-driven workflows.
sidebar_position: 1
---

# MCP in Velane

Velane includes a first-class MCP server so coding agents can create, update, publish, and invoke snippets directly from your IDE.

## Why MCP is first-class in Velane

MCP is not an addon workflow in Velane. It is a primary interface for agent-driven development:

- create snippet drafts
- update and publish versions
- invoke snippets
- read logs and metrics
- discover connected integrations

## MCP endpoint

Local default:

- `http://localhost:8090/mcp`

## Cursor setup example

Add this to your MCP config:

```json
{
  "mcpServers": {
    "velane": {
      "url": "http://localhost:8090/mcp",
      "headers": {
        "Authorization": "Bearer vl_xxxx"
      }
    }
  }
}
```

Use an API key with the minimum scope needed for your workflow.

## Typical agent workflow

1. `list_connections`
2. `get_integration_docs` for a provider
3. `create_snippet`
4. `update_draft`
5. `invoke_snippet`
6. `publish_snippet`
7. `get_logs` / `get_metrics`

## Core tools you will use often

- snippets: create, update draft, publish, list, get
- invoke: sync/async/stream invocation
- integrations: list tenant connections and provider docs
- operations: logs, metrics, secrets

## Practical guidance

- Keep API keys scoped to the least privilege needed
- Validate in `dev` first, then promote to `staging` and `prod`
- Use provider docs tooling before generating integration-heavy snippet code

## Related docs

- [Integrations Overview](../integrations/overview.md)
- [Invocation Modes](../invoke/invocation-modes.md)
- [Generated References Plan](../reference-plan.md)
