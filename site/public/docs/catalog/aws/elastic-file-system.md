---
title: "Elastic File System"
description: "Elastic File System deployment documentation"
icon: "package"
order: 100
componentName: "awselasticfilesystem"
---

# AWS Elastic File System

Deploys a managed NFS file system with mount targets, optional access points, backup policy, and resource policy. Scales storage automatically as files are added or removed, with no provisioning or capacity planning required.

## What Gets Created

When you deploy an AwsElasticFileSystem resource, OpenMCF provisions:

- **aws_efs_file_system** — the managed NFS file system
- **aws_efs_mount_target** — one per subnet (one per AZ)
- **aws_efs_access_point** — per access point entry (optional)
- **aws_efs_backup_policy** — if `backup_enabled` is true (optional)
- **aws_efs_file_system_policy** — if `policy` is provided (optional)

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **VPC with subnets** — one subnet per Availability Zone for mount targets
- **Security groups** allowing NFS traffic (TCP port 2049) from clients that will mount the file system

## Quick Start

Create a file `efs.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: my-efs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsElasticFileSystem.my-efs
spec:
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        field: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        field: status.outputs.private_subnets.[1].id
```

Deploy:

```shell
openmcf apply -f efs.yaml
```

This creates an EFS file system with mount targets in each specified subnet. The `file_system_id` and `dns_name` outputs are available for downstream references.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `subnet_ids` | `repeated StringValueOrRef` | Subnet IDs for mount targets. One mount target per subnet (one per AZ). Provide at least one subnet. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `encrypted` | `bool` | — | Enable encryption at rest. ForceNew. |
| `kms_key_id` | `StringValueOrRef` | AWS-managed key | Customer-managed KMS key for encryption. Requires `encrypted`. ForceNew. |
| `performance_mode` | `string` | generalPurpose | `generalPurpose` or `maxIO`. ForceNew. |
| `throughput_mode` | `string` | bursting | `bursting`, `provisioned`, or `elastic`. |
| `provisioned_throughput_in_mibps` | `double` | — | MiB/s when `throughput_mode` is `provisioned`. |
| `availability_zone_name` | `string` | — | AZ for One Zone storage (e.g., `us-east-1a`). ForceNew. |
| `transition_to_ia` | `string` | — | Lifecycle: transition to IA (e.g., `AFTER_7_DAYS`). |
| `transition_to_archive` | `string` | — | Lifecycle: transition to Archive. Requires `transition_to_ia`. |
| `transition_to_primary_storage_class` | `string` | — | Lifecycle: warm files on access (`AFTER_1_ACCESS`). |
| `backup_enabled` | `bool` | — | Enable automatic daily backups. |
| `security_group_ids` | `repeated StringValueOrRef` | — | Security groups for mount targets (must allow NFS TCP 2049). |
| `access_points` | `repeated AwsElasticFileSystemAccessPoint` | — | Application-specific entry points with POSIX identity. |
| `policy` | `Struct` | — | IAM resource policy for the file system (JSON). |

## Examples

### Quick Start

Minimal EFS with subnets from a VPC:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: app-efs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsElasticFileSystem.app-efs
spec:
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        field: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        field: status.outputs.private_subnets.[1].id
```

### Encrypted with Elastic Throughput

Production-ready EFS with encryption and elastic throughput:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: prod-efs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsElasticFileSystem.prod-efs
spec:
  encrypted: true
  throughput_mode: elastic
  backup_enabled: true
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        field: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        field: status.outputs.private_subnets.[1].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: efs-clients-sg
        field: status.outputs.security_group_id
```

### With Access Points and valueFrom

EFS with access points for ECS and Lambda, using `valueFrom` for cross-resource references:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: shared-efs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsElasticFileSystem.shared-efs
spec:
  encrypted: true
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: efs-key
      field: status.outputs.key_arn
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        field: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        field: status.outputs.private_subnets.[1].id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: efs-sg
        field: status.outputs.security_group_id
  access_points:
    - name: app-data
      posix_user:
        uid: 1000
        gid: 1000
      root_directory:
        path: /app-data
        creation_info:
          owner_uid: 1000
          owner_gid: 1000
          permissions: "0755"
    - name: lambda-data
      posix_user:
        uid: 1001
        gid: 1001
      root_directory:
        path: /lambda
        creation_info:
          owner_uid: 1001
          owner_gid: 1001
          permissions: "0750"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `file_system_id` | `string` | File system ID (e.g., `fs-0123456789abcdef0`). Primary identifier for EKS PersistentVolumes, ECS task definitions, and Lambda. |
| `file_system_arn` | `string` | ARN of the file system for IAM resource policies. |
| `dns_name` | `string` | Regional DNS name for NFS mount (e.g., `fs-xxx.efs.us-east-1.amazonaws.com`). |
| `mount_target_ids` | `map<string, string>` | Map of subnet ID to mount target ID. |
| `mount_target_ips` | `map<string, string>` | Map of subnet ID to mount target IP. |
| `mount_target_dns_names` | `map<string, string>` | Map of subnet ID to per-AZ mount target DNS name. |
| `access_point_ids` | `map<string, string>` | Map of access point name to access point ID. |
| `access_point_arns` | `map<string, string>` | Map of access point name to access point ARN (required for Lambda). |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides subnets for mount targets
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls NFS traffic to mount targets
- [AwsKmsKey](/docs/catalog/aws/kms-key) — customer-managed key for encryption at rest
- [AwsEksCluster](/docs/catalog/aws/eks-cluster) — uses `file_system_id` for PersistentVolumes via EFS CSI driver
- [AwsLambda](/docs/catalog/aws/lambda) — uses access point ARNs for file system configuration
