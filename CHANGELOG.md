# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Behavior contract: `docs/SPECIFICATIONS.md` (approved 2026-08-12)
- Packaging scaffold aligned with groot: Dockerfile, Dockerfile.release, Makefile/GNUmakefile, `.goreleaser.yaml`, `.golangci.yml`, `.dockerignore`
- Stub `cmd/groot-trigger` (version only; HTTP/Job via GSD next)
- Make security targets: `govulncheck`, `gocyclo`, `grype`, `security`, `docker-scan`; `COVER_MIN` default 0 until tests
- GitHub Actions CI workflow (lint + test) — runs once remote exists
- `deploy/k8s` flat manifests (ClusterIP, dual SA, collector ClusterRole)
- CONTRIBUTING.md, SECURITY.md, `.cursor` rules (English, git-flow, triad, no-delete)

## [0.1.0] — TBD

Initial release after GSD implementation.
