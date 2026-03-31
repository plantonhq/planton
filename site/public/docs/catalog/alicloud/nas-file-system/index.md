---
title: "NAS File System"
description: "NAS File System deployment documentation"
icon: "package"
order: 100
componentName: "alicloudnasfilesystem"
---

# AliCloud NAS File System

Deploys an Alibaba Cloud Network Attached Storage (NAS) file system with a VPC mount target and optional custom access control. NAS provides fully managed, elastic, shared file storage supporting NFS and SMB protocols, accessible from ECS instances, Kubernetes pods, and serverless functions within a VPC.

## What Gets Created

When you deploy an AliCloudNasFileSystem resource, OpenMCF provisions:

- **NAS File System** -- an `alicloud_nas_file_system` resource (Pulumi: `nas.FileSystem`) with the specified protocol type, storage tier, and optional encryption at rest
- **Access Group + Access Rules** -- (conditional) when `accessRules` are specified, a custom `alicloud_nas_access_group` with `alicloud_nas_access_rule` entries controlling which IP ranges can mount the file system and with what permissions
- **Mount Target** -- an `alicloud_nas_mount_target` resource (Pulumi: `nas.MountTarget`) in the specified VPC/VSwitch, producing the domain name clients use for NFS/SMB mounting
- **Tags** -- system metadata tags (`resource`, `resource_name`, `resource_kind`, `organization`, `environment`) merged with user-defined `spec.tags`, with user values taking precedence on key conflict

When no `accessRules` are specified, the mount target uses the built-in DEFAULT_VPC_GROUP_NAME access group, which allows full read-write access from all IP addresses within the VPC.

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or OpenMCF provider config
- **A VPC and VSwitch** in the target region -- the mount target is created in this VSwitch
- **OpenMCF CLI** installed with either Pulumi or Terraform (OpenTofu) backend

## Quick Start

Create a file `nas.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudNasFileSystem
metadata:
  name: shared-data
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudNasFileSystem.shared-data
spec:
  region: cn-hangzhou
  protocolType: NFS
  storageType: Performance
  vpcId: vpc-abc123
  vswitchId: vsw-abc123
```

Deploy:

```shell
openmcf apply -f nas.yaml
```

This creates a standard NFS file system with Performance storage and a mount target accessible from all VPC IPs. Mount the file system:

```shell
mount -t nfs -o vers=4,minorversion=0,noresvport <mount_target_domain>:/ /mnt/nas
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region (e.g., `cn-hangzhou`, `cn-shanghai`, `us-west-1`). | Required; non-empty |
| `protocolType` | `string` | Mount protocol: `NFS` (Linux/Unix) or `SMB` (Windows). | Required; immutable |
| `storageType` | `string` | Storage tier. Standard: `Performance`, `Capacity`, `Premium`. Extreme: `standard`, `advance`. | Required; immutable |
| `vpcId` | `StringValueOrRef` | VPC for the mount target. | Required |
| `vswitchId` | `StringValueOrRef` | VSwitch for the mount target. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `fileSystemType` | `string` | `"standard"` | `standard` (auto-scaling, general-purpose) or `extreme` (dedicated throughput, fixed capacity). **Immutable after creation.** |
| `description` | `string` | -- | Human-readable description. |
| `encryption` | `object` | -- | Encryption config: `encryptType` (`1`=NAS-managed, `2`=KMS) and optional `kmsKeyId`. **Immutable after creation.** |
| `capacity` | `int` | `0` | GiB capacity. Required for extreme NAS (min 100). Ignored for standard (auto-scales). |
| `zoneId` | `string` | -- | Availability zone. Required for extreme NAS. Format: `cn-hangzhou-a`. |
| `accessRules` | `list` | `[]` | Custom access rules. Omit for default full VPC access. |
| `resourceGroupId` | `string` | `""` | Resource group for organizational grouping. |
| `tags` | `map<string, string>` | `{}` | User-defined tags merged with system tags. |

## Examples

### Minimal NFS File System

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudNasFileSystem
metadata:
  name: dev-share
spec:
  region: cn-hangzhou
  protocolType: NFS
  storageType: Performance
  vpcId: vpc-abc123
  vswitchId: vsw-abc123
```

### Production NFS with Encryption and Access Rules

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudNasFileSystem
metadata:
  name: prod-storage
  org: my-org
  env: production
spec:
  region: cn-shanghai
  protocolType: NFS
  storageType: Performance
  encryption:
    encryptType: 1
  vpcId: vpc-prod-001
  vswitchId: vsw-prod-001
  accessRules:
    - sourceCidrIp: "10.0.1.0/24"
      rwAccessType: RDWR
    - sourceCidrIp: "10.0.2.0/24"
      rwAccessType: RDONLY
      userAccessType: root_squash
  tags:
    team: platform
```

### Extreme NAS for High-Throughput Workloads

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudNasFileSystem
metadata:
  name: hpc-scratch
spec:
  region: cn-hangzhou
  fileSystemType: extreme
  protocolType: NFS
  storageType: advance
  capacity: 500
  zoneId: cn-hangzhou-a
  encryption:
    encryptType: 2
    kmsKeyId: "cmk-abc123"
  vpcId: vpc-hpc-001
  vswitchId: vsw-hpc-001
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `file_system_id` | `string` | The NAS file system ID assigned by Alibaba Cloud. |
| `mount_target_domain` | `string` | The mount target domain name used for NFS/SMB mounting from within the VPC. |

## Related Components

- [AliCloudVpc](/docs/catalog/alicloud/vpc) -- prerequisite VPC for the mount target
- [AliCloudVswitch](/docs/catalog/alicloud/vswitch) -- prerequisite VSwitch for the mount target
- [AliCloudKmsKey](/docs/catalog/alicloud/kms-key) -- for customer-managed encryption keys
- [AliCloudStorageBucket](/docs/catalog/alicloud/oss-bucket) -- object storage alternative for unstructured data
- [AliCloudAckManagedCluster](/docs/catalog/alicloud/alicloudackmanagedcluster) -- Kubernetes clusters that mount NAS for shared persistent volumes
