---
title: Known Limitations
description: Current platform and workflow limitations to plan around.
sidebar_position: 4
---

# Known Limitations

This page lists current limitations so teams can plan around them.

## Current product limitations

- CLI `logs` command is currently a placeholder flow
- CLI stream mode is narrower than full invoke flexibility
- Some observability paths are still evolving across environments
- Firecracker mode requires host-level KVM support and extra setup
- Documentation references are being migrated to auto-generated sources

## Operational implications

- verify critical production workflows using API and admin paths, not only CLI convenience commands
- validate async and stream behavior in staging with realistic payloads
- keep a fallback runbook for invoke and release operations

## How to use this page

- review before production rollout
- review after each release for changes
- track resolved items and remove outdated entries

## Related docs

- [Production Checklist](./production-checklist.md)
- [Deployment Topology](./deployment-topology.md)
- [MCP Overview](../mcp/overview.md)
