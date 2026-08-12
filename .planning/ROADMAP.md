# GSD ROADMAP — groot-trigger

## Phase 1 — MVP HTTP → Job

**Goal:** Runnable trigger binary that serves the SPEC HTTP contract and creates collect Jobs.

**Requirements:** R1–R10

**Success:**
- `make test` / `make ci` pass
- Unit tests cover auth 401, rate 429, busy 409, HTML button, job create with fake client
- Local `GROOT_TRIGGER_API_KEY=x ./bin/groot-trigger` listens; GET shows form

## Phase 2 — Hardening / deploy polish

COVER_MIN 80 done; kind e2e (after remote optional). Flat manifests first; Helm chart only if operators ask.
