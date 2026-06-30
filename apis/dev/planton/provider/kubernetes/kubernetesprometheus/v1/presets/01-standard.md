# Standard Prometheus

This preset deploys Prometheus with persistence for metrics retention. Prometheus is the de facto standard for Kubernetes metrics collection and alerting.

## When to Use

- Cluster-wide metrics collection from Kubernetes workloads and infrastructure
- Alerting based on metric thresholds (via Alertmanager)
- Pair with `KubernetesGrafana` for visualization

## Key Configuration Choices

- **Persistence enabled** with 20Gi disk -- metrics are retained on disk across restarts; adjust based on retention period and scrape volume
- **Single replica** -- sufficient for most clusters; add a second replica for HA
- **Higher memory** (`4Gi` limit) -- Prometheus stores recent samples in memory; more memory = longer in-memory retention
- **Monitoring namespace** (`monitoring`) -- co-located with Grafana and other observability tools

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.
