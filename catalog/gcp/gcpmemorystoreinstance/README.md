# GCP Memorystore Instance

Deploys a Google Cloud Memorystore (Valkey) instance (`google_memorystore_instance`) — the new-generation, PSC-first in-memory data store with native sharding, predefined node types, RDB/AOF persistence, automated backups, and cross-region replication. Redis-compatible via the Valkey protocol.

## Overview

Google Cloud Memorystore (new-generation) replaces the legacy Memorystore for Redis API with a modern architecture built around Valkey, PSC-based networking, and first-class clustering support. It provides sub-millisecond latency for caching, session management, real-time analytics, leaderboards, and pub/sub messaging — while eliminating VPC peering complexity and adding features the legacy API never offered.

**Connectivity prerequisite**: the instance reaches your VPC through service connectivity automation, which requires a `GcpServiceConnectionPolicy` for the `gcp-memorystore` service class on the network in the instance's region — deployed BEFORE the instance. Without it, instance creation fails with a connectivity error.

## Key Differences from GcpRedisInstance (Legacy)

| Feature | GcpRedisInstance (Legacy) | GcpMemorystoreInstance (New-Gen) |
|---------|--------------------------|----------------------------------|
| **Engine** | Redis | Valkey (Redis-compatible) |
| **Networking** | VPC peering / Private Service Access | Private Service Connect (PSC) |
| **Sharding** | Not supported | Native via `shard_count` |
| **Node sizing** | `memory_size_gb` (arbitrary) | Predefined `node_type` (NANO–2XLARGE) |
| **Modes** | BASIC / STANDARD_HA | CLUSTER / CLUSTER_DISABLED |
| **Persistence** | RDB only | RDB and AOF |
| **Automated backups** | Not supported | Built-in with configurable retention |
| **Cross-region DR** | Not supported | PRIMARY/SECONDARY replication |
| **Auth** | AUTH string | IAM-based authentication |

## Quick Start

Deploy the connection policy once per (network, region), then the instance:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceConnectionPolicy
metadata:
  name: memorystore-valkey-policy
spec:
  location: us-central1
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: prod-vpc
      fieldPath: status.outputs.network_id
  serviceClass: gcp-memorystore
  pscConfig:
    subnetworks:
      - valueFrom:
          kind: GcpSubnetwork
          name: prod-subnet
          fieldPath: status.outputs.subnetwork_self_link
---
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMemorystoreInstance
metadata:
  name: dev-cache
spec:
  instanceName: dev-cache
  location: us-central1
  shardCount: 1
  mode: CLUSTER_DISABLED
  nodeType: SHARED_CORE_NANO
  pscAutoConnections:
    - network:
        valueFrom:
          kind: GcpVpcNetwork
          name: prod-vpc
          fieldPath: status.outputs.network_id
```

## Configuration Options

| Category | Options |
|----------|---------|
| **Mode** | `CLUSTER` (sharded, cluster-aware clients) or `CLUSTER_DISABLED` (standalone) |
| **Node type** | Shared-core/custom: `SHARED_CORE_NANO`, `CUSTOM_PICO`, `CUSTOM_MICRO`, `CUSTOM_MINI`; dedicated: `STANDARD_SMALL`, `STANDARD_LARGE`, `HIGHCPU_MEDIUM`, `HIGHMEM_MEDIUM`, `HIGHMEM_XLARGE`, `HIGHMEM_2XLARGE` |
| **Sharding** | `shard_count` (1+ shards; each shard handles a portion of the keyspace) |
| **Replicas** | `replica_count` (0–5 read replicas per shard for read scaling and failover) |
| **Engine** | `engine_version` (e.g., `VALKEY_8_0`, `VALKEY_7_2`); `engine_configs` for tuning |
| **Persistence** | `persistence_config.mode`: `DISABLED`, `RDB` (periodic snapshots), or `AOF` (append-only file) |
| **Encryption** | `transit_encryption_mode: SERVER_AUTHENTICATION` for TLS; `kms_key` for CMEK at rest |
| **Server CA** | `server_ca_mode` — which CA signs the TLS server certificate: `GOOGLE_MANAGED_PER_INSTANCE_CA` (default), `GOOGLE_MANAGED_SHARED_CA` (one CA to trust fleet-wide), or `CUSTOMER_MANAGED_CAS_CA` paired with `server_ca_pool` (your CA Service pool: a `GcpPrivateCaPool` reference or its full name); both immutable |
| **Auth** | `authorization_mode: IAM_AUTH` for IAM-based client authentication |
| **Networking** | `psc_auto_connections` — PSC endpoints in consumer VPCs (immutable after creation); a per-entry `project_id` omitted rides the provider's effective project |
| **Zones** | `zone_distribution_config`: `MULTI_ZONE` (HA default) or `SINGLE_ZONE` |
| **Maintenance** | `maintenance_policy` — weekly window (day + hour UTC; the API accepts only whole-hour start times); `maintenance_version` — self-service patching trigger (update-only, forward-only) |
| **Backups** | `automated_backup_config` — daily backups with configurable start hour and retention |
| **Cross-region DR** | `cross_instance_replication_config` — PRIMARY replicating to secondaries, or SECONDARY replicating from a primary (roles exchange in place during switchover) |
| **Seeding** | `gcs_source` (RDB files in GCS) XOR `managed_backup_source` (a managed backup) — creation-time only |
| **Labels** | `labels` — user labels merged beneath platform attribution labels |
| **Protection** | `deletion_protection_enabled` — defaults TRUE; destroy fails until explicitly set false. `deletion_policy` — `DELETE` (default), `PREVENT`, or `ABANDON` (evaluated only after deletion protection allows the destroy) |

**Immutable fields** (require instance replacement if changed): `instance_name`, `location`, `mode`, `authorization_mode`, `transit_encryption_mode`, `kms_key`, `zone_distribution_config`, `psc_auto_connections`, and the seed sources.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `discovery_address` | string | IP address of the PSC discovery endpoint |
| `discovery_port` | int32 | Port of the discovery endpoint (typically 6379) |
| `instance_uid` | string | Server-generated unique identifier for the instance |
| `node_size_gb` | double | Memory size per node in GB (determined by `node_type`) |
| `name` | string | Full resource path — the composition key for cross-instance replication |
| `backup_collection` | string | Where automated backups land — the source of `managed_backup_source` paths |

## Important Notes

- **Deploy the connection policy first**: a `GcpServiceConnectionPolicy` (`serviceClass: gcp-memorystore`) must exist on each `psc_auto_connections` network in this region before the instance is created, and must outlive the instance.
- **Deletion protection defaults on**: destroying an instance requires explicitly setting `deletion_protection_enabled: false` first — both engines send the flag explicitly, so destroy behavior is identical regardless of engine.
- **DR composition**: deploy the PRIMARY first (listing its secondaries by full resource path — another instance's `name` output), then each SECONDARY pointing back at the primary. A secondary is read-only until promoted.
- **Node memory is a consequence of `node_type`** — reported in `node_size_gb`, never configured directly.

### Deliberately not modeled (recorded reasons)

- **Maintenance window `minutes`/`seconds`/`nanos`** — the Memorystore API
  accepts only whole-hour window start times (a minute-carrying start is
  rejected with 400 "Invalid start time, only hours are supported" —
  live-verified; narrower than Redis); the spec models the start hour-only
  (recorded in `iac/provider-parity.yaml`).
- **`allow_fewer_zones_deployment`** — a beta-only argument (not in the GA
  provider at the pinned version); GA is the parity baseline, and beta
  surface enters only through the enumerated admission list.
- **`google_memorystore_instance_desired_user_created_endpoints`** — the
  bring-your-own-forwarding-rules alternative to auto-created endpoints; a
  Tier-2 sibling resource on concrete pull.

Every other configurable argument of `google_memorystore_instance` at the
pinned provider version is representable through this spec; the recorded
judgment lives in `iac/provider-parity.yaml`, checked by
`planton provider-parity --check`.

## When to Use GcpMemorystoreInstance vs GcpRedisInstance

**Use GcpMemorystoreInstance (this component) when:**
- You need native sharding for horizontal data distribution
- You prefer PSC networking over VPC peering
- You want AOF persistence, automated backups, or cross-region DR
- You are starting a new deployment with no legacy constraints
- You need IAM-based authentication

**Use GcpRedisInstance (legacy) when:**
- You have existing Memorystore for Redis instances and need consistency
- You require AUTH string–based authentication (not IAM)
- You depend on VPC peering or Private Service Access connectivity

## Related Kinds

- **GcpServiceConnectionPolicy** — the required PSC authorization on the network (deploy first)
- **GcpVpcNetwork** — provides the VPC network for PSC auto-connections
- **GcpKmsKey** — provides a CMEK key for encryption at rest
- **GcpProject** — provides the GCP project ID
- **GcpRedisInstance** — legacy Memorystore for Redis (VPC peering model)

## Additional Resources

- [Memorystore Documentation](https://cloud.google.com/memorystore/docs/overview)
- [Valkey Project](https://valkey.io/)
- [Service connectivity automation](https://cloud.google.com/vpc/docs/about-service-connectivity-automation)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
