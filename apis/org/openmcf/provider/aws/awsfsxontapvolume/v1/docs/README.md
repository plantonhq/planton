# AwsFsxOntapVolume — Technical Reference

## Overview

AwsFsxOntapVolume manages `aws_fsx_ontap_volume` resources via Pulumi (Go) and Terraform (HCL). It creates a single ONTAP volume within a Storage Virtual Machine (SVM), with optional tiering, SnapLock WORM compliance, and FlexGroup aggregate distribution.

## Architecture

```
FSx ONTAP File System (physical infrastructure)
└── Storage Virtual Machine (logical data server)
    ├── Volume 1 (this component)    ← /data
    ├── Volume 2                     ← /compliance
    └── Volume 3                     ← /shares
```

Volumes are the atomic data containers. Each volume has its own:
- Capacity allocation (thin-provisioned within the file system)
- Junction path (mount point in the SVM namespace)
- Security style (can differ from the SVM default)
- Tiering policy (independent data lifecycle management)
- SnapLock configuration (per-volume WORM compliance)

## Terraform Resource Mapping

| OpenMCF Field | Terraform Attribute | Notes |
|---|---|---|
| `storage_virtual_machine_id` | `storage_virtual_machine_id` | Required, ForceNew |
| `name` | `name` | Required, ForceNew, 1-203 chars |
| `size_in_megabytes` | `size_in_megabytes` | Minimum 20 |
| `junction_path` | `junction_path` | Optional, must start with `/` |
| `ontap_volume_type` | `ontap_volume_type` | RW or DP, ForceNew |
| `volume_style` | `volume_style` | FLEXVOL or FLEXGROUP, ForceNew |
| `security_style` | `security_style` | UNIX, NTFS, MIXED |
| `snapshot_policy` | `snapshot_policy` | ONTAP policy name |
| `storage_efficiency_enabled` | `storage_efficiency_enabled` | Boolean |
| `copy_tags_to_backups` | `copy_tags_to_backups` | Boolean |
| `skip_final_backup` | `skip_final_backup` | Deletion behavior |
| `bypass_snaplock_enterprise_retention` | `bypass_snaplock_enterprise_retention` | Deletion behavior |
| `tiering_policy.*` | `tiering_policy {}` | Dynamic block |
| `snaplock_configuration.*` | `snaplock_configuration {}` | Dynamic block |
| `aggregate_configuration.*` | `aggregate_configuration {}` | Dynamic block |

## Deliberate Exclusions

| Feature | Reason |
|---|---|
| `size_in_bytes` | Only needed for >2 PB volumes. `size_in_megabytes` covers 99.9% of use cases. |
| `volume_type` | Always "ONTAP" in this context. Implicit from the component name. |
| `final_backup_tags` | Deletion-time tag configuration for the final backup. Very niche. |

## Volume Types

### FLEXVOL (default)
Traditional ONTAP volume residing on a single aggregate. Simpler operations, faster metadata, lower overhead. Suitable for most workloads up to ~100 TB.

### FLEXGROUP
A volume distributed across multiple aggregates. Data is striped across constituents for parallel I/O. Ideal for:
- Data lakes (hundreds of TB to PB)
- Genomics pipelines with large file counts
- Media rendering workflows

FlexGroup requires `aggregate_configuration` specifying which aggregates to use and how many constituents per aggregate.

## Tiering Policies

| Policy | Data on SSD | Data on Capacity Pool | Use Case |
|---|---|---|---|
| NONE | All data | Nothing | Latency-sensitive, always-hot workloads |
| SNAPSHOT_ONLY | Active data | Snapshot data | Standard production with snapshot protection |
| AUTO | Recently accessed | Cold data (after cooling period) | Cost-optimized mixed-access workloads |
| ALL | Metadata only | All user data | Archive, rarely-accessed reference data |

The `cooling_period` (2-183 days) controls how long data must be unaccessed before tiering for AUTO and SNAPSHOT_ONLY policies.

## SnapLock WORM Storage

SnapLock provides Write Once Read Many (WORM) guarantees for regulatory compliance.

### ENTERPRISE vs COMPLIANCE

| Aspect | ENTERPRISE | COMPLIANCE |
|---|---|---|
| WORM guarantee | Yes | Yes |
| Admin can delete early | If privileged_delete enabled | Never |
| AWS Support can delete | No | No |
| Account owner override | No | No |
| Use case | Internal governance | SEC 17a-4, HIPAA, FINRA |

### Retention Periods

Each SnapLock volume has three retention bounds:
- **default_retention**: Applied when files are committed without explicit retention
- **minimum_retention**: Floor — no file can have shorter retention
- **maximum_retention**: Ceiling — no file can have longer retention

Duration types: SECONDS, MINUTES, HOURS, DAYS, MONTHS, YEARS, INFINITE, UNSPECIFIED.

### Autocommit

Files not modified for the autocommit period are automatically transitioned to WORM state. This eliminates the need for applications to explicitly commit files.

## Cross-Resource References

The only cross-resource dependency is the parent SVM:

```yaml
storage_virtual_machine_id:
  valueFrom:
    kind: AwsFsxOntapStorageVirtualMachine
    metadata:
      id: awsfxosvm-prod001
    fieldPath: status.outputs.svm_id
```

## Stack Outputs

| Output | Source | Consumer |
|---|---|---|
| `volume_id` | `aws_fsx_ontap_volume.id` | CloudWatch metrics, AWS APIs |
| `arn` | `aws_fsx_ontap_volume.arn` | IAM policies |
| `uuid` | `aws_fsx_ontap_volume.uuid` | SnapMirror, ONTAP REST API |
| `file_system_id` | Computed from SVM | Cross-reference with file system |
| `flexcache_endpoint_type` | Computed | FlexCache relationship identification |
| `ontap_volume_type` | Computed | Volume type confirmation (RW/DP) |
