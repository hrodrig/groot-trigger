# Contributing to groot-trigger

## Ground rules

- English only for all project artifacts.
- Security issues: [SECURITY.md](SECURITY.md) — no public issues for undisclosed vulns.
- Day-to-day work on **`develop`**. `main` is release-only (git flow).

## Planning triad

| Document | Role |
|----------|------|
| [docs/SPECIFICATIONS.md](docs/SPECIFICATIONS.md) | Behavior contract |
| [ROADMAP.md](ROADMAP.md) | Planned work |
| [CHANGELOG.md](CHANGELOG.md) | What shipped |

When shipping user-facing behavior: update SPEC if needed → mark ROADMAP item Done → CHANGELOG under `[Unreleased]`.

## Before a PR

```bash
make lint-fix
make ci
```

Maintainers before tag: `make release-check` (optional `STRICT_RELEASE=1` for image scan).

## Scope

- **In:** HTTP trigger, Job create, deploy manifests for the trigger.
- **Out:** GROOT CLI / collector logic → [groot](https://github.com/hrodrig/groot). Scheduled CronJob packaging → [groot-selfhosted](https://github.com/hrodrig/groot-selfhosted).

## Local-first

This repository may start without a GitHub remote. Prefer local commits on `develop` until `origin` exists; do not assume CI runs until the remote is connected.
