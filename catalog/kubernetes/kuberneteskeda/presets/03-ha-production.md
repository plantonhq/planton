# HA Production

This preset hardens KEDA for clusters where autoscaling is load-bearing:
two replicas of every component (warm standbys — KEDA leader-elects, so
extra replicas buy failover speed, not throughput), deliberate resource
sizing, `system-cluster-critical` priority so the engine outlives workload
evictions, cert-manager-issued internal TLS, and full Prometheus telemetry
with ServiceMonitors.

## When to Use

- Production clusters where workloads scale to zero and a stalled KEDA
  means requests queue against zero pods
- Clusters that already run KubernetesCertManager with a ClusterIssuer and
  the Prometheus operator
- Postures that require renewed, CA-chained internal TLS instead of
  operator-self-generated certificates

## Key Configuration Choices

- **`replicas: 2` on operator, metrics server, and webhooks** — per
  upstream HA guidance these are failover standbys: one instance leads or
  serves at a time, the second one cuts recovery time. The metrics server
  matters most: the HPA controller reads `external.metrics.k8s.io`
  through it on every reconcile loop.
- **Explicit resources** — requests/limits sized above the chart defaults
  so the engine is never the first eviction candidate under pressure.
- **`priorityClassName: system-cluster-critical`** — pods that scale on
  KEDA stop scaling without it; the engine belongs in the same priority
  tier as other cluster infrastructure.
- **`httpTimeoutMs: 5000`** — scalers reaching slow external HTTP metric
  sources get room before erroring (chart default: 3000).
- **`certificates.type: cert_manager` with a ClusterIssuer `valueFrom`** —
  cert-manager issues and renews the internal certificates; the issuer
  name flows from the `KubernetesClusterIssuer` resource's
  `status.outputs.cluster_issuer_name`, so the whole chain deploys in one
  Infra Chart in dependency order.
- **`prometheus.enabled` + `serviceMonitor: true`** — scaler loop
  latencies, trigger errors, and HPA interactions land in Prometheus
  automatically. The Prometheus operator CRDs MUST exist or the release
  fails to install.

## Prerequisites

- KubernetesCertManager on the cluster and a `KubernetesClusterIssuer`
  resource to reference
- Prometheus operator CRDs (kube-prometheus-stack)
- 2+ schedulable nodes for the replica pairs to spread

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `platform-ca` | Name of the `KubernetesClusterIssuer` resource whose issuer signs KEDA's internal certificates (inside `valueFrom.name`) | Your Infra Chart / `planton` resource listing |

## Related Presets

- **01-cluster-standard** — the zero-dependency baseline
- **02-eks-irsa-scalers** — combine with this preset's fields when AWS
  scalers need keyless identity
