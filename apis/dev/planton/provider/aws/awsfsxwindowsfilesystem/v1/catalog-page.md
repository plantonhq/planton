# AWS FSx for Windows File Server

Deploys an Amazon FSx for Windows File Server with Active Directory integration, configurable throughput tiers, optional audit logging, and automatic backup management. Every file system joins an AD domain (AWS Managed or self-managed) for identity-based SMB access control.

## What Gets Created

When you deploy an AwsFsxWindowsFileSystem resource, Planton provisions:

- **Windows File System** — an `aws_fsx_windows_file_system` resource with the specified deployment type (SINGLE_AZ_1, SINGLE_AZ_2, or MULTI_AZ_1), storage capacity, and throughput
- **Active Directory Join** — the file system joins either an AWS Managed Microsoft AD (via `activeDirectoryId`) or a self-managed AD domain (via `selfManagedActiveDirectory` with domain credentials)
- **DNS Aliases** — created only when `aliases` is non-empty, associates custom DNS names with the file system for DFS namespace integration
- **Audit Log Configuration** — created only when audit log levels are set, enables file access and file share access audit logging to CloudWatch Logs
- **Disk IOPS Configuration** — created only when `diskIopsConfiguration` is specified, allows USER_PROVISIONED IOPS for SSD storage

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **At least one subnet** in a VPC (two subnets in different AZs for MULTI_AZ_1)
- **An Active Directory domain** — either an AWS Managed Microsoft AD instance or a self-managed AD with DNS connectivity from the file system's subnets
- **A security group** allowing SMB traffic (TCP 445) and AD communication (TCP/UDP 53, 88, 389, 636)
- **A KMS key ARN** if using customer-managed encryption (optional — AWS-managed encryption is the default)
- **A CloudWatch Logs log group** starting with `/aws/fsx/` if configuring audit logging to a custom destination

## Quick Start

Create a file `windows-fs.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: my-windows-fs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsFsxWindowsFileSystem.my-windows-fs
spec:
  region: us-west-2
  storageCapacityGib: 32
  throughputCapacity: 32
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
  activeDirectoryId:
    value: d-0123456789
```

Deploy:

```shell
planton apply -f windows-fs.yaml
```

This creates a SINGLE_AZ_2 Windows file system with 32 GiB SSD storage, 32 MB/s throughput, joined to an AWS Managed Microsoft AD domain, with 7-day automatic backup retention.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the file system will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `storageCapacityGib` | `int` | Storage capacity in GiB. SSD: 32-65536. HDD: 2000-65536. | Minimum 32 |
| `throughputCapacity` | `int` | Throughput in MB/s. | Must be one of: 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4608, 6144, 9216, 12288 |
| `subnetIds` | `string[]` | Subnet IDs for the file system. 1 for single-AZ, 2 for MULTI_AZ_1. Can reference AwsVpc via `valueFrom`. | Minimum 1 item |
| `activeDirectoryId` OR `selfManagedActiveDirectory` | — | Exactly one must be specified. Every Windows file system must join an AD domain. | — |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `deploymentType` | `string` | `SINGLE_AZ_2` | Deployment type: `SINGLE_AZ_1`, `SINGLE_AZ_2`, or `MULTI_AZ_1`. ForceNew. |
| `storageType` | `string` | `SSD` | Storage media: `SSD` or `HDD`. HDD requires SINGLE_AZ_2/MULTI_AZ_1 and min 2000 GiB. ForceNew. |
| `preferredSubnetId` | `string` | — | Preferred file server subnet for MULTI_AZ_1. Required for multi-AZ. Can reference AwsVpc via `valueFrom`. ForceNew. |
| `securityGroupIds` | `string[]` | `[]` | Security group IDs. Must allow SMB (TCP 445) and AD ports. Can reference AwsSecurityGroup via `valueFrom`. ForceNew. |
| `kmsKeyId` | `string` | — | Customer-managed KMS key ARN. Defaults to AWS-managed FSx key. Can reference AwsKmsKey via `valueFrom`. ForceNew. |
| `aliases` | `string[]` | `[]` | DNS alias names (4-253 chars each, max 50). Requires CNAME records pointing to the file system DNS name. |
| `auditLogConfiguration.fileAccessAuditLogLevel` | `string` | `DISABLED` | File access audit: `DISABLED`, `SUCCESS_ONLY`, `FAILURE_ONLY`, `SUCCESS_AND_FAILURE`. |
| `auditLogConfiguration.fileShareAccessAuditLogLevel` | `string` | `DISABLED` | File share access audit: `DISABLED`, `SUCCESS_ONLY`, `FAILURE_ONLY`, `SUCCESS_AND_FAILURE`. |
| `auditLogConfiguration.auditLogDestination` | `string` | — | CloudWatch Logs group ARN (must start with `/aws/fsx/`). Can reference AwsCloudwatchLogGroup via `valueFrom`. |
| `diskIopsConfiguration.mode` | `string` | `AUTOMATIC` | IOPS mode: `AUTOMATIC` (scales with storage) or `USER_PROVISIONED`. |
| `diskIopsConfiguration.iops` | `int` | — | Provisioned IOPS (0-350000). Required when mode is `USER_PROVISIONED`. |
| `automaticBackupRetentionDays` | `int` | `7` | Days to retain backups (0-90). Set to 0 to disable. |
| `dailyAutomaticBackupStartTime` | `string` | — | Daily backup window in `HH:MM` UTC format. |
| `copyTagsToBackups` | `bool` | `false` | Copy tags to backup snapshots. ForceNew. |
| `skipFinalBackup` | `bool` | `true` | Skip final backup on deletion. |
| `weeklyMaintenanceStartTime` | `string` | — | Weekly maintenance window in `d:HH:MM` UTC format (1=Mon, 7=Sun). |

### Self-Managed AD Fields

When using `selfManagedActiveDirectory` instead of `activeDirectoryId`:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `selfManagedActiveDirectory.domainName` | `string` | — | FQDN of the AD domain (e.g., `corp.example.com`). Required. |
| `selfManagedActiveDirectory.dnsIps` | `string[]` | — | DNS server IPs (1-2). Must be reachable from file system subnets. Required. |
| `selfManagedActiveDirectory.username` | `string` | — | Service account username. Mutually exclusive with `domainJoinServiceAccountSecretArn`. |
| `selfManagedActiveDirectory.password` | `string` | — | Service account password. Mutually exclusive with `domainJoinServiceAccountSecretArn`. |
| `selfManagedActiveDirectory.domainJoinServiceAccountSecretArn` | `string` | — | Secrets Manager secret ARN with credentials JSON. Mutually exclusive with username/password. |
| `selfManagedActiveDirectory.fileSystemAdministratorsGroup` | `string` | `Domain Admins` | AD group with administrative privileges on the file system. |
| `selfManagedActiveDirectory.organizationalUnitDistinguishedName` | `string` | — | OU DN for the file system's computer object. |

## Examples

### AWS Managed AD

Minimal configuration using an existing AWS Directory Service managed AD:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: dev-windows-fs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsFsxWindowsFileSystem.dev-windows-fs
spec:
  region: us-west-2
  storageCapacityGib: 32
  throughputCapacity: 32
  subnetIds:
    - subnet-az1
  securityGroupIds:
    - sg-fsx
  activeDirectoryId:
    value: d-0123456789
  automaticBackupRetentionDays: 0
```

### Self-Managed AD with Audit Logging

Production file system joined to an on-premises AD with compliance audit logging:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: prod-windows-fs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsFsxWindowsFileSystem.prod-windows-fs
spec:
  region: us-east-1
  storageCapacityGib: 500
  throughputCapacity: 256
  subnetIds:
    - subnet-private-az1
  securityGroupIds:
    - sg-fsx-prod
  kmsKeyId:
    value: arn:aws:kms:us-east-1:123456789012:key/your-key
  selfManagedActiveDirectory:
    domainName: corp.example.com
    dnsIps:
      - "10.0.0.10"
      - "10.0.0.11"
    username: svc-fsx-join
    password: "<service-account-password>"
    fileSystemAdministratorsGroup: FSx Admins
  auditLogConfiguration:
    fileAccessAuditLogLevel: SUCCESS_AND_FAILURE
    fileShareAccessAuditLogLevel: FAILURE_ONLY
  automaticBackupRetentionDays: 7
  dailyAutomaticBackupStartTime: "01:00"
  copyTagsToBackups: true
  weeklyMaintenanceStartTime: "7:02:00"
```

### Multi-AZ High Availability

Mission-critical deployment with automatic failover, DNS aliases, and provisioned IOPS:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: ha-windows-fs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsFsxWindowsFileSystem.ha-windows-fs
spec:
  region: us-east-1
  deploymentType: MULTI_AZ_1
  storageCapacityGib: 1000
  throughputCapacity: 512
  subnetIds:
    - subnet-az1
    - subnet-az2
  preferredSubnetId:
    value: subnet-az1
  securityGroupIds:
    - sg-fsx-ha
  kmsKeyId:
    value: arn:aws:kms:us-east-1:123456789012:key/prod-key
  activeDirectoryId:
    value: d-0123456789
  aliases:
    - finance.corp.example.com
    - shared.corp.example.com
  auditLogConfiguration:
    fileAccessAuditLogLevel: SUCCESS_AND_FAILURE
    fileShareAccessAuditLogLevel: SUCCESS_AND_FAILURE
  diskIopsConfiguration:
    mode: USER_PROVISIONED
    iops: 100000
  automaticBackupRetentionDays: 30
  dailyAutomaticBackupStartTime: "03:00"
  copyTagsToBackups: true
  weeklyMaintenanceStartTime: "7:02:00"
```

### Cross-Resource References

Wire the file system to other Planton-managed resources using `valueFrom`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: wired-windows-fs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsFsxWindowsFileSystem.wired-windows-fs
spec:
  region: us-east-1
  storageCapacityGib: 200
  throughputCapacity: 128
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-a
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: fsx-sg
        field: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: fsx-key
      field: status.outputs.key_arn
  activeDirectoryId:
    value: d-0123456789
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `file_system_id` | `string` | The file system ID (e.g., `fs-0123456789abcdef0`) |
| `file_system_arn` | `string` | ARN for IAM policies and resource-level permissions |
| `dns_name` | `string` | DNS name for SMB mount commands (e.g., `fs-012345.corp.example.com`) |
| `preferred_file_server_ip` | `string` | IP of the active file server. For MULTI_AZ_1, follows the active server during failover |
| `remote_administration_endpoint` | `string` | Endpoint for Windows Remote PowerShell administration (`Enter-PSSession`) |
| `network_interface_ids` | `string[]` | ENI IDs (1 for single-AZ, 2 for multi-AZ). Useful for network troubleshooting |
| `vpc_id` | `string` | VPC in which the file system was created |
| `owner_id` | `string` | AWS account ID of the file system owner |

## Related Components

- [AwsVpc](/docs/catalog/aws/awsvpc) — provides the subnets for file system placement
- [AwsSecurityGroup](/docs/catalog/aws/awssecuritygroup) — controls SMB and AD traffic to the file system
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides customer-managed encryption keys
- [AwsCloudwatchLogGroup](/docs/catalog/aws/awscloudwatchloggroup) — receives audit log events
- [AwsFsxOpenzfsFileSystem](/docs/catalog/aws/awsfsxopenzfsfilesystem) — NFS alternative for Linux workloads
