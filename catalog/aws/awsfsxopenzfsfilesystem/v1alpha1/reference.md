# AwsFsxOpenzfsFileSystem

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsFsxOpenzfsFileSystemSpec defines the desired configuration for an Amazon FSx
for OpenZFS file system — a fully managed NFS file system built on the OpenZFS
file system, providing sub-millisecond latency, snapshots, cloning, and data
compression with support for NFSv3, NFSv4.0, NFSv4.1, and NFSv4.2 protocols.

FSx for OpenZFS delivers up to 10 GB/s throughput and over 1 million IOPS,
making it suitable for general-purpose NFS workloads including web serving,
content management, analytics, DevOps, CI/CD, and application data stores.

Deployment generations:
- SINGLE_AZ_1: first generation, lower throughput ceiling (64-4096 MB/s).
- SINGLE_AZ_2: current generation (160-10240 MB/s). Recommended default.
- SINGLE_AZ_HA_1 / SINGLE_AZ_HA_2: the HA variants of each generation — an
  active/standby file-server pair inside ONE availability zone (failover
  without cross-AZ data transfer costs).
- MULTI_AZ_1: active/standby across TWO availability zones — the highest
  availability tier. Requires two subnets, a preferred_subnet_id, and
  (typically) route_table_ids; supports the INTELLIGENT_TIERING storage
  class.

Key design notes:
- `deployment_type`, `subnet_ids`, `security_group_ids`, `kms_key_id`,
  `storage_type`, `backup_id`, `preferred_subnet_id`, and
  `endpoint_ip_address_range` are ForceNew — changing them replaces the
  file system.
- The root volume is created with the file system; its NFS exports,
  compression, and quotas are configured via `root_volume_configuration`
  (note: `copy_tags_to_snapshots` inside it is ForceNew).
- Child volumes are independent lifecycle resources and are NOT managed by
  this component. Use the root_volume_id output to create child volumes.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOpenzfsFileSystem
metadata:
  org: example-org
  env: dev
  name: my-openzfs-fs
  id: awsfxz-my-openzfs-fs-dev
spec:
  region: us-west-2
  # Full-surface SINGLE_AZ_2 shape so the offline plan proof covers the arms
  # the live lanes exclude: user-provisioned IOPS, quotas, backups with
  # final-backup tags, and the cascading-delete opt-in.
  deployment_type: SINGLE_AZ_2
  storage_capacity_gib: 256
  throughput_capacity: 160
  subnet_ids:
    - value: subnet-0123456789abcdef0
  security_group_ids:
    - value: sg-0123456789abcdef0
  kms_key_id:
    value: arn:aws:kms:us-west-2:123456789012:key/00000000-0000-0000-0000-000000000000
  disk_iops_configuration:
    mode: USER_PROVISIONED
    iops: 5000
  root_volume_configuration:
    data_compression_type: ZSTD
    record_size_kib: 128
    copy_tags_to_snapshots: true
    nfs_exports:
      client_configurations:
        - clients: "*"
          options:
            - rw
            - crossmnt
            - no_root_squash
    user_and_group_quotas:
      - id: 1000
        storage_capacity_quota_gib: 100
        type: USER
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "04:00"
  copy_tags_to_backups: true
  copy_tags_to_volumes: true
  skip_final_backup: false
  final_backup_tags:
    retention: decommission-audit
  delete_options:
    - DELETE_CHILD_VOLUMES_AND_SNAPSHOTS
  weekly_maintenance_start_time: "7:03:00"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.deploymentType` | `string` |  | `SINGLE_AZ_2` |  |
| `spec.storageCapacityGib` | `int32` |  |  |  |
| `spec.storageType` | `string` |  | `SSD` |  |
| `spec.throughputCapacity` | `int32` |  |  |  |
| `spec.readCacheConfiguration` | `AwsFsxOpenzfsFileSystemReadCacheConfiguration` |  |  |  |
| `spec.readCacheConfiguration.sizingMode` | `string` | yes |  |  |
| `spec.readCacheConfiguration.sizeGib` | `int32` |  |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.preferredSubnetId` | `string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.endpointIpAddressRange` | `string` |  |  |  |
| `spec.routeTableIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.route_table_id`) |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.backupId` | `string` |  |  |  |
| `spec.diskIopsConfiguration` | `AwsFsxOpenzfsFileSystemDiskIopsConfiguration` |  |  |  |
| `spec.diskIopsConfiguration.mode` | `string` |  | `AUTOMATIC` |  |
| `spec.diskIopsConfiguration.iops` | `int32` |  |  |  |
| `spec.rootVolumeConfiguration` | `AwsFsxOpenzfsFileSystemRootVolumeConfiguration` |  |  |  |
| `spec.rootVolumeConfiguration.dataCompressionType` | `string` |  | `NONE` |  |
| `spec.rootVolumeConfiguration.nfsExports` | `AwsFsxOpenzfsFileSystemNfsExports` |  |  |  |
| `spec.rootVolumeConfiguration.nfsExports.clientConfigurations` | `[]AwsFsxOpenzfsFileSystemNfsClientConfiguration` | yes |  |  |
| `spec.rootVolumeConfiguration.nfsExports.clientConfigurations[].clients` | `string` | yes |  |  |
| `spec.rootVolumeConfiguration.nfsExports.clientConfigurations[].options` | `[]string` | yes |  |  |
| `spec.rootVolumeConfiguration.readOnly` | `bool` |  |  |  |
| `spec.rootVolumeConfiguration.recordSizeKib` | `int32` |  | `128` |  |
| `spec.rootVolumeConfiguration.userAndGroupQuotas` | `[]AwsFsxOpenzfsFileSystemUserAndGroupQuota` |  |  |  |
| `spec.rootVolumeConfiguration.userAndGroupQuotas[].id` | `int32` |  |  |  |
| `spec.rootVolumeConfiguration.userAndGroupQuotas[].storageCapacityQuotaGib` | `int32` |  |  |  |
| `spec.rootVolumeConfiguration.userAndGroupQuotas[].type` | `string` | yes |  |  |
| `spec.rootVolumeConfiguration.copyTagsToSnapshots` | `bool` |  |  |  |
| `spec.automaticBackupRetentionDays` | `int32` |  | `0` |  |
| `spec.dailyAutomaticBackupStartTime` | `string` |  |  |  |
| `spec.copyTagsToBackups` | `bool` |  |  |  |
| `spec.copyTagsToVolumes` | `bool` |  |  |  |
| `spec.skipFinalBackup` | `bool` |  | `true` |  |
| `spec.finalBackupTags` | `map<string, string>` |  |  |  |
| `spec.deleteOptions` | `[]string` |  |  |  |
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

- "SINGLE_AZ_1": first-generation single-AZ. Throughput 64-4096 MB/s.
- "SINGLE_AZ_2": current-generation single-AZ. Throughput 160-10240 MB/s.
  Recommended for most workloads.
- "SINGLE_AZ_HA_1" / "SINGLE_AZ_HA_2": HA (active/standby) variants of the
  two generations within one AZ — automatic failover without cross-AZ
  data transfer charges.
- "MULTI_AZ_1": active/standby across two AZs. Requires two subnets in
  different AZs plus preferred_subnet_id; the only type supporting the
  INTELLIGENT_TIERING storage class.

Default: SINGLE_AZ_2

- default: `SINGLE_AZ_2`

### spec.storageCapacityGib

`int32` · optional (explicit presence)

Storage capacity in GiB for provisioned (SSD) storage. Range: 64-524288.
Can be increased after creation but never decreased.

Leave unset in exactly two cases: the INTELLIGENT_TIERING storage class
(capacity is elastic and never provisioned) or a backup restore
(`backup_id` — capacity comes from the backup).

- rule: {"int32":{"lte":524288,"gte":64}}

### spec.storageType

`string` · optional (explicit presence)

Storage class backing the file system. ForceNew.

- "SSD": provisioned solid-state storage (default; all deployment types).
- "INTELLIGENT_TIERING": elastic, pay-for-what-you-store capacity with a
  provisioned SSD read cache. MULTI_AZ_1 only; forbids
  storage_capacity_gib and requires read_cache_configuration.

Default: SSD

- default: `SSD`

### spec.throughputCapacity

`int32`

Throughput capacity in MB/s. Required.

Valid values by deployment generation (the values AWS accepts):
- SINGLE_AZ_1: 64, 128, 256, 512, 1024, 2048, 3072, 4096
- SINGLE_AZ_2 / MULTI_AZ_1: 160, 320, 640, 1280, 2560, 3840, 5120,
  7680, 10240
- The HA variants follow their generation's value set (validated by AWS
  at create time).

Can be changed after creation to scale performance up or down.

- rule: {"int32":{"gt":0}}

### spec.readCacheConfiguration

`AwsFsxOpenzfsFileSystemReadCacheConfiguration`

Provisioned SSD read cache for the INTELLIGENT_TIERING storage class.
Required when storage_type is INTELLIGENT_TIERING; invalid otherwise.

- rule: sizing_mode must be 'NO_CACHE', 'USER_PROVISIONED', or 'PROPORTIONAL_TO_THROUGHPUT_CAPACITY'
- rule: size_gib can only be set when sizing_mode is 'USER_PROVISIONED'

### spec.readCacheConfiguration.sizingMode

`string` · required

How the read cache is sized.

- "PROPORTIONAL_TO_THROUGHPUT_CAPACITY": AWS sizes the cache from the
  provisioned throughput (the recommended hands-off mode).
- "USER_PROVISIONED": you set the exact cache size via `size_gib`.
- "NO_CACHE": no SSD read cache (every read pays the tiered-storage
  latency; only for purely archival access patterns).

- rule: {"required":true}

### spec.readCacheConfiguration.sizeGib

`int32` · optional (explicit presence)

Read cache size in GiB when sizing_mode is "USER_PROVISIONED".

- rule: {"int32":{"gte":32}}

### spec.subnetIds

`[]string | valueFrom` · required

Subnet IDs for the file system's network interfaces. Required. ForceNew.

- Single-AZ types (incl. the HA variants): exactly one subnet.
- MULTI_AZ_1: exactly two subnets in different availability zones.

All compute resources mounting this file system must have network
connectivity to these subnets.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom`

Security groups for the file system's network interfaces. ForceNew.
Up to 50. When empty, AWS attaches the VPC's default security group.

Must allow NFS traffic between the file system and its clients:
- TCP port 111 (portmapper)
- TCP port 2049 (NFS)
- TCP ports 20001-20003 (NFS mount)

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"50"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.preferredSubnetId

`string | valueFrom`

Preferred subnet for the active file server in a MULTI_AZ_1 deployment.
ForceNew. REQUIRED for MULTI_AZ_1 (AWS's contract) and invalid for the
single-AZ types. Must be one of the subnets in subnet_ids.

In a failover event, the standby file server in the other subnet takes
over.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.endpointIpAddressRange

`string`

IP address range for the file system endpoints in a MULTI_AZ_1
deployment. ForceNew. Must be a CIDR block within the VPC's CIDR range
that does not overlap with any existing subnets; AWS assigns floating IPs
from this range for seamless failover. When omitted, AWS picks a range.

The provider does not validate this field's format for OpenZFS (unlike its
ONTAP sibling) — the CEL below mirrors the ONTAP kind so a malformed range
fails at validate instead of at the AWS API.

- rule: endpoint_ip_address_range must be an IPv4 CIDR block (e.g., '198.19.0.0/24')

### spec.routeTableIds

`[]string | valueFrom`

Route tables in which AWS manages routes to the floating file-system
endpoints of a MULTI_AZ_1 deployment. Specify every VPC route table
associated with the subnets your NFS clients live in; when omitted, AWS
uses the VPC's default route table. Up to 50.

Reference an AwsSubnet's route_table_id output when the subnet owns its
table, or the AwsVpc's main/default route-table outputs when subnets ride
the VPC main table; literals also work.

- references: AwsSubnet (`status.outputs.route_table_id`)
- rule: {"repeated":{"maxItems":"50"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.route_table_id}} -- a bare string does not parse

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key ARN for encryption at rest. ForceNew — the KMS
key cannot be changed after creation. When omitted, the file system uses
the AWS-managed FSx key. All OpenZFS file systems are encrypted at rest
by default; this field upgrades to a customer-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.backupId

`string`

ID of an FSx backup to restore this file system from ("backup-...").
ForceNew. When set, capacity and most settings come from the backup;
leave storage_capacity_gib unset.

### spec.diskIopsConfiguration

`AwsFsxOpenzfsFileSystemDiskIopsConfiguration`

SSD IOPS configuration for the file system. Controls the total
provisioned IOPS. When omitted, AWS uses AUTOMATIC mode which scales IOPS
with storage (3 IOPS per GiB).

- rule: mode must be 'AUTOMATIC' or 'USER_PROVISIONED'
- rule: iops can only be set when mode is 'USER_PROVISIONED'

### spec.diskIopsConfiguration.mode

`string` · optional (explicit presence)

IOPS provisioning mode.

- "AUTOMATIC": IOPS scale automatically based on storage capacity.
  Provides 3 IOPS per GiB of storage, up to the deployment type limit.
- "USER_PROVISIONED": you specify the exact IOPS via the `iops` field.
  Allows higher performance independent of storage size but at extra cost.

Default: AUTOMATIC

- default: `AUTOMATIC`

### spec.diskIopsConfiguration.iops

`int32` · optional (explicit presence)

Total SSD IOPS provisioned. Only valid when mode is "USER_PROVISIONED".

Ceilings by deployment generation: 160,000 (SINGLE_AZ_1) and 400,000
(SINGLE_AZ_2); MULTI_AZ_1 and the HA variants are validated by AWS at
create time.

- rule: {"int32":{"gte":0}}

### spec.rootVolumeConfiguration

`AwsFsxOpenzfsFileSystemRootVolumeConfiguration`

Configuration for the file system's root volume. The root volume is
automatically created with the file system and serves as the default NFS
mount target. Settings here control compression, NFS access, quotas, and
record size.

When omitted, the root volume uses defaults: no compression, no NFS
access restrictions, 128 KiB record size.

- rule: data_compression_type must be 'NONE', 'ZSTD', or 'LZ4'
- rule: record_size_kib must be one of: 4, 8, 16, 32, 64, 128, 256, 512, 1024

### spec.rootVolumeConfiguration.dataCompressionType

`string` · optional (explicit presence)

Data compression type applied to all data on the root volume. Reduces
storage consumption and can improve throughput for compressible data.

- "NONE": no compression (default).
- "ZSTD": Zstandard compression. Best compression ratio.
- "LZ4": LZ4 compression. Faster with lower CPU overhead.

Default: NONE

- default: `NONE`

### spec.rootVolumeConfiguration.nfsExports

`AwsFsxOpenzfsFileSystemNfsExports`

NFS export configuration for the root volume. Controls which clients can
mount the volume and with what permissions. When omitted, the volume uses
default NFS settings (accessible from within the VPC).

### spec.rootVolumeConfiguration.nfsExports.clientConfigurations

`[]AwsFsxOpenzfsFileSystemNfsClientConfiguration` · required

NFS client configurations. Each entry defines which clients can access the
volume and with what mount options. Up to 25 client configurations.

- rule: {"repeated":{"minItems":"1","maxItems":"25"}}

### spec.rootVolumeConfiguration.nfsExports.clientConfigurations[].clients

`string` · required

Client specification: an IP address, CIDR block, or wildcard (*).
1-128 characters.

Examples: "*" (all clients), "10.0.0.0/16", "192.168.1.100"

- rule: {"string":{"minLen":"1","maxLen":"128"}}

### spec.rootVolumeConfiguration.nfsExports.clientConfigurations[].options

`[]string` · required

NFS mount options for the specified clients. At least one option is
required; each option is 1-128 characters.

Common options:
- "rw" (read-write) or "ro" (read-only)
- "crossmnt" (allow traversal into child volumes)
- "root_squash" (map root to anonymous) or "no_root_squash"
- "sync" or "async"

Up to 20 options.

- rule: {"repeated":{"minItems":"1","maxItems":"20","items":{"string":{"minLen":"1","maxLen":"128"}}}}

### spec.rootVolumeConfiguration.readOnly

`bool`

Whether the root volume is read-only. When true, clients can mount the
volume but cannot write to it. Useful for shared reference data.

### spec.rootVolumeConfiguration.recordSizeKib

`int32` · optional (explicit presence)

ZFS record size in KiB. Affects read/write performance characteristics.
Larger record sizes are better for sequential I/O (analytics, streaming).
Smaller record sizes are better for random I/O (databases, OLTP).

Valid values: 4, 8, 16, 32, 64, 128, 256, 512, 1024.

Default: 128

- default: `128`

### spec.rootVolumeConfiguration.userAndGroupQuotas

`[]AwsFsxOpenzfsFileSystemUserAndGroupQuota`

Per-user and per-group storage quotas for the root volume. Limits how
much storage individual users or groups can consume.

- rule: type must be 'USER' or 'GROUP'

### spec.rootVolumeConfiguration.userAndGroupQuotas[].id

`int32`

The numeric user ID (UID) or group ID (GID). Range: 0-2147483647.

Common values: 0 (root), 1000+ (regular users/groups).

- rule: {"int32":{"gte":0}}

### spec.rootVolumeConfiguration.userAndGroupQuotas[].storageCapacityQuotaGib

`int32`

Storage capacity quota in GiB. The maximum amount of storage this user or
group can consume on the volume. Range: 0-2147483647.

Set to 0 to remove quota restrictions for this user/group.

- rule: {"int32":{"gte":0}}

### spec.rootVolumeConfiguration.userAndGroupQuotas[].type

`string` · required

Quota type: "USER" for per-user quota, "GROUP" for per-group quota.

- rule: {"string":{"minLen":"1"}}

### spec.rootVolumeConfiguration.copyTagsToSnapshots

`bool`

Copy tags from the root volume to snapshots created from it. ForceNew —
changing this replaces the whole FILE SYSTEM (a subtle provider trap:
this is the one root-volume setting that cannot change in place).

### spec.automaticBackupRetentionDays

`int32` · optional (explicit presence)

Number of days to retain automatic backups. Range: 0-90; 0 disables
automatic backups.

Default: 0 (no automatic backups)

- default: `0`
- rule: {"int32":{"lte":90,"gte":0}}

### spec.dailyAutomaticBackupStartTime

`string`

Daily UTC time to start automatic backups, in "HH:MM" format (e.g.,
"05:00"). If not specified and backups are enabled, AWS chooses a window.

- rule: daily_automatic_backup_start_time must be in 24-hour HH:MM format (e.g., '05:00')

### spec.copyTagsToBackups

`bool`

Copy tags from the file system to backups.

### spec.copyTagsToVolumes

`bool`

Copy tags from the file system to volumes. When true, tags are propagated
to the root volume and any child volumes created on this file system.

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

### spec.deleteOptions

`[]string`

Options applied when the file system is deleted.
The single supported value, "DELETE_CHILD_VOLUMES_AND_SNAPSHOTS", deletes
all child volumes and snapshots along with the file system — without it,
deletion fails while children exist. Use deliberately: it turns a
guard-railed delete into a cascading one.

### spec.weeklyMaintenanceStartTime

`string`

Weekly UTC maintenance window in the format "d:HH:MM" where d is the day
of the week (1=Monday, 7=Sunday). Example: "1:05:00" for Monday at 05:00
UTC. If not specified, AWS chooses a default window.

- rule: weekly_maintenance_start_time must be in d:HH:MM format where d is 1 (Monday) through 7 (Sunday), e.g., '1:05:00'

## Validation Rules

- `deployment_type_valid`: deployment_type must be 'SINGLE_AZ_1', 'SINGLE_AZ_2', 'SINGLE_AZ_HA_1', 'SINGLE_AZ_HA_2', or 'MULTI_AZ_1'
- `storage_type_valid`: storage_type must be 'SSD' or 'INTELLIGENT_TIERING'
- `storage_capacity_ssd_contract`: storage_capacity_gib is required for SSD storage (leave it unset only when restoring from backup_id) and must not be set for 'INTELLIGENT_TIERING'
- `intelligent_tiering_contract`: storage_type 'INTELLIGENT_TIERING' requires deployment_type 'MULTI_AZ_1' and read_cache_configuration
- `read_cache_requires_intelligent_tiering`: read_cache_configuration is only supported when storage_type is 'INTELLIGENT_TIERING'
- `throughput_capacity_single_az_1`: throughput_capacity for SINGLE_AZ_1 must be one of: 64, 128, 256, 512, 1024, 2048, 3072, 4096
- `throughput_capacity_gen2`: throughput_capacity for SINGLE_AZ_2 / MULTI_AZ_1 must be one of: 160, 320, 640, 1280, 2560, 3840, 5120, 7680, 10240
- `throughput_capacity_default_gen2`: throughput_capacity must be one of: 160, 320, 640, 1280, 2560, 3840, 5120, 7680, 10240 (the SINGLE_AZ_2 default deployment type)
- `preferred_subnet_multi_az_contract`: preferred_subnet_id is required for MULTI_AZ_1 and can only be set for MULTI_AZ_1
- `subnet_count_matches_deployment`: MULTI_AZ_1 requires exactly two subnets; single-AZ deployment types require exactly one
- `endpoint_ip_range_requires_multi_az`: endpoint_ip_address_range can only be set for MULTI_AZ_1 deployment type
- `route_tables_require_multi_az`: route_table_ids can only be set for MULTI_AZ_1 deployment type
- `backup_restore_excludes_capacity`: storage_capacity_gib cannot be set when restoring from backup_id (capacity comes from the backup)
- `delete_options_valid`: delete_options entries must be 'DELETE_CHILD_VOLUMES_AND_SNAPSHOTS'
- `disk_iops_ceiling_single_az_1`: disk_iops_configuration.iops for SINGLE_AZ_1 must be at most 160000
- `disk_iops_ceiling_single_az_2`: disk_iops_configuration.iops for SINGLE_AZ_2 must be at most 400000

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsFsxOpenzfsFileSystem, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.file_system_id` | `string` | The ID of the file system (e.g., "fs-0123456789abcdef0"). Primary identifier used by EKS PersistentVolumes, ECS task definitions, and other AWS services. |
| `status.outputs.file_system_arn` | `string` | The Amazon Resource Name of the file system. Used in IAM policies for resource-level permissions. |
| `status.outputs.dns_name` | `string` | The DNS name for the file system (e.g., "fs-0123456789abcdef0.fsx.us-east-1.amazonaws.com"). Used in NFS mount commands: mount -t nfs <dns_name>:/fsx /mnt/fsx |
| `status.outputs.endpoint_ip_address` | `string` | The endpoint IP address for the file system. For MULTI_AZ_1 deployments, this is the floating IP used for failover. For SINGLE_AZ deployments, this is the primary ENI IP. |
| `status.outputs.root_volume_id` | `string` | The root volume ID (e.g., "fsvol-0123456789abcdef0"). Required as parent_volume_id when creating child OpenZFS volumes on this file system. |
| `status.outputs.network_interface_ids` | `[]string` | The network interface IDs created for the file system, in order. SINGLE_AZ creates 1 ENI; MULTI_AZ creates 2 ENIs. Useful for security group debugging and network troubleshooting. |
| `status.outputs.vpc_id` | `string` | The VPC ID in which the file system was created. Computed from the subnets. Useful for constructing security group rules and verifying network placement. |
| `status.outputs.owner_id` | `string` | The AWS account ID of the file system owner. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.preferredSubnetId` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.routeTableIds` | AwsSubnet | `status.outputs.route_table_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
