---
title: CLI Troubleshooting
description: Resolve common CLI authentication, connectivity, and permission issues.
sidebar_position: 3
---

# CLI Troubleshooting

## `401` or authentication errors

Check:

- API key is valid (`vl_...`)
- key has required scope
- tenant slug is correct
- you passed `--api-url` for the target environment

## `404` on snippet operations

Check:

- snippet ID/slug exists in the target tenant
- you are using the same tenant where snippet was created

## Connection refused

Check:

- control-plane is running at the URL you passed
- local default is `http://localhost:8080`

Quick test:

```bash
curl http://localhost:8080/healthz
```

## Permission denied for write actions

Your key likely lacks `manage` or `admin` scope for that action.

## Stream mode confusion

Use stream mode when you expect progressive output. For standard request-response handlers, regular invoke mode is simpler.
