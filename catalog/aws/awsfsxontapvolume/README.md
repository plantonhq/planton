# AwsFsxOntapVolume

Create and manage Amazon FSx for NetApp ONTAP Volumes with Planton.

## Overview

An ONTAP Volume is a data container within an FSx for ONTAP Storage Virtual Machine (SVM). Volumes provide file-level (NFS/SMB) or block-level (iSCSI) storage with enterprise features including data tiering, SnapLock WORM compliance, and FlexGroup distribution.

Volumes sit at the bottom of the ONTAP hierarchy:

- **File System** provides physical infrastructure (storage, throughput, HA pairs)
- **SVM** provides the logical data server (protocols, endpoints, Active Directory)
- **Volume** provides data containers (this component)

## When to Use

Use AwsFsxOntapVolume when you need:

- **NFS/SMB file storage** mounted at specific paths within an SVM
- **Regulatory compliance** with SnapLock WORM (SEC 17a-4, HIPAA, FINRA)
- **Cost-optimized tiering** that automatically moves cold data to cheaper storage
- **High-throughput distributed storage** with FlexGroup across multiple aggregates

## Prerequisites

- An existing AwsFsxOntapStorageVirtualMachine (the parent SVM)
- The SVM's parent AwsFsxOntapFileSystem must have sufficient capacity

## Quick Start

### Minimal Configuration

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOntapVolume
metadata:
  name: my-data-volume
  id: awsfxov-abc123
  org: my-org
  env: dev
spec:
  storage_virtual_machine_id:
    value: svm-0123456789abcdef0
  name: vol_data
  size_in_megabytes: 1024
```

### Production Configuration

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOntapVolume
metadata:
  name: prod-data-volume
  id: awsfxov-prod456
  org: my-org
  env: prod
spec:
  storage_virtual_machine_id:
    valueFrom:
      kind: AwsFsxOntapStorageVirtualMachine
      metadata:
        id: awsfxosvm-prod789
      fieldPath: status.outputs.svm_id
  name: vol_prod_data
  size_in_megabytes: 512000
  junction_path: /data
  security_style: UNIX
  snapshot_policy: default
  storage_efficiency_enabled: true
  tiering_policy:
    name: AUTO
    cooling_period: 31
```

## Spec Fields

### Required

| Field | Type | Description |
|-------|------|-------------|
| `storage_virtual_machine_id` | StringValueOrRef | Parent SVM ID. ForceNew. |
| `name` | string | ONTAP volume name (1-203 chars, alphanumeric + underscore). ForceNew. |
| `size_in_megabytes` XOR `size_in_bytes` | int32 / int64 | Exactly one size arm: megabytes (min 20) for everyday volumes, or bytes for byte-precise sizing and volumes past 2 PiB (FlexGroup scales to ~20 PiB). |

### Optional

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `junction_path` | string | (none) | Mount point in SVM namespace. Must start with `/`. 1-255 chars. |
| `ontap_volume_type` | string | `RW` | `RW` (read-write) or `DP` (data protection). ForceNew. |
| `volume_style` | string | `FLEXVOL` | `FLEXVOL` or `FLEXGROUP`. ForceNew. |
| `security_style` | string | (inherited) | `UNIX`, `NTFS`, or `MIXED`. |
| `snapshot_policy` | string | (default) | ONTAP snapshot policy name. 1-255 chars. |
| `storage_efficiency_enabled` | bool | (ONTAP default) | Tri-state: true enables dedup/compression/compaction, false disables, unset keeps ONTAP's per-volume-type default. |
| `copy_tags_to_backups` | bool | false | Copy tags to automatic backups. |
| `skip_final_backup` | bool | false | Skip backup on deletion. Delete-time control — apply before deleting. |
| `final_backup_tags` | map | (none) | Tags applied to the final backup. Only meaningful when `skip_final_backup` is false. |
| `bypass_snaplock_enterprise_retention` | bool | false | Allow deleting SnapLock Enterprise volumes with unexpired WORM files. Delete-time control. |
| `tiering_policy` | object | (default) | Data tiering to capacity pool storage. |
| `snaplock_configuration` | object | (none) | SnapLock WORM compliance storage. |
| `aggregate_configuration` | object | (none) | FlexGroup aggregate distribution (`aggr[0-9]{1,2}` names). |

### Tiering Policy

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | `NONE`, `SNAPSHOT_ONLY`, `AUTO`, or `ALL`. |
| `cooling_period` | int32 | Days before data is tiered (2-183). Only for AUTO/SNAPSHOT_ONLY. |

### SnapLock Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `snaplock_type` | string | (required) | `ENTERPRISE` or `COMPLIANCE`. ForceNew. |
| `audit_log_volume` | bool | false | Designate as audit log volume. |
| `privileged_delete` | string | `DISABLED` | `DISABLED`, `ENABLED`, or `PERMANENTLY_DISABLED`. |
| `volume_append_mode_enabled` | bool | false | Allow appending to WORM files. |
| `autocommit_period` | object | (none) | Auto-commit files to WORM after inactivity. |
| `retention_period` | object | (none) | Default/min/max retention bounds. |

## Outputs

| Output | Description |
|--------|-------------|
| `volume_id` | Volume ID (e.g., `fsvol-0123456789abcdef0`) |
| `arn` | Volume ARN for IAM policies |
| `uuid` | ONTAP UUID for SnapMirror/REST API |
| `file_system_id` | Parent file system ID |
| `flexcache_endpoint_type` | FlexCache endpoint type (NONE/ORIGIN/CACHE) |
| `ontap_volume_type` | Confirmed volume type (RW/DP) |

## Related Kinds

- [AwsFsxOntapStorageVirtualMachine](../awsfsxontapstoragevirtualmachine/) — Parent SVM
- [AwsFsxOntapFileSystem](../awsfsxontapfilesystem/) — Grandparent file system

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
