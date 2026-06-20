---
title: "FSx Lustre File System"
description: "FSx Lustre File System deployment documentation"
icon: "package"
order: 100
componentName: "awsfsxlustrefilesystem"
---

# AWS FSx Lustre File System

Deploys an Amazon FSx for Lustre file system with configurable deployment type, storage capacity, throughput tiers, optional S3 data integration, CloudWatch audit logging, and automatic backups. The component supports both ephemeral scratch file systems for temporary high-performance processing and persistent file systems with intra-AZ data replication and backup support.

## What Gets Created

When you deploy an AwsFsxLustreFileSystem resource, OpenMCF provisions:

- **FSx for Lustre File System** — an `aws_fsx_lustre_file_system` resource placed in the specified subnet with the configured deployment type, storage capacity, encryption settings, and optional S3 import/export, CloudWatch log configuration, backup schedule, and metadata performance tuning

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A subnet** in the target Availability Zone — Lustre file systems are single-AZ, exactly one subnet is required
- **A security group** allowing Lustre traffic between the file system and its clients: TCP port 988 (Lustre protocol) and TCP ports 1018-1023 (data channels)
- **A KMS key ARN** if using customer-managed encryption at rest (all Lustre file systems are encrypted by default with an AWS-managed key)
- **A CloudWatch Logs log group** with an FSx resource policy if enabling audit logging
- **An S3 bucket** if configuring import/export paths on scratch file systems

## Quick Start

Create a file `fsx-lustre.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: my-fsx-lustre
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsFsxLustreFileSystem.my-fsx-lustre
spec:
  storageCapacityGib: 1200
  subnetId: subnet-0a1b2c3d4e5f00001
```

Deploy:

```shell
openmcf apply -f fsx-lustre.yaml
```

This creates a SCRATCH_2 SSD file system with 1200 GiB in the specified subnet. No data replication, no backups — suitable for temporary processing workloads.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the FSx Lustre file system will be created (e.g., `us-east-1`). | Required; non-empty |
| `storageCapacityGib` | `int32` | Storage capacity in GiB. Valid increments depend on deployment type and storage type. Can be increased after creation but never decreased. | Minimum 1200 |
| `subnetId` | `string` | Subnet ID for the file system's network interface. Lustre is single-AZ — exactly one subnet. ForceNew. | Required |
| `subnetId.value` | `string` | Direct subnet ID value | — |
| `subnetId.valueFrom` | `object` | Foreign key reference to an AwsVpc resource | Default kind: `AwsVpc` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `deploymentType` | `string` | `SCRATCH_2` | Deployment type controlling durability and performance. ForceNew. Valid: `SCRATCH_1`, `SCRATCH_2`, `PERSISTENT_1`, `PERSISTENT_2`. |
| `storageType` | `string` | `SSD` | Storage media type. ForceNew. `SSD` for sub-millisecond latency. `HDD` for lower cost — only available with `PERSISTENT_1`. |
| `perUnitStorageThroughput` | `int32` | — | Throughput in MB/s/TiB. Required for PERSISTENT types, invalid for SCRATCH. PERSISTENT_1 + SSD: 50, 100, 200. PERSISTENT_1 + HDD: 12, 40. PERSISTENT_2 + SSD: 125, 250, 500, 1000. |
| `dataCompressionType` | `string` | `NONE` | Data compression algorithm. `NONE` or `LZ4`. Can be changed after creation without impact to Lustre operations. |
| `fileSystemTypeVersion` | `string` | — | Lustre version (e.g., `2.12`, `2.15`). ForceNew. Leave empty for the latest version supported by the deployment type. |
| `securityGroupIds` | `string[]` | `[]` | Security group IDs attached to the file system ENI. ForceNew. Must allow TCP 988 and 1018-1023. Can reference AwsSecurityGroup resources via `valueFrom`. Up to 50. |
| `kmsKeyId` | `string` | — | Customer-managed KMS key ARN for encryption at rest. ForceNew. When omitted, uses the AWS-managed FSx key. Can reference AwsKmsKey resource via `valueFrom`. |
| `importPath` | `string` | — | S3 URI to import data from (e.g., `s3://my-bucket/prefix`). ForceNew. SCRATCH_1 and SCRATCH_2 only. |
| `exportPath` | `string` | — | S3 URI for exporting data back to S3. ForceNew. Requires `importPath` to be set. |
| `logConfiguration.destination` | `string` | — | CloudWatch Logs log group ARN for audit events. Can reference AwsCloudwatchLogGroup resource via `valueFrom`. |
| `logConfiguration.level` | `string` | `WARN_ERROR` | Audit log level. Valid: `DISABLED`, `WARN_ONLY`, `ERROR_ONLY`, `WARN_ERROR`. |
| `automaticBackupRetentionDays` | `int32` | `0` | Days to retain automatic backups. Range: 0-90. Set to 0 to disable. PERSISTENT deployments only. |
| `dailyAutomaticBackupStartTime` | `string` | — | UTC time to start daily backups in `HH:MM` format (e.g., `05:00`). |
| `copyTagsToBackups` | `bool` | `false` | Copy file system tags to automatic backups. ForceNew. |
| `skipFinalBackup` | `bool` | `true` | Skip creating a final backup on deletion. PERSISTENT deployments only. |
| `weeklyMaintenanceStartTime` | `string` | — | Weekly UTC maintenance window in `d:HH:MM` format where d is 1=Monday through 7=Sunday (e.g., `1:05:00` for Monday 05:00 UTC). |
| `metadataConfiguration.mode` | `string` | `AUTOMATIC` | Metadata IOPS mode. PERSISTENT_2 only. `AUTOMATIC` scales with storage capacity. `USER_PROVISIONED` allows explicit IOPS. |
| `metadataConfiguration.iops` | `int32` | — | Metadata IOPS when mode is `USER_PROVISIONED`. Valid values: 1500 through 192000 in documented increments. Ignored in AUTOMATIC mode. |

## Examples

### Scratch File System with S3 Import

A temporary file system that imports data from S3 for batch processing jobs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: batch-fsx
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsFsxLustreFileSystem.batch-fsx
spec:
  region: us-west-2
  storageCapacityGib: 3600
  subnetId: subnet-private-az1
  securityGroupIds:
    - sg-lustre-clients
  dataCompressionType: LZ4
  importPath: s3://my-data-bucket/training-data
  exportPath: s3://my-data-bucket/results
```

### Persistent High-Throughput for ML Training

PERSISTENT_2 with maximum throughput tier, LZ4 compression, automatic backups, and metadata IOPS scaling for production ML workloads:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: ml-training-fsx
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxLustreFileSystem.ml-training-fsx
spec:
  region: us-east-1
  deploymentType: PERSISTENT_2
  storageCapacityGib: 4800
  storageType: SSD
  perUnitStorageThroughput: 1000
  dataCompressionType: LZ4
  subnetId: subnet-private-az1
  securityGroupIds:
    - sg-lustre-ml
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/mrk-example
  automaticBackupRetentionDays: 7
  dailyAutomaticBackupStartTime: "04:00"
  copyTagsToBackups: true
  weeklyMaintenanceStartTime: "7:03:00"
  metadataConfiguration:
    mode: AUTOMATIC
```

### HDD Data Lake with Cost-Optimized Storage

PERSISTENT_1 HDD for large-capacity, sequential-throughput workloads where cost per GiB is the primary concern:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: datalake-fsx
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxLustreFileSystem.datalake-fsx
spec:
  deploymentType: PERSISTENT_1
  storageCapacityGib: 6000
  storageType: HDD
  perUnitStorageThroughput: 12
  dataCompressionType: LZ4
  subnetId: subnet-private-az1
  securityGroupIds:
    - sg-lustre-data
  automaticBackupRetentionDays: 14
  dailyAutomaticBackupStartTime: "02:00"
  copyTagsToBackups: true
  weeklyMaintenanceStartTime: "1:05:00"
  logConfiguration:
    destination: arn:aws:logs:us-east-1:123456789012:log-group:/aws/fsx/datalake
    level: WARN_ERROR
```

### Full-Featured with Logging and Custom Metadata IOPS

Production PERSISTENT_2 deployment with CloudWatch audit logging, customer-managed KMS encryption, explicit metadata IOPS, and final backup on deletion:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: prod-fsx
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxLustreFileSystem.prod-fsx
spec:
  region: us-east-1
  deploymentType: PERSISTENT_2
  storageCapacityGib: 7200
  storageType: SSD
  perUnitStorageThroughput: 500
  dataCompressionType: LZ4
  fileSystemTypeVersion: "2.15"
  subnetId: subnet-private-az1
  securityGroupIds:
    - sg-lustre-prod
    - sg-lustre-admin
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/mrk-prod-key
  automaticBackupRetentionDays: 30
  dailyAutomaticBackupStartTime: "03:00"
  copyTagsToBackups: true
  skipFinalBackup: false
  weeklyMaintenanceStartTime: "7:02:00"
  logConfiguration:
    destination: arn:aws:logs:us-east-1:123456789012:log-group:/aws/fsx/prod
    level: WARN_ERROR
  metadataConfiguration:
    mode: USER_PROVISIONED
    iops: 12000
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: ref-fsx
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxLustreFileSystem.ref-fsx
spec:
  region: us-west-2
  deploymentType: PERSISTENT_2
  storageCapacityGib: 2400
  storageType: SSD
  perUnitStorageThroughput: 250
  subnetId:
    valueFrom:
      kind: AwsSubnet
      name: my-private-subnet-a
      fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: lustre-sg
        field: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: fsx-key
      field: status.outputs.key_arn
  logConfiguration:
    destination:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: fsx-logs
        field: status.outputs.log_group_arn
    level: WARN_ERROR
```

## Presets

OpenMCF includes preset configurations for common FSx Lustre deployment patterns. Each preset is a ready-to-customize manifest with placeholder values for subnet and security group IDs.

### Scratch Development

**File**: `presets/01-scratch-development.yaml`

SCRATCH_2 SSD with 1200 GiB — the smallest and cheapest Lustre configuration. No data replication, no backups. Use for development and test environments, short-lived batch processing, CI/CD scratch space, and experiments where data loss is acceptable.

### Persistent High Throughput

**File**: `presets/02-persistent-high-throughput.yaml`

PERSISTENT_2 SSD with 2400 GiB and 1000 MB/s/TiB throughput. LZ4 compression, 7-day automatic backups at 04:00 UTC, AUTOMATIC metadata IOPS, and Sunday 03:00 UTC maintenance window. Use for distributed ML training, HPC simulations, video rendering, and production workloads requiring maximum I/O performance with data durability.

### Persistent Capacity Data Lake

**File**: `presets/03-persistent-capacity-datalake.yaml`

PERSISTENT_1 HDD with 6000 GiB and 12 MB/s/TiB throughput. LZ4 compression, 14-day automatic backups at 02:00 UTC, and Monday 05:00 UTC maintenance window. Use for data lake staging, genomics pipelines, log analysis, and workloads where cost per GiB matters more than latency.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `file_system_id` | `string` | ID of the file system (e.g., `fs-0123456789abcdef0`). Primary identifier for EKS PersistentVolumes, ECS task definitions, and AWS Batch compute environments. |
| `file_system_arn` | `string` | ARN of the file system. Used in IAM policies and for creating data repository associations. |
| `dns_name` | `string` | DNS name for the file system (e.g., `fs-0123456789abcdef0.fsx.us-east-1.amazonaws.com`). Used in mount commands with `mount_name`. |
| `mount_name` | `string` | Lustre mount name (e.g., `fsx` or `2p5wpbwj`). Auto-generated by AWS. Mount command: `mount -t lustre <dns_name>@tcp:/<mount_name> /mnt/fsx`. |
| `network_interface_ids` | `string[]` | Network interface IDs created for the file system. Lustre creates one ENI in the specified subnet. Useful for security group debugging and network troubleshooting. |
| `vpc_id` | `string` | VPC ID in which the file system was created. Computed from the subnet. |
| `file_system_type_version` | `string` | Actual Lustre version deployed (e.g., `2.12`, `2.15`). May differ from the requested version if the field was left empty. |
| `owner_id` | `string` | AWS account ID of the file system owner. |

## Related Components

- [AwsElasticFileSystem](/docs/catalog/aws/elastic-file-system) — alternative managed file system for general-purpose NFS workloads
- [AwsVpc](/docs/catalog/aws/vpc) — provides the subnet for file system placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls Lustre protocol traffic (TCP 988, 1018-1023) to the file system
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides a customer-managed encryption key for data at rest
- [AwsCloudwatchLogGroup](/docs/catalog/aws/cloudwatch-log-group) — receives audit log events from the file system
