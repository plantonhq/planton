# HA with Verified TLS (production hardening)

This preset hardens both availability and trust: two replicas guarded by a
PodDisruptionBudget keep the metrics API serving through node drains, and
cert-manager issues (and renews) the serving certificate so the API server
verifies metrics-server instead of skipping TLS verification.

## When to Use

- Production clusters where autoscaling must survive node drains and
  rolling node upgrades
- Security postures that disallow `insecureSkipTLSVerify` on APIServices
- Clusters that already run KubernetesCertManager with a ClusterIssuer

## Key Configuration Choices

- **`replicas: 2` + `podDisruptionBudget: true`** — a PodDisruptionBudget
  with minAvailable 1; the APIService fails over between healthy replicas.
  (Never enable the PDB with a single replica — it would block every
  voluntary eviction and wedge node drains.)
- **`tls.type: cert_manager` with a ClusterIssuer reference** —
  cert-manager issues and renews the serving certificate, and its CA
  injector maintains the APIService `caBundle` automatically
- **`apiService.insecureSkipTlsVerify: false`** — the API server verifies
  the serving certificate against the injected CA; only safe because the
  cert-manager arm keeps the `caBundle` wired
- **Composition via `valueFrom`** — the issuer name flows from the
  KubernetesClusterIssuer resource's `status.outputs.cluster_issuer_name`,
  so the whole chain deploys in one Infra Chart

## Prerequisites

- KubernetesCertManager installed on the cluster (the cert-manager CRDs and
  CA injector must exist, or the release fails)
- A KubernetesClusterIssuer to reference

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-cluster-issuer>` | Name of the KubernetesClusterIssuer resource whose issuer signs the serving certificate | Your Infra Chart / `planton` resource listing |

## Related Presets

- **01-managed-cloud** — the chart-default EKS / AKS posture
- **02-self-signed-kubelets** — kind / k3s / kubeadm / on-prem clusters
  (combine: `kubeletInsecureTls` is scrape-side and independent of the
  serving-side hardening here)
