# AWS FSx ONTAP Volume

Creates an Amazon FSx for NetApp ONTAP Volume within a Storage Virtual Machine (SVM). Supports data tiering to capacity pool storage, SnapLock WORM compliance for immutable record retention, and FlexGroup distribution across multiple aggregates for high-throughput workloads.

## What Gets Created

When you deploy an AwsFsxOntapVolume resource, OpenMCF provisions:

- **ONTAP Volume** — an `aws_fsx_ontap_volume` resource within the specified SVM, with configurable size, junction path, security style, and snapshot policy
- **Tiering Policy** (optional) — automatic data movement between primary SSD and capacity pool storage based on access patterns
- **SnapLock Configuration** (optional) — WORM compliance storage with configurable retention periods, autocommit, and privileged delete controls
- **Aggregate Configuration** (optional) — FlexGroup volume distribution across multiple file system aggregates for parallel I/O

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An AwsFsxOntapStorageVirtualMachine** — the parent SVM that provides protocol endpoints and namespace
- **An AwsFsxOntapFileSystem** — the grandparent file system with sufficient storage capacity
- **Sufficient file system capacity** for the requested volume size (ONTAP volumes are thin-provisioned)

## Quick Start

Create a file `ontap-volume.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapVolume
metadata:
  name: my-data-volume
  id: awsfxov-abc123
  org: my-org
  env: dev
spec:
  storageVirtualMachineId:
    value: svm-0123456789abcdef0
  name: vol_data
  sizeInMegabytes: 102400
  junctionPath: /data
  securityStyle: UNIX
  storageEfficiencyEnabled: true
```

Deploy:

```shell
openmcf apply -f ontap-volume.yaml
```

This creates a 100 GB read-write volume mounted at `/data` with UNIX security and storage efficiency enabled.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `storageVirtualMachineId` | `StringValueOrRef` | Parent SVM ID. ForceNew. | Required |
| `storageVirtualMachineId.value` | `string` | Direct SVM ID value | — |
| `storageVirtualMachineId.valueFrom` | `object` | Reference to an AwsFsxOntapStorageVirtualMachine resource | Default field: `status.outputs.svm_id` |
| `name` | `string` | ONTAP volume name. ForceNew. Alphanumeric and underscores only. | 1-203 characters, `^[a-zA-Z0-9_]+$` |
| `sizeInMegabytes` | `int32` | Volume size in megabytes. | Minimum 20 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `junctionPath` | `string` | (none) | Mount point in SVM namespace (e.g., `/data`). Must start with `/`. Volume is unmounted if omitted. |
| `ontapVolumeType` | `string` | `RW` | `RW` (read-write) or `DP` (data protection for SnapMirror). ForceNew. |
| `volumeStyle` | `string` | `FLEXVOL` | `FLEXVOL` (single aggregate) or `FLEXGROUP` (distributed). ForceNew. |
| `securityStyle` | `string` | (inherited) | `UNIX`, `NTFS`, or `MIXED`. Inherits from SVM if omitted. |
| `snapshotPolicy` | `string` | (default) | ONTAP snapshot policy name (e.g., `default`, `none`). |
| `storageEfficiencyEnabled` | `bool` | `false` | ONTAP deduplication, compression, and compaction. |
| `copyTagsToBackups` | `bool` | `false` | Copy resource tags to automatic backups. |
| `skipFinalBackup` | `bool` | `false` | Skip the backup taken when the volume is deleted. |
| `bypassSnaplockEnterpriseRetention` | `bool` | `false` | Allow deleting SnapLock Enterprise volumes with unexpired WORM files. |
| `tieringPolicy` | `object` | (none) | Data tiering configuration. See Tiering Policy below. |
| `snaplockConfiguration` | `object` | (none) | SnapLock WORM configuration. See SnapLock Configuration below. |
| `aggregateConfiguration` | `object` | (none) | FlexGroup aggregate distribution. See Aggregate Configuration below. |

### Tiering Policy

| Field | Type | Description |
|-------|------|-------------|
| `tieringPolicy.name` | `string` | `NONE`, `SNAPSHOT_ONLY`, `AUTO`, or `ALL`. |
| `tieringPolicy.coolingPeriod` | `int32` | Days before data is tiered (2-183). Only valid for `AUTO` or `SNAPSHOT_ONLY`. |

### SnapLock Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `snaplockConfiguration.snaplockType` | `string` | (required) | `ENTERPRISE` or `COMPLIANCE`. ForceNew. |
| `snaplockConfiguration.auditLogVolume` | `bool` | `false` | Designate as the SnapLock audit log volume. |
| `snaplockConfiguration.privilegedDelete` | `string` | `DISABLED` | `DISABLED`, `ENABLED`, or `PERMANENTLY_DISABLED`. |
| `snaplockConfiguration.volumeAppendModeEnabled` | `bool` | `false` | Allow appending to WORM files. |
| `snaplockConfiguration.autocommitPeriod.type` | `string` | — | `NONE`, `MINUTES`, `HOURS`, `DAYS`, `MONTHS`, `YEARS`. |
| `snaplockConfiguration.autocommitPeriod.value` | `int32` | — | Time units (1-65535). Required when type is not `NONE`. |
| `snaplockConfiguration.retentionPeriod.defaultRetention` | `object` | — | Applied to files committed without explicit retention. |
| `snaplockConfiguration.retentionPeriod.minimumRetention` | `object` | — | Floor — no file can have shorter retention. |
| `snaplockConfiguration.retentionPeriod.maximumRetention` | `object` | — | Ceiling — no file can have longer retention. |

Each retention duration has `type` (`SECONDS`/`MINUTES`/`HOURS`/`DAYS`/`MONTHS`/`YEARS`/`INFINITE`/`UNSPECIFIED`) and `value` (`int32`, 0-65535).

### Aggregate Configuration

| Field | Type | Description |
|-------|------|-------------|
| `aggregateConfiguration.aggregates` | `string[]` | Aggregate names (e.g., `aggr1`, `aggr2`). Max 12. ForceNew. |
| `aggregateConfiguration.constituentsPerAggregate` | `int32` | Constituents per aggregate (1-200). ForceNew. |

## Examples

### NFS Data Volume with Cost-Optimized Tiering

A production volume with AUTO tiering that moves cold data to cheaper capacity pool storage after 31 days:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapVolume
metadata:
  name: prod-nfs-data
  id: awsfxov-nfs001
  org: my-org
  env: prod
spec:
  storageVirtualMachineId:
    value: svm-0123456789abcdef0
  name: vol_prod_data
  sizeInMegabytes: 512000
  junctionPath: /data
  securityStyle: UNIX
  snapshotPolicy: default
  storageEfficiencyEnabled: true
  copyTagsToBackups: true
  tieringPolicy:
    name: AUTO
    coolingPeriod: 31
```

### SnapLock Compliance for Regulatory Records

Immutable storage for SEC 17a-4 compliance with 5-year default retention and 1-day autocommit:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapVolume
metadata:
  name: sec-compliance
  id: awsfxov-worm001
  org: my-org
  env: prod
spec:
  storageVirtualMachineId:
    value: svm-0123456789abcdef0
  name: vol_sec17a4
  sizeInMegabytes: 1048576
  junctionPath: /compliance/records
  securityStyle: UNIX
  storageEfficiencyEnabled: true
  tieringPolicy:
    name: SNAPSHOT_ONLY
  snaplockConfiguration:
    snaplockType: COMPLIANCE
    autocommitPeriod:
      type: DAYS
      value: 1
    retentionPeriod:
      defaultRetention:
        type: YEARS
        value: 5
      minimumRetention:
        type: YEARS
        value: 1
      maximumRetention:
        type: YEARS
        value: 10
```

### High-Throughput FlexGroup Volume

A distributed volume across 2 aggregates for data lake workloads requiring parallel I/O:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapVolume
metadata:
  name: datalake-flexgroup
  id: awsfxov-fg001
  org: my-org
  env: prod
spec:
  storageVirtualMachineId:
    value: svm-0123456789abcdef0
  name: vol_datalake
  sizeInMegabytes: 1048576
  junctionPath: /datalake
  volumeStyle: FLEXGROUP
  securityStyle: UNIX
  storageEfficiencyEnabled: true
  tieringPolicy:
    name: NONE
  aggregateConfiguration:
    aggregates:
      - aggr1
      - aggr2
    constituentsPerAggregate: 8
```

### Cross-Resource Reference with valueFrom

A volume referencing its parent SVM via `valueFrom` for infra chart dependency wiring:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapVolume
metadata:
  name: referenced-volume
  id: awsfxov-ref001
  org: my-org
  env: prod
spec:
  storageVirtualMachineId:
    valueFrom:
      kind: AwsFsxOntapStorageVirtualMachine
      metadata:
        id: awsfxosvm-prod001
      fieldPath: status.outputs.svm_id
  name: vol_app_data
  sizeInMegabytes: 102400
  junctionPath: /app
  securityStyle: UNIX
  storageEfficiencyEnabled: true
  tieringPolicy:
    name: AUTO
    coolingPeriod: 31
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `volumeId` | `string` | Volume ID (e.g., `fsvol-0123456789abcdef0`). Used in AWS APIs and CloudWatch metrics. |
| `arn` | `string` | Volume ARN. Used in IAM policies for resource-level permissions. |
| `uuid` | `string` | ONTAP UUID. Used for SnapMirror replication and ONTAP REST API operations. |
| `fileSystemId` | `string` | Parent file system ID. Useful for CloudWatch metric dimensions. |
| `flexcacheEndpointType` | `string` | FlexCache role: `NONE`, `ORIGIN`, or `CACHE`. |
| `ontapVolumeType` | `string` | Confirmed volume type: `RW` or `DP`. |

## Related Components

- [AwsFsxOntapStorageVirtualMachine](https://github.com/plantonhq/openmcf/tree/main/apis/org/openmcf/provider/aws/awsfsxontapstoragevirtualmachine/v1) — Parent SVM providing protocol endpoints and namespace
- [AwsFsxOntapFileSystem](https://github.com/plantonhq/openmcf/tree/main/apis/org/openmcf/provider/aws/awsfsxontapfilesystem/v1) — Grandparent file system providing physical infrastructure
- [AwsFsxLustreFileSystem](https://github.com/plantonhq/openmcf/tree/main/apis/org/openmcf/provider/aws/awsfsxlustrefilesystem/v1) — Alternative: HPC-optimized file system with S3 integration
- [AwsFsxOpenzfsFileSystem](https://github.com/plantonhq/openmcf/tree/main/apis/org/openmcf/provider/aws/awsfsxopenzfsfilesystem/v1) — Alternative: General-purpose NFS with OpenZFS snapshots
- [AwsElasticFileSystem](https://github.com/plantonhq/openmcf/tree/main/apis/org/openmcf/provider/aws/awselasticfilesystem/v1) — Alternative: Serverless NFS with automatic scaling
