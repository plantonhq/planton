# AlicloudNasFileSystem -- Research Documentation

## Alibaba Cloud NAS Overview

Alibaba Cloud Network Attached Storage (NAS) is a fully managed, elastic, shared file storage service supporting NFS and SMB protocols. It provides POSIX-compliant file system semantics, enabling multiple compute resources to mount and access the same file system concurrently.

### Key Characteristics

- **Shared access**: Multiple ECS instances, containers, and serverless functions can mount the same file system simultaneously with full POSIX semantics (read/write locking, permissions, etc.).
- **Elastic capacity** (standard): Standard NAS auto-scales from 0 to 10 PiB. You pay only for stored data with no pre-allocation.
- **Fixed capacity** (extreme): Extreme NAS provides dedicated throughput with pre-allocated capacity. Minimum 100 GiB.
- **Regional**: File systems are created in a region and accessed via mount targets in VPC/VSwitches within that region.
- **Immutable settings**: `file_system_type`, `protocol_type`, and `storage_type` cannot be changed after creation.

## Provider Resources

### Terraform (14 NAS resources total)

**Bundled in this component:**
- `alicloud_nas_file_system` -- the file system itself
- `alicloud_nas_access_group` -- VPC access permission group
- `alicloud_nas_access_rule` -- IP-based access rules within a group
- `alicloud_nas_mount_target` -- VPC mount point producing the NFS/SMB domain

**Not managed (available as separate resources):**
- `alicloud_nas_auto_snapshot_policy` -- automated snapshot scheduling (extreme NAS)
- `alicloud_nas_snapshot` -- point-in-time snapshots (extreme NAS)
- `alicloud_nas_lifecycle_policy` -- tiered storage lifecycle management
- `alicloud_nas_fileset` -- directory-level management unit
- `alicloud_nas_data_flow` -- OSS-to-NAS data synchronization
- `alicloud_nas_access_point` -- fine-grained mount points (NFS v4 ACL)
- `alicloud_nas_protocol_service` -- protocol acceleration (CPFS)
- `alicloud_nas_protocol_mount_target` -- CPFS mount targets
- `alicloud_nas_recycle_bin` -- soft-delete recovery
- `alicloud_nas_smb_acl_attachment` -- Active Directory integration (SMB)

### Pulumi

- **FileSystem**: `nas.FileSystem` (token: `alicloud:nas/fileSystem:FileSystem`)
- **AccessGroup**: `nas.AccessGroup` (token: `alicloud:nas/accessGroup:AccessGroup`)
- **AccessRule**: `nas.AccessRule` (token: `alicloud:nas/accessRule:AccessRule`)
- **MountTarget**: `nas.MountTarget` (token: `alicloud:nas/mountTarget:MountTarget`)
- **SDK import**: `github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/nas`

## File System Types

| Type | Capacity Model | Performance | Storage Types | VPC Placement |
|------|----------------|-------------|---------------|---------------|
| **standard** | Auto-scaling (0-10 PiB) | Shared throughput pool | Performance, Capacity, Premium | Mount target only |
| **extreme** | Fixed pre-allocated (min 100 GiB) | Dedicated throughput | standard, advance | File system + mount target |

### Standard NAS Storage Types

| Storage Type | Media | Max Throughput | Max IOPS | Latency | Use Case |
|-------------|-------|----------------|----------|---------|----------|
| Performance | SSD | Up to 10 GiB/s | 500K | <1ms | Hot data, databases, CI/CD |
| Capacity | HDD | Up to 2 GiB/s | 15K | ~10ms | Warm/cold data, archives, backups |
| Premium | Next-gen SSD | Up to 20 GiB/s | 1M | <0.2ms | High-performance workloads |

### Extreme NAS Storage Types

| Storage Type | Throughput per TiB | Max IOPS per TiB | Use Case |
|-------------|-------------------|-------------------|----------|
| standard | 150 MB/s | 50K | General extreme workloads |
| advance | 300 MB/s | 100K | ML training, HPC, media processing |

## Encryption

| encrypt_type | Description | Key Management |
|-------------|-------------|----------------|
| 0 (default) | No encryption | N/A |
| 1 | NAS-managed key | Automatic; no user-managed keys |
| 2 | Customer-managed KMS key | Requires `kms_key_id`; full key lifecycle control |

Encryption is set at creation time and is immutable (ForceNew). Data in transit is encrypted by default over NFS v4.0.

## Access Control Model

NAS uses a two-tier access control model:

1. **Access Group** -- a named collection of access rules scoped to a file system type (standard or extreme). Each mount target references one access group.
2. **Access Rules** -- IP-based allow rules within an access group. Each rule specifies a source CIDR, read-write mode, user identity mapping, and priority.

The built-in `DEFAULT_VPC_GROUP_NAME` access group allows full RDWR access from all VPC IPs and is used when no custom access group is specified.

### User Access Types (NFS identity mapping)

| Value | Behavior | Security |
|-------|----------|----------|
| `no_squash` | Preserve original user identity | Lowest security (clients set their own uid/gid) |
| `root_squash` | Map root (uid 0) to anonymous | Recommended for production (prevents root escalation) |
| `all_squash` | Map all users to anonymous | Highest security (all writes attributed to anonymous) |

## Mount Target Behavior

- Each file system can have up to 2 mount targets.
- A mount target is bound to a VPC + VSwitch and produces a domain name (e.g., `1234abc-xyz12.cn-hangzhou.nas.aliyuncs.com`).
- Standard NAS: VPC/VSwitch is set only on the mount target.
- Extreme NAS: VPC/VSwitch is set on both the file system and the mount target.
- Mount targets support NFS v3, v4.0 and SMB 2.x/3.x protocols.

## Design Decisions in This Component

1. **Composite bundling (DD07)**: File system + access group + access rules + mount target are bundled because a file system without a mount target is unreachable.
2. **Default VPC access**: When `access_rules` is omitted, the mount target uses `DEFAULT_VPC_GROUP_NAME` for simplicity.
3. **Custom access group naming**: When `access_rules` are specified, the access group is named `{metadata.name}-ag` for traceability.
4. **CPFS excluded**: CPFS (Cloud Parallel File System) is a niche HPC product with distinct requirements. It can be added as a separate component if demand arises.
5. **Snapshots excluded**: NAS snapshots are only supported on extreme file systems and have an independent lifecycle. They fit better as a separate component.
