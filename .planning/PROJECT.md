# Project: groot-trigger

## What

On-demand HTTP companion that creates Kubernetes Jobs running `groot collect`. Idle Deployment; ephemeral Jobs. Vanilla HTML button + API key.

## Why

Operators need a “Generate GROOT files” control without putting HTTP into the groot CLI (one-shot philosophy).

## Contract

Canonical: `docs/SPECIFICATIONS.md` (approved 2026-08-12).

## Constraints

- Local git first (no GitHub remote yet)
- English-only artifacts
- Logging: gghstats-style slog (not groot logx)
- Fail closed if `GROOT_TRIGGER_API_KEY` empty

## Current milestone

Phase 1 — MVP HTTP → Job (auth, rate limit, trusted proxies, single-flight 409, vanilla UI).
