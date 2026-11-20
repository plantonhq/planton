# GCP Redis Cluster

Memorystore for Redis Cluster -- Google's fully managed, sharded Redis. Data is spread across shards (each at least one node), read replicas serve reads and take over on failure, and the cluster resizes in place without downtime. Connectivity is Private Service Connect: Google places the endpoints for you when you name a consumer network, or publishes service attachments you build your own endpoints against. Choose this over `GcpRedisInstance` (the legacy single-node Memorystore for Redis over VPC peering) when the keyspace or throughput outgrows one node; choose `GcpMemorystoreInstance` when the engine should be Valkey.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Redis Cluster** -- a `redis_cluster` in your region with `shardCount` shards and `replicaCount` replicas per shard on the chosen `nodeType`, its authentication and TLS modes, persistence, backups, maintenance window, zone distribution, optional CMEK, and the Private Service Connect configuration
- **API enablement** -- `redis.googleapis.com` and `networkconnectivity.googleapis.com` on the project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/redis.admin` on the project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Networks

- **A service connection policy** -- when `pscConfigs` names a network, a `GcpServiceConnectionPolicy` with `serviceClass: gcp-memorystore-redis` must already exist on that network in the cluster's region, or creation fails with a connectivity error. This is the kind's registry prerequisite.
- **Consumer subnets** -- the policy's `pscConfig.subnetworks` supply the endpoint addresses.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRedisCluster
metadata:
  name: orders-cache
spec:
  region: us-central1
  shardCount: 3
  replicaCount: 1
  nodeType: REDIS_HIGHMEM_MEDIUM
  pscConfigs:
    - network:
        valueFrom:
          kind: GcpVpcNetwork
          name: prod-vpc
          fieldPath: status.outputs.network_id
  transitEncryptionMode: TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION
  persistenceConfig:
    mode: AOF
    aofConfig:
      appendFsync: EVERYSEC
```

```shell
planton apply -f redis-cluster.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | `string` | Region of the cluster. Immutable. |
| `shardCount` | `int32` | Shards (1-250). Resizes in place. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. Immutable. |
| `clusterName` | `string` | `metadata.name` | Name in GCP. Immutable. |
| `replicaCount` | `int32` | `0` | Replicas per shard (0-5). Resizes in place. |
| `nodeType` | `string` | Google's default (`REDIS_HIGHMEM_MEDIUM`) | `REDIS_SHARED_CORE_NANO`, `REDIS_STANDARD_SMALL`, `REDIS_HIGHCPU_MEDIUM`, `REDIS_HIGHMEM_MEDIUM`, `REDIS_STANDARD_LARGE`, `REDIS_HIGHMEM_XLARGE`, `REDIS_HIGHMEM_2XLARGE`. Removing it later keeps the current type. |
| `redisConfigs` | `map<string,string>` | — | Native Redis parameters from Google's supported subset. |
| `pscConfigs` | `[]object` | — | At most one consumer network for Google-placed endpoints (`GcpVpcNetwork` reference). Empty publishes service attachments only. |
| `authorizationMode` | `string` | `AUTH_MODE_DISABLED` | Or `AUTH_MODE_IAM_AUTH`. Immutable (the provider replaces the cluster on a change); always sent explicitly, so removing it from an IAM cluster plans a replacement. |
| `transitEncryptionMode` | `string` | `TRANSIT_ENCRYPTION_MODE_DISABLED` | Or `TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION`. Immutable (the provider replaces the cluster on a change); always sent explicitly, so removing it from a TLS cluster plans a replacement. |
| `serverCaMode` | `string` | Google's default | Which CA signs the server certificate under TLS: `SERVER_CA_MODE_GOOGLE_MANAGED_PER_INSTANCE_CA`, `SERVER_CA_MODE_GOOGLE_MANAGED_SHARED_CA`, `SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA`. |
| `serverCaPool` | `StringValueOrRef` | — | CA pool for `SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA`: a `GcpPrivateCaPool` reference or its full name. |
| `kmsKey` | `StringValueOrRef` | Google-managed | CMEK for data at rest (`GcpKmsKey` reference). |
| `persistenceConfig` | `object` | in-memory only | `mode` `DISABLED`/`RDB`/`AOF` with `rdbConfig` or `aofConfig`. Removing the block keeps Google's mode: write `mode: DISABLED` to turn persistence off. |
| `zoneDistributionConfig` | `object` | `MULTI_ZONE` | `SINGLE_ZONE` with a `zone`. Fixed at creation: a change is planned in place, never sent, and re-planned on every deploy. |
| `maintenancePolicy` | `object` | Google chooses | `weeklyMaintenanceWindow.day` and `.hour` (UTC, on the hour). Removing it lets Google choose again. |
| `automatedBackupConfig` | `object` | none | Daily backups at `startHour` kept for `retention` (`"86400s"`-`"31536000s"`). Removing it stops automated backups. |
| `crossClusterReplicationConfig` | `object` | none | `clusterRole` `PRIMARY` with `secondaryClusters`, or `SECONDARY` with `primaryCluster` (`GcpRedisCluster` references). Leave a pair with `clusterRole: NONE`; removing the block keeps the role. |
| `gcsSource` / `managedBackupSource` | `object` | — | Seed at creation from RDB files or a managed backup (exactly one). Immutable; keep it for the cluster's life, since removing it replaces the cluster. |
| `labels` | `map<string,string>` | — | Merged beneath Planton's attribution labels. |
| `deletionProtectionEnabled` | `bool` | `true` | A `DELETE` destroy fails until this is set to `false` and applied. `PREVENT` and `ABANDON` are decided before Google is called, so protection never stops an `ABANDON`. Always sent explicitly. |
| `maintenanceVersion` | `string` | Google's rollout | Requests a self-service update of an existing cluster to a version from `available_maintenance_versions`. Leave unset at creation. |
| `aclPolicy` | `string` | built-in default ACL | A Redis Cluster ACL policy's full resource name. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- **`gcsSource`** and **`managedBackupSource`** are mutually exclusive.
- **`serverCaPool`** requires `serverCaMode: SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA`; **`serverCaMode`** requires `transitEncryptionMode: TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION`.
- **`rdbConfig`** only under `mode: RDB`; **`aofConfig`** only under `mode: AOF`; **`zone`** only under `SINGLE_ZONE`.
- A **`SECONDARY`** names its `primaryCluster`; only a **`PRIMARY`** lists `secondaryClusters`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | Full resource path -- the composition key |
| `uid` | `string` | Server-generated identifier |
| `state` | `string` | `CREATING`, `ACTIVE`, `UPDATING`, or `DELETING` |
| `discovery_endpoint_address` / `discovery_endpoint_port` | `string` / `int32` | The Google-placed discovery endpoint (empty without `pscConfigs`) |
| `discovery_service_attachment` / `primary_service_attachment` / `reader_service_attachment` | `string` | The service attachments a consumer's own forwarding rules target (reader empty without replicas) |
| `size_gb` | `int32` | Redis memory across the cluster |
| `shard_count` / `replica_count` | `int32` | The counts in effect |
| `backup_collection` | `string` | Where managed backups live |
| `available_maintenance_versions` | `[]string` | The versions `maintenanceVersion` accepts now (empty when none is newer) |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Immutable after creation:** `projectId`, `clusterName`, `region`, `authorizationMode`, `transitEncryptionMode`, and the seed sources: a change replaces the cluster, and so does removing a seed source. `zoneDistributionConfig` is fixed too, but a change is planned in place and never sent, so it re-plans on every deploy. Everything else updates in place; shard and replica changes resize without downtime.
- **Removing a block is not always turning it off.** `nodeType`, `serverCaMode`, `persistenceConfig`, and `crossClusterReplicationConfig` keep Google's current value when removed; write `mode: DISABLED` or `clusterRole: NONE` instead. `maintenancePolicy` and `automatedBackupConfig` clear when removed.
- **Deletion protection guards only `DELETE`.** `PREVENT` fails a destroy before Google is called; `ABANDON` succeeds on a protected cluster and leaves it running and billing.
- **Creation takes ten to fifteen minutes** (more for many shards). Budget deploy windows for it.
- **Every node bills by the hour** from creation, serving or idle: `shardCount x (1 + replicaCount)` nodes at the `nodeType`'s rate. The smallest shape is one `REDIS_SHARED_CORE_NANO` shard with no replicas.
- **Deploy the service connection policy first** when using `pscConfigs`; without it creation fails. Without `pscConfigs`, register your own endpoints through `GcpRedisClusterEndpointSet`.
- **The maintenance window starts on the hour**; Google exposes no finer setting.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpServiceConnectionPolicy** -- the connectivity automation policy `pscConfigs` depends on
- **GcpRedisClusterEndpointSet** -- registers consumer-built PSC connections on a cluster without `pscConfigs`
- **GcpVpcNetwork** -- the consumer network the endpoints land in
- **GcpKmsKey** -- customer-managed encryption at rest
- **GcpPrivateCaPool** -- your own CA for the server certificate (`serverCaPool`)
- **GcpMemorystoreInstance** -- the Valkey-engine sibling; **GcpRedisInstance** -- the legacy single-node Redis

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
