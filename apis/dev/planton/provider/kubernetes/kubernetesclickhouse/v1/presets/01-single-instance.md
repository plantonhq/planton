# Single Instance ClickHouse

This preset deploys a single-replica ClickHouse column-oriented database with persistence. Suitable for analytics workloads, log storage, and time-series data in development or small production environments.

## When to Use

- Analytics and OLAP workloads on moderate data volumes
- Log aggregation and time-series storage
- Development or staging environments for ClickHouse-based applications

## Key Configuration Choices

- **Single replica** -- standalone ClickHouse without sharding or replication
- **Higher resources** (`2000m` CPU, `4Gi` memory limits) -- ClickHouse is CPU and memory intensive for analytical queries
- **20Gi disk** -- persistent storage; ClickHouse compresses data efficiently, so 20Gi stores significantly more than 20Gi of raw data

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
