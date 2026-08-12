# Flat Kubernetes manifests — groot-trigger (MVP)

Apply after the application implements `docs/SPECIFICATIONS.md` (GSD). Until then, image tags may not exist on GHCR.

```bash
# Create secrets first (do not commit real keys):
kubectl -n groot create secret generic groot-trigger-api \
  --from-literal=GROOT_TRIGGER_API_KEY='replace-me'

# Optional: image pull secret if GHCR (or another registry) is private:
# kubectl -n groot create secret docker-registry YOUR_PULL_SECRET \
#   --docker-server=YOUR_REGISTRY_HOST \
#   --docker-username=YOUR_USERNAME \
#   --docker-password=YOUR_PASSWORD

# Optional: object-storage creds for the collect Job (enable upload in groot-config
# and set GROOT_ENVFROM_SECRET on the Deployment):
# kubectl -n groot create secret generic groot-s3 \
#   --from-literal=AWS_ACCESS_KEY_ID=... \
#   --from-literal=AWS_SECRET_ACCESS_KEY=... \
#   --from-literal=AWS_REGION=...

kubectl apply -f deploy/k8s/manifests.yaml
kubectl -n groot port-forward svc/groot-trigger 8080:8080
# open http://127.0.0.1:8080/v1/collect
```

ClusterIP only by default. Set `GROOT_TRIGGER_TRUSTED_PROXIES` only behind a real Ingress.

The trigger container uses distroless `nonroot`; the Deployment sets numeric `runAsUser` / `runAsGroup` `65532` so `runAsNonRoot` succeeds.

RBAC:

| SA | Rights |
|----|--------|
| `groot-trigger` | create/get/list/watch Jobs (+ optional pods) in namespace |
| `groot` (Job) | read-only collector ClusterRole (same shape as groot-selfhosted) |
