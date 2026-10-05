# GCP Memorystore Instance

Deploys a Memorystore instance (Valkey/Redis-compatible) with configurable sharding, node types, Private Service Connect networking, persistence (RDB or AOF), zone distribution, TLS encryption, IAM authentication, automated backups, CMEK encryption, and cross-region replication. Cluster mode, shard topology, zone distribution, and both security modes are immutable after creation — the up-front decisions this page exists to inform.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Memorystore Instance** -- a managed in-memory data store in the specified GCP project and region, configured with the chosen node type, shard count, engine version, and cluster mode
- **Private Service Connect Endpoints** -- one or more PSC endpoints auto-created in the specified consumer VPC networks, providing private connectivity without VPC peering
- **Persistence Configuration** -- created only when `persistenceConfig` is set; configures RDB snapshots at a periodic interval or AOF logging with configurable fsync frequency
- **Zone Distribution** -- multi-zone by default for high availability; configurable as single-zone for lowest latency within one availability zone
- **Maintenance Window** -- created only when `maintenancePolicy` is set; defines a weekly 1-hour window for GCP-managed maintenance operations
- **Automated Backups** -- created only when `automatedBackupConfig` is set; daily backups at the specified hour with configurable retention
- **CMEK Encryption** -- created only when `kmsKey` is provided; encrypts data at rest using a customer-managed Cloud KMS key
- **API enablement** -- `memorystore.googleapis.com` and `networkconnectivity.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the instance will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Memorystore and Network Connectivity APIs itself.
- **A VPC network** for PSC endpoint creation. Applications connect to the instance through PSC endpoints in the consumer VPC. Provide the network self-link directly or reference a GcpVpcNetwork Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **GCP Memorystore Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev Single Shard** preset in the [Presets](#presets) tab to pre-populate a minimal development instance.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMemorystoreInstance
metadata:
  name: app-cache
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  instanceName: app-cache
  location: us-central1
  shardCount: 1
  pscAutoConnections:
    - network:
        value: "projects/acme-prod-12345/global/networks/main-vpc"
      projectId:
        value: "acme-prod-12345"
```

```shell
planton apply -f memorystore-instance.yaml
```

This creates a single-shard instance with GCP-selected defaults for node type and engine version, no persistence, no authentication, and a single PSC endpoint in the specified VPC. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the instance to a project and VPC deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  pscAutoConnections:
    - network:
        valueFrom:
          kind: GcpVpcNetwork
          name: main-vpc
          fieldPath: status.outputs.network_id
      projectId:
        valueFrom:
          kind: GcpProject
          name: production-project
          fieldPath: status.outputs.project_id
  kmsKey:
    valueFrom:
      kind: GcpKmsKey
      name: memorystore-key
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project and VPC first, then provisions the Memorystore instance with PSC connectivity and CMEK encryption.

## Key Configuration

These are the most important decisions when configuring a Memorystore instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Cluster mode and sharding** -- Set `mode` to `CLUSTER` for sharded mode with native cluster protocol (requires cluster-aware client drivers) or `CLUSTER_DISABLED` for standalone mode with a single primary endpoint. `shardCount` controls data distribution across nodes. Both are immutable after creation.

**Node type** -- Choose from `SHARED_CORE_NANO` (dev/test), `STANDARD_SMALL` (small workloads), `HIGHMEM_MEDIUM` (medium production), or `HIGHMEM_XLARGE` (large production). Memory per node is determined by the node type, not an explicit size field. The actual memory is reported in `status.outputs.node_size_gb`.

**Persistence** -- Configure `persistenceConfig.mode` as `RDB` for periodic snapshots (ONE_HOUR through TWENTY_FOUR_HOURS intervals) or `AOF` for append-only file logging (NEVER, EVERY_SEC, or ALWAYS fsync). `DISABLED` keeps data in-memory only. RDB balances durability and performance; AOF provides stronger guarantees at higher I/O cost.

**Authentication and encryption** -- Set `authorizationMode` to `IAM_AUTH` for IAM-based client authentication (immutable after creation). Set `transitEncryptionMode` to `SERVER_AUTHENTICATION` for TLS encryption of client-to-server traffic (immutable after creation). Both default to disabled.

**Zone distribution** -- Set `zoneDistributionConfig.mode` to `MULTI_ZONE` (default) for high availability across zones, or `SINGLE_ZONE` with a specific `zone` for lowest latency. Immutable after creation.

**Deletion protection** -- `deletionProtectionEnabled` defaults to TRUE (GCP's safety posture): when the field is not set, destroying the instance fails until it is explicitly set to false. Both IaC engines send the value explicitly, so destroy behavior is identical regardless of engine.

**Cross-region replication (DR)** -- Configure `crossInstanceReplicationConfig` to make this instance a `PRIMARY` replicating to secondaries in other regions, or a `SECONDARY` continuously replicating from a primary (referenced via the primary's `name` output). A secondary is read-only until promoted; roles are exchanged in place during a planned switchover.

**Data seeding** -- At creation time only, seed the instance's data from RDB files in Cloud Storage (`gcsSource.uris`) or from an existing managed backup (`managedBackupSource.backup`). At most one seed source may be set.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `pscAutoConnections[].network` | `status.outputs.network_id` |
| **GcpProject** | `pscAutoConnections[].projectId` | `status.outputs.project_id` |
| **GcpKmsKey** (optional) | `kmsKey` | `status.outputs.key_id` |
| **GcpMemorystoreInstance** (optional) | `crossInstanceReplicationConfig.primaryInstance.instance` | `status.outputs.name` |
| **GcpMemorystoreInstance** (optional) | `crossInstanceReplicationConfig.secondaryInstances[].instance` | `status.outputs.name` |

The PSC network reference resolves the VPC's `network_id` output (the relative
resource path `projects/{project}/global/networks/{network}`) — the Memorystore
API rejects full `https://` self-link URLs.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `discovery_address` | IP address of the instance's discovery endpoint | Application connection strings, client configuration |
| `discovery_port` | Port of the discovery endpoint (typically 6379) | Application connection strings, client configuration |
| `instance_uid` | Server-generated unique identifier | Correlation, audit trails, monitoring |
| `node_size_gb` | Memory size per node in GB | Capacity planning, monitoring alerts |
| `name` | Full resource path (projects/…/locations/…/instances/…) | Another instance's cross-region replication config (primary/secondary references) |
| `backup_collection` | Backup collection resource path (when automated backups are configured) | Restoring a new instance via `managedBackupSource` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev single shard** -- Single-shard standalone instance on SHARED_CORE_NANO with no persistence and a single PSC endpoint. Optimized for development and testing with minimal cost. Start from the **Dev Single Shard** preset.

**HA production** -- 3-shard cluster with HIGHMEM_MEDIUM nodes, 1 replica per shard, multi-zone distribution, TLS encryption, RDB persistence with 12-hour snapshots, weekly maintenance window, and deletion protection. Suitable for production caching and session management. Start from the **HA Production** preset.

**Enterprise cluster** -- 5-shard cluster with HIGHMEM_XLARGE nodes, 2 replicas per shard, IAM authentication, TLS encryption, CMEK encryption, AOF persistence with per-second fsync, automated daily backups with 35-day retention, and deletion protection. Suitable for regulated workloads requiring maximum durability and encryption key control. Start from the **Enterprise Cluster** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the instance is created and PSC endpoints are placed
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network where PSC endpoints are created for application connectivity
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the CMEK encryption key for data at rest