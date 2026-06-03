---
title: Local Quickstart
description: Run Velane locally and verify the core services.
sidebar_position: 1
---

# Local Quickstart

This guide gets Velane running locally so you can sign in, create a snippet, and invoke it.

## Prerequisites

- Docker + Docker Compose
- Node.js (for frontend development only)
- Go (for control-plane or CLI development only)

## 1) Clone the repository

```bash
git clone https://github.com/abskrj/velane.git
cd velane
```

## 2) Set bootstrap admin values

In `docker-compose.yml`, set:

- `BOOTSTRAP_EMAIL`
- `BOOTSTRAP_PASSWORD`
- `BOOTSTRAP_TENANT`

These create your first admin user and tenant on first boot.

## 3) Start the stack

```bash
docker compose up --build
```

On startup, Velane runs database migrations automatically and then starts all services.

## 4) Open Velane

- Admin portal: `http://localhost:8092`
- API: `http://localhost:8080`
- MCP server: `http://localhost:8090`

Sign in with your bootstrap credentials.

## 5) Verify the API health endpoint

```bash
curl http://localhost:8080/healthz
```

Expected: a healthy response payload.

## What to do next

- Continue to [Your First Snippet](./first-snippet.md)
- If you want integrations immediately, go to [Integrations Overview](../integrations/overview.md)
