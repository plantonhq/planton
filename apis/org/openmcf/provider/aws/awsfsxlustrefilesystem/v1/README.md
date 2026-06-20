# AwsFsxLustreFileSystem

Amazon FSx for Lustre — a fully managed, high-performance parallel file system built on the open-source Lustre file system. Delivers sub-millisecond latencies, hundreds of GB/s throughput, and millions of IOPS for compute-intensive workloads.

## What It Is

FSx for Lustre provides a POSIX-compliant parallel file system optimized for workloads that need to process massive datasets quickly: machine learning training, high performance computing (HPC), video processing, financial modeling, and genomics pipelines.

Lustre is a single-AZ file system. Data durability depends on the deployment type: scratch file systems provide temporary storage with no replication, while persistent file systems replicate data within the AZ and support automatic backups. All file systems are encrypted at rest by default.

This component provisions the FSx Lustre file system, its network interface (ENI) in the specified subnet, optional CloudWatch logging, backup configuration, metadata IOPS tuning, and S3 data repository integration.

## When to Use It

| Use Case | Description |
|----------|-------------|
| **ML training** | Feed training data to GPU instances at hundreds of GB/s. Lustre's parallel I/O eliminates storage bottlenecks during distributed training. |
| **HPC workloads** | CFD simulations, weather modeling, genomics — any workload that reads/writes large files across many compute nodes simultaneously. |
| **Video processing** | Transcode, render, or analyze video files with sub-millisecond latency and sustained high throughput. |
| **S3 data lake processing** | Import S3 objects as a POSIX file system. Process data with standard file I/O, then export results back to S3. |
| **Financial modeling** | Monte Carlo simulations, risk analysis, and backtesting requiring fast random I/O across large datasets. |
| **EKS persistent storage** | Use the FSx for Lustre CSI driver to provision PersistentVolumes backed by `file_system_id`. |

## When NOT to Use It

| Need | Use Instead |
|------|-------------|
| **Shared NFS file system** (POSIX, multi-AZ, auto-scaling storage) | Amazon EFS — multi-AZ, elastic storage, NFS protocol. |
| **Object storage** (blobs, backups, static assets) | Amazon S3 — cheaper, unlimited scale, REST API. |
| **Block storage** (databases, boot volumes) | Amazon EBS — single-instance, lowest latency for transactional I/O. |
| **Windows file shares** (SMB, Active Directory) | Amazon FSx for Windows File Server. |
| **General-purpose ZFS** (snapshots, clones, compression) | Amazon FSx for OpenZFS. |

## Deployment Types

| Type | Durability | Backup | Use Case |
|------|-----------|--------|----------|
| **SCRATCH_1** | None — no replication. Data lost on hardware failure. | No | Legacy scratch. Lowest cost. |
| **SCRATCH_2** | None — no replication. Higher burst throughput than SCRATCH_1. | No | Short-lived processing jobs, dev/test. Recommended over SCRATCH_1. |
| **PERSISTENT_1** | Data replicated within AZ. | Yes | Long-running workloads. Supports SSD and HDD storage. |
| **PERSISTENT_2** | Data replicated within AZ. Latest generation. | Yes | Production workloads. Highest throughput tiers, metadata IOPS configuration, SSD only. |

**Recommendation:** Use `SCRATCH_2` for ephemeral jobs and `PERSISTENT_2` for production. `PERSISTENT_1` is useful only when HDD storage is needed (large capacity at lower cost).

## Prerequisites

- **AWS account** with permissions to create FSx file systems, ENIs, and security groups.
- **VPC with a subnet** — exactly one subnet (Lustre is single-AZ). Use a private subnet in the AZ where your compute resources run.
- **Security groups** — must allow Lustre traffic between the file system and its clients:
  - TCP port 988 (Lustre protocol)
  - TCP ports 1018–1023 (Lustre data channels)

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `deployment_type` | string | No | `SCRATCH_1`, `SCRATCH_2` (default), `PERSISTENT_1`, or `PERSISTENT_2`. **ForceNew**. |
| `storage_capacity_gib` | int32 | **Yes** | Storage in GiB. Min 1200. Valid increments depend on deployment/storage type (see below). Can increase but never decrease. |
| `storage_type` | string | No | `SSD` (default) or `HDD`. **ForceNew**. HDD only with PERSISTENT_1. |
| `per_unit_storage_throughput` | int32 | Conditional | MB/s per TiB. Required for PERSISTENT_1 and PERSISTENT_2. Invalid for SCRATCH. |
| `data_compression_type` | string | No | `NONE` (default) or `LZ4`. Can be changed after creation. |
| `file_system_type_version` | string | No | Lustre version (e.g., `2.15`). **ForceNew**. Leave empty for latest. |
| `subnet_id` | StringValueOrRef | **Yes** | Subnet for the file system ENI. Exactly one subnet. **ForceNew**. |
| `security_group_ids` | []StringValueOrRef | No | Security groups for the ENI. Must allow TCP 988 + 1018–1023. **ForceNew**. |
| `kms_key_id` | StringValueOrRef | No | Customer-managed KMS key ARN. **ForceNew**. Omit for AWS-managed key. |
| `import_path` | string | No | S3 URI for data import (SCRATCH only). **ForceNew**. |
| `export_path` | string | No | S3 URI for data export. Requires `import_path`. **ForceNew**. |
| `log_configuration` | LogConfiguration | No | CloudWatch logging for audit events. |
| `automatic_backup_retention_days` | int32 | No | Backup retention (0–90 days). PERSISTENT only. Default: 0 (disabled). |
| `daily_automatic_backup_start_time` | string | No | Backup window in `HH:MM` UTC. PERSISTENT only. |
| `copy_tags_to_backups` | bool | No | Copy file system tags to backups. **ForceNew**. |
| `skip_final_backup` | bool | No | Skip final backup on deletion. Default: true. |
| `weekly_maintenance_start_time` | string | No | Maintenance window in `d:HH:MM` format (1=Mon, 7=Sun). |
| `metadata_configuration` | MetadataConfiguration | No | Metadata IOPS config. PERSISTENT_2 only. |

### Storage Capacity Rules

| Deployment Type | Storage Type | Minimum | Increment |
|----------------|-------------|---------|-----------|
| SCRATCH_2 | SSD | 1200 GiB | 2400 GiB |
| PERSISTENT_2 | SSD | 1200 GiB | 2400 GiB |
| PERSISTENT_1 | SSD | 1200 GiB | 2400 GiB |
| PERSISTENT_1 | HDD | 6000 GiB | 6000 GiB |
| SCRATCH_1 | SSD | 1200 GiB | 3600 GiB (after initial 1200, 2400, 3600) |

### Per-Unit Storage Throughput Values

| Deployment Type | Storage Type | Valid Values (MB/s/TiB) |
|----------------|-------------|------------------------|
| PERSISTENT_1 | SSD | 50, 100, 200 |
| PERSISTENT_1 | HDD | 12, 40 |
| PERSISTENT_2 | SSD | 125, 250, 500, 1000 |

### Log Configuration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `destination` | StringValueOrRef | No | CloudWatch Logs log group ARN. |
| `level` | string | No | `DISABLED`, `WARN_ONLY`, `ERROR_ONLY`, or `WARN_ERROR` (default). |

### Metadata Configuration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `mode` | string | No | `AUTOMATIC` (default) or `USER_PROVISIONED`. |
| `iops` | int32 | Conditional | Metadata IOPS when mode is `USER_PROVISIONED`. Values: 1500–192000 in defined steps. |

## Outputs

| Field | Type | Description |
|-------|------|-------------|
| `file_system_id` | string | File system ID (e.g., `fs-0123456789abcdef0`). Primary identifier for CSI drivers, ECS, Batch. |
| `file_system_arn` | string | ARN for IAM policies and data repository associations. |
| `dns_name` | string | DNS name for mount commands (e.g., `fs-xxx.fsx.us-east-1.amazonaws.com`). |
| `mount_name` | string | Lustre mount name (e.g., `fsx` or `2p5wpbwj`). Required for the mount command. |
| `network_interface_ids` | []string | ENI IDs for network troubleshooting. |
| `vpc_id` | string | VPC ID computed from the subnet. |
| `file_system_type_version` | string | Deployed Lustre version (e.g., `2.15`). |
| `owner_id` | string | AWS account ID of the file system owner. |

### Mount Command

```bash
sudo mount -t lustre <dns_name>@tcp:/<mount_name> /mnt/fsx
```

Example:

```bash
sudo mount -t lustre fs-0123456789abcdef0.fsx.us-east-1.amazonaws.com@tcp:/fsx /mnt/fsx
```

## Minimal Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: scratch-fsx
  org: my-org
spec:
  storage_capacity_gib: 1200
  subnet_id:
    value: subnet-0a1b2c3d4e5f00001
```

This creates a SCRATCH_2 SSD file system with 1200 GiB — the smallest and cheapest configuration for quick data processing.

## Production Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxLustreFileSystem
metadata:
  name: ml-training-fsx
  org: my-org
  labels:
    environment: production
    workload: ml-training
spec:
  deployment_type: PERSISTENT_2
  storage_capacity_gib: 4800
  storage_type: SSD
  per_unit_storage_throughput: 1000
  data_compression_type: LZ4
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: fsx-encryption-key
      fieldPath: status.outputs.key_arn
  subnet_id:
    valueFrom:
      kind: AwsSubnet
      name: ml-private-subnet-a
      fieldPath: status.outputs.subnet_id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: lustre-clients-sg
        fieldPath: status.outputs.security_group_id
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "05:00"
  copy_tags_to_backups: true
  weekly_maintenance_start_time: "7:03:00"
  metadata_configuration:
    mode: USER_PROVISIONED
    iops: 12000
  log_configuration:
    destination:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: fsx-audit-logs
        fieldPath: status.outputs.log_group_arn
    level: WARN_ERROR
```

## ForceNew Warnings

The following fields require **resource replacement** if changed. Plan them upfront:

| Field | Impact |
|-------|--------|
| `deployment_type` | Cannot switch between SCRATCH and PERSISTENT or between generations. |
| `storage_type` | Cannot switch between SSD and HDD. |
| `subnet_id` | Cannot move the file system to a different subnet/AZ. |
| `security_group_ids` | Cannot change security groups after creation. |
| `kms_key_id` | Cannot change the KMS key after creation. |
| `file_system_type_version` | Cannot change the Lustre version after creation. |
| `import_path` / `export_path` | Cannot change S3 paths after creation. |
| `copy_tags_to_backups` | Cannot toggle after creation. |

## Deliberately Omitted (v1)

The following FSx for Lustre features are not exposed in this API version:

- **Data repository associations** — managed as separate resources for flexible S3 integration on PERSISTENT deployments.
- **File cache** — Amazon File Cache for hybrid/on-premises workloads.
- **Storage capacity scaling** — increasing capacity is supported by AWS but not yet exposed as a spec update.
- **Root squash configuration** — NFS root squash settings.
- **Cross-account access** — VPC peering or Transit Gateway-based access from other accounts.

See [docs/README.md](docs/README.md) for architecture details and integration patterns.
