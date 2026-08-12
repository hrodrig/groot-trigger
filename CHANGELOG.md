# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- `deploy/k8s`: anonymize optional image-pull-secret sample (`YOUR_PULL_SECRET` / `YOUR_REGISTRY_HOST`; no lab-specific secret name)

## [0.1.0] — 2026-08-12

### Added

- Behavior contract: `docs/SPECIFICATIONS.md` (approved 2026-08-12)
- Packaging scaffold aligned with groot: Dockerfile, Dockerfile.release, Makefile/GNUmakefile, `.goreleaser.yaml`, `.golangci.yml`, `.dockerignore`
- Make security targets: `govulncheck`, `gocyclo`, `grype`, `security`, `docker-scan`; **`COVER_MIN=80`** gate in `make cover` / `release-check` (#10)
- GitHub Actions CI workflow (lint + test)
- GitHub Actions Release workflow: tag `v*` → GoReleaser (binaries, GHCR, cosign, SBOM) (#3, #9)
- `deploy/k8s` flat manifests (ClusterIP, dual SA, collector ClusterRole) (#4)
- CONTRIBUTING.md, SECURITY.md, `.cursor` rules (English, git-flow, triad, no-delete) (#5)
- BSD packaging: FreeBSD/OpenBSD port skeletons, `contrib/man/man1/groot-trigger.1`, `make dist-freebsd` / `dist-openbsd` / `port-*-sync` / `man-sync`; GoReleaser builds freebsd+openbsd (#5b)
- MVP application (SPEC): `GET/POST /v1/collect`, API key auth, rate limits, trusted proxies, slog access logs, client-go Job create + single-flight 409, vanilla HTML UI (#6, #7)
- Family-style README (K8s-first) + hero asset

### Changed

- `deploy/k8s`: distroless numeric `runAsUser`/`runAsGroup` `65532`; optional commented `imagePullSecrets` / `GROOT_ENVFROM_SECRET` / upload skeleton; GHCR image pin (#4)
- CI / Make: pin **golangci-lint v2.12.2** (Go 1.26.5 support; v2.5.0 binaries fail on CI)
- Make `grype`: drop in-recipe `check-docker` so the container fallback works on GitHub Actions (align with groot)

### Notes

- Helm chart for trigger (#8) deferred: ship flat `deploy/k8s` first; add a thin chart later only if operators ask or overlays become painful. Scheduled collect packaging stays in **groot-selfhosted**.
