# AwsElasticFileSystem

AWS Elastic File System (EFS) — a fully managed, elastic NFS file system that scales storage capacity automatically as files are added or removed. No provisioning or capacity planning required.

## What It Is

EFS provides a shared file system accessible over the Network File System (NFS) protocol. It is a regional, multi-AZ service by default: data is replicated across multiple Availability Zones within a region for durability and availability. Storage grows and shrinks automatically with your data; you pay only for what you use.

This component bundles the file system with its mount targets (one per subnet/AZ), optional access points, backup policy, and resource policy. Mount targets are created automatically from the provided subnet IDs.

## When to Use It

| Use Case | Description |
|----------|-------------|
| **EKS persistent storage** | Use the EFS CSI driver to provision PersistentVolumes backed by `file_system_id`. Pods share data across nodes. |
| **ECS shared volumes** | Attach EFS volumes to ECS task definitions for shared state, logs, or scratch space across tasks. |
| **Lambda file access** | Mount EFS via access points for Lambda functions that need a POSIX file system (ML models, large configs, shared caches). |
| **EC2 NFS mount** | Mount directly from EC2 instances using `mount -t nfs4 <dns_name>:/ /mnt/efs`. |

## When NOT to Use It

| Need | Use Instead |
|------|-------------|
| **Block storage** (databases, boot volumes) | Amazon EBS — lower latency, higher IOPS for single-instance workloads. |
| **Object storage** (blobs, backups, static assets) | Amazon S3 — cheaper, unlimited scale, better for unstructured data. |
| **High-performance HPC** (Lustre, parallel file systems) | Amazon FSx for Lustre or FSx for OpenZFS — sub-millisecond latency, massive throughput. |
| **Windows file shares** | Amazon FSx for Windows File Server — SMB protocol, Active Directory integration. |

## Prerequisites

- **AWS account** with permissions to create EFS file systems, mount targets, and access points.
- **VPC with subnets** — at least one subnet per Availability Zone where you need mount targets. For regional EFS, use one private subnet per AZ.
- **Security groups** — must allow inbound NFS traffic (TCP port 2049) from the clients that will mount the file system.

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `encrypted` | bool | No | Enable encryption at rest. Default: `true`. **ForceNew** — cannot be changed after creation. |
| `kms_key_id` | StringValueOrRef | No | Customer-managed KMS key ARN. Omit to use AWS-managed key `aws/elasticfilesystem`. **ForceNew**. Requires `encrypted: true`. |
| `performance_mode` | string | No | `generalPurpose` (default) or `maxIO`. **ForceNew**. |
| `throughput_mode` | string | No | `bursting` (default), `provisioned`, or `elastic`. |
| `provisioned_throughput_in_mibps` | double | No | MiB/s when `throughput_mode` is `provisioned`. Range: 1.0–3414.0 (generalPurpose), 1.0–1024.0 (maxIO). |
| `availability_zone_name` | string | No | AZ name for One Zone storage (e.g., `us-east-1a`). **ForceNew**. ~47% cheaper than Standard; single-AZ only. |
| `transition_to_ia` | string | No | Transition to Infrequent Access after period. Values: `AFTER_1_DAY`, `AFTER_7_DAYS`, …, `AFTER_365_DAYS`. |
| `transition_to_archive` | string | No | Transition IA files to Archive. Requires `transition_to_ia`. Same value set as above. |
| `transition_to_primary_storage_class` | string | No | Transition back to Standard on access. Only valid: `AFTER_1_ACCESS`. |
| `backup_enabled` | bool | No | Enable automatic daily backups via AWS Backup. Default: `false`. |
| `subnet_ids` | []StringValueOrRef | **Yes** | Subnet IDs for mount targets. One mount target per subnet; max one per AZ. Min 1. |
| `security_group_ids` | []StringValueOrRef | No | Security groups for mount targets. Must allow NFS TCP 2049. |
| `access_points` | []AwsElasticFileSystemAccessPoint | No | Application-specific entry points with POSIX identity and root directory. |
| `policy` | Struct | No | IAM resource policy (JSON). Enforce encryption in transit, restrict principals, etc. |

### Access Point Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | **Yes** | Unique name; used as key in `access_point_ids` and `access_point_arns` outputs. |
| `posix_user` | PosixUser | No | UID, GID, optional secondary GIDs enforced for all operations. |
| `root_directory` | RootDirectory | No | Path exposed as `/`; optional `creation_info` for auto-creation. |

### Root Directory Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `path` | string | **Yes** | Absolute path on EFS to expose as root. Up to 4 subdirectories deep. |
| `creation_info` | CreationInfo | No | Required when path does not exist; EFS creates directory with specified ownership. |

### Root Directory Creation Info

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `owner_uid` | int32 | **Yes** | POSIX UID for directory owner. |
| `owner_gid` | int32 | **Yes** | POSIX GID for directory owner. |
| `permissions` | string | **Yes** | Octal permissions (e.g., `0755`, `0750`). Must match `^0[0-7]{3}$`. |

## Outputs

| Field | Type | Description |
|-------|------|-------------|
| `file_system_id` | string | File system ID (e.g., `fs-0123456789abcdef0`). Primary identifier for EKS, ECS, Lambda. |
| `file_system_arn` | string | ARN for IAM resource-level permissions. |
| `dns_name` | string | Regional DNS name for NFS mount (e.g., `fs-xxx.efs.us-east-1.amazonaws.com`). |
| `mount_target_ids` | map[string]string | Subnet ID → mount target ID. |
| `mount_target_ips` | map[string]string | Subnet ID → mount target IP address. |
| `mount_target_dns_names` | map[string]string | Subnet ID → per-AZ mount target DNS name. |
| `access_point_ids` | map[string]string | Access point name → access point ID. For ECS task definitions. |
| `access_point_arns` | map[string]string | Access point name → access point ARN. Lambda requires ARN. |

## Minimal Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: app-efs
  org: my-org
spec:
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
```

## Production Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: prod-efs
  org: my-org
  labels:
    environment: production
    app: shared-storage
spec:
  encrypted: true
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: efs-encryption-key
      fieldPath: status.outputs.key_arn
  throughput_mode: elastic
  transition_to_ia: AFTER_30_DAYS
  transition_to_archive: AFTER_90_DAYS
  transition_to_primary_storage_class: AFTER_1_ACCESS
  backup_enabled: true
  subnet_ids:
    - valueFrom:
        kind: AwsSubnet
        name: prod-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: prod-private-subnet-b
        fieldPath: status.outputs.subnet_id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: efs-clients-sg
        fieldPath: status.outputs.security_group_id
  access_points:
    - name: ecs-app-data
      posix_user:
        uid: 1000
        gid: 1000
      root_directory:
        path: /app-data
        creation_info:
          owner_uid: 1000
          owner_gid: 1000
          permissions: "0755"
    - name: lambda-models
      posix_user:
        uid: 1001
        gid: 1001
      root_directory:
        path: /ml-models
        creation_info:
          owner_uid: 1001
          owner_gid: 1001
          permissions: "0755"
  policy:
    Version: "2012-10-17"
    Statement:
      - Sid: EnforceEncryptionInTransit
        Effect: Deny
        Principal: "*"
        Action: "*"
        Resource: "*"
        Condition:
          Bool:
            aws:SecureTransport: "false"
```

## ForceNew Warnings

The following fields require **resource replacement** if changed. Plan them upfront:

| Field | Impact |
|-------|--------|
| `encrypted` | Cannot enable encryption after creation. |
| `performance_mode` | Cannot switch between generalPurpose and maxIO. |
| `kms_key_id` | Cannot change the KMS key after creation. |
| `availability_zone_name` | Cannot convert between One Zone and regional storage. |

## Deliberately Omitted (v1)

The following EFS features are not exposed in this API version:

- **Replication configuration** — cross-region or same-region replication.
- **IP address type** (IPv6/dual-stack) for mount targets.
- **Per-mount-target security groups** — security groups apply to all mount targets.
- **Protection** (replication overwrite prevention).

See [docs/README.md](docs/README.md) for architecture details and integration patterns.
