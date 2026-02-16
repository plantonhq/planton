---
title: "Elastic File System"
description: "Elastic File System deployment documentation"
icon: "package"
order: 100
componentName: "awselasticfilesystem"
---

# AWS Elastic File System

Deploys an AWS Elastic File System with automatic mount target creation across specified subnets, optional access points for application-specific entry points, lifecycle policies for cost-optimized storage tiering, and an optional IAM resource policy. The component bundles everything needed to make the file system mountable immediately after deployment.

## What Gets Created

When you deploy an AwsElasticFileSystem resource, OpenMCF provisions:

- **EFS File System** — an `efs.FileSystem` resource with the configured encryption, performance mode, throughput mode, and lifecycle policies
- **Mount Targets** — one `efs.MountTarget` per subnet, placing an elastic network interface in each Availability Zone for NFS client access on TCP port 2049
- **Access Points** — one `efs.AccessPoint` per entry in `accessPoints`, each enforcing a POSIX user/group identity and optional root directory restriction
- **Backup Policy** — an `efs.BackupPolicy` enabling automatic daily backups via AWS Backup, created only when `backupEnabled` is `true`
- **File System Policy** — an `efs.FileSystemPolicy` attaching an IAM resource policy to the file system, created only when `policy` is provided

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **At least one subnet** where mount targets will be created (one subnet per AZ for multi-AZ availability)
- **A security group** allowing inbound NFS traffic (TCP port 2049) from the clients that will mount the file system
- **A KMS key ARN** if using customer-managed encryption (otherwise EFS uses the AWS-managed `aws/elasticfilesystem` key)

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
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  securityGroupIds:
    - sg-0a1b2c3d4e5f00001
```

Deploy:

```shell
openmcf apply -f efs.yaml
```

This creates an unencrypted, bursting-throughput EFS file system with mount targets in two subnets and no access points.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `subnetIds` | `StringValueOrRef[]` | Subnet IDs where mount targets are created. One mount target per subnet (one per AZ). Can reference AwsVpc resource via `valueFrom`. | Minimum 1 item required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `encrypted` | `bool` | `false` | Enable encryption at rest. ForceNew. Recommended: `true`. |
| `kmsKeyId` | `StringValueOrRef` | AWS-managed key | Customer-managed KMS key ARN for encryption. Requires `encrypted` to be `true`. ForceNew. Can reference AwsKmsKey via `valueFrom`. |
| `performanceMode` | `string` | `generalPurpose` | File system performance mode. ForceNew. Valid values: `generalPurpose`, `maxIO`. |
| `throughputMode` | `string` | `bursting` | Throughput mode. Valid values: `bursting`, `provisioned`, `elastic`. |
| `provisionedThroughputInMibps` | `double` | — | Fixed throughput in MiB/s. Required when `throughputMode` is `provisioned`. Range: 1.0-3414.0. |
| `availabilityZoneName` | `string` | — | AZ name for One Zone storage (e.g., `us-east-1a`). ForceNew. When set, only one mount target is allowed. |
| `transitionToIa` | `string` | — | Transition to Infrequent Access after period. Valid values: `AFTER_1_DAY`, `AFTER_7_DAYS`, `AFTER_14_DAYS`, `AFTER_30_DAYS`, `AFTER_60_DAYS`, `AFTER_90_DAYS`, `AFTER_180_DAYS`, `AFTER_270_DAYS`, `AFTER_365_DAYS`. |
| `transitionToArchive` | `string` | — | Transition to Archive after period. Requires `transitionToIa` to be set. Same valid values as `transitionToIa`. |
| `transitionToPrimaryStorageClass` | `string` | — | Move files back to Standard on access. Only valid value: `AFTER_1_ACCESS`. |
| `backupEnabled` | `bool` | `false` | Enable automatic daily backups via AWS Backup. |
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | Security groups for mount targets. Must allow inbound TCP 2049. Can reference AwsSecurityGroup via `valueFrom`. |
| `accessPoints` | `object[]` | `[]` | Application-specific entry points with POSIX identity enforcement and root directory restriction. |
| `accessPoints[].name` | `string` | — | Unique name for the access point. Used as the key in `access_point_ids` and `access_point_arns` outputs. Required. |
| `accessPoints[].posixUser.uid` | `int32` | — | POSIX user ID enforced for all file operations. Required when `posixUser` is set. |
| `accessPoints[].posixUser.gid` | `int32` | — | POSIX primary group ID. Required when `posixUser` is set. |
| `accessPoints[].posixUser.secondaryGids` | `int32[]` | `[]` | Secondary POSIX group IDs. |
| `accessPoints[].rootDirectory.path` | `string` | — | Absolute path exposed as root. Required when `rootDirectory` is set. |
| `accessPoints[].rootDirectory.creationInfo.ownerUid` | `int32` | — | POSIX UID for auto-created directory. Required when `creationInfo` is set. |
| `accessPoints[].rootDirectory.creationInfo.ownerGid` | `int32` | — | POSIX GID for auto-created directory. Required when `creationInfo` is set. |
| `accessPoints[].rootDirectory.creationInfo.permissions` | `string` | — | Octal permissions (e.g., `0755`). Must match `^0[0-7]{3}$`. Required when `creationInfo` is set. |
| `policy` | `object` | — | IAM resource policy for the file system as a JSON object structure. |

## Examples

### Encrypted Multi-AZ File System

Production-ready file system with encryption and mount targets across two Availability Zones:

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
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  securityGroupIds:
    - sg-nfs-access
```

### One Zone Development with Lifecycle Policies

Cost-optimized single-AZ file system with automatic tiering to Infrequent Access and Archive storage:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: dev-efs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsElasticFileSystem.dev-efs
spec:
  availabilityZoneName: us-east-1a
  throughputMode: elastic
  transitionToIa: AFTER_30_DAYS
  transitionToArchive: AFTER_90_DAYS
  transitionToPrimaryStorageClass: AFTER_1_ACCESS
  subnetIds:
    - subnet-dev-az1
  securityGroupIds:
    - sg-nfs-dev
```

### File System with Access Points

Multi-tenant file system with application-specific access points that enforce POSIX identities and root directories:

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
  backupEnabled: true
  throughputMode: elastic
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  securityGroupIds:
    - sg-nfs-access
  accessPoints:
    - name: app-data
      posixUser:
        uid: 1000
        gid: 1000
      rootDirectory:
        path: /app-data
        creationInfo:
          ownerUid: 1000
          ownerGid: 1000
          permissions: "0755"
    - name: lambda-data
      posixUser:
        uid: 1001
        gid: 1001
      rootDirectory:
        path: /lambda
        creationInfo:
          ownerUid: 1001
          ownerGid: 1001
          permissions: "0750"
```

### Using Foreign Key References

Reference OpenMCF-managed VPC subnets, security groups, and KMS keys instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticFileSystem
metadata:
  name: ref-efs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsElasticFileSystem.ref-efs
spec:
  encrypted: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: efs-key
      field: status.outputs.key_arn
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets[0].id
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets[1].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: nfs-sg
        field: status.outputs.security_group_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `file_system_id` | `string` | File system ID (e.g., `fs-0123456789abcdef0`). Used by EKS PersistentVolumes, ECS task definitions, and Lambda file system configurations. |
| `file_system_arn` | `string` | Amazon Resource Name of the file system for IAM policies |
| `dns_name` | `string` | Regional DNS name for NFS mounting (e.g., `fs-xxx.efs.us-east-1.amazonaws.com`) |
| `mount_target_ids` | `map<string, string>` | Map of subnet ID to mount target ID |
| `mount_target_ips` | `map<string, string>` | Map of subnet ID to mount target IP address |
| `mount_target_dns_names` | `map<string, string>` | Map of subnet ID to AZ-specific mount target DNS name |
| `access_point_ids` | `map<string, string>` | Map of access point name to access point ID. Only populated when `accessPoints` is configured. |
| `access_point_arns` | `map<string, string>` | Map of access point name to access point ARN. Only populated when `accessPoints` is configured. |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides the subnets for mount target placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls NFS access (TCP port 2049) to the file system
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides customer-managed encryption keys
- [AwsEksCluster](/docs/catalog/aws/eks-cluster) — consumes EFS via the EFS CSI driver for PersistentVolumes
- [AwsLambda](/docs/catalog/aws/lambda) — mounts EFS via access point ARNs for serverless file access
