# GCP AlloyDB Cluster

Deploys an AlloyDB cluster with a bundled primary instance, private networking (Private Service Access VPC peering XOR Private Service Connect), configurable PostgreSQL version, PRIMARY or cross-region SECONDARY topology, automated periodic backups, continuous backup with point-in-time recovery, optional CMEK encryption across cluster data, backups, and continuous backups independently, and a maintenance window. The cluster can also be seeded from a backup or by point-in-time recovery from another cluster's continuous backup stream — a create-time-only choice.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **AlloyDB API enablement** (`alloydb.googleapis.com`) on the target project (never disabled on destroy)
- **AlloyDB Cluster** -- a managed cluster in the specified GCP project and region with private networking, optional display name, annotations, and a subscription tier (STANDARD or TRIAL)
- **Primary Instance** -- a compute instance within the cluster configured with CPU count or explicit machine type, availability type (ZONAL or REGIONAL), database flags, query insights, SSL mode, and optional connector enforcement
- **Network Configuration** -- exactly one connectivity path: Private Service Access through the specified VPC (with optional allocated IP range for pre-planned CIDR assignments) OR Private Service Connect (`pscConfig.pscEnabled`) for endpoint-based multi-VPC access
- **Automated Backup Policy** -- created only when `automatedBackupPolicy` is specified; configures periodic snapshot backups with retention, scheduling, and optional CMEK encryption
- **Continuous Backup** -- when `continuousBackupConfig` is specified, configures WAL streaming for point-in-time recovery with configurable retention window and optional CMEK encryption
- **CMEK Encryption** -- created only when `kmsKeyName` is set on the cluster, backup policy, or continuous backup config; each can use independent KMS keys
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied at cluster and instance level for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the AlloyDB cluster will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **A VPC network** with Private Service Access configured (compose GcpGlobalAddress with VPC_PEERING purpose + GcpServiceNetworkingConnection). Provide the network as the relative resource path `projects/{project}/global/networks/{network}` -- the AlloyDB API rejects full https:// self-link URLs -- or reference a GcpVpcNetwork Infra Component via ValueFromRef (resolves to exactly that path). Not needed when using Private Service Connect instead.
- **Cloud KMS keys** (if using CMEK) -- each key must be in the same region as the cluster. The AlloyDB service account must have `roles/cloudkms.cryptoKeyEncrypterDecrypter` on each key.

## Deploy

### Console

Open the deployment store, find **GCP AlloyDB Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev Basic** preset in the [Presets](#presets) tab to pre-populate a minimal development configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAlloydbCluster
metadata:
  name: app-alloydb
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  clusterName: app-alloydb-prod
  location: us-central1
  network:
    value: "projects/acme-prod-12345/global/networks/main-vpc"
  primaryInstance:
    instanceId: app-alloydb-primary
    cpuCount: 4
    availabilityType: REGIONAL
```

```shell
planton apply -f alloydb-cluster.yaml
```

This creates a regional AlloyDB cluster with a 4-CPU primary instance, private networking via the specified VPC, and GCP default backup policies. No initial user is created. The cluster ships destroy-guarded: `deletionProtection` defaults to TRUE, so a destroy fails until the spec flips it false and that change is applied first. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the cluster to a GCP project, VPC, and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: main-vpc
      fieldPath: status.outputs.network_id
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: alloydb-cmek-key
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project, VPC, and KMS key first, then provisions the AlloyDB cluster with private connectivity and CMEK encryption.

## Key Configuration

These are the most important decisions when configuring an AlloyDB cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Primary instance sizing** -- Set `primaryInstance.cpuCount` (2, 4, 8, 16, 32, 64, 96, 128) to let GCP choose the machine family automatically, or `primaryInstance.machineType` for explicit control (e.g., `c4a-highmem-4-lssd`). These are mutually exclusive. Start with 4 CPUs for moderate production workloads.

**Availability type** -- `primaryInstance.availabilityType` set to `REGIONAL` provides multi-zone deployment with automatic failover (recommended for production). `ZONAL` is a single-zone deployment suitable for development.

**Backup strategy** -- AlloyDB supports both automated periodic backups (`automatedBackupPolicy`) and continuous backup (`continuousBackupConfig`) independently. Automated backups capture periodic snapshots with configurable retention. Continuous backup streams WAL data for point-in-time recovery within a 1-35 day window (default 14 days). Both can be independently encrypted with CMEK.

**Connection security** -- Set `primaryInstance.requireConnectors` to `true` to enforce IAM-based authentication via AlloyDB Auth Proxy or Language Connectors, rejecting direct IP connections. Set `primaryInstance.sslMode` to `ENCRYPTED_ONLY` to require TLS for all connections.

**PostgreSQL version** -- `databaseVersion` selects the PostgreSQL major version (`POSTGRES_14`, `POSTGRES_15`, `POSTGRES_16`). If omitted, GCP selects the latest stable version. Changing it later runs an in-place major upgrade; `skipAwaitMajorVersionUpgrade` lets very large clusters return early instead of blocking the deploy.

**Topology** -- `clusterType` left blank keeps GCP's PRIMARY default. `SECONDARY` builds a cross-region disaster-recovery replica that names its primary via `secondaryConfig.primaryClusterName` (the primary's `cluster_id` output). Switching a secondary's role to PRIMARY is AlloyDB's promotion lever.

**Connectivity** -- exactly one of `network` (Private Service Access) or `pscConfig.pscEnabled: true` (Private Service Connect). PSC clusters record no network; consumers create PSC endpoints after deployment.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (PSA arm) | `network` | `status.outputs.network_id` |
| **GcpKmsKey** (optional) | `kmsKeyName` | `status.outputs.key_id` |
| **GcpKmsKey** (optional) | `automatedBackupPolicy.encryptionKmsKeyName` | `status.outputs.key_id` |
| **GcpKmsKey** (optional) | `continuousBackupConfig.encryptionKmsKeyName` | `status.outputs.key_id` |
| **GcpAlloydbCluster** (PITR restore) | `restoreContinuousBackupSource.cluster` | `status.outputs.cluster_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_id` | Fully qualified cluster resource name (`projects/{p}/locations/{l}/clusters/{c}`) | Read pool instances, backup management, monitoring |
| `cluster_name` | Short cluster name | Display, logging, human-readable references |
| `primary_instance_ip` | Private IP address of the primary instance | Application connection strings (port 5432) |
| `primary_instance_name` | Fully qualified primary instance resource name | AlloyDB Auth Proxy connections, monitoring dashboards |
| `database_version` | Actual PostgreSQL version running | Application compatibility validation |
| `state` | Cluster state (`READY`, `CREATING`, `MAINTENANCE`) | Deployment validation, health checks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Development basic** -- 2-CPU ZONAL primary instance with GCP default backups. Minimal cost for development and testing. Start from the **Dev Basic** preset.

**HA production** -- 4-CPU REGIONAL primary instance with automated 7-day backup retention, initial user, and TLS-only connections. Standard production configuration for application databases. Start from the **HA Production** preset.

**Enterprise encrypted** -- 8-CPU REGIONAL primary instance with CMEK on cluster data, automated backups (30-day retention), and continuous backups (21-day PITR window), each with independent KMS keys. Query insights enabled, connector-only access enforced, and a maintenance window configured. Start from the **Enterprise Encrypted** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the AlloyDB cluster is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network for private connectivity via Private Service Access
- [**GCP Global Address**](/infra-catalog/gcp-global-address) -- reserves the VPC_PEERING range Private Service Access carves the cluster's IPs from
- [**GCP Service Networking Connection**](/infra-catalog/gcp-service-networking-connection) -- establishes the Private Service Access peering the cluster's network requires
- [**GCP AlloyDB Instance**](/infra-catalog/gcp-alloydb-instance) -- adds read pool instances that reference this cluster's `cluster_id`
- [**GCP AlloyDB User**](/infra-catalog/gcp-alloydb-user) -- manages per-application database users on this cluster
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides Cloud KMS keys for cluster, backup, and continuous backup CMEK encryption