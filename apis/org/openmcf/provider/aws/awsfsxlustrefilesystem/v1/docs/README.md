# AWS FSx for Lustre: Architecture Reference

This document provides a deep technical reference for Amazon FSx for Lustre as deployed via the AwsFsxLustreFileSystem API. It covers the Lustre protocol, deployment types, storage capacity rules, throughput models, S3 integration, encryption, backups, metadata IOPS, and integration patterns.

---

## 1. Lustre Protocol Basics

### How Lustre Works

Lustre is an open-source parallel file system designed for high performance computing. It splits file data across multiple storage targets (OSTs — Object Storage Targets) and file metadata across metadata targets (MDTs). Clients communicate with both metadata and storage servers simultaneously, enabling aggregate throughput that scales linearly with the number of storage targets.

**Architecture (in FSx):**

1. **File system** — A managed Lustre file system identified by `fs-xxxxxxxx`. AWS handles the underlying OSTs, MDTs, and management servers.
2. **Network interface (ENI)** — A single ENI in the specified subnet. All clients communicate with the file system through this endpoint.
3. **Mount name** — An auto-generated Lustre mount name (e.g., `fsx` or `2p5wpbwj`) required to construct the full mount path.

**Data flow:** Client → Lustre client kernel module → TCP port 988 → ENI → FSx backend (metadata servers + storage servers) → Durable storage (SSD or HDD).

**POSIX compliance:** Lustre provides a fully POSIX-compliant file system. Applications using standard file I/O (`open`, `read`, `write`, `close`, `stat`, `readdir`, `flock`) work without modification. Lustre supports file locking (flock and POSIX advisory locks) for coordination across clients.

**Consistency:** Lustre provides close-to-open consistency by default. After a file is closed on one client and opened on another, the second client sees the latest data. Within a single client, reads after writes are consistent. For multi-client concurrent writes to the same file, use explicit file locking.

### Lustre vs Other File System Protocols

| Protocol | Access Pattern | Typical Use |
|----------|---------------|-------------|
| Lustre | Parallel I/O, striped across OSTs | HPC, ML training, video rendering |
| NFS (EFS) | Single-stream per file, shared mount | General shared storage, ECS, Lambda |
| SMB (FSx Windows) | Windows file sharing | Active Directory environments |
| ZFS (FSx OpenZFS) | Snapshots, clones, compression | Dev environments, databases |

Lustre is optimized for workloads that read/write large files (or many files) across many compute nodes simultaneously. The parallel I/O architecture avoids the bottleneck of a single NFS server.

---

## 2. Deployment Types

### SCRATCH_1

- **Durability:** None. No data replication. If any underlying storage server fails, data on that server is lost.
- **Throughput:** 200 MB/s/TiB baseline.
- **Backup:** Not supported.
- **Storage capacity scaling:** Increasing storage forces file system replacement.
- **Status:** Legacy. Use SCRATCH_2 instead.

### SCRATCH_2

- **Durability:** None. No data replication.
- **Throughput:** 200 MB/s/TiB baseline with burst to 1300 MB/s/TiB.
- **Backup:** Not supported.
- **Storage capacity scaling:** Can be increased without replacement.
- **S3 integration:** Supports `import_path` and `export_path` for legacy S3 linking.
- **Use case:** Short-lived processing jobs, dev/test, CI pipelines. Data can be regenerated from source (e.g., S3).

### PERSISTENT_1

- **Durability:** Data replicated within the single AZ.
- **Throughput:** Depends on storage type. SSD: 50, 100, or 200 MB/s/TiB. HDD: 12 or 40 MB/s/TiB.
- **Backup:** Supported. Automatic daily backups with configurable retention.
- **Storage types:** SSD and HDD.
- **Use case:** Long-running workloads. HDD option for cost-sensitive large-capacity needs.

### PERSISTENT_2

- **Durability:** Data replicated within the single AZ. Latest generation.
- **Throughput:** 125, 250, 500, or 1000 MB/s/TiB. Higher ceiling than PERSISTENT_1.
- **Backup:** Supported.
- **Storage types:** SSD only.
- **Metadata IOPS:** Supports AUTOMATIC and USER_PROVISIONED modes.
- **Use case:** Production workloads requiring maximum performance and durability. Recommended over PERSISTENT_1 for new deployments.

### Decision Matrix

| Criteria | SCRATCH_2 | PERSISTENT_1 | PERSISTENT_2 |
|----------|-----------|-------------|-------------|
| Data durability | None | Replicated in AZ | Replicated in AZ |
| Backup support | No | Yes | Yes |
| Max throughput/TiB | 1300 (burst) | 200 (SSD) | 1000 (SSD) |
| HDD support | No | Yes | No |
| Metadata IOPS config | No | No | Yes |
| Cost | Lowest | Medium | Highest |

---

## 3. Storage Capacity Rules

### Valid Capacity Values

Storage capacity must be at least 1200 GiB and must follow specific increments depending on deployment type and storage type:

| Deployment Type | Storage Type | Minimum | Increment | Example Valid Values |
|----------------|-------------|---------|-----------|---------------------|
| SCRATCH_2 | SSD | 1200 | 2400 | 1200, 3600, 6000, 8400 |
| PERSISTENT_2 | SSD | 1200 | 2400 | 1200, 3600, 6000, 8400 |
| PERSISTENT_1 | SSD | 1200 | 2400 | 1200, 3600, 6000, 8400 |
| PERSISTENT_1 | HDD | 6000 | 6000 | 6000, 12000, 18000, 24000 |
| SCRATCH_1 | SSD | 1200 | 3600 (after 3600) | 1200, 2400, 3600, 7200, 10800 |

### Capacity Increase Rules

- Storage capacity can be **increased** after creation (except SCRATCH_1, which forces replacement).
- Storage capacity can **never be decreased**.
- Increases must follow the same increment rules.
- During a capacity increase, the file system remains available; data is rebalanced across OSTs in the background.
- Capacity increase requests may take minutes to hours depending on file system size.

### Capacity and Aggregate Throughput

Aggregate throughput = storage capacity (in TiB) × per-unit storage throughput (MB/s/TiB).

| Example Configuration | Capacity | Per-Unit | Aggregate Throughput |
|-----------------------|----------|----------|---------------------|
| SCRATCH_2, 1200 GiB | 1.17 TiB | 200 MB/s/TiB | ~234 MB/s (burst: ~1523 MB/s) |
| PERSISTENT_2, 2400 GiB, 1000 | 2.34 TiB | 1000 MB/s/TiB | ~2344 MB/s |
| PERSISTENT_2, 4800 GiB, 500 | 4.69 TiB | 500 MB/s/TiB | ~2344 MB/s |
| PERSISTENT_1 HDD, 6000 GiB, 40 | 5.86 TiB | 40 MB/s/TiB | ~234 MB/s |
| PERSISTENT_1 HDD, 12000 GiB, 40 | 11.72 TiB | 40 MB/s/TiB | ~469 MB/s |

To increase aggregate throughput, either increase storage capacity or choose a higher per-unit throughput tier. Per-unit throughput is ForceNew; capacity can be increased in place.

---

## 4. Throughput Options

### SCRATCH Throughput

SCRATCH file systems do not use `per_unit_storage_throughput`. Throughput is determined by the deployment type:

- **SCRATCH_1:** 200 MB/s/TiB baseline.
- **SCRATCH_2:** 200 MB/s/TiB baseline, burst to 1300 MB/s/TiB.

SCRATCH_2 burst throughput is automatic; no configuration needed. Burst capacity is based on a credit system similar to EBS burst. Sustained heavy I/O may exhaust credits and return to baseline.

### PERSISTENT Throughput

PERSISTENT file systems require `per_unit_storage_throughput`:

| Deployment Type | Storage Type | Valid Values (MB/s/TiB) |
|----------------|-------------|------------------------|
| PERSISTENT_1 | SSD | 50, 100, 200 |
| PERSISTENT_1 | HDD | 12, 40 |
| PERSISTENT_2 | SSD | 125, 250, 500, 1000 |

**Choosing a tier:**

- **125 MB/s/TiB:** Cost-effective for sequential workloads that don't need peak performance.
- **250 MB/s/TiB:** Good general-purpose tier for mixed workloads.
- **500 MB/s/TiB:** High-performance tier for ML training and HPC.
- **1000 MB/s/TiB:** Maximum performance. Use when throughput is the primary bottleneck (distributed training across many GPUs).

**ForceNew:** Per-unit throughput cannot be changed after creation. To change the throughput tier, create a new file system and migrate data.

### Read vs Write Throughput

FSx for Lustre reads from SSD/HDD storage (fast) and writes to both active and replica storage. Read throughput is typically higher than write throughput in practice. The per-unit throughput value applies to the aggregate I/O (reads + writes). For write-heavy workloads, benchmark against your specific throughput tier.

---

## 5. S3 Data Repository Integration

### Legacy Import/Export (SCRATCH only)

The `import_path` and `export_path` fields provide legacy S3 integration for SCRATCH file systems:

- **import_path** — S3 URI (e.g., `s3://bucket/prefix/`). At creation, FSx imports S3 object metadata into the Lustre namespace. File data is **lazy-loaded** on first access (Hierarchical Storage Management — HSM).
- **export_path** — S3 URI for exporting changes. Requires `import_path`. New and modified files are automatically exported to S3.
- Both fields are **ForceNew**. Changing them requires replacing the file system.
- Only supported on SCRATCH_1 and SCRATCH_2.

### Data Repository Associations (PERSISTENT)

For PERSISTENT file systems, S3 integration uses data repository associations (DRAs), which are **separate resources** not part of this spec:

- DRAs link a Lustre directory path to an S3 bucket/prefix.
- Multiple DRAs can be attached to a single file system (up to 8 per file system, max 25 per account).
- DRAs support auto-import (S3 → Lustre) and auto-export (Lustre → S3) modes.
- DRAs can be created, modified, and deleted without replacing the file system.

**Why separate:** DRAs have their own lifecycle (creation, update, deletion) independent of the file system. Managing them as separate resources allows flexible data pipeline configurations without file system replacement.

### HSM (Hierarchical Storage Management)

When S3 integration is configured:

1. **Import:** Only metadata (file name, size, timestamps) is imported at creation. File data remains in S3.
2. **First access:** When a client reads a file that hasn't been fetched, Lustre transparently retrieves the data from S3 (HSM recall). This adds latency to the first read.
3. **Subsequent access:** Data is served from Lustre storage at full speed.
4. **Export:** New files or modified files can be exported back to S3 (auto-export or `lfs hsm_archive`).

**Pre-loading:** To avoid first-read latency for critical data, use `lfs hsm_restore` to pre-fetch data from S3 before processing begins.

---

## 6. Encryption

### Encryption at Rest

All FSx for Lustre file systems are encrypted at rest by default using the AWS-managed FSx key. No opt-in required.

- **AWS-managed key:** Used when `kms_key_id` is omitted. Managed by AWS; no key management overhead.
- **Customer-managed key:** Set `kms_key_id` to a KMS key ARN for audit trails, key rotation, cross-account access control, and compliance requirements.
- **ForceNew:** The KMS key cannot be changed after creation.

### KMS Key Policy Requirements

When using a customer-managed KMS key, the key policy must allow the FSx service to use it:

```json
{
  "Sid": "Allow FSx",
  "Effect": "Allow",
  "Principal": {
    "Service": "fsx.amazonaws.com"
  },
  "Action": [
    "kms:Decrypt",
    "kms:GenerateDataKey",
    "kms:CreateGrant"
  ],
  "Resource": "*",
  "Condition": {
    "StringEquals": {
      "kms:CallerAccount": "<your-account-id>"
    }
  }
}
```

### Encryption in Transit

Lustre traffic between clients and the file system is encrypted in transit automatically for file systems created with Lustre version 2.12 or later. No additional configuration is needed. Earlier versions do not support encryption in transit.

---

## 7. Backup Behavior

### Automatic Backups

Automatic backups are supported only on **PERSISTENT_1** and **PERSISTENT_2** file systems. Scratch file systems cannot have backups.

- **Retention:** `automatic_backup_retention_days` controls how long backups are kept (0–90 days). Set to 0 to disable.
- **Window:** `daily_automatic_backup_start_time` in `HH:MM` UTC format. If omitted, AWS chooses a default window.
- **Tags:** `copy_tags_to_backups` copies file system tags to each backup for cost allocation and organization.
- **Final backup:** `skip_final_backup` controls whether a backup is taken when the file system is deleted. Default: true (skip).

### Manual Backups

Manual backups can be created via the AWS console or CLI at any time (for PERSISTENT file systems). They are independent of the automatic backup schedule and are not subject to `automatic_backup_retention_days`.

### Backup Mechanics

- Backups are incremental: only changed data since the last backup is stored.
- First backup captures the entire file system; subsequent backups are incremental.
- Backups are stored in the same region as the file system in an AWS-managed S3 bucket (not accessible directly).
- Restoring a backup creates a new file system. The original file system is not modified.

### Backup Duration

Backup time depends on the amount of changed data since the last backup. For large file systems with significant churn, backups may take hours. File system performance is not affected during backup.

---

## 8. Metadata IOPS Configuration

### PERSISTENT_2 Only

Metadata configuration is only available on PERSISTENT_2 file systems. It controls the performance of metadata operations: file creation, deletion, stat, listing, rename.

### AUTOMATIC Mode (Default)

FSx scales metadata IOPS based on storage capacity:

| Storage Capacity | Approximate Metadata IOPS |
|-----------------|--------------------------|
| 1.2 TiB | 1,500 |
| 2.4 TiB | 3,000 |
| 4.8 TiB | 6,000 |
| 9.6 TiB | 12,000 |
| 19.2 TiB | 24,000 |

AUTOMATIC mode is sufficient for most workloads. Metadata IOPS scales linearly with storage capacity.

### USER_PROVISIONED Mode

Allows specifying explicit metadata IOPS independent of storage capacity. Use when:

- Your workload creates, lists, or deletes millions of small files.
- You need higher metadata IOPS than AUTOMATIC provides for your storage size.
- Storage capacity is small but metadata operations are heavy (e.g., checkpointing in ML training).

**Valid values:** 1500, 3000, 6000, 12000, 24000, 36000, 48000, 60000, 72000, 84000, 96000, 108000, 120000, 132000, 144000, 156000, 168000, 180000, 192000.

**Cost:** USER_PROVISIONED mode incurs additional charges for metadata IOPS above what AUTOMATIC would provide.

---

## 9. Networking

### Single-AZ Architecture

Lustre file systems are single-AZ. Exactly one subnet is required. The file system creates one ENI in that subnet. All compute resources mounting the file system must have network connectivity to this subnet.

**Implications:**

- If the AZ becomes unavailable, the file system is unavailable.
- For PERSISTENT types, data is replicated within the AZ (not across AZs). AZ-level failure does not lose data (AWS restores from replicas after AZ recovery).
- For SCRATCH types, data is not replicated. AZ failure or hardware failure loses data.

### Security Group Requirements

Security groups attached to the file system ENI must allow Lustre traffic:

| Port | Protocol | Direction | Purpose |
|------|----------|-----------|---------|
| 988 | TCP | Inbound | Lustre protocol (LNET) |
| 1018–1023 | TCP | Inbound | Lustre data channels |

Clients must also have outbound rules allowing traffic to the file system's ENI on these ports.

**Best practice:** Create a dedicated security group for Lustre clients and reference it in both the file system's `security_group_ids` and the compute instances' security groups.

### DNS and Mount

The file system provides two values needed for mounting:

- **dns_name** — e.g., `fs-0123456789abcdef0.fsx.us-east-1.amazonaws.com`
- **mount_name** — e.g., `fsx` or `2p5wpbwj`

Mount command:

```bash
sudo mount -t lustre <dns_name>@tcp:/<mount_name> /mnt/fsx
```

### Cross-VPC and Cross-Account Access

Not directly supported. Use VPC peering or Transit Gateway with appropriate route tables and security groups to allow Lustre traffic between VPCs.

---

## 10. Logging

### CloudWatch Audit Logging

When `log_configuration` is set, FSx sends audit events to CloudWatch Logs. Events include:

- File access (read, write)
- File creation and deletion
- Permission changes

### Log Levels

| Level | Description |
|-------|-------------|
| `DISABLED` | No logging |
| `WARN_ONLY` | Warning-level events only |
| `ERROR_ONLY` | Error-level events only |
| `WARN_ERROR` | Both warning and error events (default when configured) |

### Log Group Requirements

The CloudWatch Logs log group must:

1. Exist before the file system is created.
2. Be in the same region as the file system.
3. Have a resource policy allowing FSx to write to it:

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

---

## 11. Maintenance Windows

### Weekly Maintenance

FSx requires a maintenance window for patching and updates. The `weekly_maintenance_start_time` field specifies when:

- Format: `d:HH:MM` where d is day of week (1=Monday, 7=Sunday).
- Example: `7:03:00` means Sunday at 03:00 UTC.
- If omitted, AWS chooses a default window.

### Impact

- Maintenance windows are typically brief (minutes).
- Single-AZ file systems may be briefly unavailable during maintenance.
- Schedule during low-traffic periods.

---

## 12. Lustre Client Installation and Mount

### Amazon Linux 2 / AL2023

```bash
sudo amazon-linux-extras install -y lustre
sudo mkdir -p /mnt/fsx
sudo mount -t lustre fs-xxx.fsx.us-east-1.amazonaws.com@tcp:/mount-name /mnt/fsx
```

### Ubuntu

```bash
sudo apt-get install -y lustre-client-modules-$(uname -r)
sudo mkdir -p /mnt/fsx
sudo mount -t lustre fs-xxx.fsx.us-east-1.amazonaws.com@tcp:/mount-name /mnt/fsx
```

### /etc/fstab Entry

For persistent mounts across reboots:

```
fs-xxx.fsx.us-east-1.amazonaws.com@tcp:/mount-name /mnt/fsx lustre defaults,noatime,flock,_netdev 0 0
```

**Mount options:**

- `noatime` — Do not update access times on read. Reduces metadata overhead significantly for read-heavy workloads.
- `flock` — Enable file locking.
- `_netdev` — Wait for network before mounting at boot.

---

## 13. Common Integration Patterns

### EKS with FSx CSI Driver

1. Deploy the FSx for Lustre CSI driver (or AWS-provided EKS add-on).
2. Create a StorageClass referencing the file system:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fsx-lustre-sc
provisioner: fsx.csi.aws.com
parameters:
  subnetId: <subnet-id>
  securityGroupIds: <sg-id>
  deploymentType: PERSISTENT_2
  perUnitStorageThroughput: "250"
```

Or use a static PersistentVolume referencing an existing file system:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: fsx-pv
spec:
  capacity:
    storage: 1200Gi
  accessModes:
    - ReadWriteMany
  csi:
    driver: fsx.csi.aws.com
    volumeHandle: <file_system_id>
    volumeAttributes:
      dnsname: <dns_name>
      mountname: <mount_name>
```

### AWS Batch

AWS Batch compute environments can mount FSx Lustre via launch templates:

1. Create a launch template with user data that installs the Lustre client and mounts the file system.
2. Reference `file_system_id`, `dns_name`, and `mount_name` from outputs.
3. Batch jobs access `/mnt/fsx` as a shared high-performance scratch space.

### SageMaker Training Jobs

SageMaker training jobs can use FSx Lustre as a data source:

1. Create an FSx Lustre file system with training data.
2. In the SageMaker `CreateTrainingJob` API, specify the file system as an input data channel:

```json
{
  "InputDataConfig": [{
    "ChannelName": "training",
    "DataSource": {
      "FileSystemDataSource": {
        "FileSystemId": "<file_system_id>",
        "FileSystemAccessMode": "ro",
        "FileSystemType": "FSxLustre",
        "DirectoryPath": "/mount-name"
      }
    }
  }]
}
```

### EC2 Compute Clusters

For HPC workloads using EC2 instances (e.g., with AWS ParallelCluster):

1. All instances must be in the same subnet (or a subnet that can reach the file system ENI).
2. Security groups must allow Lustre traffic (TCP 988 + 1018–1023).
3. Install Lustre client and mount on each instance.
4. Use placement groups for network-optimized instance placement.

---

## 14. Limits and Quotas

| Limit | Value |
|-------|-------|
| File systems per account per region | 100 (soft; request increase) |
| Maximum storage capacity | 100+ PiB (varies by deployment type) |
| Maximum file size | Unlimited (practical limit: storage capacity) |
| Maximum number of files | Billions (metadata capacity dependent) |
| Security groups per file system | 50 |
| Data repository associations per file system | 8 |
| Data repository associations per account | 25 |
| Automatic backup retention | 0–90 days |
| Metadata IOPS (USER_PROVISIONED) | Up to 192,000 |

---

## 15. CloudWatch Metrics and Monitoring

FSx for Lustre publishes metrics to CloudWatch:

| Metric | Description |
|--------|-------------|
| `DataReadBytes` | Bytes read from the file system |
| `DataWriteBytes` | Bytes written to the file system |
| `DataReadOperations` | Number of read operations |
| `DataWriteOperations` | Number of write operations |
| `MetadataOperations` | Number of metadata operations (stat, listdir, etc.) |
| `FreeDataStorageCapacity` | Available storage capacity in bytes |
| `AgeOfOldestQueuedMessage` | Age of oldest pending data repository task |

**Alarms:** Set alarms on `FreeDataStorageCapacity` approaching zero (plan capacity increase), `DataReadBytes`/`DataWriteBytes` approaching aggregate throughput limits (upgrade throughput tier or add capacity), and `MetadataOperations` exceeding metadata IOPS capacity (switch to USER_PROVISIONED mode).

---

## 16. Cost Optimization

### Choose the Right Deployment Type

- **Ephemeral data:** Use SCRATCH_2. No durability, lowest cost.
- **Long-running, cost-sensitive:** Use PERSISTENT_1 HDD for large capacity at low cost.
- **Production, performance-critical:** Use PERSISTENT_2 SSD.

### Right-Size Throughput

- Do not over-provision throughput. Start with 250 MB/s/TiB and upgrade only if monitoring shows saturation.
- Per-unit throughput is ForceNew; plan for peak load at creation time.

### Use LZ4 Compression

Enable `data_compression_type: LZ4` to reduce effective storage consumption. Compression ratios of 2–3x are common for text-heavy data (logs, CSVs, JSON, genomics data). This effectively increases your usable capacity without increasing the provisioned storage.

### Manage Backup Retention

- Set `automatic_backup_retention_days` to the minimum required by your recovery objectives.
- Backups consume storage in an AWS-managed S3 bucket and incur charges.
- Use `skip_final_backup: true` for test environments to avoid unnecessary final backup charges on deletion.

---

## 17. Summary

| Topic | Key Takeaway |
|-------|--------------|
| **Protocol** | Lustre parallel file system; POSIX-compliant; sub-ms latency. |
| **Deployment types** | SCRATCH_2 for ephemeral; PERSISTENT_2 for production; PERSISTENT_1 for HDD. |
| **Storage** | Min 1200 GiB; increments vary by type. Increase OK, decrease never. |
| **Throughput** | Fixed per-unit (ForceNew). Scale aggregate by adding capacity. |
| **S3 integration** | Legacy import/export for SCRATCH; DRAs for PERSISTENT. |
| **Encryption** | Always encrypted at rest. Optional customer-managed KMS. |
| **Backups** | PERSISTENT only. 0–90 day retention. Incremental. |
| **Metadata IOPS** | PERSISTENT_2 only. AUTOMATIC or USER_PROVISIONED. |
| **Networking** | Single-AZ, one subnet, TCP 988 + 1018–1023. |
| **Mount** | `mount -t lustre <dns_name>@tcp:/<mount_name> /mnt/fsx` |

For API reference and examples, see the parent [README.md](../README.md) and [examples.md](../examples.md).

---

## Appendix: Quick Reference

| Spec Field | Default | ForceNew |
|------------|---------|----------|
| deployment_type | SCRATCH_2 | Yes |
| storage_capacity_gib | (required) | No (increase only) |
| storage_type | SSD | Yes |
| per_unit_storage_throughput | (required for PERSISTENT) | Yes |
| data_compression_type | NONE | No |
| file_system_type_version | (latest) | Yes |
| subnet_id | (required) | Yes |
| security_group_ids | (optional) | Yes |
| kms_key_id | AWS-managed | Yes |
| import_path | (none) | Yes |
| export_path | (none) | Yes |
| copy_tags_to_backups | false | Yes |
| automatic_backup_retention_days | 0 | No |
| skip_final_backup | true | No |
