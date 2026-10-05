# AwsFsxLustreFileSystem

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsFsxLustreFileSystemSpec defines the desired configuration for an Amazon FSx
for Lustre file system — a fully managed, high-performance parallel file system
optimized for fast processing of workloads such as machine learning training,
high performance computing (HPC), video processing, and financial modeling.

Lustre delivers sub-millisecond latencies, hundreds of GB/s throughput, and
millions of IOPS. It integrates natively with Amazon S3, allowing transparent
access to S3 objects as files on the file system.

Choosing a shape:
- SCRATCH_2 + SSD: ephemeral processing storage (no replication, no backups).
  Cheapest way to burn through a dataset; anything worth keeping is exported
  back to S3.
- PERSISTENT_2 + SSD: durable, within-AZ-replicated storage with the highest
  throughput tiers and metadata IOPS tuning. The default choice for new
  production workloads.
- PERSISTENT_2 + INTELLIGENT_TIERING: elastic capacity that grows and shrinks
  with your data (no provisioned storage_capacity_gib); throughput is
  provisioned absolutely via throughput_capacity and reads are served from a
  provisioned SSD read cache. For large, cool datasets with hot subsets.
- PERSISTENT_1 + HDD: lowest cost per TiB for throughput-oriented, sequential
  workloads; requires drive_cache_type and per_unit_storage_throughput 12 or 40.

S3 integration comes in two generations:
- The legacy in-spec import_path/export_path arms (SCRATCH_1/SCRATCH_2/
  PERSISTENT_1 only) create one implicit S3 link at file-system creation and
  are immutable afterwards.
- AwsFsxDataRepositoryAssociation is the modern, first-class link (required
  for PERSISTENT_2): many links per file system, each with its own lifecycle
  and bidirectional auto import/export policies. Prefer it for anything
  persistent.

Key design notes:
- `deployment_type`, `storage_type`, `subnet_id`, `security_group_ids`,
  `kms_key_id`, `backup_id`, `efa_enabled`, `drive_cache_type`, and the S3
  import settings are ForceNew — changing them replaces the file system.
- Lustre file systems are single-AZ — exactly one subnet.
- Storage capacity can grow in place (never shrink; growth on SCRATCH_1
  replaces the file system).
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxLustreFileSystem
metadata:
  org: example-org
  env: dev
  name: my-lustre-fs
  id: awsfxl-my-lustre-fs-dev
spec:
  region: us-west-2
  # Full-surface PERSISTENT_2 shape so the offline plan proof covers the arms
  # the live lanes exclude: metadata IOPS, root squash, logging, backups with
  # final-backup tags, and a maintenance window.
  deployment_type: PERSISTENT_2
  storage_capacity_gib: 2400
  storage_type: SSD
  per_unit_storage_throughput: 250
  data_compression_type: LZ4
  subnet_id:
    value: subnet-0123456789abcdef0
  security_group_ids:
    - value: sg-0123456789abcdef0
  kms_key_id:
    value: arn:aws:kms:us-west-2:123456789012:key/00000000-0000-0000-0000-000000000000
  root_squash_configuration:
    root_squash: "65534:65534"
    no_squash_nids:
      - 10.0.1.6@tcp
  log_configuration:
    destination:
      value: arn:aws:logs:us-west-2:123456789012:log-group:/aws/fsx/lustre
    level: WARN_ERROR
  metadata_configuration:
    mode: USER_PROVISIONED
    iops: 3000
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "04:00"
  copy_tags_to_backups: true
  skip_final_backup: false
  final_backup_tags:
    retention: decommission-audit
  weekly_maintenance_start_time: "7:03:00"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.deploymentType` | `string` |  | `SCRATCH_2` |  |
| `spec.storageCapacityGib` | `int32` |  |  |  |
| `spec.storageType` | `string` |  | `SSD` |  |
| `spec.perUnitStorageThroughput` | `int32` |  |  |  |
| `spec.throughputCapacity` | `int32` |  |  |  |
| `spec.dataCompressionType` | `string` |  | `NONE` |  |
| `spec.fileSystemTypeVersion` | `string` |  |  |  |
| `spec.efaEnabled` | `bool` |  |  |  |
| `spec.driveCacheType` | `string` |  |  |  |
| `spec.dataReadCacheConfiguration` | `AwsFsxLustreFileSystemDataReadCacheConfiguration` |  |  |  |
| `spec.dataReadCacheConfiguration.sizingMode` | `string` | yes |  |  |
| `spec.dataReadCacheConfiguration.sizeGib` | `int32` |  |  |  |
| `spec.subnetId` | `string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.backupId` | `string` |  |  |  |
| `spec.importPath` | `string` |  |  |  |
| `spec.exportPath` | `string` |  |  |  |
| `spec.autoImportPolicy` | `string` |  |  |  |
| `spec.importedFileChunkSize` | `int32` |  |  |  |
| `spec.rootSquashConfiguration` | `AwsFsxLustreFileSystemRootSquashConfiguration` |  |  |  |
| `spec.rootSquashConfiguration.rootSquash` | `string` |  |  |  |
| `spec.rootSquashConfiguration.noSquashNids` | `[]string` |  |  |  |
| `spec.logConfiguration` | `AwsFsxLustreFileSystemLogConfiguration` |  |  |  |
| `spec.logConfiguration.destination` | `string \| valueFrom` |  |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.logConfiguration.level` | `string` |  | `WARN_ERROR` |  |
| `spec.metadataConfiguration` | `AwsFsxLustreFileSystemMetadataConfiguration` |  |  |  |
| `spec.metadataConfiguration.mode` | `string` |  | `AUTOMATIC` |  |
| `spec.metadataConfiguration.iops` | `int32` |  |  |  |
| `spec.automaticBackupRetentionDays` | `int32` |  | `0` |  |
| `spec.dailyAutomaticBackupStartTime` | `string` |  |  |  |
| `spec.copyTagsToBackups` | `bool` |  |  |  |
| `spec.skipFinalBackup` | `bool` |  | `true` |  |
| `spec.finalBackupTags` | `map<string, string>` |  |  |  |
| `spec.weeklyMaintenanceStartTime` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the file system will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.deploymentType

`string` · optional (explicit presence)

Deployment type controlling data durability and performance characteristics.
ForceNew — cannot be changed after creation.

- "SCRATCH_1": temporary storage, no replication. Legacy; fixed 200 MB/s/TiB
  throughput. Growing storage on SCRATCH_1 replaces the file system.
- "SCRATCH_2": temporary storage, no replication, burst throughput up to
  1300 MB/s/TiB. Recommended for short-lived processing jobs.
- "PERSISTENT_1": within-AZ-replicated storage with automatic backups.
  The only deployment type supporting HDD storage and the legacy S3
  import/export arms alongside SCRATCH types.
- "PERSISTENT_2": the current persistent generation — higher throughput
  tiers (125-1000 MB/s/TiB), metadata IOPS configuration, EFA/GPUDirect
  support, and the INTELLIGENT_TIERING storage class. Recommended for new
  production workloads.

Default: SCRATCH_2 (the provider's own default is the legacy SCRATCH_1;
this spec recommends SCRATCH_2 for its strictly better burst throughput at
the same price).

- default: `SCRATCH_2`

### spec.storageCapacityGib

`int32` · optional (explicit presence)

Storage capacity in GiB. Minimum 1200. Valid step sizes depend on the
deployment and storage types (AWS enforces these at create time):

- SCRATCH_2 / PERSISTENT_1 / PERSISTENT_2 (SSD): 1200, 2400, then
  increments of 2400.
- PERSISTENT_1 (HDD): increments of 6000 (12 MB/s/TiB) or 1800
  (40 MB/s/TiB).
- SCRATCH_1: 1200, 2400, 3600, then increments of 3600.

Can be increased in place (never decreased; growth on SCRATCH_1 replaces
the file system). Leave unset in exactly two cases: restoring from a
backup (`backup_id` — capacity comes from the backup), or the
INTELLIGENT_TIERING storage class (capacity is elastic and never
provisioned).

- rule: {"int32":{"gte":1200}}

### spec.storageType

`string` · optional (explicit presence)

Storage class backing the file system. ForceNew.

- "SSD": solid-state drives, sub-millisecond latency. Required for
  SCRATCH_1/SCRATCH_2 and the default for both PERSISTENT generations.
- "HDD": hard disk drives — lowest cost per TiB for sequential,
  throughput-oriented workloads. PERSISTENT_1 only; requires
  drive_cache_type and per_unit_storage_throughput 12 or 40.
- "INTELLIGENT_TIERING": elastic, pay-for-what-you-store capacity with a
  provisioned SSD read cache. PERSISTENT_2 only; requires
  throughput_capacity, data_read_cache_configuration, and
  metadata_configuration, and forbids provisioned storage_capacity_gib.

Default: SSD

- default: `SSD`

### spec.perUnitStorageThroughput

`int32` · optional (explicit presence)

Throughput per unit of storage in MB/s/TiB, for provisioned-capacity
PERSISTENT deployments (SSD and HDD storage). Invalid for SCRATCH types
and for INTELLIGENT_TIERING (which provisions throughput absolutely via
throughput_capacity instead).

Valid values by deployment and storage type:
- PERSISTENT_1 + SSD: 50, 100, 200
- PERSISTENT_1 + HDD: 12, 40
- PERSISTENT_2 + SSD: 125, 250, 500, 1000

Can be changed in place on PERSISTENT_2 (throughput scales while the file
system stays online), except when efa_enabled pins it at creation.

- rule: {"int32":{"in":[12,40,50,100,125,200,250,500,1000]}}

### spec.throughputCapacity

`int32` · optional (explicit presence)

Absolute throughput in MB/s for the INTELLIGENT_TIERING storage class.
Must be 4000 or a multiple of 4000. Required when (and only meaningful
when) storage_type is INTELLIGENT_TIERING — provisioned-capacity file
systems size their throughput per-TiB via per_unit_storage_throughput
instead.

- rule: {"int32":{"gte":4000}}

### spec.dataCompressionType

`string` · optional (explicit presence)

Enable LZ4 data compression for all data on the file system. Reduces
storage consumption and can improve throughput for compressible data.
Can be changed after creation (new writes are compressed; existing data
is not rewritten).

- "NONE": no compression (default).
- "LZ4": LZ4 compression.

- default: `NONE`

### spec.fileSystemTypeVersion

`string`

Lustre file system version, in "x.y" format (e.g., "2.12", "2.15").
Leave empty to use the latest version supported by the deployment type.
Upgrades apply in place; a downgrade replaces the file system.

- rule: file_system_type_version must be in the format x.y (e.g., '2.12', '2.15')

### spec.efaEnabled

`bool`

Enable Elastic Fabric Adapter (EFA) and GPUDirect Storage (GDS) support,
giving GPU instances a direct, OS-bypass data path to the file system.
ForceNew — must be decided at creation, and while enabled it also pins
per_unit_storage_throughput. Requires PERSISTENT_2 with
metadata_configuration, and an EFA-enabled security group attached via
security_group_ids.

### spec.driveCacheType

`string`

Read cache for HDD-backed file systems. PERSISTENT_1 + HDD only, and
REQUIRED there (AWS's contract for HDD file systems). ForceNew.

- "READ": provision an SSD read cache sized to 20% of storage capacity —
  gives HDD file systems SSD-like latency for frequently read data.
- "NONE": no read cache.

### spec.dataReadCacheConfiguration

`AwsFsxLustreFileSystemDataReadCacheConfiguration`

Provisioned SSD read cache for the INTELLIGENT_TIERING storage class.
Required when storage_type is INTELLIGENT_TIERING; invalid otherwise.

- rule: sizing_mode must be 'NO_CACHE', 'USER_PROVISIONED', or 'PROPORTIONAL_TO_THROUGHPUT_CAPACITY'
- rule: size_gib can only be set when sizing_mode is 'USER_PROVISIONED'

### spec.dataReadCacheConfiguration.sizingMode

`string` · required

How the read cache is sized.

- "PROPORTIONAL_TO_THROUGHPUT_CAPACITY": AWS sizes the cache from the
  provisioned throughput (the recommended hands-off mode).
- "USER_PROVISIONED": you set the exact cache size via `size_gib`.
- "NO_CACHE": no SSD read cache (every read pays the tiered-storage
  latency; only for purely archival access patterns).

- rule: {"required":true}

### spec.dataReadCacheConfiguration.sizeGib

`int32` · optional (explicit presence)

Read cache size in GiB when sizing_mode is "USER_PROVISIONED". The valid
range scales with throughput_capacity: for every 4000 MB/s provisioned,
AWS accepts 32 GiB to 131072 GiB of cache (e.g., 8000 MB/s allows
64-262144 GiB).

- rule: {"int32":{"gte":32}}

### spec.subnetId

`string | valueFrom` · required

Subnet for the file system's network interfaces. Required. ForceNew.

Lustre file systems are single-AZ — exactly one subnet is supported. All
compute resources mounting this file system must have network connectivity
to this subnet.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom`

Security groups for the file system's network interfaces. ForceNew.
Up to 50. When empty, AWS attaches the VPC's default security group.

Must allow Lustre traffic between the file system and its clients:
- TCP port 988 (Lustre protocol)
- TCP ports 1018-1023 (Lustre data channels)
EFA-enabled file systems additionally need an EFA-enabled security group.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"50"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key ARN for encryption at rest. ForceNew. When
omitted, the file system uses the AWS-managed FSx key (all Lustre file
systems are encrypted at rest); this field upgrades to a customer-managed
key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.backupId

`string`

ID of an FSx backup to restore this file system from ("backup-...").
ForceNew. When set, storage capacity and most file-system settings come
from the backup; leave storage_capacity_gib unset.

### spec.importPath

`string`

S3 URI to link as the file system's data repository (e.g., "s3://my-bucket"
or "s3://my-bucket/prefix"). ForceNew. Not supported on PERSISTENT_2 —
use AwsFsxDataRepositoryAssociation there (and prefer it for PERSISTENT_1
too; this arm is the legacy single-link generation).

When set, the file system imports file metadata from S3 at creation; file
data is lazy-loaded on first access.

- rule: import_path must be an S3 URI beginning with s3:// (3-900 characters)

### spec.exportPath

`string`

S3 URI where changed files are exported back to S3 (e.g.,
"s3://my-bucket/output/"). ForceNew. Requires import_path and must use the
same bucket. Set equal to import_path to overwrite objects in place; when
omitted AWS exports to "s3://{import bucket}/FSxLustre{creation timestamp}".

- rule: export_path must be an S3 URI beginning with s3:// (3-900 characters)

### spec.autoImportPolicy

`string`

How the file system stays in sync as objects change in the linked S3
bucket. Requires import_path.

- "NONE": import listings only at creation (default).
- "NEW": import metadata for objects added to the bucket.
- "NEW_CHANGED": also update metadata for changed objects.
- "NEW_CHANGED_DELETED": also delete file metadata when objects are
  deleted from the bucket.

### spec.importedFileChunkSize

`int32` · optional (explicit presence)

Stripe configuration for imported files: the maximum amount of data per
file (in MiB) stored on a single physical disk. Range: 1-512000. Requires
import_path; AWS defaults to 1024. ForceNew.

- rule: {"int32":{"lte":512000,"gte":1}}

### spec.rootSquashConfiguration

`AwsFsxLustreFileSystemRootSquashConfiguration`

Root squash configuration — maps root (UID/GID 0) clients to an
unprivileged identity so no mounting host has automatic root access to
the file system's contents. A POSIX-security hardening measure for
multi-tenant compute fleets. Can be changed after creation.

### spec.rootSquashConfiguration.rootSquash

`string`

The UID:GID pair that root users are squashed to (e.g., "65534:65534" for
nobody:nogroup). Both values range 0-4294967294. Setting this enables
root squash; omit the whole block to leave root access unrestricted.

- rule: root_squash must be in UID:GID format (e.g., '65534:65534')

### spec.rootSquashConfiguration.noSquashNids

`[]string`

Lustre NIDs (network identifiers) of clients EXEMPT from root squash —
administrative hosts that keep real root access. Format: an IPv4 address
(ranges allowed in brackets) followed by "@tcp", e.g. "10.0.1.6@tcp" or
"10.0.[2-10].[1-255]@tcp".

- rule: {"repeated":{"items":{"cel":[{"id":"nid_format","message":"each NID must be an IPv4 address or bracketed range followed by '@tcp' (e.g., '10.0.1.6@tcp')","expression":"this.matches('^([0-9\\\\[\\\\]-]*\\\\.){3}([0-9\\\\[\\\\]-]*)@tcp$')"}]}}}

### spec.logConfiguration

`AwsFsxLustreFileSystemLogConfiguration`

CloudWatch logging for data repository events (imports/exports between
the file system and its linked S3 repositories). Useful for auditing
repository task failures and lifecycle debugging.

- rule: level must be 'DISABLED', 'WARN_ONLY', 'ERROR_ONLY', or 'WARN_ERROR'

### spec.logConfiguration.destination

`string | valueFrom`

CloudWatch Logs log group ARN to receive the events. When set, the log
group must exist and have a resource policy allowing FSx to write to it.
When left empty with a level set, FSx logs to its DEFAULT log group
(/aws/fsx/lustre) -- logging is NOT disabled by omitting the
destination. Logging is off only when the whole log_configuration
message is absent (or the level is DISABLED).

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.logConfiguration.level

`string` · optional (explicit presence)

Log level controlling which events are logged.

- "DISABLED": no logging (default when log_configuration is omitted).
- "WARN_ONLY": log warning-level events only.
- "ERROR_ONLY": log error-level events only.
- "WARN_ERROR": log both warning and error events.

Default: WARN_ERROR

- default: `WARN_ERROR`

### spec.metadataConfiguration

`AwsFsxLustreFileSystemMetadataConfiguration`

Metadata performance configuration. PERSISTENT_2 only (required there
when storage_type is INTELLIGENT_TIERING or efa_enabled is set). Controls
the metadata IOPS available for file creation, listing, and similar
operations. Most workloads perform well with AUTOMATIC mode.

- rule: mode must be 'AUTOMATIC' or 'USER_PROVISIONED'
- rule: iops can only be set when mode is 'USER_PROVISIONED'

### spec.metadataConfiguration.mode

`string` · optional (explicit presence)

Metadata IOPS mode.

- "AUTOMATIC": FSx scales metadata IOPS with the file system's storage
  capacity. The right choice for most workloads.
- "USER_PROVISIONED": you specify exact metadata IOPS — higher
  metadata performance independent of storage size, at additional cost.

Default: AUTOMATIC

- default: `AUTOMATIC`

### spec.metadataConfiguration.iops

`int32` · optional (explicit presence)

Metadata IOPS when mode is "USER_PROVISIONED". IOPS can be increased in
place; a decrease replaces the file system.

Valid values: 1500, 3000, 6000, then multiples of 12000 up to 192000.

- rule: {"int32":{"in":[1500,3000,6000,12000,24000,36000,48000,60000,72000,84000,96000,108000,120000,132000,144000,156000,168000,180000,192000]}}

### spec.automaticBackupRetentionDays

`int32` · optional (explicit presence)

Number of days to retain automatic backups. Range: 0-90; 0 disables
automatic backups. Backups are only supported on PERSISTENT deployments.

Default: 0 (no automatic backups)

- default: `0`
- rule: {"int32":{"lte":90,"gte":0}}

### spec.dailyAutomaticBackupStartTime

`string`

Daily UTC time to start automatic backups, in "HH:MM" format (e.g.,
"05:00"). Only meaningful when automatic backups are enabled; when
omitted AWS chooses a window.

- rule: daily_automatic_backup_start_time must be in 24-hour HH:MM format (e.g., '05:00')

### spec.copyTagsToBackups

`bool`

Copy the file system's tags to its automatic backups. ForceNew.

### spec.skipFinalBackup

`bool` · optional (explicit presence)

Skip creating a final backup when the file system is deleted. Applies to
PERSISTENT deployments (SCRATCH file systems have no backups).

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

Weekly UTC maintenance window in "d:HH:MM" format where d is the day of
the week (1=Monday, 7=Sunday). Example: "1:05:00" for Monday 05:00 UTC.

- rule: weekly_maintenance_start_time must be in d:HH:MM format where d is 1 (Monday) through 7 (Sunday), e.g., '1:05:00'

## Validation Rules

- `deployment_type_valid`: deployment_type must be 'SCRATCH_1', 'SCRATCH_2', 'PERSISTENT_1', or 'PERSISTENT_2'
- `storage_type_valid`: storage_type must be 'SSD', 'HDD', or 'INTELLIGENT_TIERING'
- `storage_capacity_required`: storage_capacity_gib is required (leave it unset only when restoring from backup_id or when storage_type is 'INTELLIGENT_TIERING')
- `hdd_requires_persistent_1`: storage_type 'HDD' is only supported with deployment_type 'PERSISTENT_1'
- `intelligent_tiering_contract`: storage_type 'INTELLIGENT_TIERING' requires deployment_type 'PERSISTENT_2' plus throughput_capacity, data_read_cache_configuration, and metadata_configuration
- `drive_cache_type_hdd_contract`: drive_cache_type ('READ' or 'NONE') is required when storage_type is 'HDD' and invalid otherwise
- `per_unit_throughput_requires_persistent`: per_unit_storage_throughput applies only to PERSISTENT_1 or PERSISTENT_2 deployments with provisioned (SSD/HDD) storage
- `throughput_capacity_intelligent_tiering_contract`: throughput_capacity (a multiple of 4000 MB/s) is required when storage_type is 'INTELLIGENT_TIERING' and invalid otherwise
- `data_read_cache_requires_intelligent_tiering`: data_read_cache_configuration is only supported when storage_type is 'INTELLIGENT_TIERING'
- `efa_requires_persistent_2_metadata`: efa_enabled requires deployment_type 'PERSISTENT_2' and metadata_configuration
- `data_compression_type_valid`: data_compression_type must be 'NONE' or 'LZ4'
- `import_path_not_persistent_2`: import_path is not supported on PERSISTENT_2 deployments — use an AwsFsxDataRepositoryAssociation instead
- `export_requires_import`: export_path requires import_path to be set
- `auto_import_policy_valid`: auto_import_policy must be 'NONE', 'NEW', 'NEW_CHANGED', or 'NEW_CHANGED_DELETED', and requires import_path
- `chunk_size_requires_import`: imported_file_chunk_size requires import_path to be set
- `metadata_config_requires_persistent_2`: metadata_configuration is only supported on PERSISTENT_2 deployment type
- `backup_restore_excludes_capacity`: storage_capacity_gib cannot be set when restoring from backup_id (capacity comes from the backup)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsFsxLustreFileSystem, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.file_system_id` | `string` | The ID of the file system (e.g., "fs-0123456789abcdef0"). This is the primary identifier used by EKS PersistentVolumes, ECS task definitions, and AWS Batch compute environments. |
| `status.outputs.file_system_arn` | `string` | The Amazon Resource Name of the file system. Used in IAM policies for resource-level permissions and for creating data repository associations. |
| `status.outputs.dns_name` | `string` | The DNS name for the file system (e.g., "fs-0123456789abcdef0.fsx.us-east-1.amazonaws.com"). Used in mount commands together with `mount_name`: mount -t lustre <dns_name>@tcp:/<mount_name> /mnt/fsx |
| `status.outputs.mount_name` | `string` | The Lustre mount name for the file system (e.g., "fsx" or "2p5wpbwj"). Used in the mount command together with `dns_name`. This value is auto-generated by AWS and is required to construct the full mount path. |
| `status.outputs.network_interface_ids` | `[]string` | The network interface IDs created for the file system, in order. Lustre file systems create one ENI in the specified subnet. Useful for security group debugging and network troubleshooting. |
| `status.outputs.vpc_id` | `string` | The VPC ID in which the file system was created. Computed from the subnet. Useful for constructing security group rules and verifying network placement. |
| `status.outputs.file_system_type_version` | `string` | The Lustre file system type version (e.g., "2.12", "2.15"). Reflects the actual version deployed, which may differ from the requested version if the field was left empty (AWS chooses the latest). |
| `status.outputs.owner_id` | `string` | The AWS account ID of the file system owner. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetId` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.logConfiguration.destination` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsFsxDataRepositoryAssociation | `spec.fileSystemId` | `status.outputs.file_system_id` |

## See Also

- [Overview](../README.md)
