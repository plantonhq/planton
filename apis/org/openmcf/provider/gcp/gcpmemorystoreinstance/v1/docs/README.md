# GcpMemorystoreInstance — Research and Design Documentation

## 1. Introduction

### The Memorystore Deployment Landscape

Google Cloud offers two distinct APIs for managed in-memory data stores under the Memorystore brand:

1. **Memorystore for Redis (Legacy API)** — The original offering, backed by the `google_redis_instance` Terraform resource (or Pulumi `redis.Instance`). It provisions standalone or HA Redis instances connected via VPC peering. This API is modeled in OpenMCF as **GcpRedisInstance**.

2. **Memorystore (New-Generation API)** — A fundamentally redesigned service using the `google_memorystore_instance` Terraform resource (or Pulumi `memorystore.Instance`). It introduces the Valkey engine, Private Service Connect (PSC) networking, native sharding, predefined node types, AOF persistence, and automated backups. This API is modeled in OpenMCF as **GcpMemorystoreInstance**.

The new-generation API is not a drop-in replacement for the legacy API. It uses a different resource model, different networking primitives, and a different engine. Google is steering new deployments toward the new-generation API, but legacy instances remain fully supported.

### Why a Separate Component?

Despite both services being "Memorystore," the APIs are different enough to warrant separate OpenMCF components:

- Different Terraform/Pulumi resource types (`redis.Instance` vs. `memorystore.Instance`)
- Different networking models (VPC peering vs. PSC)
- Different spec fields (memory_size_gb vs. node_type, tier vs. mode)
- Different output shapes (host/port vs. discovery_address/discovery_port)
- Different authentication models (AUTH string vs. IAM)

Merging them into a single component would create a confusing union type with conditional validation. Separate components give each API a clean, focused spec.

---

## 2. Valkey vs. Redis

### History

Redis was created by Salvatore Sanfilippo in 2009 as an open-source, in-memory data structure store. It became the dominant caching layer for modern applications. In March 2024, Redis Ltd. changed the Redis license from BSD to a dual Server Side Public License (SSPL) / Redis Source Available License (RSAL), restricting cloud providers from offering Redis as a managed service without a commercial agreement.

In response, the Linux Foundation forked Redis 7.2.4 and created **Valkey** — a community-driven, BSD-licensed continuation of Redis. Major cloud providers (Google, AWS, Oracle) immediately backed Valkey as the engine for their managed services.

### Compatibility

Valkey maintains wire-protocol compatibility with Redis. Existing Redis clients, libraries, and tools work with Valkey without modification. The command set is identical for all practical purposes. Applications migrating from Redis to Valkey need no code changes — only connection endpoint updates.

### Why Google Chose Valkey

Google's new-generation Memorystore uses Valkey as the engine because:

1. **License freedom**: BSD license allows Google to offer a fully managed service without SSPL restrictions
2. **Community governance**: Linux Foundation stewardship ensures no single vendor controls the project
3. **Feature parity**: Valkey 7.2+ matches Redis 7.2 feature-for-feature
4. **Forward development**: Valkey 8.0 introduces new features (e.g., improved multi-threading) not available in Redis OSS
5. **Industry alignment**: AWS (ElastiCache), Oracle (OCI Cache), and others also adopted Valkey

For OpenMCF users, the engine choice is transparent: applications use the same commands, the same client libraries, and the same data structures. The only user-visible difference is the `engine_version` field using `VALKEY_8_0` or `VALKEY_7_2` instead of `REDIS_7_0`.

---

## 3. Private Service Connect (PSC) Networking

### How PSC Works

Private Service Connect is Google Cloud's private networking model for connecting consumer VPCs to managed services without VPC peering. Instead of peering two networks (which shares route tables and has limits), PSC creates a forwarding rule in the consumer VPC that tunnels traffic to the service producer's network via Google's internal backbone.

For Memorystore, the flow is:

```
Application Pod/VM (Consumer VPC)
   ↓ connect to PSC endpoint IP
PSC Forwarding Rule (auto-created in consumer VPC)
   ↓ internal tunnel via Google backbone
Memorystore Instance (Producer VPC, managed by Google)
```

Each PSC auto-connection creates a forwarding rule and an internal IP address in the consumer VPC. Applications connect to this IP address as if the Memorystore instance were a local resource.

### PSC vs. VPC Peering

| Aspect | VPC Peering (Legacy) | PSC (New-Gen) |
|--------|---------------------|---------------|
| **Route table sharing** | Yes — both VPCs see each other's routes | No — consumer only sees the PSC endpoint IP |
| **IP conflicts** | Possible if CIDR ranges overlap | Impossible — PSC uses forwarding rules |
| **Peering limits** | 25 peering connections per VPC | No peering limits; PSC scales independently |
| **Cross-project** | Requires peering in both projects | PSC endpoint created in consumer project only |
| **Transitive routing** | Not supported (no peering-of-peering) | N/A — PSC endpoints are directly routable |
| **Setup complexity** | Reserve IP range, create peering, wait for propagation | Declare PSC connection in spec; GCP creates endpoints |
| **Security** | Firewall rules on both sides | Consumer VPC only; service producer is isolated |

### Implications for OpenMCF

PSC connections are **immutable after instance creation**. This means:

- The consumer VPC and project must be determined at deployment time
- Adding or changing PSC connections requires instance replacement
- For cross-project access, define multiple `psc_auto_connections` entries upfront

The OpenMCF spec models PSC connections as a repeated `GcpMemorystoreInstancePscAutoConnection` message with `network` and `project_id` fields, both supporting `StringValueOrRef` for cross-resource references.

---

## 4. Node Types and Capacity Planning

### Predefined Node Types

Unlike the legacy API where you specify arbitrary `memory_size_gb`, the new-generation API uses predefined node types that bundle CPU and memory:

| Node Type | Category | Approximate Memory | Use Case |
|-----------|----------|-------------------|----------|
| `SHARED_CORE_NANO` | Shared core | ~1.5 GB | Development, testing, tiny workloads |
| `STANDARD_SMALL` | Dedicated core | ~6.5 GB | Small production, low-traffic caching |
| `HIGHMEM_MEDIUM` | High memory | ~13 GB | Medium production, session stores |
| `HIGHMEM_XLARGE` | High memory | ~58 GB | Large production, high-throughput caching |

**Note:** Exact memory per node is determined by GCP and may change. The actual value is reported in the `node_size_gb` stack output after provisioning.

### Capacity Calculation

Total instance memory depends on node type, shard count, and replica count:

```
Total memory = node_size_gb × shard_count × (1 + replica_count)
```

Examples:
- 1 shard × SHARED_CORE_NANO × 0 replicas = ~1.5 GB total
- 3 shards × HIGHMEM_MEDIUM × 1 replica = ~78 GB total (39 GB primary + 39 GB replica)
- 5 shards × HIGHMEM_XLARGE × 2 replicas = ~870 GB total

### Choosing a Node Type

- **SHARED_CORE_NANO**: Cost-optimized for non-production. Not suitable for latency-sensitive workloads due to shared CPU.
- **STANDARD_SMALL**: Good starting point for production. Dedicated CPU provides consistent performance.
- **HIGHMEM_MEDIUM**: Recommended for most production workloads. Balances memory capacity with cost.
- **HIGHMEM_XLARGE**: For large-scale caching, high cardinality datasets, or when minimizing shard count is important.

---

## 5. Sharding and Replication Architecture

### Cluster Mode (CLUSTER)

In CLUSTER mode, the instance uses the native cluster protocol (compatible with Redis Cluster):

- Data is partitioned across shards using hash slots (16,384 slots total)
- Each shard owns a range of hash slots and stores the corresponding key-value pairs
- Clients must use cluster-aware drivers that handle slot discovery and command routing
- Adding shards redistributes hash slots and data automatically (online resharding)

**When to use CLUSTER:**
- Datasets larger than a single node's memory
- Write-heavy workloads that benefit from distributed writes
- When horizontal scaling is more important than simplicity

### Standalone Mode (CLUSTER_DISABLED)

In CLUSTER_DISABLED mode, the instance has a single primary with optional replicas:

- All data resides on a single shard (one primary node)
- Any Valkey/Redis client works — no cluster-aware driver needed
- Replicas provide read scaling and automatic failover (but all data is on each replica)
- Simpler operational model for workloads that fit in a single node

**When to use CLUSTER_DISABLED:**
- Datasets that fit in a single node type's memory
- Simpler client configuration requirements
- Applications using multi-key operations across arbitrary keys (which require same-slot in CLUSTER mode)

### Replication

Replicas (0–5 per shard) provide:

1. **Read scaling**: Replicas handle read-only traffic, reducing load on the primary
2. **Automatic failover**: If the primary fails, a replica is promoted automatically
3. **Zone redundancy**: With MULTI_ZONE distribution, replicas are placed in different zones

Each replica is a full copy of its shard's data. Write operations always go to the primary; replication is asynchronous.

---

## 6. Persistence Options

### Overview

Memorystore (new-gen) supports three persistence modes:

| Mode | Description | Data Loss Window | Performance Impact |
|------|-------------|-----------------|-------------------|
| `DISABLED` | No persistence; data is in-memory only | All data lost on restart | None |
| `RDB` | Periodic point-in-time snapshots | Up to snapshot interval | Low (background fork) |
| `AOF` | Append-only file logging every write | Depends on fsync setting | Medium to High |

### RDB (Redis Database) Snapshots

RDB creates a binary snapshot of the entire dataset at configurable intervals:

| Snapshot Period | Interval | Typical Use Case |
|----------------|----------|-----------------|
| `ONE_HOUR` | Every hour | High-value data, minimal acceptable loss |
| `SIX_HOURS` | Every 6 hours | Production caches with moderate durability needs |
| `TWELVE_HOURS` | Every 12 hours | Less critical data, lower I/O overhead |
| `TWENTY_FOUR_HOURS` | Daily | Cost-optimized persistence for non-critical data |

Optional `rdb_snapshot_start_time` allows pinning the first snapshot to a specific RFC3339 timestamp.

**Pros:**
- Lower performance impact than AOF (single background fork)
- Compact storage format
- Fast recovery (load binary snapshot)

**Cons:**
- Data loss up to the snapshot interval on crash
- Memory spike during fork (copy-on-write overhead)

### AOF (Append-Only File)

AOF logs every write command to disk, providing stronger durability:

| Fsync Mode | Behavior | Data Loss Risk | Performance |
|------------|----------|---------------|-------------|
| `NEVER` | OS decides when to flush | Up to OS buffer (seconds) | Best performance |
| `EVERY_SEC` | Flush once per second | Up to 1 second of writes | Good balance |
| `ALWAYS` | Flush on every write | Minimal (single write) | Highest latency |

**Pros:**
- Finer-grained durability than RDB
- Configurable trade-off between durability and performance

**Cons:**
- Higher I/O than RDB (continuous disk writes)
- Larger file size than RDB snapshots
- Slower recovery (replay all commands)

### RDB vs. AOF Comparison

| Dimension | RDB | AOF |
|-----------|-----|-----|
| **Durability** | Up to interval gap | Up to fsync gap (1 sec typical) |
| **Performance impact** | Low (periodic fork) | Medium (continuous writes) |
| **Recovery speed** | Fast (binary load) | Slower (command replay) |
| **Storage size** | Compact binary | Larger (append log) |
| **Best for** | Caches with moderate durability | Session stores, critical state |

### Recommendation

- **Pure cache** (data can be rebuilt): `DISABLED`
- **Cache with nice-to-have durability**: `RDB` with `SIX_HOURS` or `TWELVE_HOURS`
- **Session store or critical state**: `AOF` with `EVERY_SEC`
- **Maximum durability**: `AOF` with `ALWAYS` (accept performance cost)

---

## 7. CMEK Encryption

### How It Works

By default, Memorystore encrypts data at rest using Google-managed encryption keys. For compliance requirements that mandate customer control over encryption keys, the `kms_key` field specifies a Cloud KMS CryptoKey:

```
projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
```

When CMEK is configured:
- GCP encrypts the Memorystore data with the specified key
- The Memorystore service account must have `cloudkms.cryptoKeyEncrypterDecrypter` role on the key
- The KMS key must be in the same region as the Memorystore instance
- The key is immutable after instance creation — changing it requires instance replacement

### OpenMCF Integration

The `kms_key` field supports `StringValueOrRef`, enabling references to a `GcpKmsKey` resource:

```yaml
kmsKey:
  value: "projects/my-project/locations/us-central1/keyRings/cache-keys/cryptoKeys/memorystore-cmek"
```

Or via cross-resource reference:

```yaml
kmsKey:
  valueFrom:
    kind: GcpKmsKey
    name: memorystore-cmek
    field: status.outputs.key_id
```

---

## 8. Automated Backups

### Overview

The new-generation API supports automated daily backups — a feature not available in the legacy Redis API. Backups are stored in a GCP-managed backup collection and can be used to restore data to a new instance.

### Configuration

| Field | Description | Constraints |
|-------|-------------|-------------|
| `start_hour` | Hour of day (UTC) when backup starts | 0–23 |
| `retention` | How long backups are kept | `86400s` (1 day) to `31536000s` (365 days) |

### Behavior

- Backups run daily at the specified hour
- Each backup captures the full dataset (all shards)
- Older backups are automatically deleted when they exceed the retention period
- Backups are incremental internally (GCP manages storage efficiency)
- Backup and restore operations do not affect instance availability

### Recommendation

For production instances, enable automated backups with 7–35 days of retention. Use `3024000s` (35 days) as a reasonable default that covers most incident response windows.

---

## 9. 80/20 Scoping Rationale

### What We Included

The OpenMCF spec covers the fields needed for ~80% of Memorystore deployments:

| Feature | Rationale |
|---------|-----------|
| `instance_name`, `location`, `project_id` | Core identity — every instance needs these |
| `shard_count`, `mode` | Fundamental topology decision (cluster vs. standalone) |
| `node_type` | Primary sizing lever |
| `engine_version`, `engine_configs` | Version pinning and tuning |
| `replica_count` | HA and read scaling |
| `psc_auto_connections` | Only connectivity model for new-gen API |
| `authorization_mode` | Security: IAM auth |
| `transit_encryption_mode` | Security: TLS in transit |
| `kms_key` | Security: CMEK at rest |
| `persistence_config` (RDB + AOF) | Data durability |
| `zone_distribution_config` | HA: multi-zone vs. single-zone |
| `maintenance_policy` | Operational: controlled maintenance windows |
| `automated_backup_config` | Operational: daily backups with retention |
| `deletion_protection_enabled` | Safety: prevent accidental deletion |

### What We Excluded and Why

| Excluded Feature | Reason |
|-----------------|--------|
| **Cross-instance replication** | Niche feature for disaster recovery across regions. Complex topology that most users don't need. Can be added in a future version. |
| **`gcs_source`** | Used to import data from a GCS bucket during instance creation. One-time migration operation, not a steady-state configuration. Better handled by a separate migration workflow. |
| **`managed_backup_source`** | Used to restore an instance from a managed backup. One-time restore operation, not a declarative spec field. Better handled imperatively. |
| **User-created endpoints** | The API supports `DesiredUserCreatedEndpoints` for manually managed PSC endpoints. Auto-created endpoints (`DesiredAutoCreatedEndpoints`) cover the common case. User-created endpoints are for advanced networking scenarios (e.g., custom forwarding rules). |
| **`display_name`** | Cosmetic field. Instance identification is handled by `instance_name` and OpenMCF labels. |
| **`secondary_zone`** | Zone pinning within MULTI_ZONE. GCP handles zone placement optimally; manual zone pinning is rarely needed. |
| **`ondemand_maintenance`** | Triggers immediate maintenance. Imperative action, not a declarative field. |

### Design Principle

The spec follows the 80/20 rule: cover the majority of production use cases with a clean, focused API. Advanced features (cross-instance replication, GCS import, user-created endpoints) are deferred until user demand justifies the additional complexity. All excluded features can be accessed directly via Terraform or Pulumi if needed for a specific deployment.

---

## 10. Comparison with GcpRedisInstance

### Architectural Differences

| Dimension | GcpRedisInstance (Legacy) | GcpMemorystoreInstance (New-Gen) |
|-----------|--------------------------|----------------------------------|
| **GCP API** | `redis.googleapis.com/v1` | `memorystore.googleapis.com/v1` |
| **Terraform resource** | `google_redis_instance` | `google_memorystore_instance` |
| **Pulumi resource** | `redis.Instance` | `memorystore.Instance` |
| **Engine** | Redis (SSPL-licensed upstream) | Valkey (BSD-licensed) |
| **Networking** | VPC peering / Private Service Access | Private Service Connect (PSC) |
| **Topology** | BASIC (standalone) / STANDARD_HA (primary+replica) | CLUSTER (sharded) / CLUSTER_DISABLED (standalone) |
| **Sizing** | Arbitrary `memory_size_gb` | Predefined `node_type` |
| **Sharding** | Not supported | Native via `shard_count` |
| **Persistence** | RDB only | RDB and AOF |
| **Authentication** | AUTH string (GCP-managed) | IAM-based |
| **Automated backups** | Not supported | Built-in with configurable retention |
| **Read endpoint** | Separate `read_endpoint` + `read_endpoint_port` | Single `discovery_address` + `discovery_port` |

### Output Differences

| GcpRedisInstance Outputs | GcpMemorystoreInstance Outputs |
|--------------------------|-------------------------------|
| `host` | `discovery_address` |
| `port` | `discovery_port` |
| `read_endpoint` | (N/A — unified discovery endpoint) |
| `read_endpoint_port` | (N/A) |
| `current_location_id` | (N/A — zone info not exposed) |
| `auth_string` | (N/A — IAM auth, no string) |
| (N/A) | `instance_uid` |
| (N/A) | `node_size_gb` |

### Migration Path

There is no in-place migration from legacy Redis to new-gen Memorystore. Migration involves:

1. Provision a new GcpMemorystoreInstance
2. Use application-level data migration (dual-write, or export/import via RDB)
3. Update application connection strings to the new discovery endpoint
4. Decommission the old GcpRedisInstance

This is a deliberate architectural decision by Google — the APIs are fundamentally different services.

---

## 11. Production Best Practices

### Sizing

- **Start small**: Use `SHARED_CORE_NANO` for dev/test. Upgrade node type as workload grows.
- **Monitor memory usage**: Use Cloud Monitoring to track memory utilization. Size for 60–70% peak utilization to leave headroom for traffic spikes and RDB fork overhead.
- **Shard for scale**: If dataset exceeds single-node capacity, use `CLUSTER` mode with multiple shards. Prefer fewer large shards over many small shards for simplicity.

### Networking

- **Always use PSC**: PSC is the only networking option. Define at least one `psc_auto_connections` entry for the instance to be reachable.
- **Plan PSC connections upfront**: PSC connections are immutable. If multi-VPC or cross-project access is needed, include all connections at creation time.
- **Firewall rules**: PSC endpoints are internal IPs in the consumer VPC. Standard VPC firewall rules apply — ensure the application subnet can reach port 6379.

### Security

- **Enable IAM auth** (`authorization_mode: IAM_AUTH`) for production. IAM integrates with Google's identity platform and provides auditable access control.
- **Enable TLS** (`transit_encryption_mode: SERVER_AUTHENTICATION`) for encrypted client connections. Especially important when transit crosses network boundaries.
- **Use CMEK** when compliance requires customer-managed encryption keys. Ensure proper key rotation policies.

### Persistence

- **Pure cache**: Set `persistence_config.mode: DISABLED`. Data can be rebuilt from the source of truth.
- **Session store**: Use `AOF` with `EVERY_SEC` for sub-second durability.
- **Critical state**: Use `AOF` with `ALWAYS` and accept the performance overhead.
- **General durability**: Use `RDB` with `SIX_HOURS` for a good balance.

### Operational

- **Maintenance windows**: Always set a `maintenance_policy` for production. Choose a low-traffic period. Expect brief connectivity interruptions during the 1-hour window.
- **Automated backups**: Enable with 7–35 day retention for production instances. Backups provide a safety net for accidental data loss or corruption.
- **Deletion protection**: Enable `deletion_protection_enabled: true` for all production instances. Prevents accidental destruction during `openmcf destroy` or `pulumi destroy`.
- **Zone distribution**: Use `MULTI_ZONE` (default) for production. `SINGLE_ZONE` only for latency-sensitive workloads where all consumers are in the same zone.

### Monitoring

- **Cloud Monitoring**: Memorystore exports metrics including memory usage, connected clients, commands per second, and cache hit ratio. Create alerts for high memory utilization (>80%), connection limits, and elevated error rates.
- **Uptime checks**: Verify the PSC endpoint is reachable from the application network.
- **Client-side metrics**: Instrument your application's cache client to track hit/miss ratios, latency percentiles, and connection pool utilization.

---

## 12. Implementation Landscape

### Pulumi Module Architecture

The OpenMCF GcpMemorystoreInstance Pulumi module lives at:

```
apis/org/openmcf/provider/gcp/gcpmemorystoreinstance/v1/iac/pulumi/
```

**Resource**: `github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/memorystore.Instance`

**Key mappings:**

| Spec Field | Pulumi Property |
|------------|-----------------|
| instanceName | InstanceId |
| projectId | Project |
| location | Location |
| shardCount | ShardCount |
| mode | Mode |
| nodeType | NodeType |
| engineVersion | EngineVersion |
| engineConfigs | EngineConfigs |
| replicaCount | ReplicaCount |
| pscAutoConnections | DesiredAutoCreatedEndpoints |
| authorizationMode | AuthorizationMode |
| transitEncryptionMode | TransitEncryptionMode |
| kmsKey | KmsKey |
| persistenceConfig | PersistenceConfig |
| zoneDistributionConfig | ZoneDistributionConfig |
| maintenancePolicy | MaintenancePolicy |
| automatedBackupConfig | AutomatedBackupConfig |
| deletionProtectionEnabled | DeletionProtectionEnabled |

**Output extraction:**

The discovery endpoint is extracted from the nested PSC structure using `ApplyT`:

```
createdInstance.Endpoints → []InstanceEndpoint
  → Connections → []InstanceEndpointConnection
    → PscAutoConnection → { IpAddress, Port, ConnectionType }
```

The module searches for `CONNECTION_TYPE_DISCOVERY` and falls back to any available connection. Node memory size is extracted from `NodeConfigs[0].SizeGb`.

### Terraform Module Architecture

The Terraform module lives at:

```
apis/org/openmcf/provider/gcp/gcpmemorystoreinstance/v1/iac/tf/
```

It provisions a `google_memorystore_instance` resource with equivalent field mappings and outputs `discovery_address`, `discovery_port`, `instance_uid`, and `node_size_gb`.

---

## 13. Common Pitfalls

1. **Immutable field changes**: Changing `instance_name`, `location`, `mode`, `authorization_mode`, `transit_encryption_mode`, `kms_key`, `zone_distribution_config`, or `psc_auto_connections` forces instance replacement. Plan for data migration or use blue-green deployment.

2. **PSC connection planning**: PSC connections cannot be added after creation. If you anticipate needing cross-project or multi-VPC access, include all connections in the initial manifest.

3. **Cluster-aware clients**: `CLUSTER` mode requires cluster-aware Valkey/Redis clients. Standard (non-cluster) clients will fail to route commands to the correct shard. Verify your client library supports cluster mode before deploying.

4. **Node type memory assumptions**: Do not assume specific memory values for node types. The actual memory per node is determined by GCP and reported in `node_size_gb` after provisioning. Use stack outputs for capacity calculations.

5. **CMEK key region**: The KMS key must be in the same region as the Memorystore instance. Cross-region keys are not supported.

6. **Deletion protection**: Enable for production. If you need to destroy the instance, first update the manifest with `deletion_protection_enabled: false`, apply the change, then destroy.

7. **SHARED_CORE_NANO performance**: Shared-core nodes have variable CPU performance. Do not use for latency-sensitive production workloads.

8. **AOF ALWAYS mode**: `append_fsync: ALWAYS` provides maximum durability but significantly increases write latency. Use only when every write must be durable; prefer `EVERY_SEC` for most workloads.

---

## 14. Conclusion

### When to Use GcpMemorystoreInstance

- New deployments that need managed in-memory caching or session storage on GCP
- Workloads requiring native sharding for horizontal scale
- Environments where PSC networking is preferred over VPC peering
- Use cases needing AOF persistence or automated backups
- Teams standardizing on Valkey as the engine

### When to Use GcpRedisInstance Instead

- Existing deployments on the legacy Memorystore for Redis API
- Workloads requiring AUTH string authentication (not IAM)
- Environments dependent on VPC peering or Private Service Access
- Applications that need a separate read endpoint (legacy API model)

### References

- [Memorystore Documentation](https://cloud.google.com/memorystore/docs/overview)
- [Memorystore Instance REST API](https://cloud.google.com/memorystore/docs/reference/rest/v1/projects.locations.instances)
- [Valkey Project](https://valkey.io/)
- [Valkey 8.0 Release Notes](https://valkey.io/blog/valkey-8-0-ga/)
- [Private Service Connect Overview](https://cloud.google.com/vpc/docs/private-service-connect)
- [Pulumi GCP Memorystore Instance](https://www.pulumi.com/registry/packages/gcp/api-docs/memorystore/instance/)
- [Terraform google_memorystore_instance](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/memorystore_instance)
- [Memorystore for Redis Documentation (Legacy)](https://cloud.google.com/memorystore/docs/redis)
