---
title: Testing and Quality
description: Validation checklist for backend, frontend, and documentation changes.
sidebar_position: 3
---

# Testing and Quality

Use this checklist when contributing changes.

## Control-plane and backend services

- run Go tests for touched services
- verify API behavior with local calls for changed endpoints

## Frontend apps

- run TypeScript type checks for changed app(s)
- validate affected user flow in browser

## Docs quality

- keep docs concise and task-oriented
- prefer examples that match real command/API behavior
- update related pages when terminology changes

## Suggested pre-PR checklist

- tests pass in touched areas
- no debug-only logs left behind
- docs updated for user-visible changes
- security-sensitive changes reviewed with extra care
