# groot-trigger — specifications

**Status:** Approved (2026-08-12). Behavior contract for implementation (GSD next).  
**Repo:** local only until GitHub remote exists (`github.com/hrodrig/groot-trigger` planned).  
**Design history:** [docs/superpowers/specs/2026-08-12-groot-trigger-design.md](superpowers/specs/2026-08-12-groot-trigger-design.md)  
**Not in scope:** GROOT CLI behavior; Helm CronJob packaging (**groot** / **groot-selfhosted**).

---

## 1. Problem

Operators want a **“Generate GROOT files”** control (browser or HTTP client) that:

1. Runs an in-cluster `groot collect`
2. Uploads the archive to external object storage (e.g. Contabo S3)
3. Returns quickly without blocking the client for minutes

MVP ships the button **inside groot-trigger**: `GET /v1/collect` serves a minimal HTML page; `POST /v1/collect` starts the Job. No separate web app required.

The **groot** product is a one-shot CLI by design (no HTTP server, no long-lived collector daemon). Putting an API inside **groot** would break that philosophy.

## 2. Goals / non-goals

### Goals

- Idle **HTTP** Deployment that creates a Kubernetes **Job** running `ghcr.io/hrodrig/groot`
- Built-in UI: `GET /v1/collect` → **vanilla** HTML/CSS (embedded; no CSS framework) with API key field + button **“Generate GROOT files”** (English)
- Fire-and-forget: `POST /v1/collect` → `202 Accepted` + `run_id` (JSON or HTML result page)
- **API key auth** on `POST /v1/collect` (and on the HTML form); refuse to start if key unset
- **Rate limit** on POST (per client IP; optional global) → `429`
- **Trusted proxies** opt-in (CIDRs); default ignore forwarded headers
- Single-flight: `409 Conflict` if a collect Job is already Pending/Running
- Reuse operator config patterns from **groot-selfhosted** (ConfigMap `groot.yml`, Secret for `AWS_*`, image pin `vX.Y.Z`)
- English-only artifacts; companion to groot, not a fork of the collector

### Non-goals (MVP)

- OIDC / mTLS / per-user identity (Phase 2)
- Status poll / download proxy / presigned URL API
- Live watch / event-driven collect (upstream ROADMAP **#55** — separate)
- Embedding collector code or serving HTTP from the **groot** binary
- Multi-cluster
- Replacing Helm CronJob schedules (CronJob remains optional scheduled path)

## 3. Architecture

```
┌─────────────┐  GET  /v1/collect     ┌──────────────────────┐
│  Browser /  │ ─────────────────────► │  HTML: API key +     │
│  curl       │ ◄──── page             │  “Generate GROOT…” │
│             │                        │                      │
│             │  POST + API key        │  groot-trigger       │
│             │ ─────────────────────► │  Deployment (idle)   │
│             │ ◄──── 202 / 401 / 409  │                      │
└─────────────┘                        └──────────┬───────────┘
                                                  │ batch/v1 Job create
                                                  ▼
                                       ┌──────────────────────┐
                                       │  Job Pod             │
                                       │  image: groot:vX.Y.Z │
                                       │  collect → upload S3 │
                                       │  → exit              │
                                       └──────────────────────┘
```

**Model A (locked):** no idle `groot` pod. Only **groot-trigger** stays up. Each authenticated POST spawns a one-shot Job.

| Component | Lives in | Role |
|-----------|----------|------|
| CLI + image | **groot** | Unchanged one-shot collect / upload / notify |
| CronJob / Helm | **groot-selfhosted** | Optional schedule; Job template reference |
| Trigger | **groot-trigger** | GET page + auth + POST → Job; concurrency gate |

## 4. HTTP contract (MVP)

### Authentication (required)

Shared **API key** from env `GROOT_TRIGGER_API_KEY` (Kubernetes Secret → env). Process **exits on startup** if the key is empty (fail closed).

**How clients send the key (any one accepted):**

| Client | Mechanism |
|--------|-----------|
| Browser form | Field `api_key` (`input type=password`) on POST body (`application/x-www-form-urlencoded`) |
| curl / automation | Header `Authorization: Bearer <key>` **or** `X-API-Key: <key>` |

Rules:

- Constant-time compare (`crypto/subtle`)
- Missing / wrong key → **`401`** (JSON or HTML); no Job create
- **Never** accept the key via query string (leaks in access logs / Referer)
- `/healthz` and `/readyz` stay **unauthenticated** (probes)
- `GET /v1/collect` may stay unauthenticated (page only; no collect). Collect action is always POST + key

API key is a **shared secret**, not per-user identity. Still prefer ClusterIP / no public Ingress. Key ≠ network isolation.

### Rate limit (MVP)

In-process limiter (no Redis). Complements **409** single-flight (concurrency) with request throttling (auth brute-force / spam).

| Scope | Default (configurable) |
|-------|------------------------|
| Per client IP on `POST /v1/collect` | e.g. **10 req / minute** |
| Optional global cap on `POST` | e.g. **30 req / minute** |

- Exceeded → **`429 Too Many Requests`** (+ `Retry-After` when practical)
- `/healthz` / `/readyz` **not** rate-limited
- `GET /v1/collect` lightly limited or unlimited (static page); priority = protect **POST**
- Implementation: `golang.org/x/time/rate` or equivalent token bucket; memory keyed by client IP

### Client IP / trusted proxies

Default (**safe for ClusterIP / port-forward**): client IP = `RemoteAddr` only. **Ignore** `X-Forwarded-For` / `X-Real-IP`.

When behind Ingress / reverse proxy, set trusted proxy CIDRs. Only then peel forwarded headers from a peer in that set.

| Env | Purpose |
|-----|---------|
| `GROOT_TRIGGER_TRUSTED_PROXIES` | Comma-separated CIDRs (e.g. `10.0.0.0/8,192.168.0.0/16`). Empty = do not trust forwarded headers |
| `GROOT_TRIGGER_RATE_LIMIT_POST` | POST per-IP limit (e.g. `10/1m`); `0` disables |
| `GROOT_TRIGGER_RATE_LIMIT_GLOBAL` | Optional global POST cap; `0` = off |

Wrong trusted-proxy config → spoofed IPs → broken rate limits / misleading logs. Document: leave empty unless Ingress is intentional.

### `GET /v1/collect`

Serves a **minimal HTML** page (English UI strings):

- **Stack:** vanilla HTML + CSS only — **no** Tailwind, Bootstrap, JS framework, or CDN stylesheets. Embed templates/CSS in the Go binary (`embed`)
- Title / brand: GROOT trigger
- Password field: **API key**
- One primary control: button label **“Generate GROOT files”**
- Form: `method=POST`, `action=/v1/collect`, fields `api_key` (+ optional `message`)
- Visual: operator utility (CSS custom properties, sober palette, monospace for `run_id` / status). No marketing hero, cards, or stat strips
- No status poll, no download list

Optional later: if `Accept: application/json`, return a short JSON description of the endpoint (not required for MVP). Default for browsers = `text/html`.

### `POST /v1/collect`

Starts a collect Job **after** successful API key check. Clients: browser form, `fetch`, or `curl`.

**Request body:**

- Form: `api_key` (required for browser), optional `message`
- Optional JSON (`Content-Type: application/json`) when using headers for auth:

```json
{
  "message": "optional operator note for archive / notify"
}
```

(Do not put the API key in JSON if a header is used; form field `api_key` or header still required.)

**Responses (content negotiation):**

Prefer JSON when `Accept` includes `application/json` or request used JSON / `X-API-Key` / `Authorization`. Otherwise return a **simple HTML** result page with a link back to `GET /v1/collect`.

| Code | When | JSON body | HTML |
|------|------|-----------|------|
| `202` | Job created | `{"run_id":"<id>","job":"<job-name>"}` | “Collect started” + `run_id` + link back |
| `401` | Missing / invalid API key | `{"error":"unauthorized"}` | “Unauthorized” + link back |
| `409` | Collect Job with label `app.kubernetes.io/name=groot-trigger-collect` is Pending or Running | `{"error":"collect_in_progress","job":"<existing>"}` | “Collect already in progress” + link back |
| `429` | Rate limit exceeded | `{"error":"rate_limited"}` | “Too many requests” + link back |
| `400` | Malformed JSON | `{"error":"bad_request"}` | Short error + link back |
| `500` | API / RBAC / apiserver failure | `{"error":"internal","detail":"..."}` (no secrets) | Short error + link back |

**No** `GET /v1/collect/{id}` in MVP. Completion signal = notify channels and/or object appearing in the bucket.

### `GET /healthz`

Liveness: `200` if process up (no apiserver check required).

### `GET /readyz`

Readiness: `200` if in-cluster config / Job client can be constructed (lightweight).

## 5. Job shape

- **Name:** `groot-collect-<run_id_short>` (DNS-1123 safe)
- **Labels:**
  - `app.kubernetes.io/name=groot-trigger-collect`
  - `app.kubernetes.io/part-of=groot-trigger`
  - `groot-trigger/run_id=<run_id>`
- **Image:** configurable; default `ghcr.io/hrodrig/groot:v1.1.1` (GHCR publishes **`v`-prefixed** tags only)
- **Args:** `collect --config /config/groot.yml` (+ optional `--verbose` via values)
- **ServiceAccount:** dedicated Job SA with **read-only** ClusterRole (same shape as groot-selfhosted collector RBAC)
- **Volumes:** ConfigMap (groot.yml), PVC or emptyDir for `/out` (operator choice)
- **envFrom:** optional Secret (e.g. `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`) for `upload.s3`
- **TTL:** `ttlSecondsAfterFinished` set so completed Jobs are garbage-collected

### Single-flight

Trigger lists Jobs with label `app.kubernetes.io/name=groot-trigger-collect` in Active (not succeeded/failed). If any → **409**.

**Note (lab spike):** CronJob `concurrencyPolicy: Forbid` does **not** block `kubectl create job --from=…`. The **409 gate must live in groot-trigger**, not rely on Forbid alone.

## 6. Configuration

Trigger Deployment env / ConfigMap (examples):

| Key | Purpose |
|-----|---------|
| `GROOT_IMAGE` | Job container image |
| `GROOT_NAMESPACE` | Namespace for Jobs (default: pod namespace) |
| `GROOT_CONFIGMAP` | ConfigMap name mounting `groot.yml` |
| `GROOT_CONFIG_KEY` | Key inside CM (default `groot.yml`) |
| `GROOT_OUT_PVC` | Optional PVC claim name for `/out` |
| `GROOT_JOB_SA` | ServiceAccount name for Job pods |
| `GROOT_EXTRA_ARGS` | Extra CLI args (e.g. `--verbose`) |
| `GROOT_ENVFROM_SECRET` | Optional Secret name for Job `envFrom` |
| `GROOT_TRIGGER_API_KEY` | **Required.** Shared secret; empty → process exit |
| `GROOT_TRIGGER_TRUSTED_PROXIES` | Optional CIDR list; empty = ignore `X-Forwarded-*` |
| `GROOT_TRIGGER_RATE_LIMIT_POST` | Per-IP POST limit (default e.g. `10/1m`; `0` = off) |
| `GROOT_TRIGGER_RATE_LIMIT_GLOBAL` | Optional global POST limit (`0` = off) |
| `GROOT_TRIGGER_LOG_FORMAT` | `json` (default) or `text` |
| `GROOT_TRIGGER_LOG_LEVEL` | `info` (default), `debug`, `warn`, `error` |
| `LISTEN_ADDR` | Default `:8080` |

Upload/bucket settings stay in **groot.yml** (ConfigMap), not in trigger code. Credentials stay in Secrets.

## 7. RBAC

**Trigger SA** (API pod):

- `create`, `get`, `list`, `watch` on `batch/jobs` in the target namespace
- `get`, `list`, `watch` on `pods` (optional; for debugging / future status)
- **Not** cluster-wide collect rights

**Job SA** (collect pod):

- Same read-only collector ClusterRole as **groot-selfhosted** (pods/logs, events, nodes, workloads, metrics, …)

## 8. Error handling & observability

- Failed Job create → `500` + structured log (no AWS keys)
- Collect/upload failures inside Job → Job `Failed`; operator sees notify / Job events; API already returned `202`
- Metrics (stretch): `groot_trigger_collect_requests_total{result=accepted|conflict|unauthorized|rate_limited|error}`

### Logging

**Model: gghstats** (HTTP service), **not** groot CLI `logx`.

| Piece | gghstats | groot-trigger |
|-------|----------|---------------|
| Library | `log/slog` | same |
| Level env | `GGHSTATS_LOG_LEVEL` | `GROOT_TRIGGER_LOG_LEVEL` (`debug`/`info`/`warn`/`error`, default `info`) |
| Format | JSON/text via handler | `GROOT_TRIGGER_LOG_FORMAT` = `json` (default) or `text` |
| HTTP access | msg `"http"` + `method`, `path`, `status`, `ip`, `dur` | same (+ `run_id` / `result` on collect) |
| **Level by status** | `<400` Info · `4xx` Warn · `5xx` Error (`httpAccessLogLevel`) | **same** |
| Probes | skip `/healthz` in access log | skip `/healthz` + `/readyz` |
| Prefix | line prefix `gghstats ` | optional `groot-trigger ` (grep in shared streams) |
| Startup | one banner line (version, listen, **masked** secrets) | same (mask API key; never print full key) |
| Trusted IP | `TrustedProxies` + `clientIP` | same as rate-limit / access log |

**Never log:** API key, `Authorization`, form `api_key`, AWS keys, full Secret values.

Startup config summary: image, namespace, rate limits, trusted-proxy on/off — not the key.

## 9. Security (MVP)

- **API key required** for collect (`POST`); fail closed if unset
- **Rate limit** on `POST` (per-IP ± global); **429**
- **Trusted proxies** opt-in via CIDRs; default ignore forwarded headers
- Service ClusterIP only; no Ingress in default manifests (defense in depth with the key)
- Document: anyone with the key + network path can start a full-cluster read collect + upload — treat the key like a credential; rotate via Secret
- Phase 2: OIDC / mTLS / short-lived tokens
- Distroless / nonroot image for trigger binary
- No shell in Job image (official groot distroless)

## 10. Testing

| Layer | What |
|-------|------|
| Unit | run_id; busy check; auth 401; rate limit 429; trusted-proxy IP picking; GET HTML form; POST JSON/HTML |
| Integration | envtest or kind: POST without key → 401; with key → Job; burst → 429 |
| Lab | Browser / port-forward → form + key → POST; Job+S3 path already validated on a lab cluster |

## 11. Repo layout

```
groot-trigger/
  docs/SPECIFICATIONS.md          # this contract
  docs/superpowers/specs/…        # design history
  cmd/groot-trigger/
  internal/
  deploy/                         # Helm/manifests (GSD)
  Dockerfile / Dockerfile.release
  .goreleaser.yaml / Makefile
```

Stack: Go **1.26.x** (align with groot), client-go, stdlib `net/http` (keep deps small). Packaging mirrors **groot** (distroless, GoReleaser, `v`-prefixed image tags).

## 12. Relationship to siblings

| Repo | Boundary |
|------|----------|
| **groot** | Product CLI + `ghcr.io/hrodrig/groot` |
| **groot-selfhosted** | How to schedule/deploy collect (Helm CronJob, docker, examples) |
| **groot-trigger** | On-demand HTTP → Job |
| **groot-share** | Archive inbox / share UI (out of scope here) |

## 13. Lab spike notes (2026-08-12)

Validated on a lab cluster before this design:

1. Manual Job from CronJob template → collect OK (~15–25s lab)
2. Overlapping manual Jobs **both run** despite `Forbid`
3. Contabo S3 upload OK with Secret `AWS_*` + `upload.s3` endpoint **without** bucket path suffix
4. Upload success log line requires **`--verbose`** (`logger.OK` is verbose-gated)
5. Image tag must be **`v1.1.1`**, not `1.1.1`
6. Helm chart **groot-selfhosted 0.1.13** ships `extraEnvFrom` / `extraArgs` and v-tag normalization

## 14. Implementation notes

Application code is implemented via **GSD** against this SPEC. Packaging (Docker, Makefile, GoReleaser) may land before feature code.

---

**Decision log**

| Decision | Choice |
|----------|--------|
| Where | New repo **groot-trigger**, not inside groot |
| Runtime model | **A** — Job on demand; trigger idle |
| UI | **GET `/v1/collect`** = vanilla HTML (API key + **“Generate GROOT files”**); **POST** starts Job |
| Frontend | **Vanilla** embed (HTML/CSS); no Tailwind/Bootstrap/CDN |
| Auth MVP | **Shared API key** (`GROOT_TRIGGER_API_KEY`); Bearer / `X-API-Key` / form `api_key`; fail closed |
| Rate limit | In-process per-IP (+ optional global) on POST → **429** |
| Trusted proxies | Opt-in CIDRs; default ignore `X-Forwarded-*` |
| Logging | **gghstats-style `slog`**: level env + HTTP access level-by-status (4xx warn / 5xx error); not groot `logx` |
| Response | **202** fire-and-forget (+ `run_id`); HTML or JSON by Accept |
| Concurrency | **409** if collect Job active |
