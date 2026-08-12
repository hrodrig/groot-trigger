# AGENTS.md — groot-trigger

Companion HTTP service: create in-cluster **Jobs** running **groot** collect.

| Do | Do not |
|----|--------|
| Implement against `docs/SPECIFICATIONS.md` | Put collector logic / product SPEC here |
| Mirror groot packaging patterns | Put long-lived HTTP inside **groot** CLI |
| English-only artifacts | Invent remotes or push secrets |

| Repo | Role |
|------|------|
| **groot** | CLI + image |
| **groot-selfhosted** | Scheduled / bastion deploy |
| **groot-trigger** | `GET`/`POST /v1/collect` → Job |

**Remote:** [github.com/hrodrig/groot-trigger](https://github.com/hrodrig/groot-trigger). Release via PR `develop` → `main`, then annotated tag on `main`.
