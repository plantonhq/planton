# Free Instance

This preset provisions a zero-cost Cloud Spanner instance using the FREE_INSTANCE type. It is ideal for development, prototyping, CI/CD testing, and learning Spanner without incurring any charges.

## When to Use

- Local development and integration testing
- CI/CD pipelines that need a temporary Spanner instance
- Proof-of-concept or prototyping new applications
- Learning Cloud Spanner features without cost

## Key Configuration

- **FREE_INSTANCE type** -- zero-cost instance with ~10 GB storage and limited throughput
- **No capacity fields** -- num_nodes, processing_units, and autoscaling_config must not be set
- **No edition** -- edition cannot be configured for free instances
- **No automatic backups** -- AUTOMATIC backup schedule is not available for free instances
- **One per billing account** -- GCP limits free instances to one per billing account

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the instance will be created | GCP Console or `GcpProject` outputs |
| `<instance-name>` | Name for this Spanner instance (6-30 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `dev-spanner`) |
| `<display-name>` | Human-readable display name (4-30 chars) | Choose a descriptive name (e.g., `Dev Spanner`) |

## Related Presets

- **02-regional-production** -- ENTERPRISE edition with fixed capacity for predictable workloads
- **03-autoscaling-production** -- ENTERPRISE edition with autoscaling for variable workloads
