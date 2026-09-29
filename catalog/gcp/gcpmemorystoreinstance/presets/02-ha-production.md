# HA Production

This preset provisions a production-ready Memorystore instance in CLUSTER mode with 3 shards, 1 replica per shard, TLS encryption, RDB persistence, multi-zone distribution, a maintenance window, and deletion protection. It is suitable for production workloads that require high availability, data durability, and security controls.

## When to Use

- Production application caching with high availability requirements
- Session storage for stateless web applications
- Workloads requiring encrypted client connections
- Environments where accidental deletion must be prevented
- Applications using cluster-aware Valkey/Redis drivers

## Key Configuration

- **CLUSTER mode** — sharded topology with native cluster protocol; clients must use cluster-aware drivers
- **3 shards, 1 replica** — data distributed across 3 shards, each with 1 read replica for failover and read scaling
- **HIGHMEM_MEDIUM node type** — high-memory dedicated nodes for moderate production workloads
- **transitEncryptionMode: SERVER_AUTHENTICATION** — TLS for client-to-server connections
- **persistenceConfig: RDB** — point-in-time snapshots every 12 hours for data durability
- **zoneDistributionConfig: MULTI_ZONE** — nodes spread across multiple availability zones for high availability
- **maintenancePolicy** — Sunday 3:00 UTC; GCP applies patches during this window
- **deletionProtectionEnabled** — prevents accidental destruction of the instance

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | Consumer project ID where the PSC endpoint is created (`pscAutoConnections[].projectId`) | GCP Console or `GcpProject` outputs |
| `my-prod-memorystore` (`instanceName`) | Name for this Memorystore instance (4-63 chars, lowercase, hyphens) | Choose a descriptive name |
| `us-central1` (`location`) | GCP region for the instance | [GCP regions](https://cloud.google.com/about/locations) |
| `prod-vpc` | Name of the GcpVpcNetwork resource the PSC connection attaches to (its `network_id` is read) | Your network manifest's `metadata.name` |

## Related Presets

- **01-dev-single-shard** — Minimal standalone instance for dev/test
- **03-enterprise-cluster** — CLUSTER mode with 5 shards, IAM auth, CMEK, AOF persistence, and automated backups
