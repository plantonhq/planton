# Valkey Cache

This preset creates a single-node Valkey 8 cluster for caching workloads, with an LRU eviction policy and VPC-private access on the smallest managed-database size.

## When to Use

- Application-level caching (sessions, computed results, hot lookups)
- Workloads that tolerate key eviction under memory pressure
- Private-network caching for apps already running inside a VPC

## Key Configuration Choices

- **Valkey 8** (`engine: valkey`, `engineVersion: "8"`) -- the caching engine DigitalOcean offers for new clusters. Redis is no longer creatable: the `redis` engine slug is kept only so existing Redis clusters can be adopted, and a create with it fails at the API (measured 2026-09-16). Valkey is Redis-protocol compatible, so clients and eviction policies carry over unchanged.
- **LRU eviction** (`evictionPolicy: allkeys_lru`) -- evicts the least-recently-used keys when memory fills, the right default for pure caches. Eviction policies apply only to the caching engines.
- **Single node** (`nodeCount: 1`) -- caches usually tolerate a brief failover gap; add standbys only when cache warm-up is expensive.
- **VPC placement** (`vpc.valueFrom`) -- references a `DigitalOceanVpc` resource named `my-vpc`; rename it to your VPC resource, or replace the block with `value: <uuid>` for an unmanaged VPC.

## Related Presets

- **01-postgresql-ha** -- Use for durable relational data
- **02-postgresql-dev** -- Use for dev/test relational databases
