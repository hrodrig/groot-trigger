# ROADMAP — groot-trigger

**Current focus:** Dual-install with groot-selfhosted (`always/` vs `job-sa/`); Helm chart for trigger (#8) only if needed.

| ID | Item | Status |
|----|------|--------|
| #1 | SPEC + packaging scaffold (Docker, Make, GoReleaser + cosign/SBOM) | Done |
| #2 | Make security targets (`govulncheck`, `gocyclo`, `grype`) | Done |
| #3 | `.github/workflows/ci.yml` + `release.yml` (tag → GHCR) | Done |
| #4 | `deploy/k8s` flat manifests + README | Done |
| #5 | CONTRIBUTING / SECURITY / `.cursor` rules | Done |
| #5b | BSD ports (FreeBSD/OpenBSD) + man + dist-* Make targets | Done |
| #6 | GSD: HTTP server, API key, rate limit, trusted proxies, vanilla UI | Done |
| #7 | GSD: client-go Job create + single-flight 409 | Done |
| #8 | Helm chart (thin wrap of flat manifests) | Deferred — flat `deploy/k8s` enough for now; revisit if operators ask or copy-paste hurts |
| #9 | First tagged release (`v0.1.0`) + GHCR | Done (v0.1.0) |
| #10 | Raise `COVER_MIN` to 80 after tests exist | Done |
| #11 | Collect Job `readOnlyRootFilesystem` + `/tmp` emptyDir | Done (v0.1.1) |
| #12 | Split `deploy/k8s` into `always/` + `job-sa/` (Helm coexistence) | Done (unreleased) |

Contract: [docs/SPECIFICATIONS.md](docs/SPECIFICATIONS.md)
