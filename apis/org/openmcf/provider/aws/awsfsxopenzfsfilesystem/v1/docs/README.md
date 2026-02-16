# AwsFsxOpenzfsFileSystem — Technical Reference

## Service Overview

Amazon FSx for OpenZFS is a fully managed file storage service built on the OpenZFS file system. It provides sub-millisecond latency, up to 10 GB/s throughput, and over 1 million IOPS through standard NFS protocols (NFSv3, NFSv4.0, NFSv4.1, NFSv4.2). Key OpenZFS features include instant snapshots, data cloning, ZSTD/LZ4 compression, and per-user/group quotas.

## Terraform Resource Mapping

| OpenMCF Concept | Terraform Resource | Notes |
|-----------------|-------------------|-------|
| File system | `aws_fsx_openzfs_file_system` | Primary resource |
| Root volume | Inline `root_volume_configuration` block | Configured on the file system resource |
| Child volumes | `aws_fsx_openzfs_volume` | NOT managed by this component (independent lifecycle) |
| Snapshots | `aws_fsx_openzfs_snapshot` | NOT managed by this component |

## Deployment Types

### SINGLE_AZ_1

First-generation single-AZ deployment. Lower throughput ceiling (max 4,096 MB/s). Uses one subnet and one ENI. Lowest cost option. Best for development and non-critical workloads.

### SINGLE_AZ_2

Latest single-AZ deployment (recommended for new workloads). Higher throughput ceiling (max 10,240 MB/s). More features than SINGLE_AZ_1. One subnet, one ENI.

### MULTI_AZ_1

Multi-AZ deployment with automatic failover. Data is synchronously replicated across two AZs. Requires two subnets in different AZs, a preferred subnet (active node), route table IDs for automatic route management, and optionally an endpoint IP address range for floating IPs. Same throughput range as SINGLE_AZ_2.

Failover is transparent to NFS clients — the DNS name resolves to a floating IP that follows the active file server.

## Storage Types

### SSD (v1 supported)

Solid-state drives. Sub-millisecond latency. Available for all deployment types. Storage capacity: 64–524,288 GiB.

### INTELLIGENT_TIERING (v1 deferred)

Automatic data tiering between SSD and capacity pool storage. Only available with MULTI_AZ_1. No explicit storage capacity — AWS manages capacity automatically. Requires `read_cache_configuration`. Deferred from v1 due to complexity and limited adoption.

## Disk IOPS

- **AUTOMATIC** (default): 3 IOPS per GiB of storage, up to deployment type limit
- **USER_PROVISIONED**: Explicit IOPS independent of storage. SINGLE_AZ_1: up to 160,000. SINGLE_AZ_2/MULTI_AZ_1: up to 400,000.

## Root Volume Configuration

The root volume is automatically created with the file system. Key settings:

- **data_compression_type**: NONE, ZSTD (best ratio, ~2-3x), LZ4 (fastest, ~1.5-2x)
- **record_size_kib**: 4–1024 KiB. Default 128. Smaller for random I/O (databases), larger for sequential (analytics).
- **nfs_exports**: Client configurations with IP/CIDR/wildcard + mount options (rw, ro, crossmnt, root_squash, etc.)
- **user_and_group_quotas**: Per-UID/GID storage limits. Up to 100 entries.
- **read_only**: Makes the entire root volume read-only.

## Networking

- **Ports required**: TCP 111 (portmapper), TCP 2049 (NFS), TCP 20001-20003 (NFS mount)
- **SINGLE_AZ**: 1 ENI in the specified subnet
- **MULTI_AZ**: 2 ENIs (one per subnet), floating IP for failover

## Backup

- Automatic backups: 0–90 day retention (0 = disabled)
- Daily backup window: HH:MM UTC format
- Final backup on deletion: controlled by `skip_final_backup`
- Tag propagation: `copy_tags_to_backups`, `copy_tags_to_volumes`

## ForceNew Attributes

Changing these requires replacing the file system (destructive):

- `deployment_type`
- `subnet_ids`
- `security_group_ids`
- `kms_key_id`
- `preferred_subnet_id`
- `endpoint_ip_address_range`
- `root_volume_configuration.copy_tags_to_snapshots`

## v1 Scope and Exclusions

### Included
- File system creation with all deployment types
- Root volume configuration (compression, NFS, quotas, record size)
- Disk IOPS configuration
- Backup configuration
- Customer-managed KMS encryption
- Multi-AZ networking (preferred subnet, route tables, endpoint IP range)

### Excluded (future versions)
- `INTELLIGENT_TIERING` storage type and `read_cache_configuration`
- `backup_id` (create from existing backup)
- `delete_options` (child volume deletion behavior)
- `final_backup_tags`
- Child volumes (`aws_fsx_openzfs_volume`) — separate component
- Snapshots (`aws_fsx_openzfs_snapshot`) — separate component
