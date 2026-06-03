---
title: Deployment Topology
description: Overview of Velane service topology and runtime architecture.
sidebar_position: 2
---

# Deployment Topology

Velane is built around a control-plane with separate execution runtimes.

## Core components

- control-plane (API, auth, scheduling, tenancy)
- bun executor runtime
- python executor runtime
- postgres
- redis
- mcp server
- admin portal
- embed dashboard
- nango (internal)

## Recommended flow

1. Clients call control-plane APIs
2. Control-plane authorizes and resolves snippet/version
3. Scheduler dispatches execution to Bun/Python runtimes
4. Results return through control-plane

This keeps policy and orchestration centralized while execution stays isolated.

## Local defaults

- control-plane: `:8080`
- MCP: `:8090`
- admin: `:8092`
- Bun executor: mapped via compose
- Python executor: mapped via compose

## Operational guidance

- Keep provider auth and integration internals behind control-plane boundaries
- Avoid direct browser access to internal integration backends
- Treat executors as stateless compute and scale horizontally
