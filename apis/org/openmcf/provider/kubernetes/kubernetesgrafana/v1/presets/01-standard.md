# Standard Grafana

This preset deploys Grafana with ingress for external access to dashboards. Grafana provides visualization and alerting for metrics from Prometheus, Elasticsearch, Loki, and many other data sources.

## When to Use

- You need a dashboarding and visualization platform for your monitoring stack
- You want the Grafana UI accessible via a hostname
- Pair with `KubernetesPrometheus` for a complete metrics pipeline

## Key Configuration Choices

- **Ingress enabled** -- exposes Grafana at the specified hostname
- **Monitoring namespace** (`monitoring`) -- co-located with Prometheus and other observability tools
- **Default resources** -- Grafana is lightweight; increase for dashboards with many concurrent users

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-grafana.example.com>` | Hostname for the Grafana web UI | Your DNS provider |
