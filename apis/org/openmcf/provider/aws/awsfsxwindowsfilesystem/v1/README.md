# AwsFsxWindowsFileSystem

Amazon FSx for Windows File Server — a fully managed, enterprise-grade Windows file system accessible over the industry-standard SMB (Server Message Block) protocol. Built on Windows Server, it integrates natively with Microsoft Active Directory for identity-based access control, Windows ACLs, DFS namespaces, and shadow copies.

## What It Is

FSx for Windows File Server provides shared file storage for Windows-based workloads. It delivers up to 12 GB/s throughput and millions of IOPS on SSD storage, with full support for SMB 2.0–3.1.1, NTFS, Windows ACLs, and DFS namespaces. Every file system joins an Active Directory domain, enabling seamless identity-based access for Windows and Linux SMB clients.

This component provisions the FSx for Windows file system, its network interfaces (ENIs) in the specified subnets, Active Directory domain join, optional audit logging, backup configuration, DNS aliases, and disk IOPS tuning.

## When to Use It

| Use Case | Description |
|----------|-------------|
| **Windows home directories** | Centralized user profiles and home folders accessible from domain-joined Windows desktops. |
| **.NET applications** | Shared configuration, assets, and data files for ASP.NET or Windows-native applications running on EC2 or ECS. |
| **SQL Server databases** | SMB file shares for SQL Server database files, transaction logs, and backups. |
| **Enterprise content management** | SharePoint, document management systems, or media pipelines requiring Windows ACLs. |
| **EKS Windows containers** | Use the SMB CSI driver to mount FSx Windows file shares as PersistentVolumes in Windows pods. |
| **Migration from on-premises** | Replace on-premises Windows file servers with managed FSx. DNS aliases and DFS namespaces enable transparent migration. |

## When NOT to Use It

| Need | Use Instead |
|------|-------------|
| **Linux shared NFS storage** (POSIX, multi-AZ, auto-scaling) | Amazon EFS — NFS protocol, elastic storage. |
| **High performance parallel I/O** (HPC, ML training) | Amazon FSx for Lustre — parallel file system, sub-ms latency, hundreds of GB/s. |
| **Object storage** (blobs, backups, static assets) | Amazon S3 — unlimited scale, REST API. |
| **Block storage** (databases, boot volumes) | Amazon EBS — single-instance, lowest latency. |
| **General-purpose ZFS** (snapshots, clones, compression) | Amazon FSx for OpenZFS. |

## Prerequisites

- **AWS account** with permissions to create FSx file systems, ENIs, and security groups.
- **VPC with subnets** — one subnet for SINGLE_AZ deployments, two subnets in different AZs for MULTI_AZ_1.
- **Security groups** — must allow:
  - TCP port 445 (SMB)
  - TCP port 5985 (WinRM / PowerShell remote administration)
  - TCP/UDP port 53 (DNS), TCP/UDP port 88 (Kerberos), TCP port 389 (LDAP), TCP port 636 (LDAPS) for AD communication
- **Active Directory** — exactly one of:
  - AWS Managed Microsoft AD (Directory Service) — `active_directory_id`
  - Self-managed / on-premises AD — `self_managed_active_directory`

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `deployment_type` | string | No | `SINGLE_AZ_1`, `SINGLE_AZ_2` (default), or `MULTI_AZ_1`. **ForceNew**. |
| `storage_capacity_gib` | int32 | **Yes** | Storage in GiB. SSD: 32–65536. HDD: 2000–65536. Can increase but never decrease. |
| `storage_type` | string | No | `SSD` (default) or `HDD`. **ForceNew**. HDD only with SINGLE_AZ_2 or MULTI_AZ_1. |
| `throughput_capacity` | int32 | **Yes** | MB/s. Valid: 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4608, 6144, 9216, 12288. Can be changed after creation. |
| `subnet_ids` | []StringValueOrRef | **Yes** | Subnet IDs. One for SINGLE_AZ, two for MULTI_AZ_1. **ForceNew**. |
| `preferred_subnet_id` | StringValueOrRef | Conditional | Active file server subnet for MULTI_AZ_1. Must be in `subnet_ids`. **ForceNew**. |
| `security_group_ids` | []StringValueOrRef | No | Security groups for the ENIs (up to 50). **ForceNew**. |
| `kms_key_id` | StringValueOrRef | No | Customer-managed KMS key ARN. **ForceNew**. Omit for AWS-managed key. |
| `active_directory_id` | StringValueOrRef | Conditional | AWS Managed AD ID. Mutually exclusive with `self_managed_active_directory`. |
| `self_managed_active_directory` | Object | Conditional | Self-managed AD config. Mutually exclusive with `active_directory_id`. |
| `aliases` | []string | No | DNS alias names (up to 50). Create CNAME records pointing to the file system's DNS name. |
| `audit_log_configuration` | Object | No | Audit logging for file and share access events. |
| `disk_iops_configuration` | Object | No | SSD IOPS config. AUTOMATIC (default) or USER_PROVISIONED. |
| `automatic_backup_retention_days` | int32 | No | Backup retention 0–90 days. Default: 7. Set 0 to disable. |
| `daily_automatic_backup_start_time` | string | No | Backup window in `HH:MM` UTC. |
| `copy_tags_to_backups` | bool | No | Copy tags to backups. **ForceNew**. |
| `skip_final_backup` | bool | No | Skip final backup on deletion. Default: true. |
| `weekly_maintenance_start_time` | string | No | Maintenance window in `d:HH:MM` format (1=Mon, 7=Sun). |

### Self-Managed Active Directory Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `domain_name` | string | **Yes** | FQDN of the AD domain (e.g., `corp.example.com`). |
| `dns_ips` | []string | **Yes** | DNS server IPs for the domain (1–2 IPs). |
| `username` | string | Conditional | Service account username. Mutually exclusive with `domain_join_service_account_secret_arn`. |
| `password` | string | Conditional | Service account password. Mutually exclusive with `domain_join_service_account_secret_arn`. |
| `domain_join_service_account_secret_arn` | StringValueOrRef | Conditional | Secrets Manager ARN with domain join credentials. Recommended for production. |
| `file_system_administrators_group` | string | No | AD group for file system admin privileges. Default: `Domain Admins`. |
| `organizational_unit_distinguished_name` | string | No | OU where the computer object is created. |

### Audit Log Configuration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file_access_audit_log_level` | string | No | `DISABLED` (default), `SUCCESS_ONLY`, `FAILURE_ONLY`, or `SUCCESS_AND_FAILURE`. |
| `file_share_access_audit_log_level` | string | No | `DISABLED` (default), `SUCCESS_ONLY`, `FAILURE_ONLY`, or `SUCCESS_AND_FAILURE`. |
| `audit_log_destination` | StringValueOrRef | No | CloudWatch Logs log group ARN (must start with `/aws/fsx/`). |

### Disk IOPS Configuration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `mode` | string | No | `AUTOMATIC` (default) or `USER_PROVISIONED`. |
| `iops` | int32 | Conditional | Total SSD IOPS (0–350000). Only valid when mode is `USER_PROVISIONED`. |

## Outputs

| Field | Type | Description |
|-------|------|-------------|
| `file_system_id` | string | File system ID (e.g., `fs-0123456789abcdef0`). Used by EKS SMB CSI driver, AWS Backup. |
| `file_system_arn` | string | ARN for IAM policies. |
| `dns_name` | string | DNS name for SMB mount commands: `net use Z: \\<dns_name>\share`. |
| `preferred_file_server_ip` | string | IP of the active file server. For MULTI_AZ_1, follows failover. |
| `remote_administration_endpoint` | string | PowerShell remote admin endpoint: `Enter-PSSession -ComputerName <endpoint> -ConfigurationName FsxRemoteAdmin`. |
| `network_interface_ids` | []string | ENI IDs. SINGLE_AZ: 1 ENI. MULTI_AZ: 2 ENIs. |
| `vpc_id` | string | VPC ID computed from subnets. |
| `owner_id` | string | AWS account ID of the file system owner. |

### Mount Command

```cmd
net use Z: \\<dns_name>\share
```

Example:

```cmd
net use Z: \\fs-0123456789abcdef0.corp.example.com\share
```

From Linux (with `cifs-utils` installed):

```bash
sudo mount -t cifs //<dns_name>/share /mnt/fsx -o username=user,password=pass,domain=CORP
```

## Minimal Example (AWS Managed AD)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: dev-windows-fs
  org: my-org
spec:
  storage_capacity_gib: 32
  throughput_capacity: 32
  subnet_ids:
    - value: subnet-0a1b2c3d4e5f00001
  security_group_ids:
    - value: sg-0123456789abcdef0
  active_directory_id:
    value: d-0123456789
```

This creates a SINGLE_AZ_2 SSD file system with 32 GiB and 32 MB/s throughput joined to an AWS Managed AD — the smallest and cheapest configuration for development.

## Production Example (Self-Managed AD + Audit Logging)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: prod-windows-fs
  org: my-org
  labels:
    environment: production
spec:
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 500
  storage_type: SSD
  throughput_capacity: 256
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: fsx-prod-key
      fieldPath: status.outputs.key_arn
  subnet_ids:
    - valueFrom:
        kind: AwsSubnet
        name: prod-private-subnet-a
        fieldPath: status.outputs.subnet_id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: smb-clients-sg
        fieldPath: status.outputs.security_group_id
  self_managed_active_directory:
    domain_name: corp.example.com
    dns_ips:
      - "10.0.0.10"
      - "10.0.0.11"
    username: svc-fsx-join
    password: "<your-service-account-password>"
    file_system_administrators_group: FSx Admins
  audit_log_configuration:
    file_access_audit_log_level: SUCCESS_AND_FAILURE
    file_share_access_audit_log_level: FAILURE_ONLY
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "01:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "7:02:00"
```

## High Availability Example (Multi-AZ)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxWindowsFileSystem
metadata:
  name: ha-windows-fs
  org: my-org
  labels:
    environment: production
    tier: critical
spec:
  deployment_type: MULTI_AZ_1
  storage_capacity_gib: 1000
  storage_type: SSD
  throughput_capacity: 512
  subnet_ids:
    - valueFrom:
        kind: AwsSubnet
        name: ha-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: ha-private-subnet-b
        fieldPath: status.outputs.subnet_id
  preferred_subnet_id:
    valueFrom:
      kind: AwsSubnet
      name: ha-private-subnet-a
      fieldPath: status.outputs.subnet_id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: smb-sg
        fieldPath: status.outputs.security_group_id
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: fsx-ha-key
      fieldPath: status.outputs.key_arn
  active_directory_id:
    value: d-0123456789
  aliases:
    - finance.corp.example.com
    - shared.corp.example.com
  audit_log_configuration:
    file_access_audit_log_level: SUCCESS_AND_FAILURE
    file_share_access_audit_log_level: SUCCESS_AND_FAILURE
  disk_iops_configuration:
    mode: USER_PROVISIONED
    iops: 100000
  automatic_backup_retention_days: 30
  daily_automatic_backup_start_time: "03:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "7:02:00"
```

## ForceNew Warnings

The following fields require **resource replacement** if changed. Plan them upfront:

| Field | Impact |
|-------|--------|
| `deployment_type` | Cannot switch between SINGLE_AZ and MULTI_AZ or between generations. |
| `storage_type` | Cannot switch between SSD and HDD. |
| `subnet_ids` | Cannot move the file system to different subnets/AZs. |
| `preferred_subnet_id` | Cannot change the preferred subnet for MULTI_AZ. |
| `security_group_ids` | Cannot change security groups after creation. |
| `kms_key_id` | Cannot change the KMS key after creation. |
| `copy_tags_to_backups` | Cannot toggle after creation. |

## Deliberately Omitted (v1)

The following FSx for Windows features are not exposed in this API version:

- **Storage capacity scaling** — increasing capacity is supported by AWS but not yet exposed as a spec update.
- **Shadow copies** — configured via PowerShell on the file system, not via the API.
- **Data deduplication** — configured via PowerShell on the file system.
- **File share creation** — managed via PowerShell or Windows administrative tools.
- **Cross-account access** — VPC peering or Transit Gateway-based access from other accounts.

See [docs/README.md](docs/README.md) for architecture details and integration patterns.
