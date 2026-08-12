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
- **Code:** stub only (`groot-trigger version`); full app via GSD
- **Remote:** local git only until GitHub repo exists

## Build (local)

```bash
make build
./bin/groot-trigger version
```

## Language

English only for all project artifacts.
