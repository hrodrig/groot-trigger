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

Maintainers before tag: `make release-check` (includes `make cover` with **`COVER_MIN=80`**; optional `STRICT_RELEASE=1` for image scan).

Release: open PR `develop` → `main` (`gh pr create`), merge on GitHub, then annotated tag `vX.Y.Z` on `main` → push tag → `.github/workflows/release.yml` runs GoReleaser (binaries + `ghcr.io/hrodrig/groot-trigger:vX.Y.Z`).

## Scope

- **In:** HTTP trigger, Job create, deploy manifests for the trigger.
- **Out:** GROOT CLI / collector logic → [groot](https://github.com/hrodrig/groot). Scheduled CronJob packaging → [groot-selfhosted](https://github.com/hrodrig/groot-selfhosted).

## Branches

Day-to-day work on **`develop`**. Do not commit features on **`main`**. CI runs on push/PR to `main` and `develop`.
