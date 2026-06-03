---
title: Installation and Auth
description: Build the Velane CLI and authenticate with an API key.
sidebar_position: 1
---

# CLI Installation and Auth

The Velane CLI is useful for quick snippet workflows from your terminal.

## Build the CLI locally

```bash
cd services/cli
go build -o velane .
```

You can now run `./velane` from that directory.

## Authenticate with an API key

Use a `vl_` key from your tenant.

```bash
./velane login --key vl_xxxx
```

## Point CLI to your local stack

Most local workflows should include:

- `--api-url http://localhost:8080`
- `--tenant <your-tenant-slug>`

Example:

```bash
./velane --api-url http://localhost:8080 --tenant myorg snippets list
```

## Best practice

- keep keys out of shell history when possible
- use least-privilege keys for automation
- validate changes in `dev` before publish to higher environments
