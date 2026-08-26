# Contributing to groot-trigger

## Ground rules

- English only for all project artifacts.
- Security issues: [SECURITY.md](SECURITY.md) — no public issues for undisclosed vulns.
- **Branch policy:** `develop` is the integration branch; `main` is stable releases.
- **PR base:** open normal feature/fix PRs against **`develop`**.
- **No direct push** to `develop` or `main` (GitHub rulesets `protect-develop` / `protect-main`).

## How to contribute

- **Bug reports and feature requests:** Open an [issue](https://github.com/hrodrig/groot-trigger/issues).
- **Code changes:** From an updated `develop`, create a local branch (`feat/short-name`, `fix/description`, `chore/…`), push it, and open a pull request **into `develop`**.
- **Scope:** Keep PRs focused and small when possible.

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

## Release flow

1. Merge feature PRs into **`develop`**.
2. Open PR **`develop` → `main`** (`gh pr create --base main --head develop`). Merge on GitHub.
3. Annotated tag `vX.Y.Z` on **`main`** → `git push origin vX.Y.Z` → `.github/workflows/release.yml` runs GoReleaser (binaries + `ghcr.io/hrodrig/groot-trigger:vX.Y.Z`).

## Scope

- **In:** HTTP trigger, Job create, deploy manifests for the trigger.
- **Out:** GROOT CLI / collector logic → [groot](https://github.com/hrodrig/groot). Scheduled CronJob packaging → [groot-selfhosted](https://github.com/hrodrig/groot-selfhosted).
