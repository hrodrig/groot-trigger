# ROADMAP — groot-trigger

**Current focus:** Local scaffold complete through CI/deploy/docs. Next: **GSD** application code. Remote GitHub deferred.

| ID | Item | Status |
|----|------|--------|
| #1 | SPEC + packaging scaffold (Docker, Make, GoReleaser) | Done (local) |
| #2 | Make security targets (`govulncheck`, `gocyclo`, `grype`) | Done (local) |
| #3 | `.github/workflows/ci.yml` | Done (local; needs remote to run) |
| #4 | `deploy/k8s` flat manifests + README | Done (sketch; needs app) |
| #5 | CONTRIBUTING / SECURITY / `.cursor` rules | Done (local) |
| #6 | GSD: HTTP server, API key, rate limit, trusted proxies, vanilla UI | Pending |
| #7 | GSD: client-go Job create + single-flight 409 | Pending |
| #8 | Helm chart (optional thin wrap of flat manifests) | Pending |
| #9 | First tagged release (`v0.1.0`) + GHCR | Pending (needs remote) |
| #10 | Raise `COVER_MIN` to 80 after tests exist | Pending |

Contract: [docs/SPECIFICATIONS.md](docs/SPECIFICATIONS.md)
