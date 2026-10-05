# GCP Filestore Instance

Deploys a fully managed NFS file server on Google Cloud Filestore with configurable tiers (BASIC_SSD through ENTERPRISE), VPC network connectivity, CMEK encryption, NFS export controls, and IOPS performance tuning. Each instance provides a single file share mountable via NFSv3 or NFSv4.1, and can be seeded from a backup or paired with a peer instance for cross-instance replication at create time.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Filestore Instance** -- a managed NFS file server in the specified project and location, configured with the chosen tier, capacity, and protocol
- **File Share** -- a single NFS export with the specified name and capacity, mountable at `<ip>:/<share_name>`
- **VPC Network Attachment** -- connects the instance to the specified VPC network using DIRECT_PEERING, PRIVATE_SERVICE_ACCESS, or PRIVATE_SERVICE_CONNECT mode
- **NFS Export Options** -- created only when export options are provided; control client access by IP range, read/write mode, and root squash settings
- **Performance Configuration** -- created only when `performanceConfig` is set; configures fixed IOPS or per-TB IOPS scaling (ZONAL, REGIONAL, ENTERPRISE tiers only)
- **CMEK Encryption** -- applied only when `kmsKeyName` is provided; encrypts data at rest with a customer-managed Cloud KMS key
- **Filestore API enablement** -- `file.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Filestore instance will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **A VPC network** for the instance to connect to. Provide the network NAME directly or reference a GcpVpcNetwork Infra Component via ValueFromRef — the Filestore API rejects self-link URLs for same-project networks. The network configuration is immutable after creation.
- **Private Services Access** (only for PRIVATE_SERVICE_ACCESS connect mode) -- the VPC must have a private services connection configured.

## Deploy

### Console

Open the deployment store, find **GCP Filestore Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev Basic** preset in the [Presets](#presets) tab to pre-populate a development-ready configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFilestoreInstance
metadata:
  name: shared-nfs
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  instanceName: shared-nfs
  location: us-central1-a
  tier: BASIC_SSD
  fileShare:
    name: data
    capacityGb: 2560
  networkConfig:
    network:
      value: "projects/acme-prod-12345/global/networks/main-vpc"
```

```shell
planton apply -f filestore.yaml
```

This creates a BASIC_SSD Filestore instance with a 2.5 TiB file share named `data`, connected to the specified VPC via DIRECT_PEERING. No CMEK, deletion protection, or performance tuning is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Filestore instance to infrastructure deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  networkConfig:
    network:
      valueFrom:
        kind: GcpVpcNetwork
        name: production-vpc
        fieldPath: status.outputs.network_name
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: filestore-key
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project, VPC, and KMS key first, then provisions the Filestore instance with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Filestore instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Tier selection** -- Determines performance, availability, and pricing. BASIC_SSD (2.5 TiB minimum) for development and moderate workloads. ZONAL for modern SSD with IOPS tuning. ENTERPRISE for regional HA with automatic failover. The tier is immutable after creation.

**File share capacity** -- Set `fileShare.capacityGb` to meet your storage needs. Minimums depend on tier: 1024 GiB for STANDARD/ZONAL/ENTERPRISE, 2560 GiB for BASIC_SSD, 10240 GiB for HIGH_SCALE_SSD. Capacity can be increased after creation but not decreased.

**Network connectivity** -- Choose the `connectMode` in `networkConfig`: DIRECT_PEERING (default, simplest), PRIVATE_SERVICE_ACCESS (for Shared VPC or enterprise networks), or PRIVATE_SERVICE_CONNECT. The network configuration is immutable after creation.

**NFS export options** -- Control client access to the file share via `fileShare.nfsExportOptions`. Configure `ipRanges` to restrict which CIDR blocks can mount, `accessMode` for READ_WRITE or READ_ONLY, and `squashMode` for root user handling (ROOT_SQUASH maps root to anonymous for security).

**Performance tuning** -- Available on ZONAL, REGIONAL, and ENTERPRISE tiers. Choose between `performanceConfig.fixedIops` (constant IOPS regardless of capacity) and `performanceConfig.iopsPerTb` (IOPS scales with capacity). Omit for tier-default performance.

**Restore and replication** -- Seed the share from an existing backup at creation via `fileShare.sourceBackup` (the share's capacity must cover the backup's source capacity). For cross-instance DR, `initialReplication` pairs this instance with a peer at creation — usually as the STANDBY replica of an existing ACTIVE instance (referenced via its `instance_id` output). Both are create-time only, and backups cannot be taken from a standby replica.

**Labels and tags** -- `labels` are mutable organizational metadata merged beneath Planton's attribution labels. `tags` bind Resource Manager tag values (`tagKeys/{id}` → `tagValues/{id}`) for org-policy and IAM conditions — create-time only.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `networkConfig.network` | `status.outputs.network_name` |
| **GcpKmsKey** (optional) | `kmsKeyName` | `status.outputs.key_id` |
| **GcpFilestoreInstance** (optional) | `initialReplication.peerInstances[]` | `status.outputs.instance_id` |

The network reference resolves the VPC's plain NAME output — the Filestore API
rejects self-link URLs for same-project networks.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `instance_id` | Fully qualified resource ID (`projects/{p}/locations/{l}/instances/{i}`) | Another instance's `initialReplication.peerInstances`, monitoring, audit logs |
| `instance_name` | Short name of the Filestore instance | Application configuration |
| `ip_addresses` | IP addresses assigned on the connected VPC network | NFS mount commands (`mount <ip>:/<share> /mnt`) |
| `file_share_name` | Name of the file share (NFS export path) | NFS mount path construction |
| `reserved_ip_range` | The /29 CIDR range in use (set or auto-selected by GCP) | Planning non-overlapping address space for the next instance |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev basic** -- A BASIC_SSD instance with 2.5 TiB capacity, default VPC, and DIRECT_PEERING. Minimal configuration for development, testing, and CI/CD pipelines. Start from the **Dev Basic** preset.

**Production enterprise** -- An ENTERPRISE-tier instance with regional HA, PRIVATE_SERVICE_ACCESS networking, ROOT_SQUASH security, IP range restrictions, and deletion protection. For production workloads that cannot tolerate zone-level outages. Start from the **Production Enterprise** preset.

**High performance zonal** -- A ZONAL-tier instance with 2.5 TiB capacity, 20,000 fixed IOPS, CMEK encryption, and PRIVATE_SERVICE_ACCESS networking. Optimized for media rendering, EDA, genomics, and ML training data staging. Start from the **High Performance Zonal** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the instance is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network for instance connectivity
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the encryption key for customer-managed encryption at rest