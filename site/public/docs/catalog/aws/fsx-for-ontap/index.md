---
title: "FSx for ONTAP"
description: "FSx for ONTAP deployment documentation"
icon: "package"
order: 100
componentName: "awsfsxontapfilesystem"
---

# AWS FSx for ONTAP

Deploys an Amazon FSx for NetApp ONTAP file system — a fully managed shared storage with multi-protocol access (NFS, SMB, iSCSI), HA pairs for scale-out throughput, and optional multi-AZ failover. This component provisions the file system only; Storage Virtual Machines (SVMs) and volumes are managed by separate components. Use the management endpoint for ONTAP CLI access and SnapMirror replication.

## What Gets Created

When you deploy an AwsFsxOntapFileSystem resource, OpenMCF provisions:

- **FSx for ONTAP File System** — an `aws_fsx_ontap_file_system` resource placed in the specified subnets with encryption at rest, throughput and storage capacity as configured, tagged with OpenMCF resource metadata
- **Disk IOPS Configuration** — created only when `diskIopsConfiguration` is specified; controls SSD IOPS in AUTOMATIC or USER_PROVISIONED mode
- **Multi-AZ Route Entries** — created only for MULTI_AZ_1 or MULTI_AZ_2 deployments; AWS manages routes in the specified route tables for automatic failover

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **One subnet** for SINGLE_AZ deployments, or **two subnets in different AZs** for MULTI_AZ_1 or MULTI_AZ_2
- **A security group** allowing traffic between clients and the file system: TCP 111 (portmapper), 635 (mountd), 2049 (NFS), 4045–4046 (NFS lock/status), 445 (SMB), 3260 (iSCSI), 443 (ONTAP REST API)
- **A KMS key ARN** if using customer-managed encryption at rest (optional — AWS-managed key used by default)
- **Preferred subnet, endpoint IP range, and route table IDs** if deploying multi-AZ (for automatic failover routing)
- **ONTAP admin password** (8–50 characters) if ONTAP CLI or REST API access is needed

## Quick Start

Create a file `ontap.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapFileSystem
metadata:
  name: my-ontap
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsFsxOntapFileSystem.my-ontap
spec:
  region: us-east-1
  storageCapacityGib: 1024
  throughputCapacityPerHaPair: 128
  subnetIds:
    - subnet-0123456789abcdef0
```

Deploy:

```shell
openmcf apply -f ontap.yaml
```

This creates a SINGLE_AZ_2 ONTAP file system with 1024 GiB SSD storage, 128 MB/s throughput per HA pair, and one HA pair in the specified subnet.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the FSx ONTAP file system will be created (e.g., `us-east-1`). | Required; non-empty |
| `storageCapacityGib` | `int32` | Storage capacity in GiB. Can be increased after creation but never decreased. | 1024–1048576 (1 TiB – 1 PiB) |
| `throughputCapacityPerHaPair` | `int32` | Throughput per HA pair in MB/s. Total throughput = this value × number of HA pairs. | One of: 128, 256, 384, 512, 768, 1024, 1536, 2048, 3072, 4096, 6144 |
| `subnetIds` | `StringValueOrRef[]` | Subnet IDs. 1 for single-AZ, 2 for multi-AZ. Can reference AwsVpc via `valueFrom`. | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `deploymentType` | `string` | `SINGLE_AZ_2` | `SINGLE_AZ_1`, `SINGLE_AZ_2`, `MULTI_AZ_1`, or `MULTI_AZ_2`. ForceNew. SINGLE_AZ_2 recommended for most workloads; MULTI_AZ_2 for HA. |
| `storageType` | `string` | `SSD` | `SSD` (sub-millisecond latency) or `HDD` (throughput-oriented with SSD cache). ForceNew. |
| `haPairs` | `int32` | `1` | Number of HA pairs (1–12). Single-AZ only; multi-AZ is fixed at 1. Scale-out for throughput. |
| `preferredSubnetId` | `StringValueOrRef` | — | Active file server subnet. Required for MULTI_AZ_1 and MULTI_AZ_2. ForceNew. Can reference AwsVpc via `valueFrom`. |
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | Security group IDs. Can reference AwsSecurityGroup via `valueFrom`. ForceNew. Up to 50. |
| `endpointIpAddressRange` | `string` | — | CIDR range for endpoint floating IPs. MULTI_AZ only. ForceNew. |
| `routeTableIds` | `StringValueOrRef[]` | `[]` | Route tables for failover routing. MULTI_AZ only. Up to 50. |
| `kmsKeyId` | `StringValueOrRef` | AWS-managed | Customer-managed KMS key ARN. Can reference AwsKmsKey via `valueFrom`. ForceNew. |
| `fsxAdminPassword` | `string` | — | Password for fsxadmin (ONTAP CLI and REST API). 8–50 characters. Sensitive; not returned in reads. |
| `diskIopsConfiguration.mode` | `string` | `AUTOMATIC` | `AUTOMATIC` (3 IOPS per GiB) or `USER_PROVISIONED` (explicit IOPS). |
| `diskIopsConfiguration.iops` | `int32` | — | Total SSD IOPS. Only when mode is `USER_PROVISIONED`. Range: 0–2,400,000. |
| `automaticBackupRetentionDays` | `int32` | `0` | Days to retain automatic backups (0–90). 0 disables. |
| `dailyAutomaticBackupStartTime` | `string` | — | Backup window in HH:MM UTC format. Requires `automaticBackupRetentionDays` > 0. |
| `copyTagsToBackups` | `bool` | `false` | Propagate file system tags to backups. |
| `skipFinalBackup` | `bool` | `true` | Skip final backup on deletion. |
| `weeklyMaintenanceStartTime` | `string` | — | Maintenance window in d:HH:MM UTC (1=Mon, 7=Sun). |

## Examples

### Single-AZ with Provisioned IOPS

A SINGLE_AZ_2 file system with USER_PROVISIONED IOPS, security groups, and daily backups:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapFileSystem
metadata:
  name: app-storage
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxOntapFileSystem.app-storage
spec:
  region: us-east-1
  deploymentType: SINGLE_AZ_2
  storageCapacityGib: 2048
  storageType: SSD
  throughputCapacityPerHaPair: 512
  haPairs: 1
  subnetIds:
    - subnet-private-az1
  securityGroupIds:
    - sg-ontap-access
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/my-key
  diskIopsConfiguration:
    mode: USER_PROVISIONED
    iops: 50000
  automaticBackupRetentionDays: 7
  dailyAutomaticBackupStartTime: "05:00"
  copyTagsToBackups: true
```

### Scale-Out Single-AZ with Multiple HA Pairs

A SINGLE_AZ_2 deployment with four HA pairs for higher aggregate throughput:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapFileSystem
metadata:
  name: throughput-tier
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxOntapFileSystem.throughput-tier
spec:
  region: us-east-1
  deploymentType: SINGLE_AZ_2
  storageCapacityGib: 8192
  throughputCapacityPerHaPair: 1024
  haPairs: 4
  subnetIds:
    - subnet-private-az1
  securityGroupIds:
    - sg-ontap-access
  fsxAdminPassword: <sensitive>
  weeklyMaintenanceStartTime: "7:02:00"
```

### Multi-AZ High Availability

A MULTI_AZ_2 deployment with automatic failover across two availability zones:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapFileSystem
metadata:
  name: ha-ontap
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxOntapFileSystem.ha-ontap
spec:
  region: us-east-1
  deploymentType: MULTI_AZ_2
  storageCapacityGib: 4096
  throughputCapacityPerHaPair: 512
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  preferredSubnetId: subnet-private-az1
  endpointIpAddressRange: 10.0.100.0/24
  routeTableIds:
    - rtb-private-az1
    - rtb-private-az2
  securityGroupIds:
    - sg-ontap-access
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/prod-key
  automaticBackupRetentionDays: 14
  dailyAutomaticBackupStartTime: "03:00"
  copyTagsToBackups: true
```

### Using Foreign Key References

Reference subnets, security groups, and KMS key from other OpenMCF-managed resources:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapFileSystem
metadata:
  name: ref-ontap
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxOntapFileSystem.ref-ontap
spec:
  region: us-east-1
  storageCapacityGib: 2048
  throughputCapacityPerHaPair: 256
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-a
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: ontap-sg
        field: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-key
      field: status.outputs.key_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `file_system_id` | `string` | File system ID (e.g., `fs-0123456789abcdef0`). Primary identifier for SVMs, volumes, and other integrations. |
| `file_system_arn` | `string` | ARN for IAM resource-level permissions and cross-account access. |
| `dns_name` | `string` | DNS name for the file system (e.g., `fs-xxx.fsx.us-east-1.amazonaws.com`). |
| `management_dns_name` | `string` | Management endpoint DNS name for ONTAP CLI (SSH) and REST API. Connect via `ssh fsxadmin@<management_dns_name>`. |
| `management_ip_addresses` | `string[]` | Management endpoint IP addresses. Alternative to DNS for direct IP access. |
| `intercluster_dns_name` | `string` | Intercluster endpoint DNS name for NetApp SnapMirror replication between FSx for ONTAP file systems. |
| `intercluster_ip_addresses` | `string[]` | Intercluster endpoint IP addresses for SnapMirror peering when DNS is unavailable. |
| `network_interface_ids` | `string[]` | ENI IDs created for the file system. Single-AZ: 1 per HA pair; multi-AZ: 2. |
| `vpc_id` | `string` | VPC where the file system resides. |
| `owner_id` | `string` | AWS account ID of the file system owner. |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides the subnets for file system placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls NFS, SMB, iSCSI, and management traffic to/from the file system
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides customer-managed encryption keys
- [AwsFsxOpenzfsFileSystem](/docs/catalog/aws/fsx-for-openzfs) — NFS-only FSx option with ZFS compression and quotas
- [AwsFsxLustreFileSystem](/docs/catalog/aws/fsx-lustre-file-system) — HPC-optimized file system with S3 integration
