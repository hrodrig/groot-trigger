# AGENTS.md — groot-trigger

Companion HTTP service: create in-cluster **Jobs** running **groot** collect.

| Do | Do not |
|----|--------|
| Implement against `docs/SPECIFICATIONS.md` | Put collector logic / product SPEC here |
| Mirror groot packaging patterns | Put long-lived HTTP inside **groot** CLI |
| English-only artifacts | Push to GitHub until remote exists (local first) |

| Repo | Role |
|------|------|
| **groot** | CLI + image |
| **groot-selfhosted** | Scheduled / bastion deploy |
| **groot-trigger** | `GET`/`POST /v1/collect` → Job |

Next: GSD implementation of SPEC (auth, rate limit, Job create, vanilla UI).
