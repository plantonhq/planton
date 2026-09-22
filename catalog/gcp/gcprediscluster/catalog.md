# GCP Redis Cluster

Memorystore for Redis Cluster -- Google's fully managed, sharded Redis. Your keyspace is spread across shards, read replicas serve reads and take over on failure, and the cluster grows or shrinks in place without downtime. Clients reach it over Private Service Connect: name a consumer network and Google places the endpoints, or build your own endpoints against the service attachments the cluster publishes. It is the cache and session store for a service that has outgrown one Redis node, and the data structure server for leaderboards, rate limiters, and queues at scale.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Redis Cluster** -- a `redis.Cluster` in the chosen region with the declared shards and replicas on the chosen node type, authentication and TLS modes, persistence, backups, maintenance window, zone distribution, optional customer-managed encryption, and Private Service Connect configuration
- **API enablement** -- the Memorystore for Redis and Network Connectivity APIs on the project, never disabled on destroy

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/redis.admin` on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Networks

- **A service connection policy** -- a `GcpServiceConnectionPolicy` with `serviceClass: gcp-memorystore-redis` on the consumer network in the cluster's region, deployed before the cluster when `pscConfigs` is set. Its subnets supply the endpoint addresses.

## Deploy

### Console

Open the deployment store, find **GCP Redis Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Development Cache** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRedisCluster
metadata:
  name: orders-cache
  org: acme-corp
  env: prod
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
  automatedBackupConfig:
    startHour: 2
    retention: "3024000s"
```

```shell
planton apply -f redis-cluster.yaml
```

This creates a three-shard cluster with one replica per shard, TLS, append-only persistence, and daily backups kept for 35 days, reachable from the production VPC. A Stack Job tracks the provisioning in real time; creation takes ten to fifteen minutes.

### InfraChart

When deploying as part of a multi-resource environment, wire `pscConfigs[].network` to the `GcpVpcNetwork` deployed in the same InfraPipeline and place the `GcpServiceConnectionPolicy` before the cluster; downstream services read the `discovery_endpoint_address` output.

## Key Configuration

These are the most important decisions when configuring a cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Size** -- `shardCount` splits the keyspace; `replicaCount` (0-5) adds per-shard replicas for reads and failover; `nodeType` fixes each node's memory and hourly rate. All three resize in place.

**Connectivity** -- `pscConfigs` names one consumer network and Google places the endpoints (a `GcpServiceConnectionPolicy` for `gcp-memorystore-redis` must exist there first). Leave it empty to publish service attachments only and register hand-built endpoints through `GcpRedisClusterEndpointSet` -- the multi-VPC pattern.

**Security** -- `authorizationMode: AUTH_MODE_IAM_AUTH` puts the data plane under IAM; `transitEncryptionMode: TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION` turns on TLS; `kmsKey` brings your own key for data at rest. The first two are immutable.

**Durability** -- `persistenceConfig` (RDB snapshots or AOF log) makes data survive restarts; `automatedBackupConfig` takes daily backups with a retention you choose; `crossClusterReplicationConfig` replicates to another region.

**Destroy semantics** -- `deletionProtectionEnabled` (default `true`) blocks destroys until flipped; `deletionPolicy: PREVENT` is a second guard.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `pscConfigs[].network` | `status.outputs.network_id` |
| **GcpKmsKey** | `kmsKey` | `status.outputs.key_id` |
| **GcpRedisCluster** | `crossClusterReplicationConfig.primaryCluster.cluster`, `secondaryClusters[].cluster` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | Full resource path | A `SECONDARY`'s `primaryCluster`; a `GcpRedisClusterEndpointSet`'s `cluster` |
| `discovery_endpoint_address` / `discovery_endpoint_port` | The Google-placed discovery endpoint | The `REDIS_HOST` a service connects to |
| `discovery_service_attachment` / `primary_service_attachment` / `reader_service_attachment` | Service attachments consumer forwarding rules target | A consumer `GcpGlobalForwardingRule`'s `target` |
| `state` | Lifecycle state | Readiness checks |
| `size_gb`, `shard_count`, `replica_count` | The shape in effect | Capacity dashboards |
| `backup_collection` | Where managed backups live | `managedBackupSource` for a new cluster |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Development cache** -- One nano shard, no replicas, deletion protection off. Start from the **Development Cache** preset.

**Production cache** -- Multi-zone shards with replicas, TLS, IAM auth, AOF persistence, daily backups, a maintenance window. Start from the **Production Cache** preset.

**Cross-region secondary** -- A read-only replica cluster in another region for disaster recovery. Start from the **Cross-Region Secondary** preset.

## Works With

- [**GCP Service Connection Policy**](/cloud-catalog/gcp-service-connection-policy) -- the automation policy Google-placed endpoints need
- [**GCP Redis Cluster Endpoint Set**](/cloud-catalog/gcp-redis-cluster-endpoint-set) -- registers consumer-built PSC connections
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the consumer network
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption at rest
- [**GCP Cloud Run Worker Pool**](/cloud-catalog/gcp-cloud-run-worker-pool) -- a typical client, reaching the cluster over direct VPC egress
