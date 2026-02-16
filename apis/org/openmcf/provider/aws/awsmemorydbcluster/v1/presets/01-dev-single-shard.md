# Dev Single Shard

This preset creates a single-shard, single-node MemoryDB cluster with TLS encryption and the open-access ACL. It is the simplest way to get a durable Redis-compatible database running for development.

## When to Use

- Local development and testing environments
- Prototyping applications that need a durable Redis-compatible store
- CI/CD pipeline data stores
- Feature branch environments where HA is unnecessary

## Key Configuration Choices

- **Single shard, zero replicas** (`numShards: 1`, `numReplicasPerShard: 0`) — one node only, minimizing cost
- **db.t4g.small** — burstable Graviton instance, sufficient for development workloads
- **open-access ACL** — no authentication required, simplifying local development
- **TLS enabled** — in-transit encryption on by default, establishing security baseline
- **Redis 7.1** — latest stable version with ACL support and function libraries

## Placeholders to Replace

Rename `metadata.name` to match your use case (e.g., `dev-session-store`, `ci-memorydb`).

## Common Additions

- Add `subnetIds` and `securityGroupIds` for VPC-based deployments
- Switch to a custom ACL name for authentication
- Increase `numReplicasPerShard` to 1 or 2 for read scaling
- Add `parameterGroupFamily: memorydb_redis7` with `parameters` for engine tuning

## Related Presets

- **02-production-ha** — multi-shard production setup with replicas, snapshots, and custom ACL
- **03-high-throughput** — large-scale cluster with data tiering for cost-efficient large datasets
