# AwsFsxWindowsFileSystem

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsFsxWindowsFileSystemSpec defines the desired configuration for an Amazon FSx
for Windows File Server — a fully managed, enterprise-grade Windows file system
accessible over the industry-standard SMB (Server Message Block) protocol. It is
built on Windows Server and integrates with Microsoft Active Directory for
identity-based access control, Windows ACLs, DFS namespaces, and shadow copies.

FSx for Windows delivers up to 12 GB/s throughput and millions of IOPS on SSD
storage, making it suitable for Windows-native workloads including .NET
applications, SQL Server databases, home directories, media processing, and
enterprise content management systems.

Key design notes:
- Every Windows file system MUST join an Active Directory domain. Specify
  exactly one of `active_directory_id` (AWS Managed Microsoft AD) or
  `self_managed_active_directory` (self-managed / on-premises AD).
- `deployment_type`, `storage_type`, `subnet_ids`, `preferred_subnet_id`,
  `security_group_ids`, `kms_key_id`, and `copy_tags_to_backups` are ForceNew —
  changing them requires replacing the file system.
- SINGLE_AZ_2 is the recommended deployment type for new workloads (latest
  generation, higher throughput ceiling). MULTI_AZ_1 provides automatic failover
  across AZs for high availability.
- HDD storage is only supported on SINGLE_AZ_2 and MULTI_AZ_1. HDD requires a
  minimum of 2000 GiB storage capacity.
- Throughput capacity is an absolute MB/s value from a fixed set of valid tiers.
- DNS aliases allow the file system to be accessed via custom DNS names
  (e.g., for DFS namespace integration or migration from on-premises filers).
- Audit logging tracks file access and file share access events to CloudWatch
  Logs for compliance and security monitoring.
- Credentials, region, and deployment workflow live outside this spec in stack inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxWindowsFileSystem
metadata:
  org: test-org
  env: test
  name: test-windows-fs
  id: awsfxw-test-windows-fs-test
spec:
  region: us-west-2
  # Full-surface SINGLE_AZ_2 shape so the offline plan proof covers the arms
  # the live lanes exclude: the self-managed AD Secrets Manager join, audit
  # logging, user-provisioned IOPS, aliases, and backups with final-backup
  # tags.
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 64
  storage_type: SSD
  throughput_capacity: 32
  subnet_ids:
    - value: subnet-test123
  security_group_ids:
    - value: sg-test123
  kms_key_id:
    value: arn:aws:kms:us-west-2:123456789012:key/00000000-0000-0000-0000-000000000000
  self_managed_active_directory:
    domain_name: corp.example.com
    dns_ips:
      - 10.0.0.1
      - 10.0.0.2
    domain_join_service_account_secret_arn:
      value: arn:aws:secretsmanager:us-west-2:123456789012:secret:fsx-ad-join-abc123
    organizational_unit_distinguished_name: OU=FSx,DC=corp,DC=example,DC=com
  aliases:
    - files.corp.example.com
  audit_log_configuration:
    file_access_audit_log_level: SUCCESS_AND_FAILURE
    file_share_access_audit_log_level: FAILURE_ONLY
    audit_log_destination:
      value: arn:aws:logs:us-west-2:123456789012:log-group:/aws/fsx/windows
  disk_iops_configuration:
    mode: USER_PROVISIONED
    iops: 5000
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "03:00"
  copy_tags_to_backups: true
  skip_final_backup: false
  final_backup_tags:
    retention: decommission-audit
  weekly_maintenance_start_time: "7:02:00"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.deploymentType` | `string` |  | `SINGLE_AZ_2` |  |
| `spec.storageCapacityGib` | `int32` |  |  |  |
| `spec.storageType` | `string` |  | `SSD` |  |
| `spec.throughputCapacity` | `int32` |  |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.preferredSubnetId` | `string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.backupId` | `string` |  |  |  |
| `spec.activeDirectoryId` | `string \| valueFrom` |  |  |  |
| `spec.selfManagedActiveDirectory` | `AwsFsxWindowsFileSystemSelfManagedActiveDirectory` |  |  |  |
| `spec.selfManagedActiveDirectory.domainName` | `string` | yes |  |  |
| `spec.selfManagedActiveDirectory.dnsIps` | `[]string` | yes |  |  |
| `spec.selfManagedActiveDirectory.username` | `string` |  |  |  |
| `spec.selfManagedActiveDirectory.password` | `string` (sensitive) |  |  |  |
| `spec.selfManagedActiveDirectory.domainJoinServiceAccountSecretArn` | `string \| valueFrom` |  |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.selfManagedActiveDirectory.fileSystemAdministratorsGroup` | `string` |  | `Domain Admins` |  |
| `spec.selfManagedActiveDirectory.organizationalUnitDistinguishedName` | `string` |  |  |  |
| `spec.aliases` | `[]string` |  |  |  |
| `spec.auditLogConfiguration` | `AwsFsxWindowsFileSystemAuditLogConfiguration` |  |  |  |
| `spec.auditLogConfiguration.fileAccessAuditLogLevel` | `string` |  | `DISABLED` |  |
| `spec.auditLogConfiguration.fileShareAccessAuditLogLevel` | `string` |  | `DISABLED` |  |
| `spec.auditLogConfiguration.auditLogDestination` | `string \| valueFrom` |  |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.diskIopsConfiguration` | `AwsFsxWindowsFileSystemDiskIopsConfiguration` |  |  |  |
| `spec.diskIopsConfiguration.mode` | `string` |  | `AUTOMATIC` |  |
| `spec.diskIopsConfiguration.iops` | `int32` |  |  |  |
| `spec.automaticBackupRetentionDays` | `int32` |  | `7` |  |
| `spec.dailyAutomaticBackupStartTime` | `string` |  |  |  |
| `spec.copyTagsToBackups` | `bool` |  |  |  |
| `spec.skipFinalBackup` | `bool` |  | `true` |  |
| `spec.finalBackupTags` | `map<string, string>` |  |  |  |
| `spec.weeklyMaintenanceStartTime` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.deploymentType

`string` · optional (explicit presence)

Deployment type controlling availability and performance characteristics.
ForceNew — cannot be changed after creation.

- "SINGLE_AZ_1": first-generation single-AZ. Limited throughput tiers.
- "SINGLE_AZ_2": latest single-AZ with higher throughput ceiling and HDD
  support. Recommended for most workloads.
- "MULTI_AZ_1": multi-AZ with automatic failover across two AZs. Requires
  two subnets and a preferred_subnet_id. Supports HDD storage.

Default: SINGLE_AZ_2

- default: `SINGLE_AZ_2`

### spec.storageCapacityGib

`int32` · optional (explicit presence)

Storage capacity in GiB.

Valid ranges depend on storage type:
- SSD: 32–65536 GiB
- HDD: 2000–65536 GiB

Storage can be increased after creation but never decreased. Leave unset
only when restoring from a backup (`backup_id` — capacity comes from the
backup).

- rule: {"int32":{"lte":65536,"gte":32}}

### spec.storageType

`string` · optional (explicit presence)

Storage media type. ForceNew — cannot be changed after creation.

- "SSD": solid-state drives. Sub-millisecond latency. Required for
  SINGLE_AZ_1. Recommended for most workloads.
- "HDD": hard disk drives. Lower cost, higher latency. Only available for
  SINGLE_AZ_2 and MULTI_AZ_1 deployment types. Requires minimum 2000 GiB
  storage capacity.

The provider's shared storage-type enum also accepts INTELLIGENT_TIERING,
but AWS's CreateFileSystem contract scopes Intelligent-Tiering to OpenZFS
multi-AZ and Lustre PERSISTENT_2 file systems only — it is not valid for
Windows. The two-value domain here is deliberate (spec mirrors AWS where
the provider is loose).

Default: SSD

- default: `SSD`

### spec.throughputCapacity

`int32`

Throughput capacity in MB/s. Required.

Valid values: 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4608, 6144, 9216, 12288.
The maximum available throughput depends on the deployment type, and the
4608+ tiers are available only in a handful of regions (per the FSx for
Windows performance documentation).

Known provider defect at the pinned release: the provider's value list
carries a digit-transposition typo — it validates 12228 where AWS's
documented top tier is 12288 — so the 12288 tier fails the provider's
plan-time validation on both engines until a provider release fixes the
list. The spec mirrors AWS's contract, never the typo.

Throughput can be changed after creation to scale performance up or down.

- rule: {"int32":{"gt":0}}

### spec.subnetIds

`[]string | valueFrom` · required

Subnet IDs for the file system's network interfaces. Required. ForceNew.

- SINGLE_AZ_1 / SINGLE_AZ_2: exactly one subnet.
- MULTI_AZ_1: exactly two subnets in different availability zones.

All compute resources mounting this file system must have SMB network
connectivity to these subnets (TCP port 445).

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.preferredSubnetId

`string | valueFrom`

Preferred subnet for the active file server in a MULTI_AZ_1 deployment.
ForceNew. Required when deployment_type is MULTI_AZ_1. Must be one of the
subnets specified in subnet_ids.

In a failover event, the standby file server in the other subnet takes over.
Ignored for SINGLE_AZ deployments.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom`

Security groups for the file system's network interfaces. ForceNew.

Must allow SMB traffic between the file system and its clients:
- TCP port 445 (SMB)
- TCP port 5985 (WinRM for PowerShell remote administration)

Additionally, for Active Directory communication:
- TCP/UDP port 53 (DNS), TCP/UDP port 88 (Kerberos),
  TCP port 389 (LDAP), TCP port 636 (LDAPS)

Up to 50 security groups. When empty, AWS attaches the VPC's default
security group.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"50"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key ARN for encryption at rest. ForceNew — the KMS key
cannot be changed after creation. When omitted, the file system uses the
AWS-managed FSx key. All Windows file systems are encrypted at rest by
default; this field upgrades to a customer-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.backupId

`string`

ID of an FSx backup to restore this file system from ("backup-...").
ForceNew. When set, storage capacity and most settings come from the
backup; leave storage_capacity_gib unset.

### spec.activeDirectoryId

`string | valueFrom`

ID of an existing AWS Managed Microsoft AD (Directory Service) to join.
ForceNew. Mutually exclusive with `self_managed_active_directory`.

Use this when you have an AWS Directory Service managed AD already
provisioned. The file system joins the domain automatically.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.selfManagedActiveDirectory

`AwsFsxWindowsFileSystemSelfManagedActiveDirectory`

Self-managed Active Directory configuration for joining an on-premises or
EC2-hosted AD domain. Mutually exclusive with `active_directory_id`.

Use this when your AD domain controller runs outside AWS Directory Service
(e.g., on-premises AD, AD on EC2, or Azure AD DS).

- rule: specify either username/password or domain_join_service_account_secret_arn, not both
- rule: when using direct credentials, both username and password must be provided
- rule: file_system_administrators_group must be 1-256 characters when provided

### spec.selfManagedActiveDirectory.domainName

`string` · required

Fully qualified domain name of the self-managed AD directory.
Example: "corp.example.com"

- rule: {"string":{"minLen":"1"}}

### spec.selfManagedActiveDirectory.dnsIps

`[]string` · required

IP addresses of the DNS servers for the AD domain. Required.
Must be reachable from the file system's subnets (same VPC CIDR or
RFC 1918 private ranges). Minimum 1, maximum 2 IP addresses.

- rule: {"repeated":{"minItems":"1","maxItems":"2","items":{"string":{"ip":true}}}}

### spec.selfManagedActiveDirectory.username

`string`

Service account username for domain join operations. Mutually exclusive
with `domain_join_service_account_secret_arn`. Length: 1-256 characters.

- rule: {"string":{"maxLen":"256"}}

### spec.selfManagedActiveDirectory.password

`string` · sensitive

Service account password for domain join operations. Mutually exclusive
with `domain_join_service_account_secret_arn`. Length: 1-256 characters.

For production workloads, prefer `domain_join_service_account_secret_arn`
to avoid storing credentials in the resource manifest.

- rule: {"string":{"maxLen":"256"}}

### spec.selfManagedActiveDirectory.domainJoinServiceAccountSecretArn

`string | valueFrom`

ARN of an AWS Secrets Manager secret containing the service account
credentials for domain join. Mutually exclusive with `username`/`password`.

The secret must contain a JSON object with "username" and "password" keys.
This is the recommended approach for production deployments. A valueFrom
defaults to an AwsSecretsManagerSecret and reads its
status.outputs.secret_arn.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.selfManagedActiveDirectory.fileSystemAdministratorsGroup

`string` · optional (explicit presence)

Name of the AD group whose members are granted administrative privileges
on the file system. Members can administer the file system from a remote
PowerShell endpoint using the FSx Remote PowerShell.

Default: Domain Admins

- default: `Domain Admins`

### spec.selfManagedActiveDirectory.organizationalUnitDistinguishedName

`string`

Organizational Unit (OU) distinguished name within the AD directory where
the file system's computer object is created.

Example: "OU=FSx,DC=corp,DC=example,DC=com"

Only the OU immediately above the computer object can be specified.
If not provided, the computer object is created in the default "Computers"
container in the AD domain. Length: 1-2000 characters.

- rule: {"string":{"maxLen":"2000"}}

### spec.aliases

`[]string`

DNS alias names to associate with the file system. Up to 50 aliases.

Aliases allow the file system to be accessed via custom DNS names
(e.g., "finance.corp.example.com") in addition to the default DNS name.
Useful for DFS namespace integration, migration from on-premises filers,
or providing user-friendly mount points.

Each alias must be a valid DNS name (4-253 characters). You must create a
DNS CNAME record pointing each alias to the file system's DNS name.

- rule: {"repeated":{"maxItems":"50","items":{"string":{"minLen":"4","maxLen":"253"}}}}

### spec.auditLogConfiguration

`AwsFsxWindowsFileSystemAuditLogConfiguration`

Audit log configuration for tracking file access and file share access
events. Logs are sent to CloudWatch Logs for compliance and security
monitoring. When omitted, audit logging is disabled.

- rule: file_access_audit_log_level must be 'DISABLED', 'SUCCESS_ONLY', 'FAILURE_ONLY', or 'SUCCESS_AND_FAILURE'
- rule: file_share_access_audit_log_level must be 'DISABLED', 'SUCCESS_ONLY', 'FAILURE_ONLY', or 'SUCCESS_AND_FAILURE'

### spec.auditLogConfiguration.fileAccessAuditLogLevel

`string` · optional (explicit presence)

Logging level for individual file access events (open, read, write, delete,
rename, change permissions on files and folders).

- "DISABLED": no file access logging.
- "SUCCESS_ONLY": log successful access events only.
- "FAILURE_ONLY": log failed access attempts only (e.g., access denied).
- "SUCCESS_AND_FAILURE": log all access events.

Default: DISABLED

- default: `DISABLED`

### spec.auditLogConfiguration.fileShareAccessAuditLogLevel

`string` · optional (explicit presence)

Logging level for file share access events (connect to share, disconnect,
change share permissions).

- "DISABLED": no file share access logging.
- "SUCCESS_ONLY": log successful share access events only.
- "FAILURE_ONLY": log failed share access attempts only.
- "SUCCESS_AND_FAILURE": log all share access events.

Default: DISABLED

- default: `DISABLED`

### spec.auditLogConfiguration.auditLogDestination

`string | valueFrom`

CloudWatch Logs log group ARN to receive audit events. The log group must
start with "/aws/fsx/" as required by AWS. If not set when audit levels are
enabled, FSx creates a default log stream in the "/aws/fsx/windows" group.

Only valid when at least one audit log level is not DISABLED.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.diskIopsConfiguration

`AwsFsxWindowsFileSystemDiskIopsConfiguration`

SSD IOPS configuration for the file system. Controls the total provisioned
IOPS. When omitted, AWS uses AUTOMATIC mode which scales IOPS with storage.
Only applicable to SSD storage type.

- rule: mode must be 'AUTOMATIC' or 'USER_PROVISIONED'
- rule: iops can only be set when mode is 'USER_PROVISIONED'

### spec.diskIopsConfiguration.mode

`string` · optional (explicit presence)

IOPS provisioning mode.

- "AUTOMATIC": IOPS scale automatically based on storage capacity.
  Provides 3 IOPS per GiB of storage.
- "USER_PROVISIONED": you specify the exact IOPS via the `iops` field.
  Allows higher performance independent of storage size but at extra cost.

Default: AUTOMATIC

- default: `AUTOMATIC`

### spec.diskIopsConfiguration.iops

`int32` · optional (explicit presence)

Total SSD IOPS provisioned. Only valid when mode is "USER_PROVISIONED".

Valid range: 0–350000.

- rule: {"int32":{"lte":350000,"gte":0}}

### spec.automaticBackupRetentionDays

`int32` · optional (explicit presence)

Number of days to retain automatic backups. Range: 0-90. Set to 0 to
disable automatic backups.

Default: 7

- default: `7`

### spec.dailyAutomaticBackupStartTime

`string`

Daily UTC time to start automatic backups, in HH:MM format (e.g., "01:00").
If not specified and backups are enabled, AWS chooses a default window.

- rule: daily_automatic_backup_start_time must be in 24-hour HH:MM format (e.g., '01:00')

### spec.copyTagsToBackups

`bool`

Copy tags from the file system to backups. ForceNew.

### spec.skipFinalBackup

`bool` · optional (explicit presence)

Skip creating a final backup when the file system is deleted.

Default: true — deletion is clean by default; set to false to keep a
last-resort restore point (the final backup outlives the file system and
keeps billing until deleted).

- default: `true`

### spec.finalBackupTags

`map<string, string>`

Tags applied to the final backup taken on deletion. Only meaningful when
skip_final_backup is false.

### spec.weeklyMaintenanceStartTime

`string`

Weekly UTC maintenance window in the format "d:HH:MM" where d is the day of
the week (1=Monday, 7=Sunday). Example: "7:02:00" for Sunday at 02:00 UTC.
If not specified, AWS chooses a default window.

- rule: weekly_maintenance_start_time must be in d:HH:MM format where d is 1 (Monday) through 7 (Sunday), e.g., '7:02:00'

## Validation Rules

- `deployment_type_valid`: deployment_type must be 'SINGLE_AZ_1', 'SINGLE_AZ_2', or 'MULTI_AZ_1'
- `storage_type_valid`: storage_type must be 'SSD' or 'HDD'
- `hdd_requires_compatible_deployment`: storage_type 'HDD' is only supported with deployment_type 'SINGLE_AZ_2' or 'MULTI_AZ_1'
- `hdd_minimum_storage`: storage_type 'HDD' requires storage_capacity_gib >= 2000
- `storage_capacity_required`: storage_capacity_gib is required (leave it unset only when restoring from backup_id)
- `backup_restore_excludes_capacity`: storage_capacity_gib cannot be set when restoring from backup_id (capacity comes from the backup)
- `throughput_capacity_valid`: throughput_capacity must be one of: 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4608, 6144, 9216, 12288
- `preferred_subnet_multi_az_contract`: preferred_subnet_id is required for MULTI_AZ_1 and can only be set for MULTI_AZ_1
- `subnet_count_matches_deployment`: MULTI_AZ_1 requires exactly two subnets; SINGLE_AZ deployment types require exactly one
- `ad_required`: exactly one of active_directory_id or self_managed_active_directory must be specified (Active Directory is mandatory for Windows File Server)
- `audit_destination_requires_enabled_level`: audit_log_configuration.audit_log_destination cannot be set when both log levels are DISABLED
- `backup_time_requires_retention`: daily_automatic_backup_start_time requires automatic_backup_retention_days > 0

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsFsxWindowsFileSystem, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.file_system_id` | `string` | The ID of the file system (e.g., "fs-0123456789abcdef0"). Primary identifier used by EKS SMB CSI driver, AWS Backup, and other AWS services. |
| `status.outputs.file_system_arn` | `string` | The Amazon Resource Name of the file system. Used in IAM policies for resource-level permissions. |
| `status.outputs.dns_name` | `string` | The DNS name for the file system (e.g., "fs-0123456789abcdef0.corp.example.com" for AD-joined file systems). Used in SMB mount commands: net use Z: \\<dns_name>\share |
| `status.outputs.preferred_file_server_ip` | `string` | The IP address of the preferred (active) file server. For MULTI_AZ_1, this is the active server's IP; in failover, the standby takes over. For SINGLE_AZ, this is the primary ENI IP. Useful for DNS record creation and network troubleshooting. |
| `status.outputs.remote_administration_endpoint` | `string` | The endpoint for remote administration via Windows Remote PowerShell. For MULTI_AZ_1: a floating endpoint that follows the active file server. For SINGLE_AZ: the file system's DNS name. Connect with: Enter-PSSession -ComputerName <endpoint> -ConfigurationName FsxRemoteAdmin |
| `status.outputs.network_interface_ids` | `[]string` | The network interface IDs created for the file system, in order. SINGLE_AZ creates 1 ENI; MULTI_AZ creates 2 ENIs. Useful for security group debugging and network troubleshooting. |
| `status.outputs.vpc_id` | `string` | The VPC ID in which the file system was created. Computed from the subnets. Useful for constructing security group rules and verifying network placement. |
| `status.outputs.owner_id` | `string` | The AWS account ID of the file system owner. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.preferredSubnetId` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.selfManagedActiveDirectory.domainJoinServiceAccountSecretArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.auditLogConfiguration.auditLogDestination` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |

## See Also

- [Overview](../README.md)
