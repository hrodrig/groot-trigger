# AGENTS.md — groot-trigger

Companion HTTP service: create in-cluster **Jobs** running **groot** collect.

| Do | Do not |
|----|--------|
| Implement against `docs/SPECIFICATIONS.md` | Put collector logic / product SPEC here |
| Mirror groot packaging patterns | Put long-lived HTTP inside **groot** CLI |
| English-only artifacts | Invent remotes or push secrets |
| Local branch → PR into **`develop`** | Direct push to `develop` or `main` |

| Repo | Role |
|------|------|
| **groot** | CLI + image |
| **groot-selfhosted** | Scheduled / bastion deploy |
| **groot-trigger** | `GET`/`POST /v1/collect` → Job |

**Remote:** [github.com/hrodrig/groot-trigger](https://github.com/hrodrig/groot-trigger).

Work on a **local branch**, then PR into **`develop`**. Never push to `develop` or `main`. Release: PR `develop` → `main`, then annotated tag on `main`. After every merge into `main`, sync **`main` → `develop`** (merge commit) so the next release PR is not **"out-of-date with the base branch"** — see `.cursor/rules/git-flow.mdc`.
