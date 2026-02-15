---
title: "Memorystore Instance"
description: "Provision and manage Google Cloud Memorystore instances"
icon: "database"
order: 100
componentName: "gcpmemorystoreinstance"
---

# GCP Memorystore Instance

Deploys a Google Cloud Memorystore instance using the new-generation API with the Valkey engine. This component provisions a fully managed, in-memory data store with native sharding, Private Service Connect (PSC) networking, configurable persistence, automated backups, and multi-zone distribution. It supports both standalone (CLUSTER_DISABLED) and sharded cluster (CLUSTER) topologies, making it suitable for everything from development caching to enterprise-scale real-time analytics.

## What Gets Created

When you deploy a GcpMemorystoreInstance resource, OpenMCF provisions:

- **Memorystore Instance** — a `google-native:memorystore/v1:Instance` resource with the specified shard count, node type, and engine version (Valkey)
- **PSC Auto-Connections** — Private Service Connect endpoints auto-created in the consumer VPC, providing private connectivity without VPC peering
- **Persistence Configuration** — RDB snapshots or AOF logging when configured, ensuring data survives restarts and failovers
- **Automated Backups** — daily backups with configurable retention when `automatedBackupConfig` is provided
- **Zone Distribution** — MULTI_ZONE or SINGLE_ZONE placement of nodes across availability zones
- **Encryption** — TLS transit encryption and/or CMEK encryption at rest when configured

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **An existing GCP project** — referenced via `projectId`
- **Memorystore API enabled** (`memorystore.googleapis.com`) on the target project
- **A VPC network** — required for PSC auto-connections; the instance is only reachable via PSC endpoints
- **IAM permissions** — `roles/memorystore.admin` or equivalent on the target project
- **A KMS key** (optional) — required only when configuring CMEK encryption at rest

## Quick Start

Create a file `memorystore.yaml`:

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
  projectId:
    value: my-gcp-project-123
  instanceName: my-cache
  location: us-central1
  shardCount: 1
  mode: CLUSTER_DISABLED
  nodeType: SHARED_CORE_NANO
  pscAutoConnections:
    - network:
        value: "projects/my-gcp-project-123/global/networks/dev-vpc"
      projectId:
        value: "my-gcp-project-123"
```

Deploy:

```shell
openmcf apply -f memorystore.yaml
```

This creates a minimal standalone Memorystore instance with a single shard, the smallest node type, and a PSC endpoint in the specified VPC.

## Key Features

- **Native sharding** — distribute data across multiple shards for horizontal scaling; each shard handles a portion of the keyspace
- **Predefined node types** — choose from SHARED_CORE_NANO, STANDARD_SMALL, HIGHMEM_MEDIUM, and HIGHMEM_XLARGE based on workload needs
- **Private Service Connect** — instances are reachable only via PSC endpoints; no VPC peering required
- **Dual persistence modes** — RDB snapshots for periodic point-in-time recovery or AOF for per-write durability
- **Automated backups** — daily backups with configurable retention (1–365 days)
- **IAM authentication** — authenticate clients using GCP IAM credentials instead of shared secrets
- **TLS encryption** — SERVER_AUTHENTICATION mode for encrypted client-to-server traffic
- **CMEK support** — encrypt data at rest using customer-managed Cloud KMS keys
- **Multi-zone distribution** — spread nodes across availability zones for resilience
- **Cluster and standalone modes** — CLUSTER for sharded topology with native protocol, CLUSTER_DISABLED for single-endpoint simplicity
- **Valkey engine** — Redis-compatible protocol with configurable engine version and tuning parameters

## Configuration Highlights

| Feature | Field | Values / Notes |
|---------|-------|----------------|
| Topology | `mode` | `CLUSTER` (sharded, cluster-aware clients) or `CLUSTER_DISABLED` (standalone) |
| Scaling | `shardCount` | 1+ shards; each shard adds capacity and throughput |
| Node size | `nodeType` | `SHARED_CORE_NANO`, `STANDARD_SMALL`, `HIGHMEM_MEDIUM`, `HIGHMEM_XLARGE` |
| Read replicas | `replicaCount` | 0–5 replicas per shard for read scaling and failover |
| Networking | `pscAutoConnections` | PSC endpoints in consumer VPCs; supports cross-project access |
| Authentication | `authorizationMode` | `AUTH_DISABLED` (default) or `IAM_AUTH` |
| Transit encryption | `transitEncryptionMode` | `TRANSIT_ENCRYPTION_DISABLED` or `SERVER_AUTHENTICATION` |
| Encryption at rest | `kmsKey` | Cloud KMS key resource name for CMEK |
| Persistence | `persistenceConfig.mode` | `DISABLED`, `RDB`, or `AOF` |
| RDB snapshots | `persistenceConfig.rdbConfig.rdbSnapshotPeriod` | `ONE_HOUR`, `SIX_HOURS`, `TWELVE_HOURS`, `TWENTY_FOUR_HOURS` |
| AOF durability | `persistenceConfig.aofConfig.appendFsync` | `NEVER`, `EVERY_SEC`, `ALWAYS` |
| Zone placement | `zoneDistributionConfig.mode` | `MULTI_ZONE` or `SINGLE_ZONE` |
| Maintenance | `maintenancePolicy.weeklyMaintenanceWindow` | Day of week + hour (UTC) |
| Backups | `automatedBackupConfig` | Start hour + retention duration in seconds |
| Engine tuning | `engineConfigs` | Key-value map (e.g., `maxmemory-policy: volatile-lru`) |
| Deletion guard | `deletionProtectionEnabled` | `true` to prevent accidental deletion |

## Presets

OpenMCF provides three ready-to-use presets for common scenarios:

- **01-dev-single-shard** — Minimal standalone instance with SHARED_CORE_NANO nodes, no persistence or encryption. Ideal for development and testing.
- **02-ha-production** — CLUSTER mode with 3 shards, 1 replica, HIGHMEM_MEDIUM nodes, TLS, RDB persistence, multi-zone distribution, and deletion protection. Suitable for production workloads.
- **03-enterprise-cluster** — CLUSTER mode with 5 shards, 2 replicas, HIGHMEM_XLARGE nodes, IAM auth, TLS, CMEK, AOF persistence, automated backups (35-day retention), and deletion protection. Designed for enterprise and mission-critical environments.

## GcpRedisInstance vs GcpMemorystoreInstance

Google Cloud offers two Memorystore APIs. OpenMCF models each as a separate component:

| Aspect | GcpRedisInstance | GcpMemorystoreInstance |
|--------|-----------------|----------------------|
| **API generation** | Legacy Memorystore for Redis API | New-generation Memorystore API |
| **Engine** | Redis | Valkey (Redis-compatible) |
| **Networking** | VPC peering (`authorizedNetwork`) | Private Service Connect (PSC) |
| **Sharding** | Not supported; single primary | Native sharding via `shardCount` |
| **Node sizing** | `memorySizeGb` (explicit memory) | `nodeType` (predefined CPU + memory tiers) |
| **Cluster mode** | Not available | `CLUSTER` or `CLUSTER_DISABLED` |
| **Persistence** | RDB only | RDB or AOF |
| **Automated backups** | Not available | Daily backups with configurable retention |
| **Authentication** | Redis AUTH string | IAM_AUTH or AUTH_DISABLED |
| **Read replicas** | Separate `readReplicasMode` + `replicaCount` | `replicaCount` per shard (built into cluster topology) |

**When to use GcpRedisInstance:** You have existing workloads on the legacy API, need VPC peering connectivity, or require the traditional Redis AUTH model.

**When to use GcpMemorystoreInstance:** New deployments that benefit from native sharding, PSC networking, AOF persistence, automated backups, or IAM-based authentication. This is Google's recommended path for new Memorystore instances.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `discoveryAddress` | `string` | IP address of the instance's discovery endpoint. Clients connect here for cluster topology discovery and command routing. |
| `discoveryPort` | `int32` | Port of the discovery endpoint (typically 6379). |
| `instanceUid` | `string` | Server-generated unique identifier for the instance. |
| `nodeSizeGb` | `double` | Memory size per node in GB, determined by the chosen `nodeType`. |

## Related Components

- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project where the instance is created
- [GcpVpc](/docs/catalog/gcp/vpc) — provides the VPC network for PSC auto-connections
- [GcpKmsKeyRing](/docs/catalog/gcp/kms-key-ring) — contains the key ring for CMEK encryption keys
- [GcpRedisInstance](/docs/catalog/gcp/gcpredisinstance-research-and-design-documentation) — legacy Memorystore for Redis API; use for existing VPC-peering-based workloads
