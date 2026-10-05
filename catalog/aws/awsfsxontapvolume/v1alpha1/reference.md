# AwsFsxOntapVolume

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsFsxOntapVolumeSpec defines the desired configuration for an Amazon FSx for
NetApp ONTAP Volume. A volume is a data container within an ONTAP Storage
Virtual Machine (SVM) that provides file-level (NFS/SMB) or block-level
(iSCSI) storage.

Volumes sit at the bottom of the FSx ONTAP hierarchy:
- **File System** provides physical infrastructure (storage, throughput, HA)
- **SVM** provides the logical data server (protocols, endpoints, AD)
- **Volume** provides data containers (capacity, tiering, snapshots, SnapLock)

Key features:
- **Tiering policy** moves cold data to lower-cost capacity pool storage.
- **SnapLock** provides WORM (Write Once Read Many) storage for regulatory
  compliance (SEC 17a-4, HIPAA, FINRA).
- **FlexGroup** distributes a single volume across multiple aggregates for
  high-throughput large-scale workloads.
- **Storage efficiency** enables ONTAP deduplication, compression, and
  compaction to reduce physical storage consumption.

Key design notes:
- `storage_virtual_machine_id` and `name` are ForceNew — changing either
  requires replacing the volume.
- `ontap_volume_type` and `volume_style` are ForceNew (they define the
  volume's fundamental architecture), as are the aggregate layout fields and
  the SnapLock type. Everything else — size, junction path, security style,
  snapshot policy, storage efficiency, tiering, and the mutable SnapLock
  settings — updates in place.
- `junction_path` is the mount point in the SVM namespace. Without it, the
  volume exists but is not accessible via NFS/SMB.
- `security_style` can differ from the SVM's root volume security style,
  allowing mixed workloads within a single SVM.
- Size is set through exactly one arm: `size_in_megabytes` for everyday
  volumes, or `size_in_bytes` for byte-precise sizing and volumes beyond
  2 PiB.
- The deletion controls (`skip_final_backup`,
  `bypass_snaplock_enterprise_retention`, `final_backup_tags`) take effect
  at delete time and must be applied to the volume BEFORE it is destroyed —
  set them early, not in the same change that deletes the volume.
- Credentials, region, and deployment workflow live outside this spec in IaC
  inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOntapVolume
metadata:
  name: test-ontap-volume
  id: awsfxov-test123
  org: test-org
  env: test
spec:
  region: us-west-2
  storage_virtual_machine_id:
    value: svm-0123456789abcdef0
  name: vol_test
  # The byte-precise size arm (1 TiB) so the offline plan proof exercises the
  # int64 → string conversion path alongside the FLEXGROUP aggregate layout,
  # SnapLock, tiering, and the delete-time controls.
  size_in_bytes: 1099511627776
  junction_path: /test
  ontap_volume_type: RW
  volume_style: FLEXGROUP
  security_style: UNIX
  snapshot_policy: default
  storage_efficiency_enabled: true
  copy_tags_to_backups: true
  skip_final_backup: false
  final_backup_tags:
    retention: decommission-audit
  bypass_snaplock_enterprise_retention: false
  tiering_policy:
    name: AUTO
    cooling_period: 31
  snaplock_configuration:
    snaplock_type: ENTERPRISE
    audit_log_volume: false
    privileged_delete: DISABLED
    volume_append_mode_enabled: false
    autocommit_period:
      type: DAYS
      value: 7
    retention_period:
      default_retention:
        type: YEARS
        value: 1
      minimum_retention:
        type: DAYS
        value: 30
      maximum_retention:
        type: YEARS
        value: 5
  aggregate_configuration:
    aggregates:
      - aggr1
      - aggr2
    constituents_per_aggregate: 8
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.storageVirtualMachineId` | `string \| valueFrom` | yes |  | AwsFsxOntapStorageVirtualMachine (`status.outputs.svm_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.sizeInMegabytes` | `int32` |  |  |  |
| `spec.sizeInBytes` | `int64` |  |  |  |
| `spec.junctionPath` | `string` |  |  |  |
| `spec.ontapVolumeType` | `string` |  | `RW` |  |
| `spec.volumeStyle` | `string` |  | `FLEXVOL` |  |
| `spec.securityStyle` | `string` |  |  |  |
| `spec.snapshotPolicy` | `string` |  |  |  |
| `spec.storageEfficiencyEnabled` | `bool` |  |  |  |
| `spec.copyTagsToBackups` | `bool` |  | `false` |  |
| `spec.skipFinalBackup` | `bool` |  | `false` |  |
| `spec.finalBackupTags` | `map<string, string>` |  |  |  |
| `spec.bypassSnaplockEnterpriseRetention` | `bool` |  | `false` |  |
| `spec.tieringPolicy` | `AwsFsxOntapVolumeTieringPolicy` |  |  |  |
| `spec.tieringPolicy.name` | `string` |  |  |  |
| `spec.tieringPolicy.coolingPeriod` | `int32` |  |  |  |
| `spec.snaplockConfiguration` | `AwsFsxOntapVolumeSnaplockConfiguration` |  |  |  |
| `spec.snaplockConfiguration.snaplockType` | `string` | yes |  |  |
| `spec.snaplockConfiguration.auditLogVolume` | `bool` |  | `false` |  |
| `spec.snaplockConfiguration.privilegedDelete` | `string` |  | `DISABLED` |  |
| `spec.snaplockConfiguration.volumeAppendModeEnabled` | `bool` |  | `false` |  |
| `spec.snaplockConfiguration.autocommitPeriod` | `AwsFsxOntapVolumeAutocommitPeriod` |  |  |  |
| `spec.snaplockConfiguration.autocommitPeriod.type` | `string` |  |  |  |
| `spec.snaplockConfiguration.autocommitPeriod.value` | `int32` |  |  |  |
| `spec.snaplockConfiguration.retentionPeriod` | `AwsFsxOntapVolumeRetentionPeriod` |  |  |  |
| `spec.snaplockConfiguration.retentionPeriod.defaultRetention` | `AwsFsxOntapVolumeRetentionDuration` |  |  |  |
| `spec.snaplockConfiguration.retentionPeriod.defaultRetention.type` | `string` |  |  |  |
| `spec.snaplockConfiguration.retentionPeriod.defaultRetention.value` | `int32` |  |  |  |
| `spec.snaplockConfiguration.retentionPeriod.minimumRetention` | `AwsFsxOntapVolumeRetentionDuration` |  |  |  |
| `spec.snaplockConfiguration.retentionPeriod.minimumRetention.type` | `string` |  |  |  |
| `spec.snaplockConfiguration.retentionPeriod.minimumRetention.value` | `int32` |  |  |  |
| `spec.snaplockConfiguration.retentionPeriod.maximumRetention` | `AwsFsxOntapVolumeRetentionDuration` |  |  |  |
| `spec.snaplockConfiguration.retentionPeriod.maximumRetention.type` | `string` |  |  |  |
| `spec.snaplockConfiguration.retentionPeriod.maximumRetention.value` | `int32` |  |  |  |
| `spec.aggregateConfiguration` | `AwsFsxOntapVolumeAggregateConfiguration` |  |  |  |
| `spec.aggregateConfiguration.aggregates` | `[]string` |  |  |  |
| `spec.aggregateConfiguration.constituentsPerAggregate` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.storageVirtualMachineId

`string | valueFrom` · required

The ID of the Storage Virtual Machine that this volume belongs to. Required.
ForceNew — the volume cannot be moved to a different SVM after creation.

The SVM provides the network endpoints, protocol configuration, and Active
Directory integration. All volumes within an SVM share its protocol stack.

- references: AwsFsxOntapStorageVirtualMachine (`status.outputs.svm_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsFsxOntapStorageVirtualMachine, name: <that resource's name>, fieldPath: status.outputs.svm_id}} -- a bare string does not parse

### spec.name

`string` · required

The name of the volume within the ONTAP file system. Required. ForceNew.

This is the ONTAP volume name (not the Planton metadata name). ONTAP volume
names must be alphanumeric plus underscores — hyphens are not allowed. This
name appears in junction paths, SnapMirror relationships, and ONTAP CLI
output.

Constraints: 1-203 characters, alphanumeric and underscore only.

- rule: {"string":{"minLen":"1","maxLen":"203"}}

### spec.sizeInMegabytes

`int32` · optional (explicit presence)

The size of the volume in megabytes. Exactly one of `size_in_megabytes`
and `size_in_bytes` must be set.

Minimum 20 MB. Maximum is constrained by the file system's total storage
capacity. ONTAP volumes support thin provisioning, so the logical size can
exceed the physical capacity available — ONTAP handles overcommit at the
aggregate level. Size can be increased or decreased in place.

Covers volumes up to ~2 PiB (int32 megabytes). For larger volumes or
byte-precise sizing, use `size_in_bytes` instead.

- rule: {"int32":{"gte":20}}

### spec.sizeInBytes

`int64` · optional (explicit presence)

The size of the volume in bytes — the byte-precise sizing arm, and the
only way to size volumes beyond 2 PiB (FlexGroup volumes scale to ~20 PiB).
Exactly one of `size_in_megabytes` and `size_in_bytes` must be set.

Maximum: 22,517,998,000,000,000 bytes (~20 PiB). FLEXGROUP volumes
require at least 100 GiB per constituent.

- rule: {"int64":{"lte":"22517998000000000","gte":"20971520"}}

### spec.junctionPath

`string`

The location in the SVM namespace where this volume is mounted. Clients
access the volume at this path (e.g., mount nfs.svm.example.com:/vol1).

If omitted, the volume is created but not mounted — it exists in ONTAP
but is not accessible via NFS/SMB until a junction path is set.

Must start with "/" and be unique within the SVM. Examples: "/vol1",
"/data/prod", "/shares/finance". Can be changed after creation (the
volume remounts at the new path).

Constraints: 1-255 characters.

- rule: {"string":{"maxLen":"255"}}

### spec.ontapVolumeType

`string` · optional (explicit presence)

The ONTAP volume type. ForceNew.

- "RW": Read-write volume. The standard type for serving data to clients.
- "DP": Data protection volume. A read-only destination for SnapMirror
  replication. DP volumes cannot be mounted until the SnapMirror
  relationship is broken or the volume is converted.

Default: RW

- default: `RW`

### spec.volumeStyle

`string` · optional (explicit presence)

The volume style. ForceNew.

- "FLEXVOL": Traditional ONTAP volume on a single aggregate. Suitable for
  most workloads. Simpler operations and faster metadata performance.
- "FLEXGROUP": A volume distributed across multiple aggregates for high
  throughput and large-scale workloads (hundreds of TBs to PBs). Requires
  aggregate_configuration. Ideal for data lakes, genomics, and media.

Default: FLEXVOL

- default: `FLEXVOL`

### spec.securityStyle

`string`

The security style for this volume's root directory. Controls how file
permissions are evaluated. Can be changed after creation.

- "UNIX": UNIX permissions (mode bits, uid/gid). Best for Linux/NFS.
- "NTFS": Windows ACLs. Best for Windows/SMB with Active Directory.
- "MIXED": Both permission systems coexist. The effective security style
  depends on which protocol last set permissions on a file.

If omitted, inherits from the parent SVM's root_volume_security_style.

### spec.snapshotPolicy

`string`

The name of the ONTAP snapshot policy to apply to this volume. Snapshot
policies control automatic snapshot creation and retention. Can be changed
after creation.

Common policies: "default" (6 hourly + 2 daily + 2 weekly), "none"
(no automatic snapshots). Custom policies can be created via the ONTAP CLI.

Constraints: 1-255 characters.

- rule: {"string":{"maxLen":"255"}}

### spec.storageEfficiencyEnabled

`bool` · optional (explicit presence)

Enable ONTAP storage efficiency features: deduplication, compression, and
compaction. These features reduce physical storage consumption by
identifying and eliminating redundant data blocks. Can be changed after
creation.

Recommended for most workloads (set true). Disable only for workloads
that are already compressed or deduplicated (e.g., encrypted data,
pre-compressed media files) where the CPU overhead provides no benefit.
When omitted, ONTAP applies its own per-volume-type default.

### spec.copyTagsToBackups

`bool` · optional (explicit presence)

Whether to copy resource tags to automatic volume backups.

Default: false

- default: `false`

### spec.skipFinalBackup

`bool` · optional (explicit presence)

Whether to skip the automatic backup that AWS takes when the volume is
deleted. Set to true for development/test volumes where the backup is
unnecessary.

Default: false (a final backup is taken)

- default: `false`

### spec.finalBackupTags

`map<string, string>`

Tags applied to the final backup taken on deletion. Only meaningful when
skip_final_backup is false.

### spec.bypassSnaplockEnterpriseRetention

`bool` · optional (explicit presence)

Whether to allow deletion of a SnapLock Enterprise volume that contains
WORM files with unexpired retention periods. Only relevant for SnapLock
Enterprise volumes — Compliance volumes can never bypass retention.

Default: false

- default: `false`

### spec.tieringPolicy

`AwsFsxOntapVolumeTieringPolicy`

Data tiering policy that controls when and how data moves from primary
SSD storage to lower-cost capacity pool storage. If omitted, the volume
uses the default tiering policy (SNAPSHOT_ONLY). Can be changed after
creation.

- rule: tiering policy name must be 'NONE', 'SNAPSHOT_ONLY', 'AUTO', or 'ALL'
- rule: cooling_period is only valid when tiering policy name is 'AUTO' or 'SNAPSHOT_ONLY'
- rule: cooling_period must be between 2 and 183 days

### spec.tieringPolicy.name

`string`

The tiering policy name. Required when a tiering policy is declared —
declaring the block is choosing a policy; to keep AWS's default
(SNAPSHOT_ONLY), omit tiering_policy entirely.

- "NONE": All data remains on primary SSD storage. No tiering. Use for
  latency-sensitive workloads where all data must be instantly accessible.
- "SNAPSHOT_ONLY": Only snapshot data (point-in-time copies) is tiered.
  Active file system data stays on SSD. The safest tiering option.
- "AUTO": Data not accessed for the cooling period is automatically tiered.
  The most cost-effective option for mixed-access workloads.
- "ALL": All data (including active data) is stored on capacity pool. Only
  metadata stays on SSD. Lowest cost, highest latency for first access.

### spec.tieringPolicy.coolingPeriod

`int32`

The number of days before data is considered "cold" and eligible for
tiering to capacity pool storage. Only applicable when name is "AUTO" or
"SNAPSHOT_ONLY".

Range: 2-183 days. Lower values tier data more aggressively (lower cost,
potentially higher latency for recently accessed data).

### spec.snaplockConfiguration

`AwsFsxOntapVolumeSnaplockConfiguration`

SnapLock configuration for WORM (Write Once Read Many) compliance storage.
When configured, files committed to this volume become immutable for their
retention period. ForceNew for snaplock_type.

SnapLock has two modes:
- ENTERPRISE: Admins can delete WORM files before retention expiry (if
  privileged_delete is enabled). Suitable for internal governance.
- COMPLIANCE: No one — not even the root/admin user or AWS support — can
  delete WORM files before retention expiry. Required for SEC 17a-4, HIPAA,
  and similar regulations.

Once set, the snaplock_type cannot be changed. Choosing the wrong type
requires deleting and recreating the volume.

- rule: snaplock_type must be 'ENTERPRISE' or 'COMPLIANCE'
- rule: privileged_delete must be 'DISABLED', 'ENABLED', or 'PERMANENTLY_DISABLED'

### spec.snaplockConfiguration.snaplockType

`string` · required

The SnapLock retention mode. Required. ForceNew — cannot be changed after
volume creation.

- "ENTERPRISE": Administrative deletion of WORM files is possible (if
  privileged_delete is enabled). Suitable for internal governance policies
  where an escape hatch is acceptable.
- "COMPLIANCE": Immutable. No one can delete WORM files before retention
  expiry — not the root user, not AWS Support, not even the account owner.
  Required for SEC 17a-4 and similar strict regulatory mandates.

- rule: {"string":{"minLen":"1"}}

### spec.snaplockConfiguration.auditLogVolume

`bool` · optional (explicit presence)

Whether this volume is designated as the SnapLock audit log volume.
A single audit log volume per SVM records all SnapLock operations
(file commits, retention changes, privileged deletions).

Default: false

- default: `false`

### spec.snaplockConfiguration.privilegedDelete

`string` · optional (explicit presence)

Controls whether privileged deletion of WORM files is allowed before their
retention period expires. Only meaningful for ENTERPRISE SnapLock.

- "DISABLED": Privileged delete is not allowed (default).
- "ENABLED": Administrators can delete WORM files early.
- "PERMANENTLY_DISABLED": Privileged delete is permanently disabled and
  cannot be re-enabled. Use this for Enterprise volumes that must never
  allow early deletion.

Default: DISABLED

- default: `DISABLED`

### spec.snaplockConfiguration.volumeAppendModeEnabled

`bool` · optional (explicit presence)

Whether volume-append mode is enabled. When enabled, files can be appended
to (new data added at the end) even after being committed to WORM state.
The existing content remains immutable. Useful for log files and audit
trails that need continuous appending.

Default: false

- default: `false`

### spec.snaplockConfiguration.autocommitPeriod

`AwsFsxOntapVolumeAutocommitPeriod`

Configures automatic commitment of files to WORM state after a period of
inactivity. When autocommit is configured, files that have not been modified
for the specified duration are automatically transitioned to WORM state.

This eliminates the need for applications to explicitly commit files.

- rule: autocommit period type must be 'NONE', 'MINUTES', 'HOURS', 'DAYS', 'MONTHS', or 'YEARS'
- rule: autocommit value must match its unit's AWS range (MINUTES 5-65535, HOURS 1-65535, DAYS 1-3650, MONTHS 1-120, YEARS 1-10) and must be unset for NONE

### spec.snaplockConfiguration.autocommitPeriod.type

`string`

The unit of time for the autocommit period. Required when autocommit_period
is declared — the AWS API's AutocommitPeriod.Type member is required, and a
unit-less period cannot be expressed. To disable autocommit, set "NONE"
(or omit autocommit_period entirely; NONE is also AWS's default).

- "NONE": Autocommit is disabled.
- "MINUTES", "HOURS", "DAYS", "MONTHS", "YEARS": The time unit for the
  value field.

### spec.snaplockConfiguration.autocommitPeriod.value

`int32`

The number of time units before an unmodified file is auto-committed to
WORM state. Required for every unit type except "NONE" (which takes no
value). Each unit carries its own AWS range:

- MINUTES: 5-65,535
- HOURS: 1-65,535
- DAYS: 1-3,650
- MONTHS: 1-120
- YEARS: 1-10

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.snaplockConfiguration.retentionPeriod

`AwsFsxOntapVolumeRetentionPeriod`

Configures the default, minimum, and maximum retention periods for WORM
files on this volume. These bounds constrain how long files must be retained
and provide guardrails for retention policy enforcement.

### spec.snaplockConfiguration.retentionPeriod.defaultRetention

`AwsFsxOntapVolumeRetentionDuration`

The default retention period applied to files committed to WORM state
without an explicit retention period.

- rule: retention duration type must be 'SECONDS', 'MINUTES', 'HOURS', 'DAYS', 'MONTHS', 'YEARS', 'INFINITE', or 'UNSPECIFIED'
- rule: retention value must match its unit's AWS range (SECONDS/MINUTES 0-65535, HOURS 0-24, DAYS 0-365, MONTHS 0-12, YEARS 0-100) and must be unset for INFINITE/UNSPECIFIED

### spec.snaplockConfiguration.retentionPeriod.defaultRetention.type

`string`

The unit of time for the retention duration. Required when the duration is
declared — the AWS API's RetentionPeriod.Type member is required, and a
unit-less duration cannot be expressed.

- "SECONDS", "MINUTES", "HOURS", "DAYS", "MONTHS", "YEARS": Standard
  time units. The value field specifies the count.
- "INFINITE": Files are retained forever. The value field takes no value.
- "UNSPECIFIED": No retention period is set. The value field takes no value.

### spec.snaplockConfiguration.retentionPeriod.defaultRetention.value

`int32`

The number of time units for the retention duration. Each unit carries its
own AWS range (0 is legal and means a zero-length duration):

- SECONDS: 0-65,535
- MINUTES: 0-65,535
- HOURS: 0-24
- DAYS: 0-365
- MONTHS: 0-12
- YEARS: 0-100

INFINITE and UNSPECIFIED take no value.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.snaplockConfiguration.retentionPeriod.minimumRetention

`AwsFsxOntapVolumeRetentionDuration`

The minimum retention period. Files cannot have retention periods shorter
than this value.

- rule: retention duration type must be 'SECONDS', 'MINUTES', 'HOURS', 'DAYS', 'MONTHS', 'YEARS', 'INFINITE', or 'UNSPECIFIED'
- rule: retention value must match its unit's AWS range (SECONDS/MINUTES 0-65535, HOURS 0-24, DAYS 0-365, MONTHS 0-12, YEARS 0-100) and must be unset for INFINITE/UNSPECIFIED

### spec.snaplockConfiguration.retentionPeriod.minimumRetention.type

`string`

The unit of time for the retention duration. Required when the duration is
declared — the AWS API's RetentionPeriod.Type member is required, and a
unit-less duration cannot be expressed.

- "SECONDS", "MINUTES", "HOURS", "DAYS", "MONTHS", "YEARS": Standard
  time units. The value field specifies the count.
- "INFINITE": Files are retained forever. The value field takes no value.
- "UNSPECIFIED": No retention period is set. The value field takes no value.

### spec.snaplockConfiguration.retentionPeriod.minimumRetention.value

`int32`

The number of time units for the retention duration. Each unit carries its
own AWS range (0 is legal and means a zero-length duration):

- SECONDS: 0-65,535
- MINUTES: 0-65,535
- HOURS: 0-24
- DAYS: 0-365
- MONTHS: 0-12
- YEARS: 0-100

INFINITE and UNSPECIFIED take no value.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.snaplockConfiguration.retentionPeriod.maximumRetention

`AwsFsxOntapVolumeRetentionDuration`

The maximum retention period. Files cannot have retention periods longer
than this value.

- rule: retention duration type must be 'SECONDS', 'MINUTES', 'HOURS', 'DAYS', 'MONTHS', 'YEARS', 'INFINITE', or 'UNSPECIFIED'
- rule: retention value must match its unit's AWS range (SECONDS/MINUTES 0-65535, HOURS 0-24, DAYS 0-365, MONTHS 0-12, YEARS 0-100) and must be unset for INFINITE/UNSPECIFIED

### spec.snaplockConfiguration.retentionPeriod.maximumRetention.type

`string`

The unit of time for the retention duration. Required when the duration is
declared — the AWS API's RetentionPeriod.Type member is required, and a
unit-less duration cannot be expressed.

- "SECONDS", "MINUTES", "HOURS", "DAYS", "MONTHS", "YEARS": Standard
  time units. The value field specifies the count.
- "INFINITE": Files are retained forever. The value field takes no value.
- "UNSPECIFIED": No retention period is set. The value field takes no value.

### spec.snaplockConfiguration.retentionPeriod.maximumRetention.value

`int32`

The number of time units for the retention duration. Each unit carries its
own AWS range (0 is legal and means a zero-length duration):

- SECONDS: 0-65,535
- MINUTES: 0-65,535
- HOURS: 0-24
- DAYS: 0-365
- MONTHS: 0-12
- YEARS: 0-100

INFINITE and UNSPECIFIED take no value.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.aggregateConfiguration

`AwsFsxOntapVolumeAggregateConfiguration`

Aggregate configuration for FLEXGROUP volumes. Controls how the volume is
distributed across the file system's aggregates. Ignored for FLEXVOL
volumes.

All fields in this block are ForceNew — changing the aggregate layout
requires recreating the volume.

- rule: constituents_per_aggregate must be between 1 and 200

### spec.aggregateConfiguration.aggregates

`[]string`

The list of aggregate names to use for the FlexGroup volume. Each name
must match the pattern "aggr" followed by 1-2 digits (e.g., "aggr1",
"aggr2") — aggregate numbering follows the file system's HA pairs.
Maximum 12 aggregates.

ForceNew — changing this requires volume recreation.

- rule: {"repeated":{"maxItems":"12","items":{"cel":[{"id":"aggregate_name_format","message":"each aggregate must be 'aggr' followed by 1-2 digits (e.g., 'aggr1')","expression":"this.matches('^aggr[0-9]{1,2}$')"}]}}}

### spec.aggregateConfiguration.constituentsPerAggregate

`int32`

The number of FlexGroup constituents (member volumes) to create per
aggregate. The total number of constituents equals
constituents_per_aggregate * len(aggregates).

Higher values increase parallelism but also metadata overhead. Default
in AWS is typically 8.

Range: 1-200. ForceNew.

## Validation Rules

- `name_format`: name must contain only alphanumeric characters and underscores (no hyphens, spaces, or special characters)
- `size_exactly_one_arm`: exactly one of size_in_megabytes and size_in_bytes must be set
- `ontap_volume_type_valid`: ontap_volume_type must be 'RW' or 'DP'
- `volume_style_valid`: volume_style must be 'FLEXVOL' or 'FLEXGROUP'
- `security_style_valid`: security_style must be 'UNIX', 'NTFS', or 'MIXED'
- `junction_path_format`: junction_path must start with '/'
- `aggregate_requires_flexgroup`: aggregate_configuration is only valid for FLEXGROUP volumes (set volume_style to 'FLEXGROUP')

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsFsxOntapVolume, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.volume_id` | `string` | The ID of the volume (e.g., "fsvol-0123456789abcdef0"). Primary identifier used in AWS APIs and CloudWatch metrics for this volume. |
| `status.outputs.arn` | `string` | The Amazon Resource Name of the volume. Used in IAM policies for resource-level permissions (e.g., restricting backup or snapshot operations to specific volumes). |
| `status.outputs.uuid` | `string` | The universally unique identifier of the volume in ONTAP. Used for SnapMirror replication relationships, ONTAP REST API operations, and cross-cluster volume identification. |
| `status.outputs.file_system_id` | `string` | The file system ID that this volume belongs to (e.g., "fs-0123456789abcdef0"). Computed from the SVM's parent file system — useful for constructing CloudWatch metric dimensions or cross-referencing with the file system component. |
| `status.outputs.flexcache_endpoint_type` | `string` | The FlexCache endpoint type for this volume. - "NONE": Not participating in any FlexCache relationship. - "ORIGIN": This volume is the origin (source of truth) for one or more FlexCache volumes. - "CACHE": This volume is a FlexCache (read cache) of a remote origin. |
| `status.outputs.ontap_volume_type` | `string` | The ONTAP volume type confirmation. Either "RW" (read-write) or "DP" (data protection / SnapMirror destination). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.storageVirtualMachineId` | AwsFsxOntapStorageVirtualMachine | `status.outputs.svm_id` |

## See Also

- [Overview](../README.md)
