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
| `my-gcp-project-123` | Consumer project ID where the PSC endpoint is created (`pscAutoConnections[].projectId`) | GCP Console or `GcpProject` outputs |
| `my-dev-memorystore` (`instanceName`) | Name for this Memorystore instance (4-63 chars, lowercase, hyphens) | Choose a descriptive name |
| `us-central1` (`location`) | GCP region for the instance | [GCP regions](https://cloud.google.com/about/locations) |
| `dev-vpc` | Name of the GcpVpcNetwork resource the PSC connection attaches to (its `network_id` is read) | Your network manifest's `metadata.name` |

## Related Presets

- **02-ha-production** — CLUSTER mode with 3 shards, replicas, TLS, persistence, and deletion protection
- **03-enterprise-cluster** — CLUSTER mode with 5 shards, IAM auth, CMEK, AOF persistence, and automated backups
