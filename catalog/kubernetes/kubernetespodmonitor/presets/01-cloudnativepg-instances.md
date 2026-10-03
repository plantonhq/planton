# CloudNativePG Instances

Scrapes every instance of one CloudNativePG cluster on the `metrics` port its instance pods declare (9187). Each series carries the cluster's name as `job`, and the instance's role (primary or replica) and name. Instances are scraped directly because no Service exposes their metrics port, and because a replica failing readiness is exactly the one worth watching.

## When to Use

- A PostgreSQL cluster runs on the CloudNativePG operator (`KubernetesPostgres` on `KubernetesCloudNativePgOperator`), and you want its replication, connection and transaction metrics in Prometheus.
- You need to tell the primary's series from the replicas' (`cnpg.io/instanceRole`).

## How It Works

The selector matches the instance pods by `cnpg.io/cluster`, which the operator stamps on every instance. `job_label` turns the same label into `job`, so every instance of one cluster is one job. `pod_target_labels` copies the role and instance name onto the series, so a query can isolate the primary (`cnpg_pg_replication_lag` on replicas, `cnpg_backends_total` by role). The metric relabeling drops one series per PostgreSQL setting, which nobody alerts on and which multiplies the series count by the number of settings.

## Key Configuration Choices

- **`port: metrics`.** The port name CloudNativePG declares on its instance containers; a name, not 9187, so a change of port keeps working.
- **`sample_limit: 100000`.** A cluster with many databases and the default queries stays well under it; a custom query that returns per-row series trips it loudly.
- **No credentials.** The instance's metrics port serves plain HTTP inside the cluster by default. When the cluster serves its metrics over TLS (KubernetesPostgres `monitoring.tls_enabled`), add `scheme: https` and a `tls_config` whose `ca` references the cluster's CA Secret.

## Prerequisites

- A CloudNativePG cluster named `<cluster>` in `<namespace>`, with its instance pods declaring the `metrics` port.
- A Prometheus that selects this object (`KubernetesKubePrometheusStack`; under its default discovery no label is needed).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<cluster>` | The CloudNativePG cluster's name (its pods' `cnpg.io/cluster` label), also used as the monitor's name. |
| `<namespace>` | Namespace of the cluster and the monitor. |
