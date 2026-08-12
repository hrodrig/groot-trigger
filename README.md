# groot-trigger

On-demand HTTP companion that creates a Kubernetes **Job** running [`groot`](https://github.com/hrodrig/groot) (`collect` → optional upload).

| Endpoint | Behavior |
|----------|----------|
| `GET /v1/collect` | Vanilla HTML: API key + **Generate GROOT files** |
| `POST /v1/collect` | Start Job (`202` / `401` / `409` / `429`); API key required |

**Not** the GROOT CLI. Product stays one-shot; this companion stays idle until called.

| Repo | Role |
|------|------|
| **[groot](https://github.com/hrodrig/groot)** | CLI, SPEC, `ghcr.io/hrodrig/groot` |
| **[groot-selfhosted](https://github.com/hrodrig/groot-selfhosted)** | CronJob / Helm / bastion |
| **groot-trigger** (this repo) | HTTP → Job |

## Status

- **Contract:** [docs/SPECIFICATIONS.md](docs/SPECIFICATIONS.md) (approved)
- **Code:** MVP HTTP → Job (`docs/SPECIFICATIONS.md`); GSD Phase 1
- **Scaffold:** Docker / Make / GoReleaser / CI / BSD / `deploy/k8s`
- **Remote:** local git only until GitHub repo exists

## Build / run (local)

```bash
make build
export GROOT_TRIGGER_API_KEY='replace-me'
./bin/groot-trigger                 # listens on :8080
# open http://127.0.0.1:8080/v1/collect
make ci
```

Deploy sketch: [deploy/k8s/README.md](deploy/k8s/README.md)

BSD ports: [contrib/README.md](contrib/README.md) (`make dist-freebsd`, `make dist-openbsd`)
## Language

English only for all project artifacts.
