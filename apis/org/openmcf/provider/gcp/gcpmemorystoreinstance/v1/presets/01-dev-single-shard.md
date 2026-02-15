# Dev Single Shard

This preset provisions a minimal Memorystore instance in standalone (CLUSTER_DISABLED) mode with a single shard and the smallest available node type. It is ideal for development, testing, or prototyping where high availability and data durability are not required.

## When to Use

- Local development and integration testing
- CI/CD pipelines that need a temporary in-memory data store
- Proof-of-concept or prototyping environments
- Lightweight caching with minimal cost

## Key Configuration

- **CLUSTER_DISABLED mode** — standalone instance with a single primary endpoint; any Valkey/Redis client works without cluster-aware drivers
- **1 shard** — single partition; no data distribution across nodes
- **SHARED_CORE_NANO node type** — smallest available; shared-core instance suitable for low-throughput workloads
- **No persistence** — data is in-memory only; lost on restart
- **No encryption** — transit encryption and CMEK are not configured
- **PSC networking** — a single Private Service Connect endpoint connects the instance to the specified VPC

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the instance will be created | GCP Console or `GcpProject` outputs |
| `<instance-name>` | Name for this Memorystore instance (4-63 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `dev-cache`) |
| `<gcp-region>` | GCP region for the instance (e.g., `us-central1`) | [GCP regions](https://cloud.google.com/about/locations) |
| `<vpc-network-path>` | Full path of the VPC network (e.g., `projects/my-project/global/networks/dev-vpc`) | `GcpVpc` status outputs or GCP Console |

## Related Presets

- **02-ha-production** — CLUSTER mode with 3 shards, replicas, TLS, persistence, and deletion protection
- **03-enterprise-cluster** — CLUSTER mode with 5 shards, IAM auth, CMEK, AOF persistence, and automated backups
