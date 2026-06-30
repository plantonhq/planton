# GCP Memorystore Instance (New-Generation API)

Deploys a Google Cloud Memorystore instance via the new-generation API using Pulumi `memorystore.Instance`. Memorystore is a fully managed, in-memory data store supporting the Valkey protocol (Redis-compatible), with native sharding, Private Service Connect (PSC) networking, and predefined node types.

## Overview

Google Cloud Memorystore (new-generation) replaces the legacy Memorystore for Redis API with a modern architecture built around Valkey, PSC-based networking, and first-class clustering support. It provides sub-millisecond latency for caching, session management, real-time analytics, leaderboards, and pub/sub messaging — while eliminating VPC peering complexity and adding features that the legacy API never offered.

## Key Differences from GcpRedisInstance (Legacy)

| Feature | GcpRedisInstance (Legacy) | GcpMemorystoreInstance (New-Gen) |
|---------|--------------------------|----------------------------------|
| **Engine** | Redis | Valkey (Redis-compatible) |
| **Networking** | VPC peering / Private Service Access | Private Service Connect (PSC) |
| **Sharding** | Not supported | Native via `shard_count` |
| **Node sizing** | `memory_size_gb` (arbitrary) | Predefined `node_type` (NANO–XLARGE) |
| **Modes** | BASIC / STANDARD_HA | CLUSTER / CLUSTER_DISABLED |
| **Persistence** | RDB only | RDB and AOF |
| **Automated backups** | Not supported | Built-in with configurable retention |
| **Auth** | AUTH string | IAM-based authentication |

## Quick Start

A minimal manifest to create a development Memorystore instance:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpMemorystoreInstance
metadata:
  name: dev-cache
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpMemorystoreInstance.dev-cache
spec:
  projectId:
    value: <gcp-project-id>
  instanceName: dev-cache
  location: us-central1
  shardCount: 1
  mode: CLUSTER_DISABLED
  nodeType: SHARED_CORE_NANO
  pscAutoConnections:
    - network:
        value: "projects/<gcp-project-id>/global/networks/<vpc-name>"
      projectId:
        value: "<gcp-project-id>"
```

## Configuration Options

| Category | Options |
|----------|---------|
| **Mode** | `CLUSTER` (sharded, cluster-aware clients) or `CLUSTER_DISABLED` (standalone) |
| **Node type** | `SHARED_CORE_NANO`, `STANDARD_SMALL`, `HIGHMEM_MEDIUM`, `HIGHMEM_XLARGE` |
| **Sharding** | `shard_count` (1+ shards; each shard handles a portion of the keyspace) |
| **Replicas** | `replica_count` (0–5 read replicas per shard for read scaling and failover) |
| **Engine** | `engine_version` (e.g., `VALKEY_8_0`, `VALKEY_7_2`); `engine_configs` for tuning |
| **Persistence** | `persistence_config.mode`: `DISABLED`, `RDB` (periodic snapshots), or `AOF` (append-only file) |
| **Encryption** | `transit_encryption_mode: SERVER_AUTHENTICATION` for TLS; `kms_key` for CMEK at rest |
| **Auth** | `authorization_mode: IAM_AUTH` for IAM-based client authentication |
| **Networking** | `psc_auto_connections` — PSC endpoints in consumer VPCs (immutable after creation) |
| **Zones** | `zone_distribution_config`: `MULTI_ZONE` (HA default) or `SINGLE_ZONE` |
| **Maintenance** | `maintenance_policy` — weekly window (day + hour UTC) |
| **Backups** | `automated_backup_config` — daily backups with configurable start hour and retention |
| **Protection** | `deletion_protection_enabled: true` to prevent accidental deletion |

**Immutable fields** (require instance replacement if changed): `instance_name`, `location`, `mode`, `authorization_mode`, `transit_encryption_mode`, `kms_key`, `zone_distribution_config`, `psc_auto_connections`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `discovery_address` | string | IP address of the PSC discovery endpoint |
| `discovery_port` | int32 | Port of the discovery endpoint (typically 6379) |
| `instance_uid` | string | Server-generated unique identifier for the instance |
| `node_size_gb` | double | Memory size per node in GB (determined by `node_type`) |

## When to Use GcpMemorystoreInstance vs GcpRedisInstance

**Use GcpMemorystoreInstance (this component) when:**
- You need native sharding for horizontal data distribution
- You prefer PSC networking over VPC peering
- You want AOF persistence or automated backups
- You are starting a new deployment with no legacy constraints
- You need IAM-based authentication

**Use GcpRedisInstance (legacy) when:**
- You have existing Memorystore for Redis instances and need consistency
- You require AUTH string–based authentication (not IAM)
- You depend on VPC peering or Private Service Access connectivity
- You need read replicas with a separate read endpoint (legacy API model)

## Related Components

- **GcpProject** — provides the GCP project ID
- **GcpVpc** — provides the VPC network for PSC auto-connections
- **GcpKmsKey** — provides a CMEK key for encryption at rest
- **GcpRedisInstance** — legacy Memorystore for Redis (VPC peering model)

## Additional Resources

- [Memorystore Documentation](https://cloud.google.com/memorystore/docs/overview)
- [Memorystore Instance REST API](https://cloud.google.com/memorystore/docs/reference/rest/v1/projects.locations.instances)
- [Valkey Project](https://valkey.io/)
- [Private Service Connect Overview](https://cloud.google.com/vpc/docs/private-service-connect)
