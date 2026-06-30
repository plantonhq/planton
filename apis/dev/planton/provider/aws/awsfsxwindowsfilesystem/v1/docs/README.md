# AWS FSx for Windows File Server: Architecture Reference

This document provides a deep technical reference for Amazon FSx for Windows File Server as deployed via the AwsFsxWindowsFileSystem API. It covers the SMB protocol, deployment types, Active Directory integration, storage options, networking, audit logging, backup behavior, PowerShell administration, DNS aliases, and pricing considerations.

---

## 1. SMB Protocol Basics

### How FSx for Windows Works

FSx for Windows File Server is built on Windows Server and provides fully managed file storage accessible via the Server Message Block (SMB) protocol. Unlike file systems that use a custom or Linux-native protocol, FSx Windows is a true Windows file server — it runs NTFS, supports Windows ACLs, shadow copies, DFS namespaces, and integrates natively with Active Directory.

**Architecture:**

1. **File system** — A managed Windows file server identified by `fs-xxxxxxxx`. AWS handles the underlying Windows Server instances, storage, and networking.
2. **Network interfaces (ENIs)** — SINGLE_AZ creates one ENI; MULTI_AZ creates two ENIs (one per AZ). Clients connect via TCP port 445 (SMB).
3. **Active Directory** — Every file system MUST join an AD domain. AD provides identity-based access control, Kerberos authentication, and group policy integration.

**Data flow:** Client → SMB 3.x → TCP port 445 → ENI → FSx Windows Server backend → NTFS on SSD/HDD storage.

**SMB versions:** FSx supports SMB 2.0, 2.1, 3.0, 3.0.2, and 3.1.1. SMB 3.x provides encryption in transit, multi-channel I/O, and transparent failover for MULTI_AZ deployments.

**Windows ACLs:** NTFS permissions are enforced on all file and folder access. ACLs are managed via standard Windows security dialogs, `icacls`, or PowerShell.

### SMB vs Other File System Protocols

| Protocol | Access Pattern | Typical Use |
|----------|---------------|-------------|
| SMB (FSx Windows) | Windows file sharing, AD-integrated | Windows workloads, .NET apps, SQL Server |
| NFS (EFS) | POSIX shared storage, multi-AZ | Linux apps, ECS, Lambda |
| Lustre (FSx Lustre) | Parallel I/O, striped across OSTs | HPC, ML training, video rendering |
| ZFS (FSx OpenZFS) | Snapshots, clones, compression | Dev environments, databases |

---

## 2. Deployment Types

### SINGLE_AZ_1

- **Availability:** Single AZ. If the AZ fails, the file system is unavailable.
- **Throughput ceiling:** Lower than SINGLE_AZ_2. Limited throughput tiers.
- **HDD support:** No — SSD only.
- **Status:** First generation. Use SINGLE_AZ_2 for new deployments.

### SINGLE_AZ_2

- **Availability:** Single AZ. Latest generation.
- **Throughput ceiling:** Higher than SINGLE_AZ_1. Supports all throughput tiers up to 12,288 MB/s.
- **HDD support:** Yes.
- **Use case:** Most workloads that don't require cross-AZ failover. Recommended default.

### MULTI_AZ_1

- **Availability:** Multi-AZ with automatic failover. Active file server runs in the preferred subnet; standby file server runs in the second subnet.
- **Failover:** Automatic. Typically completes in under 30 seconds. The DNS name and `remote_administration_endpoint` follow the active server.
- **Throughput ceiling:** Same as SINGLE_AZ_2.
- **HDD support:** Yes.
- **Requirements:** Two subnets in different AZs. `preferred_subnet_id` must be set.
- **Use case:** Mission-critical workloads requiring high availability.

### Decision Matrix

| Criteria | SINGLE_AZ_1 | SINGLE_AZ_2 | MULTI_AZ_1 |
|----------|-------------|-------------|------------|
| Generation | First | Latest | Latest |
| Cross-AZ failover | No | No | Yes (auto) |
| Max throughput | Lower tiers | 12,288 MB/s | 12,288 MB/s |
| HDD support | No | Yes | Yes |
| Subnets required | 1 | 1 | 2 |
| Cost | Lowest | Low | Higher (2x ENIs, standby server) |

**Recommendation:** Use SINGLE_AZ_2 for development and non-critical production. Use MULTI_AZ_1 for business-critical workloads where AZ failure would cause unacceptable downtime.

---

## 3. Active Directory Integration

### Why AD is Mandatory

Windows file systems rely on Active Directory for authentication, authorization, and identity management. Every FSx for Windows file system MUST join an AD domain. There is no "standalone" mode.

AD provides:
- **Kerberos authentication** — Clients authenticate using Kerberos tickets, not stored passwords.
- **Identity-based access control** — File and folder permissions are set using AD users and groups.
- **Group Policy** — Domain policies apply to the file system.
- **DFS namespaces** — Organize multiple file shares under a single namespace tree.

### AWS Managed Microsoft AD

Use `active_directory_id` to join an AWS Directory Service Managed AD:

- AWS manages the AD domain controllers, replication, patching, and backups.
- The file system joins automatically using the directory ID.
- Simplest setup — no credentials needed in the manifest.
- Supports AD trusts for accessing resources in other domains.

**When to use:** New environments without existing AD, or when you want AWS to manage the AD infrastructure.

### Self-Managed Active Directory

Use `self_managed_active_directory` to join an on-premises or EC2-hosted AD domain:

- You manage the AD domain controllers.
- FSx needs a service account with permissions to create computer objects in the domain.
- Two authentication methods:
  1. **Direct credentials:** `username` + `password` in the manifest. Simple but stores credentials in the resource spec.
  2. **Secrets Manager:** `domain_join_service_account_secret_arn` pointing to a secret with `{"username": "...", "password": "..."}`. Recommended for production.

**Required AD permissions for the service account:**
- Create computer objects
- Delete computer objects (for cleanup on file system deletion)
- Reset passwords
- Write `servicePrincipalName` and `dNSHostName` attributes

**DNS resolution:** The `dns_ips` must be reachable from the file system's subnets. If your AD DNS servers are on-premises, ensure network connectivity via VPN or Direct Connect.

**Organizational Unit:** Use `organizational_unit_distinguished_name` to place the computer object in a specific OU for targeted Group Policy application.

---

## 4. Storage

### SSD Storage

- **Latency:** Sub-millisecond.
- **Capacity range:** 32–65,536 GiB.
- **Supported deployment types:** All (SINGLE_AZ_1, SINGLE_AZ_2, MULTI_AZ_1).
- **IOPS:** Automatically scales with capacity (3 IOPS/GiB) or manually provisioned up to 350,000 IOPS.
- **Use case:** Most workloads. Required for SINGLE_AZ_1.

### HDD Storage

- **Latency:** Higher than SSD (single-digit milliseconds).
- **Capacity range:** 2,000–65,536 GiB (minimum 2,000 GiB).
- **Supported deployment types:** SINGLE_AZ_2 and MULTI_AZ_1 only.
- **IOPS:** Not configurable. AWS manages HDD IOPS.
- **Use case:** Large-capacity workloads where cost per GiB is the primary concern: file archival, bulk storage, infrequently accessed data.

### Storage Type Selection

| Criteria | SSD | HDD |
|----------|-----|-----|
| Latency | Sub-millisecond | Single-digit milliseconds |
| Minimum capacity | 32 GiB | 2,000 GiB |
| IOPS configuration | Yes (AUTOMATIC or USER_PROVISIONED) | No |
| Cost per GiB | Higher | Lower |
| SINGLE_AZ_1 support | Yes | No |

**ForceNew:** Storage type cannot be changed after creation.

### Disk IOPS Configuration (SSD Only)

| Mode | Behavior |
|------|----------|
| AUTOMATIC (default) | 3 IOPS per GiB of storage. Scales automatically with capacity increases. |
| USER_PROVISIONED | Specify exact IOPS (0–350,000). Higher performance independent of storage size. Extra cost. |

**Example:** A 500 GiB SSD file system in AUTOMATIC mode provides 1,500 IOPS. In USER_PROVISIONED mode, you could set 50,000 IOPS for the same 500 GiB.

### Throughput Capacity

Throughput capacity is specified in MB/s as an absolute value (not per-TiB like Lustre):

| Valid Tiers (MB/s) | Typical Use |
|---------------------|-------------|
| 8, 16, 32 | Development, small workloads |
| 64, 128, 256 | Production, moderate I/O |
| 512, 1024, 2048 | High-performance production |
| 4608, 6144, 9216, 12288 | Extreme workloads (media, large databases) |

**Not ForceNew:** Throughput capacity can be changed after creation. Scale up or down as workload demands change.

---

## 5. Networking

### Ports and Protocols

| Port | Protocol | Direction | Purpose |
|------|----------|-----------|---------|
| 445 | TCP | Inbound | SMB file sharing |
| 5985 | TCP | Inbound | WinRM (PowerShell remote administration) |
| 53 | TCP/UDP | Outbound | DNS resolution (AD) |
| 88 | TCP/UDP | Outbound | Kerberos authentication (AD) |
| 389 | TCP | Outbound | LDAP (AD) |
| 636 | TCP | Outbound | LDAPS (AD) |
| 3268 | TCP | Outbound | Global catalog (AD, if using multi-domain forest) |
| 9389 | TCP | Outbound | AD web services (AD) |

### Security Group Best Practices

1. **Dedicated security group:** Create a security group specifically for the FSx file system ENIs.
2. **Client security group:** Create a separate security group for SMB clients. Allow outbound to the FSx security group on port 445.
3. **Bi-directional rules:** The FSx security group must allow inbound from the client security group. The client security group must allow outbound to the FSx security group.
4. **AD communication:** If using self-managed AD, the FSx security group must allow outbound to the AD domain controllers on ports 53, 88, 389, 636.

### Multi-AZ Networking

For MULTI_AZ_1 deployments:

- Two ENIs are created, one in each subnet.
- The preferred subnet hosts the active file server; the other hosts the standby.
- During failover, DNS resolves to the standby's ENI. Clients reconnect automatically via SMB 3.x transparent failover.
- Both subnets must be in different AZs.
- Both subnets must have the same security group access.

### Cross-VPC Access

Not directly supported. Use VPC peering or Transit Gateway with appropriate route tables and security groups to allow SMB traffic (port 445) between VPCs.

---

## 6. Audit Logging

### What Gets Logged

FSx for Windows supports two independent audit log streams:

| Level | Events |
|-------|--------|
| **File access** | Open, read, write, delete, rename, change permissions on individual files and folders. |
| **File share access** | Connect to share, disconnect from share, change share permissions. |

### Log Levels

Each stream can be set independently:

| Value | Behavior |
|-------|----------|
| `DISABLED` | No logging for this stream. |
| `SUCCESS_ONLY` | Log successful operations only. |
| `FAILURE_ONLY` | Log failed attempts only (e.g., access denied). Useful for security monitoring. |
| `SUCCESS_AND_FAILURE` | Log all operations. Most comprehensive but highest volume. |

### Log Destination

Audit logs are sent to CloudWatch Logs. The log group ARN must start with `/aws/fsx/` as required by AWS.

- If `audit_log_destination` is not set, FSx creates a default log stream in `/aws/fsx/windows`.
- For custom log groups, create the log group first and set the appropriate resource policy:

```json
{
  "Effect": "Allow",
  "Principal": {
    "Service": "fsx.amazonaws.com"
  },
  "Action": [
    "logs:CreateLogStream",
    "logs:PutLogEvents"
  ],
  "Resource": "<log-group-arn>:*"
}
```

### Compliance Use Cases

- **SOC 2:** Enable `SUCCESS_AND_FAILURE` on both streams. Retain logs for the audit period.
- **HIPAA:** Track all file access to PHI shares. Use `FAILURE_ONLY` on file share access to detect unauthorized connection attempts.
- **Security monitoring:** Use `FAILURE_ONLY` on file access to detect brute-force or unauthorized access patterns. Route to CloudWatch Alarms or SIEM.

---

## 7. Backup and Maintenance

### Automatic Backups

- **Retention:** `automatic_backup_retention_days` controls retention (0–90 days). Default: 7. Set to 0 to disable.
- **Window:** `daily_automatic_backup_start_time` in `HH:MM` UTC. If omitted, AWS chooses a default.
- **Tags:** `copy_tags_to_backups` copies file system tags to each backup for cost allocation.
- **Final backup:** `skip_final_backup` controls whether a backup is taken on deletion. Default: true (skip).

### Backup Mechanics

- Backups are incremental (only changed data since last backup).
- First backup captures the entire file system.
- Backups are stored in an AWS-managed location (not directly accessible via S3).
- Restoring a backup creates a new file system.
- File system performance is not affected during backup.

### Manual Backups

Manual backups can be created at any time via the AWS console or CLI. They are independent of automatic backups and not subject to `automatic_backup_retention_days`.

### Weekly Maintenance

- **Format:** `d:HH:MM` where d = day of week (1=Monday, 7=Sunday).
- **Example:** `7:02:00` = Sunday at 02:00 UTC.
- Maintenance windows are typically brief (minutes).
- MULTI_AZ_1 maintains availability during maintenance by failing over to the standby.
- SINGLE_AZ deployments may be briefly unavailable during maintenance.
- Schedule during low-traffic periods.

---

## 8. PowerShell Remote Administration

### Connecting

FSx for Windows provides a remote PowerShell endpoint for administrative operations:

```powershell
Enter-PSSession -ComputerName <remote_administration_endpoint> -ConfigurationName FsxRemoteAdmin
```

For MULTI_AZ_1, the endpoint follows the active file server across failovers.

### Available Operations

From the FSx Remote PowerShell session, you can:

- **Create file shares:** `New-FSxSmbShare -Name "Engineering" -Path "D:\Shares\Engineering" -Description "Engineering team files"`
- **Manage shadow copies:** Enable, configure schedules, and manage VSS snapshots.
- **Configure data deduplication:** Enable dedup to reduce storage consumption for files with redundant data.
- **Set quotas:** Configure NTFS disk quotas per user or group.
- **Manage DFS namespaces:** Create namespace folders pointing to file system shares.
- **View active sessions:** `Get-FSxSmbSession` to see connected clients.
- **View open files:** `Get-FSxSmbOpenFile` to see files currently opened by clients.

### Administrative Access

Only members of the `file_system_administrators_group` (default: `Domain Admins`) can connect to the remote PowerShell endpoint. Customize this group in the AD configuration to limit administrative access.

---

## 9. DNS Aliases and DFS Namespaces

### DNS Aliases

DNS aliases allow the file system to be accessed via custom DNS names instead of the auto-generated DNS name:

1. Add alias names to the `aliases` field (up to 50).
2. Create DNS CNAME records pointing each alias to the file system's `dns_name` output.
3. Clients access the file system using the alias: `\\finance.corp.example.com\share`.

**Use cases:**
- **Migration:** Replace an on-premises file server. Point the old DNS name to FSx as a CNAME.
- **User-friendly names:** Give teams memorable mount points (`\\engineering.corp.example.com`).
- **DFS integration:** Use aliases as DFS namespace targets.

### DFS Namespaces

DFS (Distributed File System) namespaces organize multiple file shares under a unified path:

```
\\corp.example.com\DfsRoot\
├── Engineering → \\fsx-1.corp.example.com\engineering
├── Finance → \\fsx-2.corp.example.com\finance
└── Marketing → \\fsx-1.corp.example.com\marketing
```

DFS is configured via PowerShell on the file system. The `aliases` field helps register the file system under the appropriate DFS targets.

---

## 10. Pricing Model Overview

FSx for Windows pricing has several components:

| Component | Unit | Notes |
|-----------|------|-------|
| **Storage capacity** | $/GiB-month | SSD is more expensive than HDD. |
| **Throughput capacity** | $/MB/s-month | Higher tiers cost more. Can be changed after creation. |
| **Backups** | $/GiB-month | Incremental. Only changed data stored. |
| **Data transfer** | Standard AWS data transfer rates | Cross-AZ and cross-region transfer charges apply. |
| **Multi-AZ premium** | Built into pricing | MULTI_AZ_1 is more expensive due to standby server and second ENI. |
| **Provisioned IOPS** | $/IOPS-month (USER_PROVISIONED only) | No extra charge for AUTOMATIC mode. |

### Cost Optimization Strategies

1. **Right-size throughput:** Start with the lowest tier that meets your needs. Throughput can be scaled up after creation.
2. **Use HDD for cold data:** Large file archives benefit from HDD's lower cost per GiB.
3. **Manage backup retention:** Set `automatic_backup_retention_days` to the minimum required by your recovery objectives.
4. **Avoid over-provisioned IOPS:** Use AUTOMATIC mode unless you have a demonstrated need for higher IOPS.
5. **Single-AZ for non-critical workloads:** MULTI_AZ_1 adds cost for the standby server and cross-AZ data replication.
6. **Monitor and adjust:** Use CloudWatch metrics to identify under-utilized throughput or IOPS, and scale down.

---

## 11. CloudWatch Metrics and Monitoring

FSx for Windows publishes metrics to CloudWatch:

| Metric | Description |
|--------|-------------|
| `DataReadBytes` | Bytes read from the file system. |
| `DataWriteBytes` | Bytes written to the file system. |
| `DataReadOperations` | Number of read operations. |
| `DataWriteOperations` | Number of write operations. |
| `FreeStorageCapacity` | Available storage capacity in bytes. |
| `NetworkThroughputUtilization` | Percentage of throughput capacity used. |
| `FileServerDiskThroughputUtilization` | Percentage of disk throughput used. |
| `CPUUtilization` | File server CPU utilization. |

**Recommended alarms:**
- `FreeStorageCapacity` approaching zero → plan storage increase.
- `NetworkThroughputUtilization` sustained above 80% → increase throughput capacity.
- `CPUUtilization` sustained above 80% → increase throughput capacity (also increases CPU).

---

## 12. Integration Patterns

### EKS with SMB CSI Driver

1. Deploy the SMB CSI driver in your EKS cluster.
2. Create a Kubernetes Secret with AD credentials for SMB mount.
3. Create a StorageClass or static PersistentVolume referencing the file system's `dns_name`:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: fsx-windows-pv
spec:
  capacity:
    storage: 100Gi
  accessModes:
    - ReadWriteMany
  csi:
    driver: smb.csi.k8s.io
    volumeHandle: fsx-windows-unique-id
    volumeAttributes:
      source: //<dns_name>/share
    nodeStageSecretRef:
      name: smb-creds
      namespace: default
```

### EC2 Windows Instances

1. Join EC2 instances to the same AD domain.
2. Map the drive:

```cmd
net use Z: \\<dns_name>\share /persistent:yes
```

3. Alternatively, use Group Policy to map drives automatically at logon.

### EC2 Linux Instances (CIFS)

1. Install `cifs-utils`.
2. Mount:

```bash
sudo mount -t cifs //<dns_name>/share /mnt/fsx \
  -o username=svc-mount,password=<password>,domain=CORP,vers=3.0
```

3. For persistent mounts, add to `/etc/fstab`:

```
//<dns_name>/share /mnt/fsx cifs username=svc-mount,password=<password>,domain=CORP,vers=3.0,_netdev 0 0
```

### SQL Server

SQL Server can use FSx Windows file shares for:
- Database files (`.mdf`, `.ndf`) on SMB shares with continuously available (CA) shares.
- Transaction log files (`.ldf`) on SMB shares.
- Backup files (`.bak`) for centralized backup storage.

Enable the CA share feature via PowerShell on the FSx file system.

---

## 13. Limits and Quotas

| Limit | Value |
|-------|-------|
| File systems per account per region | 100 (soft; request increase) |
| Maximum SSD storage capacity | 65,536 GiB |
| Maximum HDD storage capacity | 65,536 GiB |
| Maximum throughput capacity | 12,288 MB/s |
| Maximum provisioned IOPS (SSD) | 350,000 |
| DNS aliases per file system | 50 |
| Security groups per file system | 50 |
| Automatic backup retention | 0–90 days |
| Maximum file size | 64 TiB (NTFS limit) |

---

## 14. Summary

| Topic | Key Takeaway |
|-------|--------------|
| **Protocol** | SMB 2.0–3.1.1 on NTFS. Windows ACLs. Kerberos auth. |
| **Deployment types** | SINGLE_AZ_2 for most workloads; MULTI_AZ_1 for HA. |
| **Active Directory** | Mandatory. AWS Managed AD or self-managed. |
| **Storage** | SSD (32–65536 GiB) or HDD (2000–65536 GiB). HDD requires SINGLE_AZ_2 or MULTI_AZ_1. |
| **Throughput** | Absolute MB/s value (8–12288). Can be changed after creation. |
| **IOPS** | SSD: AUTOMATIC (3/GiB) or USER_PROVISIONED (up to 350K). |
| **Encryption** | Always encrypted at rest. Optional customer-managed KMS. |
| **Backups** | 0–90 day retention. Incremental. Default: 7 days. |
| **Audit logging** | File access + share access → CloudWatch Logs. |
| **Administration** | PowerShell remote endpoint for shares, dedup, shadow copies. |
| **Networking** | TCP 445 (SMB), TCP 5985 (WinRM), AD ports for authentication. |
| **DNS aliases** | Up to 50 custom DNS names per file system. |

For API reference and examples, see the parent [README.md](../README.md) and [examples.md](../examples.md).

---

## Appendix: Quick Reference

| Spec Field | Default | ForceNew |
|------------|---------|----------|
| deployment_type | SINGLE_AZ_2 | Yes |
| storage_capacity_gib | (required) | No (increase only) |
| storage_type | SSD | Yes |
| throughput_capacity | (required) | No |
| subnet_ids | (required) | Yes |
| preferred_subnet_id | (conditional) | Yes |
| security_group_ids | (optional) | Yes |
| kms_key_id | AWS-managed | Yes |
| active_directory_id | (conditional) | No |
| self_managed_active_directory | (conditional) | Partial |
| aliases | (none) | No |
| audit_log_configuration | (disabled) | No |
| disk_iops_configuration | AUTOMATIC | No |
| automatic_backup_retention_days | 7 | No |
| copy_tags_to_backups | false | Yes |
| skip_final_backup | true | No |
