# Redis Single Node

This preset creates a single-node Redis 7.1 cluster with encryption enabled. It is the fastest way to get a development or testing cache running with secure defaults.

## When to Use

- Local development and testing environments
- Low-traffic applications that don't need high availability
- Prototyping cache-backed features before scaling up
- CI/CD pipeline caches

## Key Configuration Choices

- **Single node** (`numCacheClusters: 1`) — no replicas, no automatic failover; keeps cost minimal for non-production use
- **cache.t3.micro** — smallest burstable instance; sufficient for development workloads
- **Encryption enabled** — both at-rest and in-transit encryption are on by default; establishes security baseline from day one
- **Redis 7.1** — latest stable major version with ACL support, multi-part AOF, and function libraries

## Placeholders to Replace

Rename `metadata.name` to match your use case (e.g., `dev-session-cache`, `ci-redis`).

## Common Additions

- Add `subnetIds` and `securityGroupIds` for VPC-based deployments
- Increase to `numCacheClusters: 3` with `automaticFailoverEnabled: true` for production readiness
- Add `parameterGroupFamily: redis7` with `parameters` to tune `maxmemory-policy`
- Set `snapshotRetentionLimit` to enable automatic backups

## Related Presets

- **02-redis-ha-cluster** — production non-clustered setup with failover, multi-AZ, and snapshots
- **03-redis-clustered-production** — horizontally sharded cluster for large-scale workloads
