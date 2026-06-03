---
title: Development Workflow
description: Recommended local workflow for contributing changes safely.
sidebar_position: 2
---

# Development Workflow

This is the recommended local development flow for contributors.

## 1) Bring up the stack

```bash
docker compose up --build
```

## 2) Make focused changes

- keep changes scoped to one feature area
- update docs when behavior changes
- avoid mixing unrelated refactors in one branch

## 3) Validate locally

- run relevant service tests
- run frontend type checks when touching UI
- verify end-to-end behavior for affected flows

## 4) Prefer feature promotion flow

For snippet lifecycle work, validate behavior in `dev`, then `staging`, then `prod`.

## 5) Keep docs and behavior aligned

If API behavior, auth requirements, or UX changes, update docs in the same change set.
