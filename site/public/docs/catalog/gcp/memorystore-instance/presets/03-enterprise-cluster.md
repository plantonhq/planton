---
title: "Enterprise Cluster"
description: "This preset provisions a fully-featured Memorystore instance in CLUSTER mode with 5 shards, 2 replicas per shard, IAM authentication, TLS encryption, customer-managed encryption keys (CMEK), AOF..."
type: "preset"
rank: "03"
presetSlug: "03-enterprise-cluster"
componentSlug: "memorystore-instance"
componentTitle: "Memorystore Instance"
provider: "gcp"
icon: "package"
order: 3
---

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
| `<gcp-project-id>` | GCP project ID where the instance will be created | GCP Console or `GcpProject` outputs |
| `<instance-name>` | Name for this Memorystore instance (4-63 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `enterprise-cache`) |
| `<gcp-region>` | GCP region for the instance (e.g., `us-central1`) | [GCP regions](https://cloud.google.com/about/locations) |
| `<vpc-network-path>` | Full path of the VPC network (e.g., `projects/my-project/global/networks/prod-vpc`) | `GcpVpc` status outputs or GCP Console |
| `<kms-key-resource-name>` | Full KMS key resource name (e.g., `projects/my-project/locations/us-central1/keyRings/cache-keys/cryptoKeys/cache-cmek`) or reference via `valueFrom` | `GcpKmsKey` status outputs or GCP Console |

## Related Presets

- **01-dev-single-shard** — Minimal standalone instance for dev/test
- **02-ha-production** — CLUSTER mode with 3 shards; simpler production setup without IAM auth or CMEK
