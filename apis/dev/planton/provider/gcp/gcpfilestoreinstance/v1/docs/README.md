# GcpFilestoreInstance Research Document

## Service Overview

Google Cloud Filestore is a managed NFS file storage service built on ZFS (for newer tiers) and NetApp-derived technology (for legacy tiers). It provides high-performance file storage that mounts natively on Compute Engine VMs, GKE nodes, and any NFS-compatible client within the connected VPC.

Filestore fills the gap between block storage (Persistent Disks, which attach to individual VMs) and object storage (Cloud Storage, which uses an HTTP API). When applications need a POSIX filesystem interface with shared access across multiple clients, Filestore is the standard GCP answer.

## Deployment Landscape

### Terraform

The Terraform resource `google_filestore_instance` manages the instance, its single file share, and its network attachment as a single resource. Provider version `~> 6.0` is required for the full feature set including `performance_config`, `deletion_protection_enabled`, and `protocol` fields.

Key characteristics:
- `file_shares` block: MaxItems 1 (only one share per instance)
- `networks` block: MaxItems 1 (only one network per instance)
- `tier` and `location` are ForceNew (immutable after creation)
- `capacity_gb` can be increased but not decreased
- `nfs_export_options` can be updated after creation

### Pulumi

The Pulumi resource `filestore.Instance` mirrors the Terraform schema. The Go SDK uses `InstanceFileSharesArgs` (singular struct, not array) for the file share and `InstanceNetworkArray` for networks.

### GCP Tiers Explained

| Tier | Storage | Availability | Min Capacity | Performance Tuning | Location Type |
|------|---------|-------------|-------------|-------------------|---------------|
| STANDARD (fka BASIC_HDD) | HDD | Single-zone | 1 TiB | No | Zone |
| PREMIUM (fka BASIC_SSD) | SSD | Single-zone | 2.5 TiB | No | Zone |
| HIGH_SCALE_SSD | SSD | Single-zone | 10 TiB | No | Zone |
| ZONAL | SSD | Single-zone | 1 TiB | Yes (fixed_iops, iops_per_tb) | Zone |
| REGIONAL | SSD | Multi-zone | 1 TiB | Yes | Region |
| ENTERPRISE | SSD | Multi-zone | 1 TiB | Yes | Region |

STANDARD/PREMIUM are the rebranded names for BASIC_HDD/BASIC_SSD. Both names are accepted by the API. ZONAL is the modern replacement for BASIC_SSD and HIGH_SCALE_SSD, offering better performance tuning and lower minimum capacity.

## 80/20 Scoping Rationale

### What We Include

1. **All 8 tiers**: Users need the full range from cost-effective HDD to enterprise HA.
2. **File share configuration**: Name, capacity, NFS export options cover 100% of share config.
3. **Network configuration**: VPC attachment with connect_mode and reserved IP range cover the three standard connectivity patterns.
4. **Protocol selection**: NFS_V3 and NFS_V4_1 are both widely used.
5. **CMEK encryption**: Required for compliance in regulated industries.
6. **Deletion protection**: Safety guard for production instances.
7. **Performance configuration**: Fixed IOPS and per-TB IOPS are the primary tuning knobs.
8. **Description**: Standard metadata field.

### What We Exclude (and Why)

1. **Directory services (LDAP)**: Only valid with NFSv4.1. Requires external LDAP infrastructure we don't model. Niche use case. Defer to v2.
2. **Initial replication**: Cross-instance replication is an advanced DR feature that creates circular dependencies. Same exclusion rationale as AlloyDB SECONDARY clusters.
3. **Source backup restoration**: Operational restore concern, not infrastructure definition. Same exclusion as GcpMemorystoreInstance's managed_backup_source.
4. **Resource Manager tags**: Not GCP labels. Advanced organizational feature excluded by all other components.
5. **IP modes configuration**: MODE_IPV4 is the overwhelming default (99%+). IPv6 for NFS is extremely niche. Hardcoded in IaC modules.
6. **PSC config (endpoint_project)**: Only relevant for shared VPC setups. Niche. Can add in v2 if demanded.

### Design Decisions

**Singular file_share and network_config (not repeated)**: GCP Filestore supports exactly one file share and one network per instance. Making these repeated fields would be dishonest and confuse users. We use singular sub-messages that accurately reflect the API constraint.

**location instead of zone**: The `zone` field is deprecated in both Terraform and Pulumi. The `location` field accepts either a zone or region depending on the tier, providing a single unified field.

**Performance config as optional sub-message**: Most users don't need IOPS tuning (tier defaults are reasonable). Making it optional keeps the common case clean while enabling advanced tuning for ZONAL/REGIONAL/ENTERPRISE users.

**No ip_modes exposure**: MODE_IPV4 is hardcoded in both IaC modules. IPv6 NFS is not a realistic use case for our target audience.

## StringValueOrRef Fields

| Field | Default Kind | Field Path | Purpose |
|-------|-------------|------------|---------|
| `projectId` | GcpProject | `status.outputs.project_id` | GCP project for instance creation |
| `networkConfig.network` | GcpVpc | `status.outputs.network_self_link` | VPC network attachment |
| `kmsKeyName` | GcpKmsKey | `status.outputs.key_id` | CMEK encryption key |

## Stack Outputs

| Output | Source | Downstream Use |
|--------|--------|----------------|
| `instanceId` | Pulumi: `instance.ID()`, TF: `this.id` | Fully qualified path for API references |
| `instanceName` | Pulumi: `instance.Name`, TF: `this.name` | Short name for display and logging |
| `ipAddresses` | Pulumi: `instance.Networks[0].IpAddresses`, TF: `this.networks[0].ip_addresses` | NFS mount point IP addresses |
| `fileShareName` | Spec passthrough | NFS mount path construction |
| `createTime` | Pulumi: `instance.CreateTime`, TF: `this.create_time` | Instance creation timestamp |

## Infra-Chart Composition

Filestore instances are commonly composed with:

- **GcpGkeCluster**: GKE pods mount Filestore via PersistentVolume/PersistentVolumeClaim using the NFS CSI driver
- **GcpComputeInstance**: VMs mount Filestore shares directly via `mount` command
- **GcpVpc + GcpSubnetwork**: Network foundation for Filestore connectivity
- **GcpKmsKey + GcpKmsKeyRing**: CMEK encryption chain

Filestore sits at DAG layer 2 (depends on VPC/network at layer 1, consumed by compute/workloads at layer 3+).
