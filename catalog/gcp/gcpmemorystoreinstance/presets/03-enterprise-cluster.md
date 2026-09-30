# Enterprise Cluster

This preset provisions a fully-featured Memorystore instance in CLUSTER mode with 5 shards, 2 replicas per shard, IAM authentication, TLS encryption, customer-managed encryption keys (CMEK), AOF persistence, multi-zone distribution, automated backups with 35-day retention, and deletion protection. It is designed for enterprise and mission-critical workloads that demand maximum durability, security, and compliance.

## When to Use

- Mission-critical applications requiring the highest availability and durability
- Workloads subject to compliance or governance requirements (CMEK, IAM auth)
- Large-scale caching or real-time analytics with high throughput demands
- Environments requiring automated daily backups with extended retention
- Organizations that mandate IAM-based authentication over shared secrets

## Key Configuration

- **CLUSTER mode** — sharded topology with native cluster protocol; clients must use cluster-aware drivers
- **5 shards, 2 replicas** — data distributed across 5 shards, each with 2 read replicas for maximum read throughput and resilience
- **HIGHMEM_XLARGE node type** — largest high-memory nodes for demanding production workloads
- **authorizationMode: IAM_AUTH** — clients authenticate using GCP IAM credentials instead of shared passwords
- **transitEncryptionMode: SERVER_AUTHENTICATION** — TLS for client-to-server connections
- **kmsKey** — customer-managed encryption key (CMEK) for encryption at rest; use full KMS key resource name or `valueFrom` to reference a GcpKmsKey
- **engineConfigs** — Valkey engine tuning; `maxmemory-policy: volatile-lru` evicts keys with an expiry set, using least-recently-used ordering
- **persistenceConfig: AOF (EVERY_SEC)** — append-only file flushed every second; stronger durability than RDB with minimal performance impact
- **zoneDistributionConfig: MULTI_ZONE** — nodes spread across multiple availability zones
- **automatedBackupConfig** — daily backups at 2:00 UTC with 35-day retention (3,024,000 seconds)
- **maintenancePolicy** — Sunday 3:00 UTC maintenance window
- **deletionProtectionEnabled** — prevents accidental destruction of the instance

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | Consumer project ID where the PSC endpoint is created (`pscAutoConnections[].projectId`) | GCP Console or `GcpProject` outputs |
| `my-enterprise-memorystore` (`instanceName`) | Name for this Memorystore instance (4-63 chars, lowercase, hyphens) | Choose a descriptive name |
| `us-central1` (`location`) | GCP region for the instance | [GCP regions](https://cloud.google.com/about/locations) |
| `prod-vpc` | Name of the GcpVpcNetwork resource the PSC connection attaches to (its `network_id` is read) | Your network manifest's `metadata.name` |
| `memorystore-cmek` | Name of the GcpKmsKey resource that encrypts the instance (its `key_id` is read) | Your KMS key manifest's `metadata.name` |

## Related Presets

- **01-dev-single-shard** — Minimal standalone instance for dev/test
- **02-ha-production** — CLUSTER mode with 3 shards; simpler production setup without IAM auth or CMEK
