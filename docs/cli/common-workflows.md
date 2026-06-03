---
title: Common Workflows
description: Daily CLI workflows for snippets, versions, and invocation.
sidebar_position: 2
---

# CLI Common Workflows

This page covers the most common day-to-day CLI flows.

## 1) List snippets

```bash
./velane --api-url http://localhost:8080 --tenant myorg snippets list
```

## 2) Push code and publish to dev

```bash
./velane --api-url http://localhost:8080 --tenant myorg snippets push handler.ts --publish dev
```

If the snippet does not exist yet, the CLI creates it and then pushes a new version.

## 3) List versions

```bash
./velane --api-url http://localhost:8080 --tenant myorg versions list <snippet-id>
```

## 4) Publish a specific version

```bash
./velane --api-url http://localhost:8080 --tenant myorg versions publish <snippet-id> <version-number> staging
```

## 5) Invoke a snippet

```bash
./velane --api-url http://localhost:8080 --tenant myorg invoke <snippet-slug-or-id> --input '{"name":"world"}'
```

## 6) Stream invocation output

```bash
./velane --api-url http://localhost:8080 --tenant myorg invoke <snippet-slug-or-id> --stream
```

## Suggested team flow

1. Push to `dev`
2. Validate behavior with invoke
3. Publish to `staging`
4. Promote to `prod` after checks
