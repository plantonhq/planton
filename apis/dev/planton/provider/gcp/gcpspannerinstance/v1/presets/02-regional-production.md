# Regional Production

This preset provisions a production-ready Cloud Spanner instance with a single node, ENTERPRISE edition, and automatic backup scheduling. It is suitable for production workloads with predictable capacity needs deployed in a single GCP region.

## When to Use

- Production applications with predictable, steady traffic patterns
- Workloads that do not need multi-region replication
- Cost-conscious production deployments (single-region is significantly cheaper than multi-region)
- Applications that need strong consistency within a region

## Key Configuration

- **1 node** -- provides ~10,000 QPS reads, ~2,000 QPS writes, 10 TB storage. Scale up by increasing numNodes.
- **ENTERPRISE edition** -- granular instance sizing and advanced features
- **AUTOMATIC backup schedule** -- GCP automatically creates backup schedules for new databases in this instance
- **Regional config** -- all replicas in one region; 99.99% availability SLA

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the instance will be created | GCP Console or `GcpProject` outputs |
| `<instance-name>` | Name for this Spanner instance (6-30 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `prod-spanner`) |
| `<instance-config>` | Instance configuration (e.g., `regional-us-central1`, `regional-europe-west1`) | [Spanner configurations](https://cloud.google.com/spanner/docs/instance-configurations) |
| `<display-name>` | Human-readable display name (4-30 chars) | Choose a descriptive name (e.g., `Production Spanner`) |

## Related Presets

- **01-free-instance** -- Zero-cost instance for development/testing
- **03-autoscaling-production** -- Autoscaling for variable workloads
