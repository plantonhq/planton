# Development Cache

## Use Case

The cheapest Memorystore for Redis Cluster Google sells: one shared-core shard, no replicas, plaintext, in-memory only. For a development or test environment that needs a real Redis Cluster endpoint (Cluster-protocol clients, slot-aware drivers) without paying for a production shape.

## When to Use

- A development environment where the cache can be rebuilt at any time
- Integration tests against the OSS Cluster protocol
- A sandbox for trying `redisConfigs` or an ACL policy before production

## What This Creates

- A one-shard, zero-replica `REDIS_SHARED_CORE_NANO` cluster in `us-central1`
- Google-placed Private Service Connect endpoints in the `dev-vpc` network
- Deletion protection off and `deletionPolicy: DELETE`, so a destroy succeeds first time

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `pscConfigs[].network` | `dev-vpc` | Your VPC; a `GcpServiceConnectionPolicy` with `serviceClass: gcp-memorystore-redis` must exist on it in this region. |
| `shardCount` | `1` | More shards when a test needs a multi-slot topology. |
| `nodeType` | `REDIS_SHARED_CORE_NANO` | A dedicated-core type when the test needs realistic throughput. |
| `persistenceConfig` | none | `RDB` to keep test data across restarts. |

One nano node bills by the hour from creation, serving or idle; destroy the cluster when the environment is idle.
