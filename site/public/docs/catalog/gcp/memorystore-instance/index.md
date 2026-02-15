---
title: "Memorystore Instance"
description: "Memorystore Instance deployment documentation"
icon: "package"
order: 100
componentName: "gcpmemorystoreinstance"
---

# GCP Memorystore Instance

Deploys a fully managed GCP Memorystore instance supporting the Valkey protocol (Redis-compatible) with configurable sharding, Private Service Connect (PSC) networking, persistence modes (RDB/AOF), automated backups, and customer-managed encryption keys. The instance connects to consumer VPCs exclusively through PSC endpoints.

## What Gets Created

When you deploy a GcpMemorystoreInstance resource, OpenMCF provisions:

- **Memorystore Instance** — a `memorystore.Instance` resource with the specified shard count, node type, and engine version
- **PSC Auto-Created Endpoints** — one per entry in `pscAutoConnections`, each creating a Private Service Connect endpoint in the specified consumer VPC network for application connectivity
- **Persistence Layer** — when configured, either RDB snapshots at a configurable interval or AOF write logging for data durability across restarts
- **Automated Backup Schedule** — when configured, daily backups at a specified hour with configurable retention duration
- **Maintenance Window** — when configured, a weekly maintenance window controlling when GCP performs scheduled maintenance

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** where the Memorystore instance will be created
- **A VPC network** in the consumer project for PSC endpoint creation (format: `projects/{project_id}/global/networks/{network_id}`)
- **A Cloud KMS key** if using customer-managed encryption at rest (CMEK)

## Quick Start

Create a file `memorystore-instance.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpMemorystoreInstance
metadata:
  name: my-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpMemorystoreInstance.my-cache
spec:
  projectId: my-gcp-project
  instanceName: my-cache
  location: us-central1
  shardCount: 1
  pscAutoConnections:
    - network: projects/my-gcp-project/global/networks/default
      projectId: my-gcp-project
```

Deploy:

```shell
openmcf apply -f memorystore-instance.yaml
```

This creates a single-shard Memorystore instance in `us-central1` with a PSC endpoint in the default VPC network.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` | GCP project where the instance is created. Can reference GcpProject via `valueFrom`. | Required |
| `instanceName` | `string` | Name of the Memorystore instance. Becomes the GCP resource name. Immutable after creation. | 4-63 chars, lowercase letters/numbers/hyphens, must start with letter and end with letter or number |
| `location` | `string` | GCP region for deployment (e.g., `us-central1`). Immutable after creation. | Required |
| `shardCount` | `int` | Number of shards. Each shard handles a portion of the keyspace. | Minimum 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mode` | `string` | — | Instance topology. `CLUSTER`: sharded mode requiring cluster-aware clients. `CLUSTER_DISABLED`: standalone single-primary mode. Immutable after creation. |
| `nodeType` | `string` | GCP default | CPU and memory per node. `SHARED_CORE_NANO`, `STANDARD_SMALL`, `HIGHMEM_MEDIUM`, or `HIGHMEM_XLARGE`. |
| `engineVersion` | `string` | Latest | Valkey engine version (e.g., `VALKEY_8_0`, `VALKEY_7_2`). |
| `engineConfigs` | `map<string, string>` | `{}` | Engine configuration parameters as key-value pairs (e.g., `maxmemory-policy`, `notify-keyspace-events`). |
| `replicaCount` | `int` | `0` | Read replicas per shard (0-5). Replicas provide read scaling and automatic failover. |
| `pscAutoConnections` | `object[]` | `[]` | PSC endpoints for VPC connectivity. Each entry creates a PSC endpoint in the specified consumer VPC. Immutable after creation. |
| `pscAutoConnections[].network` | `string` | — | Consumer VPC network. Format: `projects/{project}/global/networks/{network}`. Can reference GcpVpc via `valueFrom`. |
| `pscAutoConnections[].projectId` | `string` | — | Consumer project ID for the PSC endpoint. Can reference GcpProject via `valueFrom`. |
| `authorizationMode` | `string` | `AUTH_DISABLED` | Authentication mode. `AUTH_DISABLED`: no auth. `IAM_AUTH`: GCP IAM credentials required. Immutable after creation. |
| `transitEncryptionMode` | `string` | `TRANSIT_ENCRYPTION_DISABLED` | TLS encryption mode. `TRANSIT_ENCRYPTION_DISABLED`: no encryption. `SERVER_AUTHENTICATION`: clients verify server via TLS. Immutable after creation. |
| `kmsKey` | `string` | Google-managed | Cloud KMS key for customer-managed encryption at rest. Format: `projects/{p}/locations/{l}/keyRings/{kr}/cryptoKeys/{k}`. Can reference GcpKmsKey via `valueFrom`. Immutable after creation. |
| `persistenceConfig.mode` | `string` | — | Persistence mode. `DISABLED`: in-memory only. `RDB`: periodic snapshots. `AOF`: append-only file logging. |
| `persistenceConfig.rdbConfig.rdbSnapshotPeriod` | `string` | — | Snapshot frequency. `ONE_HOUR`, `SIX_HOURS`, `TWELVE_HOURS`, or `TWENTY_FOUR_HOURS`. Required when mode is `RDB`. |
| `persistenceConfig.rdbConfig.rdbSnapshotStartTime` | `string` | GCP default | RFC3339 timestamp for the first snapshot. |
| `persistenceConfig.aofConfig.appendFsync` | `string` | — | AOF flush frequency. `NEVER`, `EVERY_SEC`, or `ALWAYS`. Required when mode is `AOF`. |
| `zoneDistributionConfig.mode` | `string` | — | Zone distribution. `MULTI_ZONE`: nodes across zones for HA. `SINGLE_ZONE`: all nodes in one zone. Immutable after creation. |
| `zoneDistributionConfig.zone` | `string` | — | Zone for `SINGLE_ZONE` mode (e.g., `us-central1-a`). Required when mode is `SINGLE_ZONE`. |
| `maintenancePolicy.weeklyMaintenanceWindow.day` | `string` | — | Day of week for maintenance (`MONDAY` through `SUNDAY`). |
| `maintenancePolicy.weeklyMaintenanceWindow.hour` | `int` | — | Hour of day (0-23, UTC) when the 1-hour maintenance window starts. |
| `automatedBackupConfig.startHour` | `int` | — | Hour of day (0-23, UTC) when the daily backup starts. |
| `automatedBackupConfig.retention` | `string` | — | Backup retention duration in seconds (e.g., `3024000s` for 35 days). Min: `86400s` (1 day). Max: `31536000s` (365 days). |
| `deletionProtectionEnabled` | `bool` | `false` | Prevents deletion of the instance until this flag is disabled. |

## Examples

### Standalone Instance with RDB Persistence

A single-shard standalone instance with periodic RDB snapshots and deletion protection:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpMemorystoreInstance
metadata:
  name: session-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpMemorystoreInstance.session-cache
spec:
  projectId: my-gcp-project
  instanceName: session-cache
  location: us-central1
  shardCount: 1
  mode: CLUSTER_DISABLED
  nodeType: STANDARD_SMALL
  replicaCount: 1
  deletionProtectionEnabled: true
  pscAutoConnections:
    - network: projects/my-gcp-project/global/networks/prod-vpc
      projectId: my-gcp-project
  persistenceConfig:
    mode: RDB
    rdbConfig:
      rdbSnapshotPeriod: SIX_HOURS
```

### Sharded Cluster with AOF and Automated Backups

A multi-shard cluster with AOF persistence, automated daily backups, and a maintenance window:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpMemorystoreInstance
metadata:
  name: realtime-store
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpMemorystoreInstance.realtime-store
spec:
  projectId: my-gcp-project
  instanceName: realtime-store
  location: us-east1
  shardCount: 3
  mode: CLUSTER
  nodeType: HIGHMEM_MEDIUM
  engineVersion: VALKEY_8_0
  replicaCount: 2
  deletionProtectionEnabled: true
  pscAutoConnections:
    - network: projects/my-gcp-project/global/networks/prod-vpc
      projectId: my-gcp-project
  persistenceConfig:
    mode: AOF
    aofConfig:
      appendFsync: EVERY_SEC
  automatedBackupConfig:
    startHour: 3
    retention: "3024000s"
  maintenancePolicy:
    weeklyMaintenanceWindow:
      day: SUNDAY
      hour: 5
  zoneDistributionConfig:
    mode: MULTI_ZONE
```

### Encrypted Instance with IAM Auth and Foreign Key References

A production instance using CMEK encryption, IAM authentication, TLS, and foreign key references to other OpenMCF-managed resources:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpMemorystoreInstance
metadata:
  name: secure-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpMemorystoreInstance.secure-cache
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  instanceName: secure-cache
  location: us-central1
  shardCount: 2
  mode: CLUSTER
  nodeType: HIGHMEM_XLARGE
  replicaCount: 1
  authorizationMode: IAM_AUTH
  transitEncryptionMode: SERVER_AUTHENTICATION
  deletionProtectionEnabled: true
  kmsKey:
    valueFrom:
      kind: GcpKmsKey
      name: my-kms-key
      field: status.outputs.key_id
  pscAutoConnections:
    - network:
        valueFrom:
          kind: GcpVpc
          name: prod-vpc
          field: status.outputs.network_self_link
      projectId:
        valueFrom:
          kind: GcpProject
          name: my-project
          field: status.outputs.project_id
  persistenceConfig:
    mode: RDB
    rdbConfig:
      rdbSnapshotPeriod: TWELVE_HOURS
  engineConfigs:
    maxmemory-policy: allkeys-lru
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `discovery_address` | `string` | IP address of the instance's discovery endpoint. Clients connect to this address for cluster topology discovery and command routing. |
| `discovery_port` | `int` | Port of the instance's discovery endpoint (typically 6379). |
| `instance_uid` | `string` | Server-generated unique identifier for the instance. Stable across updates. |
| `node_size_gb` | `double` | Memory size per node in GB, determined by the chosen `nodeType`. Useful for capacity planning. |

## Related Components

- [GcpVpc](/docs/catalog/gcp/vpc) — provides the VPC network for PSC endpoint creation
- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project for instance deployment
- [GcpKmsKey](/docs/catalog/gcp/kms-key) — provides the Cloud KMS key for customer-managed encryption at rest
- [GcpRedisInstance](/docs/catalog/gcp/redis-instance) — legacy Memorystore for Redis API using VPC peering instead of PSC
