# AWS FSx Windows File System

Deploys a fully managed Windows file system on Amazon FSx for Windows File Server with SMB protocol access, Active Directory integration for identity-based access control, configurable deployment types (single-AZ or multi-AZ with automatic failover), Windows ACLs, DNS aliases, and audit logging to CloudWatch. Every Windows file system must join an Active Directory domain -- AWS Managed Microsoft AD or self-managed -- and the deployment type is a one-way door that fixes the availability posture and which storage media are available.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **FSx Windows File System** -- an enterprise-grade SMB file system joined to an Active Directory domain, with configurable deployment type (SINGLE_AZ_1, SINGLE_AZ_2, MULTI_AZ_1), storage media (SSD or HDD), throughput capacity, and DNS aliases
- **Active Directory Join** -- the file system joins either an AWS Managed Microsoft AD (via `activeDirectoryId`) or a self-managed AD domain (via `selfManagedActiveDirectory`) for Windows ACL-based access control
- **Audit Log Configuration** -- created only when `auditLogConfiguration` is provided; tracks file access and file share access events to CloudWatch Logs for compliance monitoring
- **Disk IOPS Configuration** -- configured only when `diskIopsConfiguration` is provided; controls provisioned SSD IOPS in AUTOMATIC or USER_PROVISIONED mode
- **Automatic Backups** -- enabled by default with 7-day retention; configurable via `automaticBackupRetentionDays`
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **One or two subnets** in the target VPC. Single-AZ deployments require exactly one subnet. Multi-AZ (MULTI_AZ_1) requires two subnets in different Availability Zones plus a `preferredSubnetId`. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef.
- **A security group** that allows SMB traffic (TCP 445), WinRM (TCP 5985), and Active Directory communication (TCP/UDP 53, 88, 389, 636). Provide the ID directly or reference an AwsSecurityGroup Infra Component.
- **An Active Directory domain** -- every Windows file system must join an AD domain. Provide either an AWS Managed Microsoft AD directory ID (`activeDirectoryId`) or configure `selfManagedActiveDirectory` with domain name, DNS IPs, and join credentials (direct or via Secrets Manager ARN).
- **A KMS key** (optional) -- for customer-managed encryption at rest instead of the default AWS-managed FSx key. Provide the ARN directly or reference an AwsKmsKey Infra Component.
- **A CloudWatch log group** (optional) -- required only when enabling audit logging. The log group name must start with `/aws/fsx/`. Reference an AwsCloudwatchLogGroup Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **AWS FSx Windows File System**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single-AZ Development FSx Windows** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxWindowsFileSystem
metadata:
  name: corp-file-server
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  deploymentType: SINGLE_AZ_2
  storageCapacityGib: 256
  throughputCapacity: 32
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
  securityGroupIds:
    - value: "sg-0a1b2c3d4e5f00001"
  activeDirectoryId:
    value: "d-0123456789"
```

```shell
planton apply -f fsx-windows.yaml
```

This creates a single-AZ Windows file system with 256 GiB SSD storage, 32 MB/s throughput, joined to an AWS Managed Microsoft AD, 7-day automatic backup retention, and no audit logging. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the file system to a VPC and security group deployed in the same InfraPipeline:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-subnet-a
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: windows-smb-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-encryption-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the VPC, security group, and KMS key first, then provisions the Windows file system with the resolved values.

## Key Configuration

These are the most important decisions when configuring an FSx Windows file system. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Active Directory integration** -- Every Windows file system must join an AD domain. Use `activeDirectoryId` for AWS Managed Microsoft AD (simplest setup). Use `selfManagedActiveDirectory` for on-premises or EC2-hosted AD, providing the domain name, DNS server IPs, and join credentials. For production, use `domainJoinServiceAccountSecretArn` (Secrets Manager) instead of inline `username`/`password`.

**Deployment type** -- SINGLE_AZ_2 is the recommended default, supporting HDD storage and higher throughput tiers. MULTI_AZ_1 provides automatic failover across two AZs with floating DNS -- clients reconnect transparently after failover. This is a ForceNew attribute.

**Storage type and capacity** -- SSD provides sub-millisecond latency for active workloads. HDD is available on SINGLE_AZ_2 and MULTI_AZ_1 with a minimum of 2000 GiB, suitable for home directories and archival storage. Storage can be increased but never decreased. Alternatively, restore from an FSx backup with `backupId` (ForceNew) -- the file system inherits the backup's size, so capacity is not specified on a restore.

**Final backup tags** -- When `skipFinalBackup` is off, `finalBackupTags` are applied to the backup taken on deletion, so cost-allocation and retention markers survive the file system itself.

**Audit logging** -- Configure `auditLogConfiguration` to track file access events (open, read, write, delete) and file share access events (connect, disconnect, permission changes) at SUCCESS_ONLY, FAILURE_ONLY, or SUCCESS_AND_FAILURE levels. Logs are sent to a CloudWatch log group for compliance and security monitoring.

**DNS aliases** -- Add custom DNS names (up to 50) via `aliases` for DFS namespace integration, migration from on-premises filers, or user-friendly SMB mount points. Create CNAME records pointing each alias to the file system's DNS name.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSubnet** (optional, multi-AZ) | `preferredSubnetId` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsCloudwatchLogGroup** (optional) | `auditLogConfiguration.auditLogDestination` | `status.outputs.log_group_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `file_system_id` | FSx file system identifier | EKS SMB CSI driver, AWS Backup integration |
| `file_system_arn` | Amazon Resource Name of the file system | IAM policies for resource-level permissions |
| `dns_name` | AD-joined DNS name | SMB mount command: `net use Z: \\<dns_name>\share` |
| `preferred_file_server_ip` | Active file server IP address | DNS record creation, network troubleshooting |
| `remote_administration_endpoint` | Windows Remote PowerShell endpoint | `Enter-PSSession -ComputerName <endpoint> -ConfigurationName FsxRemoteAdmin` |
| `network_interface_ids` | ENI IDs for the file system | Security group debugging, network troubleshooting |
| `vpc_id` | VPC containing the file system | Security group rules, network placement verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single AZ development** -- SINGLE_AZ_2 with minimal SSD storage and throughput. Cost-effective for development, testing, and small team file shares that do not require cross-AZ redundancy. Start from the **Single-AZ Development FSx Windows** preset.

**Single AZ production** -- SINGLE_AZ_2 with larger storage, higher throughput, customer-managed KMS encryption, audit logging enabled, and automatic backups. Suitable for production home directories, .NET application data, and SQL Server databases. Start from the **Single-AZ Production FSx Windows** preset.

**Multi AZ high availability** -- MULTI_AZ_1 with automatic failover across two AZs, audit logging, and automatic backups. Designed for business-critical Windows workloads requiring transparent failover with floating DNS. Start from the **Multi-AZ High Availability FSx Windows** preset.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) -- provides subnets for file system network interface placement
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls SMB (TCP 445), WinRM (TCP 5985), and AD traffic access to the file system
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for encryption at rest
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- receives file access and share access audit events for compliance
