# AwsFsxOpenzfsFileSystem

Deploys an Amazon FSx for OpenZFS file system — a fully managed, high-performance NFS file system built on the OpenZFS file system with support for snapshots, cloning, data compression (ZSTD/LZ4), per-user/group quotas, and Multi-AZ high availability.

## When to Use

Use FSx for OpenZFS when you need:

- **General-purpose NFS storage** for web serving, content management, analytics, DevOps, or application data
- **Drop-in replacement** for self-managed NFS servers (ZFS on Linux, FreeNAS/TrueNAS)
- **Snapshots and cloning** for rapid testing, CI/CD, or disaster recovery
- **Data compression** to reduce storage costs (ZSTD achieves 50-70% reduction for compressible data)
- **Multi-AZ high availability** with automatic failover for mission-critical workloads
- **Per-user/group quotas** to control storage consumption in shared environments

### When NOT to Use

- **HPC / ML training workloads** → Use [AwsFsxLustreFileSystem](../awsfsxlustrefilesystem/) (Lustre protocol, S3 integration)
- **Windows SMB file shares** → Use AwsFsxWindowsFileSystem (Active Directory, SMB protocol)
- **Enterprise NAS/SAN** → Use AwsFsxOntapFileSystem (NetApp ONTAP, iSCSI + NFS + SMB)
- **Simple EFS workloads** → Use [AwsElasticFileSystem](../awselasticfilesystem/) (simpler, serverless, lower throughput)

## Prerequisites

- **VPC with subnets** in the target region. SINGLE_AZ deployments need 1 subnet; MULTI_AZ needs 2 subnets in different AZs.
- **Security group** allowing NFS traffic: TCP 111 (portmapper), TCP 2049 (NFS), TCP 20001-20003 (NFS mount).
- **KMS key** (optional) for customer-managed encryption at rest.

## Deployment Types

| Type | AZs | Throughput Range | Use Case |
|------|-----|------------------|----------|
| SINGLE_AZ_1 | 1 | 64 – 4,096 MB/s | Development, cost-sensitive |
| SINGLE_AZ_2 | 1 | 160 – 10,240 MB/s | Production, recommended |
| MULTI_AZ_1 | 2 | 160 – 10,240 MB/s | HA, mission-critical |

## Spec Fields

### File System Core

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `deployment_type` | string | No | SINGLE_AZ_2 | SINGLE_AZ_1, SINGLE_AZ_2, SINGLE_AZ_HA_1, SINGLE_AZ_HA_2, or MULTI_AZ_1 |
| `storage_capacity_gib` | int32 | Conditional | — | 64–524288 GiB. Required for SSD (unless restoring from `backup_id`); forbidden for INTELLIGENT_TIERING |
| `storage_type` | string | No | SSD | SSD or INTELLIGENT_TIERING (MULTI_AZ_1 only, elastic capacity) |
| `throughput_capacity` | int32 | Yes | — | MB/s (see deployment type table). Scales in place |
| `read_cache_configuration` | message | Conditional | — | Provisioned SSD read cache. Required for INTELLIGENT_TIERING, invalid otherwise |
| `backup_id` | string | No | — | FSx backup to restore from. ForceNew; excludes `storage_capacity_gib` |

### Networking

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `subnet_ids` | repeated StringValueOrRef | Yes | Exactly 1 for single-AZ types (incl. HA), exactly 2 for MULTI_AZ_1 |
| `security_group_ids` | repeated StringValueOrRef | No | Up to 50 |
| `preferred_subnet_id` | StringValueOrRef | MULTI_AZ_1 (required) | Active file server subnet; invalid for single-AZ types |
| `endpoint_ip_address_range` | string | No | CIDR for MULTI_AZ_1 floating endpoints |
| `route_table_ids` | repeated StringValueOrRef | No | MULTI_AZ_1 route management (up to 50; AWS uses the VPC default table when omitted) |

### Encryption

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `kms_key_id` | StringValueOrRef | No | Customer-managed KMS key ARN |

### Disk IOPS

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `disk_iops_configuration.mode` | string | AUTOMATIC | AUTOMATIC or USER_PROVISIONED |
| `disk_iops_configuration.iops` | int32 | — | Total IOPS (USER_PROVISIONED only) |

### Root Volume Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `root_volume_configuration.data_compression_type` | string | NONE | NONE, ZSTD, or LZ4 |
| `root_volume_configuration.nfs_exports` | message | — | NFS client configurations |
| `root_volume_configuration.read_only` | bool | false | Read-only root volume |
| `root_volume_configuration.record_size_kib` | int32 | 128 | 4, 8, 16, 32, 64, 128, 256, 512, 1024 |
| `root_volume_configuration.user_and_group_quotas` | repeated | — | Per-user/group storage quotas |
| `root_volume_configuration.copy_tags_to_snapshots` | bool | false | Propagate tags to snapshots |

### Backup

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `automatic_backup_retention_days` | int32 | 0 | 0–90 days (0 disables) |
| `daily_automatic_backup_start_time` | string | — | HH:MM UTC |
| `copy_tags_to_backups` | bool | false | Tag propagation |
| `copy_tags_to_volumes` | bool | false | Tag propagation to volumes |
| `skip_final_backup` | bool | true | Skip backup on deletion |
| `final_backup_tags` | map | — | Tags applied to the final backup when one is taken |
| `delete_options` | repeated string | — | `DELETE_CHILD_VOLUMES_AND_SNAPSHOTS` opts into cascading deletion (without it, deletion fails while children exist) |

### Maintenance

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `weekly_maintenance_start_time` | string | — | d:HH:MM UTC (1=Mon, 7=Sun) |

## Outputs

| Output | Description |
|--------|-------------|
| `file_system_id` | File system ID (e.g., fs-0123456789abcdef0) |
| `file_system_arn` | ARN for IAM policies |
| `dns_name` | DNS name for NFS mount |
| `endpoint_ip_address` | Endpoint IP (floating for MULTI_AZ) |
| `root_volume_id` | Root volume ID for child volume creation |
| `network_interface_ids` | ENI IDs for network debugging |
| `vpc_id` | VPC where file system resides |
| `owner_id` | AWS account ID |

## Mounting

```bash
# NFS v4.1 mount (recommended)
sudo mount -t nfs -o nfsvers=4.1 <dns_name>:/fsx /mnt/fsx

# For child volumes
sudo mount -t nfs -o nfsvers=4.1 <dns_name>:/fsx/child-vol /mnt/child-vol
```

## Presets

- **01-single-az-development** — SINGLE_AZ_1, 64 GiB, 64 MB/s. Lowest cost for dev/test.
- **02-single-az-production** — SINGLE_AZ_2, 512 GiB, 640 MB/s, ZSTD, KMS, 7-day backups.
- **03-multi-az-high-availability** — MULTI_AZ_1, 1024 GiB, 1280 MB/s, 100K IOPS, quotas, 14-day backups.
- **04-multi-az-intelligent-tiering** — MULTI_AZ_1 on INTELLIGENT_TIERING, 1280 MB/s, elastic capacity, proportional read cache.

## Scope

This component creates the file system and configures its root volume. **Not included** (independent lifecycle resources, referenced via the `root_volume_id` / `file_system_id` outputs):

- **Child volumes** (`aws_fsx_openzfs_volume`) — a separate composable resource surface
- **Snapshots** (`aws_fsx_openzfs_snapshot`) — point-in-time operations, not declarative infrastructure

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
