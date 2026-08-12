# Flat Kubernetes manifests — groot-trigger (MVP sketch)

Apply after the application implements `docs/SPECIFICATIONS.md` (GSD). Until then, image tags may not exist on GHCR.

```bash
# Create secrets first (do not commit real keys):
kubectl -n groot create secret generic groot-trigger-api \
  --from-literal=GROOT_TRIGGER_API_KEY='replace-me'

# Optional: S3 creds for collect Job (same pattern as groot-selfhosted lab):
# kubectl -n groot create secret generic groot-s3 \
#   --from-literal=AWS_ACCESS_KEY_ID=... \
#   --from-literal=AWS_SECRET_ACCESS_KEY=... \
#   --from-literal=AWS_REGION=EU

kubectl apply -f deploy/k8s/manifests.yaml
kubectl -n groot port-forward svc/groot-trigger 8080:8080
# open http://127.0.0.1:8080/v1/collect
```

ClusterIP only by default. Set `GROOT_TRIGGER_TRUSTED_PROXIES` only behind a real Ingress.

RBAC:

| SA | Rights |
|----|--------|
| `groot-trigger` | create/get/list/watch Jobs (+ optional pods) in namespace |
| `groot` (Job) | read-only collector ClusterRole (same shape as groot-selfhosted) |
