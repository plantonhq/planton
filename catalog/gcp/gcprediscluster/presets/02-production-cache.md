# Production Cache

## Use Case

A production Memorystore for Redis Cluster: sharded for capacity, replicated for failover, spread across zones, IAM-authenticated, TLS-encrypted, persisted with an append-only log, backed up daily, and guarded twice against an accidental destroy. The cache in front of a service's database, or the data structure server behind rate limiters and leaderboards.

## When to Use

- Any cache or session store a production service depends on
- A workload that needs read scaling (replicas serve reads) and zero-downtime resizing
- A regulated environment that needs IAM on the data plane and TLS in transit

## What This Creates

- A three-shard cluster with one replica per shard (six `REDIS_HIGHMEM_MEDIUM` nodes) across zones in `us-central1`
- Google-placed Private Service Connect endpoints in `prod-vpc`
- IAM authentication, TLS with Google's per-cluster CA, `allkeys-lru` eviction
- AOF persistence flushed every second, daily backups kept 35 days, a Sunday 03:00 UTC maintenance window
- Deletion protection on and `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `shardCount` / `replicaCount` | `3` / `1` | Size for the keyspace and read load; both resize in place. `replicaCount: 2` for a second failover target. |
| `nodeType` | `REDIS_HIGHMEM_MEDIUM` | `REDIS_HIGHMEM_XLARGE` or `_2XLARGE` for large keyspaces; `REDIS_STANDARD_SMALL` for a small production cache. |
| `kmsKey` | Google-managed | A `GcpKmsKey` reference for customer-managed encryption at rest. |
| `serverCaMode` | per-cluster Google CA | `SERVER_CA_MODE_GOOGLE_MANAGED_SHARED_CA` so a fleet trusts one CA; `SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA` with `serverCaPool` for your own CA. |
| `persistenceConfig` | AOF every second | `RDB` with a snapshot period when write throughput matters more than the last second of writes. |
| `automatedBackupConfig.retention` | 35 days | Your retention policy, between `"86400s"` and `"31536000s"`. |
| `aclPolicy` | built-in default ACL | A Redis Cluster ACL policy's full resource name for per-user command and key rules. |

`authorizationMode` and `transitEncryptionMode` cannot be changed after creation -- a different posture is a new cluster.
