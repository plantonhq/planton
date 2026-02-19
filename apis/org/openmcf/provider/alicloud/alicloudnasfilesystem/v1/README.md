# AlicloudNasFileSystem

Manages an Alibaba Cloud Network Attached Storage (NAS) file system with a VPC mount target and optional custom access control.

## Overview

NAS provides fully managed, elastic, shared file storage supporting NFS and SMB protocols. This component bundles a file system, an optional access group with access rules, and a VPC mount target into a single deployable unit. A file system without a mount target is unreachable, so these resources are always deployed together.

### What Gets Created

- **NAS File System** -- an `alicloud_nas_file_system` resource (Pulumi: `nas.FileSystem`) with the specified protocol, storage type, and optional encryption
- **Access Group + Access Rules** -- (conditional) an `alicloud_nas_access_group` with `alicloud_nas_access_rule` entries, created only when `accessRules` are specified in the spec
- **Mount Target** -- an `alicloud_nas_mount_target` resource (Pulumi: `nas.MountTarget`) in the specified VPC/VSwitch, producing the domain name used to mount the file system
- **Tags** -- system metadata tags merged with user-defined tags

### File System Types

Two file system types are supported:

| Type | Capacity | Storage Types | Use Case |
|------|----------|---------------|----------|
| **standard** (default) | Auto-scaling | `Performance`, `Capacity`, `Premium` | General-purpose workloads, shared config, logs |
| **extreme** | Fixed, pre-allocated | `standard`, `advance` | High-throughput: ML training, media processing, HPC |

### Important: Immutable Settings

`fileSystemType`, `protocolType`, and `storageType` are **immutable after creation**. Changing any of them requires destroying and recreating the file system and mount target.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., "cn-hangzhou") |
| `protocolType` | string | Mount protocol: `NFS` or `SMB` |
| `storageType` | string | Storage tier (see table above) |
| `vpcId` | StringValueOrRef | VPC for the mount target |
| `vswitchId` | StringValueOrRef | VSwitch for the mount target |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `fileSystemType` | string | `"standard"` | `standard` or `extreme` |
| `description` | string | -- | Human-readable description |
| `encryption` | object | -- | Encryption config: `encryptType` (1=NAS-managed, 2=KMS) + optional `kmsKeyId` |
| `capacity` | int | 0 | GiB capacity (required for extreme NAS, ignored for standard) |
| `zoneId` | string | -- | Availability zone (required for extreme NAS) |
| `accessRules` | list | `[]` | Custom access rules; omit for default VPC-wide RDWR access |
| `resourceGroupId` | string | `""` | Resource group for organizational grouping |
| `tags` | map | `{}` | Key-value tags |

### Access Rule Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sourceCidrIp` | string | (required) | Source IP/CIDR (e.g., `10.0.0.0/8`, `0.0.0.0/0`) |
| `rwAccessType` | string | `"RDWR"` | `RDWR` (read-write) or `RDONLY` (read-only) |
| `userAccessType` | string | `"no_squash"` | `no_squash`, `root_squash`, or `all_squash` |
| `priority` | int | 1 | Rule priority (1-100, lower = higher precedence) |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `file_system_id` | NAS file system ID |
| `mount_target_domain` | Domain name for NFS/SMB mounting |

## Dependencies

- **AlicloudVpc** -- provides `vpcId` for the mount target
- **AlicloudVswitch** -- provides `vswitchId` for the mount target
- **AlicloudKmsKey** -- (optional) provides `kmsKeyId` for customer-managed encryption

## Related Components

- **AlicloudOssBucket** -- object storage (S3-compatible), for unstructured data at scale
- **AlicloudAckManagedCluster** -- Kubernetes clusters that can mount NAS for shared persistent storage
- **AlicloudEcsInstance** -- compute instances that mount NAS via NFS/SMB
