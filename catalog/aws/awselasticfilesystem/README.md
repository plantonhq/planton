# AwsElasticFileSystem

AWS Elastic File System (EFS) — a fully managed, elastic NFS file system that scales storage capacity automatically as files are added or removed. No provisioning or capacity planning required.

## What It Is

EFS provides a shared file system accessible over the Network File System (NFS) protocol. It is a regional, multi-AZ service by default: data is replicated across multiple Availability Zones within a region for durability and availability. Storage grows and shrinks automatically with your data; you pay only for what you use.

This kind bundles the file system with its mount targets (one per subnet/AZ, with optional static IPv4/IPv6 addressing), backup policy, resource policy, replication-overwrite protection, and cross-region/cross-AZ replication. Access points are a separate, first-class resource — see [AwsEfsAccessPoint](../awsefsaccesspoint/README.md) — that references this file system and is itself referenced by Lambda and ECS task definitions.

## When to Use It

| Use Case | Description |
|----------|-------------|
| **EKS persistent storage** | Use the EFS CSI driver to provision PersistentVolumes backed by `file_system_id`. Pods share data across nodes. |
| **ECS shared volumes** | Attach EFS volumes to ECS task definitions for shared state, logs, or scratch space across tasks (via an AwsEfsAccessPoint for least-privilege access). |
| **Lambda file access** | Mount EFS via an AwsEfsAccessPoint for Lambda functions that need a POSIX file system (ML models, large configs, shared caches). |
| **EC2 NFS mount** | Mount directly from EC2 instances using `mount -t nfs4 <dns_name>:/ /mnt/efs`. |

## When NOT to Use It

| Need | Use Instead |
|------|-------------|
| **Block storage** (databases, boot volumes) | Amazon EBS — lower latency, higher IOPS for single-instance workloads. |
| **Object storage** (blobs, backups, static assets) | Amazon S3 — cheaper, unlimited scale, better for unstructured data. |
| **High-performance HPC** (Lustre, parallel file systems) | Amazon FSx for Lustre or FSx for OpenZFS — sub-millisecond latency, massive throughput. |
| **Windows file shares** | Amazon FSx for Windows File Server — SMB protocol, Active Directory integration. |

## Prerequisites

- **VPC with subnets** — one subnet per Availability Zone where you need mount targets. For regional EFS, use one private subnet per AZ.
- **Security groups** — must allow inbound NFS traffic (TCP port 2049) from the clients that will mount the file system. When omitted, AWS attaches the VPC's default security group.

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | string | **Yes** | AWS region for the file system. |
| `encrypted` | bool | No | Enable encryption at rest. Recommended: `true`. **ForceNew** — cannot be changed after creation. |
| `kms_key_id` | StringValueOrRef | No | Customer-managed KMS key ARN. Omit to use AWS-managed key `aws/elasticfilesystem`. **ForceNew**. Requires `encrypted: true`. |
| `performance_mode` | string | No | `generalPurpose` (AWS default) or `maxIO`. **ForceNew**. |
| `throughput_mode` | string | No | `bursting` (AWS default), `provisioned`, or `elastic`. Elastic requires generalPurpose. |
| `provisioned_throughput_in_mibps` | double | No | MiB/s when `throughput_mode` is `provisioned`. AWS accepts 1.0–3414.0 (generalPurpose), 1.0–1024.0 (maxIO). |
| `availability_zone_name` | string | No | AZ name for One Zone storage (e.g., `us-east-1a`). **ForceNew**. ~47% cheaper than Standard; single-AZ only. |
| `transition_to_ia` | string | No | Transition to Infrequent Access after period. Values: `AFTER_1_DAY`, `AFTER_7_DAYS`, …, `AFTER_365_DAYS`. |
| `transition_to_archive` | string | No | Transition IA files to Archive. Requires `transition_to_ia`. Same value set as above. |
| `transition_to_primary_storage_class` | string | No | Transition back to Standard on access. Only valid: `AFTER_1_ACCESS`. |
| `backup_enabled` | bool | No | Enable automatic daily backups via AWS Backup. Default: `false`. |
| `replication_overwrite_protection` | string | No | `ENABLED` (AWS default) or `DISABLED`. Must be `DISABLED` before this file system can be a replication destination. |
| `mount_targets` | []MountTarget | **Yes** | One NFS endpoint per subnet (max one per AZ). Min 1. |
| `security_group_ids` | []StringValueOrRef | No | Security groups applied to all mount targets. Must allow NFS TCP 2049. |
| `policy` | Struct | No | IAM resource policy (JSON). Enforce encryption in transit, restrict principals, etc. |
| `bypass_policy_lockout_safety_check` | bool | No | Skip AWS's lockout check when putting a policy that denies the deploying principal future policy updates. Requires `policy`. |
| `replication` | Replication | No | Replicate to another region and/or AZ for disaster recovery. |

### Mount Target Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `subnet_id` | StringValueOrRef | **Yes** | Subnet for the mount target; its AZ determines which clients it serves. **ForceNew**. |
| `ip_address` | string | No | Static IPv4 address from the subnet's CIDR. **ForceNew**. |
| `ip_address_type` | string | No | `IPV4_ONLY` (AWS default), `IPV6_ONLY`, or `DUAL_STACK`. **ForceNew**. |
| `ipv6_address` | string | No | Static IPv6 address; requires an IPv6-capable `ip_address_type`. **ForceNew**. |

### Replication Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `destination_region` | string | No* | Region for the replica. *At least one of region / AZ is required. |
| `destination_availability_zone_name` | string | No* | AZ for a One Zone replica — the cheaper DR shape. |
| `destination_kms_key_id` | StringValueOrRef | No | KMS key for the replica (replicas are always encrypted). |
| `destination_file_system_id` | StringValueOrRef | No | Replicate into an existing file system (its overwrite protection must be `DISABLED`). |

## Outputs

| Field | Type | Description |
|-------|------|-------------|
| `file_system_id` | string | File system ID (e.g., `fs-0123456789abcdef0`). Primary identifier for EKS, ECS, and AwsEfsAccessPoint. |
| `file_system_arn` | string | ARN for IAM resource-level permissions. |
| `dns_name` | string | Regional DNS name for NFS mount (e.g., `fs-xxx.efs.us-east-1.amazonaws.com`). |
| `mount_target_ids` | map[string]string | Subnet ID → mount target ID. |
| `mount_target_ips` | map[string]string | Subnet ID → mount target IPv4 address (empty for IPV6_ONLY targets). |
| `mount_target_ipv6_addresses` | map[string]string | Subnet ID → mount target IPv6 address (IPV6_ONLY / DUAL_STACK targets). |
| `mount_target_dns_names` | map[string]string | Subnet ID → per-AZ mount target DNS name. |
| `replication_destination_file_system_id` | string | Replica's file system ID; empty when replication is not configured. |

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticFileSystem
metadata:
  name: app-efs
  org: my-org
spec:
  region: us-east-1
  mountTargets:
    - subnetId:
        value: subnet-0a1b2c3d4e5f00001
    - subnetId:
        value: subnet-0a1b2c3d4e5f00002
```

## Production Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticFileSystem
metadata:
  name: prod-efs
  org: my-org
  labels:
    environment: production
    app: shared-storage
spec:
  region: us-east-1
  encrypted: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: efs-encryption-key
      fieldPath: status.outputs.key_arn
  throughputMode: elastic
  transitionToIa: AFTER_30_DAYS
  transitionToArchive: AFTER_90_DAYS
  transitionToPrimaryStorageClass: AFTER_1_ACCESS
  backupEnabled: true
  mountTargets:
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: prod-private-subnet-a
          fieldPath: status.outputs.subnet_id
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: prod-private-subnet-b
          fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: efs-clients-sg
        fieldPath: status.outputs.security_group_id
  replication:
    destinationRegion: us-west-2
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

Application entry points are declared as separate `AwsEfsAccessPoint` resources referencing this file system:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEfsAccessPoint
metadata:
  name: ecs-app-data
  org: my-org
spec:
  region: us-east-1
  fileSystemId:
    valueFrom:
      kind: AwsElasticFileSystem
      name: prod-efs
      fieldPath: status.outputs.file_system_id
  posixUser:
    uid: 1000
    gid: 1000
  rootDirectory:
    path: /app-data
    creationInfo:
      ownerUid: 1000
      ownerGid: 1000
      permissions: "0755"
```

## ForceNew Warnings

The following fields require **resource replacement** if changed. Plan them upfront:

| Field | Impact |
|-------|--------|
| `encrypted` | Cannot enable encryption after creation. |
| `performance_mode` | Cannot switch between generalPurpose and maxIO. |
| `kms_key_id` | Cannot change the KMS key after creation. |
| `availability_zone_name` | Cannot convert between One Zone and regional storage. |
| `mount_targets[].*` | A mount target's subnet, addresses, and address family are all create-time; only its security groups mutate. |
| `replication` | The whole replication configuration replaces on any change (the destination file system survives). |

## Deliberately Omitted (v1)

- **Per-mount-target security groups** — security groups apply to all mount targets; they gate the same NFS clients regardless of AZ, and AWS's own console offers no per-AZ distinction. Compose separate file systems if isolation is genuinely per-AZ.
- **`creation_token` as a field** — the token is pinned to `metadata.name` by both engines (it is the idempotency key, not configuration).

See GUIDE.md for architecture details and integration patterns.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
