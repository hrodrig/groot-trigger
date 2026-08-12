# Phase 1 PLAN — MVP HTTP → Job

## Tracer slice

End-to-end: POST with API key → fake JobStarter creates job → 202 JSON; without key → 401; busy → 409.

## Tasks

1. `internal/config` — env load, fail closed on empty API key, rate-limit parse
2. `internal/logging` — slog JSON/text + level
3. `internal/proxy` — trusted proxies + client IP
4. `internal/auth` — constant-time key check
5. `internal/ratelimit` — per-IP + global token buckets
6. `internal/jobs` — JobStarter interface + k8s impl + labels/single-flight
7. `internal/server` — mux, HTML embed, handlers, access log
8. `cmd/groot-trigger` — wire serve; version still works
9. Tests + `go mod tidy`; bump COVER_MIN if feasible

## Out of scope

OIDC, status poll, Helm chart polish, kind e2e.
