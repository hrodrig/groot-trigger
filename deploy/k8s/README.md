# Flat Kubernetes manifests — groot-trigger

Pin `ghcr.io/hrodrig/groot-trigger:v0.1.2` (GoReleaser; **`v`-prefixed** tags). Behavior: [docs/SPECIFICATIONS.md](../../docs/SPECIFICATIONS.md).

Namespace `groot` is assumed (`kubectl create namespace groot` if missing). Do not commit real API keys.

## Which directory to apply

| Directory | What | When |
|-----------|------|------|
| [`always/`](always/) | Trigger SA, Role/RoleBinding, sample ConfigMap, Deployment, Service | **Every** install |
| [`job-sa/`](job-sa/) | Job SA `groot` + collector ClusterRole/Binding | Standalone only |

**Standalone** (no groot-selfhosted Helm in this namespace):

```bash
kubectl apply -f deploy/k8s/always -f deploy/k8s/job-sa
```

**Beside groot-selfhosted Helm** (release already created the collect Job SA — default name `groot` when the release is `groot`):

```bash
kubectl apply -f deploy/k8s/always
```

Do **not** apply `job-sa/` on top of Helm: that ServiceAccount is Helm-owned (`kubectl apply` will warn about a missing `last-applied-configuration` annotation and fight the next `helm upgrade`). Helm already binds collector RBAC to that SA. Keep `GROOT_JOB_SA` equal to the Helm ServiceAccount name (default `groot`).

`imagePullSecrets` for the collect image belong on the Helm Job SA / CronJob values, not in `job-sa/`. The trigger Deployment may set a pull secret for the **trigger** image (commented in [`always/deployment.yaml`](always/deployment.yaml)).

Do not `kubectl apply -f deploy/k8s/` (parent): that would apply both directories.

## Secrets, then apply

```bash
# Create secrets first (do not commit real keys):
kubectl -n groot create secret generic groot-trigger-api \
  --from-literal=GROOT_TRIGGER_API_KEY='replace-me'

# Optional: image pull secret if GHCR (or another registry) is private:
# kubectl -n groot create secret docker-registry YOUR_PULL_SECRET \
#   --docker-server=YOUR_REGISTRY_HOST \
#   --docker-username=YOUR_USERNAME \
#   --docker-password=YOUR_PASSWORD

# Optional: upload creds for the collect Job (enable upload in groot-config
# — S3 / GCS / SFTP per groot — and set GROOT_ENVFROM_SECRET on the Deployment):
# kubectl -n groot create secret generic YOUR_UPLOAD_SECRET \
#   --from-literal=AWS_ACCESS_KEY_ID=... \
#   --from-literal=AWS_SECRET_ACCESS_KEY=... \
#   --from-literal=AWS_REGION=...
#   # SFTP: GROOT_UPLOAD_SFTP_IDENTITY_FILE (and related GROOT_UPLOAD_SFTP_*)

kubectl apply -f deploy/k8s/always -f deploy/k8s/job-sa   # or always/ only; see table above
kubectl -n groot port-forward svc/groot-trigger 8080:8080
# open http://127.0.0.1:8080/v1/collect
```

ClusterIP only by default. Set `GROOT_TRIGGER_TRUSTED_PROXIES` only behind a real Ingress.

The trigger container uses distroless `nonroot`; the Deployment sets numeric `runAsUser` / `runAsGroup` `65532` so `runAsNonRoot` succeeds.

RBAC:

| SA | Rights | Created by |
|----|--------|------------|
| `groot-trigger` | create/get/list/watch Jobs (+ optional pods) in namespace | `always/` |
| `groot` (Job) | read-only collector ClusterRole (same shape as groot-selfhosted) | `job-sa/` **or** Helm |
