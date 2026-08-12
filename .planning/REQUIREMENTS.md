# Requirements — groot-trigger MVP

Derived from `docs/SPECIFICATIONS.md`.

| ID | Requirement |
|----|-------------|
| R1 | `GET /healthz` / `GET /readyz` unauthenticated |
| R2 | `GET /v1/collect` vanilla HTML (API key field + Generate GROOT files) |
| R3 | `POST /v1/collect` requires API key (Bearer / X-API-Key / form `api_key`) |
| R4 | Empty API key at startup → process exit |
| R5 | POST rate limit per-IP (+ optional global) → 429 |
| R6 | Trusted proxies opt-in; default ignore X-Forwarded-* |
| R7 | Busy collect Job → 409 |
| R8 | Success → 202 + run_id (JSON or HTML) |
| R9 | slog JSON default; HTTP access level by status |
| R10 | Create batch/v1 Job running `GROOT_IMAGE` with configured mounts/env |
