# AwsFsxOntapFileSystem

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsFsxOntapFileSystemSpec defines the desired configuration for an Amazon FSx
for NetApp ONTAP file system — an enterprise-grade, fully managed shared
storage built on NetApp's ONTAP file system. It provides multi-protocol access
(NFS, SMB, iSCSI) with features like instant snapshots, cloning, replication,
data tiering, compression, and deduplication.

FSx for ONTAP supports scale-out deployments with up to 12 HA pairs for
petabyte-scale single-AZ configurations, and multi-AZ deployments with
automatic failover for high availability. It is the most feature-rich FSx
type, suitable for enterprise workloads, VMware Cloud on AWS, database
storage, and hybrid cloud scenarios via NetApp SnapMirror.

Choosing a shape:
- SINGLE_AZ_2: the current single-AZ generation and the only deployment type
  with scale-out HA pairs (1-12) — up to 1 PiB and tens of GB/s in one
  namespace. The default and the recommended starting point.
- MULTI_AZ_2: the current multi-AZ generation — a standby file server in a
  second AZ takes over automatically on failure. For production workloads
  that must survive an AZ outage.
- SINGLE_AZ_1 / MULTI_AZ_1: the first generation. Choose these only to match
  an existing estate; they cap at 192 TiB, a single HA pair, and lower
  throughput tiers.

Key design notes:
- This component manages the ONTAP **file system** only. Storage Virtual
  Machines (SVMs) and volumes are independent lifecycle resources managed
  by AwsFsxOntapStorageVirtualMachine and AwsFsxOntapVolume respectively.
  Backups (automatic and final) are configured on volumes, not here — the
  file system itself has no backup settings.
- `deployment_type`, `subnet_ids`, `preferred_subnet_id`, `security_group_ids`,
  `kms_key_id`, `storage_type`, and `endpoint_ip_address_range` are ForceNew —
  changing them replaces the file system.
- On the first generation the scaling knobs are also effectively frozen:
  increasing `throughput_capacity_per_ha_pair` on SINGLE_AZ_1/MULTI_AZ_1, or
  `ha_pairs` on SINGLE_AZ_1, replaces the file system. Only SINGLE_AZ_2 and
  MULTI_AZ_2 scale these in place.
- The `fsx_admin_password` provides access to the ONTAP CLI (SSH) and REST API
  for advanced administration (e.g., LIF management, SnapMirror configuration).
- Endpoints are computed outputs: management (ONTAP CLI/API) and intercluster
  (SnapMirror replication between file systems). Data access endpoints (NFS/
  SMB/iSCSI DNS names) live on the SVM, not the file system.
- Credentials, region, and deployment workflow live outside this spec in IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOntapFileSystem
metadata:
  name: ontap-fs-test
  id: awsfxo-test
  org: test-org
  env: dev
spec:
  region: us-west-2
  # Full-surface multi-AZ shape so the offline plan proof exercises every
  # module arm, including the multi-AZ-only floating endpoint range and
  # managed route tables.
  deployment_type: MULTI_AZ_2
  storage_capacity_gib: 4096
  storage_type: SSD
  throughput_capacity_per_ha_pair: 768
  ha_pairs: 1
  subnet_ids:
    - value: subnet-test123
    - value: subnet-test456
  preferred_subnet_id:
    value: subnet-test123
  security_group_ids:
    - value: sg-test123
  endpoint_ip_address_range: 198.19.255.0/24
  route_table_ids:
    - value: rtb-test123
  kms_key_id:
    value: arn:aws:kms:us-west-2:123456789012:key/00000000-0000-0000-0000-000000000000
  fsx_admin_password: OntapAdmin2024!
  disk_iops_configuration:
    mode: USER_PROVISIONED
    iops: 50000
  automatic_backup_retention_days: 7
  daily_automatic_backup_start_time: "04:00"
  weekly_maintenance_start_time: "7:02:00"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.deploymentType` | `string` |  | `SINGLE_AZ_2` |  |
| `spec.storageCapacityGib` | `int32` |  |  |  |
| `spec.storageType` | `string` |  | `SSD` |  |
| `spec.throughputCapacity` | `int32` |  |  |  |
| `spec.throughputCapacityPerHaPair` | `int32` |  |  |  |
| `spec.haPairs` | `int32` |  | `1` |  |
| `spec.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.preferredSubnetId` | `string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.endpointIpAddressRange` | `string` |  |  |  |
| `spec.routeTableIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.route_table_id`) |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.fsxAdminPassword` | `string` (sensitive) |  |  |  |
| `spec.diskIopsConfiguration` | `AwsFsxOntapFileSystemDiskIopsConfiguration` |  |  |  |
| `spec.diskIopsConfiguration.mode` | `string` |  | `AUTOMATIC` |  |
| `spec.diskIopsConfiguration.iops` | `int32` |  |  |  |
| `spec.automaticBackupRetentionDays` | `int32` |  | `0` |  |
| `spec.dailyAutomaticBackupStartTime` | `string` |  |  |  |
| `spec.weeklyMaintenanceStartTime` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.deploymentType

`string` · optional (explicit presence)

Deployment type controlling availability, performance, and scale-out
characteristics. ForceNew — cannot be changed after creation.

- "SINGLE_AZ_1": first-generation single-AZ. Single HA pair only; adding
  HA pairs or increasing per-pair throughput replaces the file system.
- "SINGLE_AZ_2": current single-AZ generation with scale-out HA pairs
  (1-12) that can be added in place. Recommended for most workloads.
- "MULTI_AZ_1": first-generation multi-AZ with automatic failover across
  two AZs. Fixed at 1 HA pair. Requires two subnets and preferred_subnet_id.
- "MULTI_AZ_2": current multi-AZ generation with automatic failover.
  Fixed at 1 HA pair. Recommended for high-availability workloads.

Default: SINGLE_AZ_2

- default: `SINGLE_AZ_2`

### spec.storageCapacityGib

`int32`

Storage capacity in GiB. Required.

The valid range scales with HA pairs: minimum 1024 GiB per HA pair,
maximum 524288 GiB (512 TiB) per HA pair, up to the 1048576 GiB (1 PiB)
absolute ceiling on SINGLE_AZ_2 scale-out deployments. First-generation
deployments (SINGLE_AZ_1 / MULTI_AZ_1) cap at 196608 GiB (192 TiB).

Storage can be increased after creation but never decreased. Choose based
on data size; ONTAP's built-in compression and deduplication typically
achieve 2-5x data reduction, and volume tiering to capacity-pool storage
stretches the SSD tier further.

- rule: {"int32":{"lte":1048576,"gte":1024}}

### spec.storageType

`string` · optional (explicit presence)

Storage media type. ForceNew — cannot be changed after creation.

ONTAP file systems support only "SSD" (sub-millisecond-latency solid-state
primary storage). The other FSx storage classes do not apply here: HDD is
a Windows/Lustre option and INTELLIGENT_TIERING an OpenZFS/Lustre option —
AWS rejects both for ONTAP at create time. Cost tiering on ONTAP is
instead achieved per volume via `tiering_policy` on AwsFsxOntapVolume,
which moves cold data to the built-in elastic capacity pool.

Default: SSD

- default: `SSD`

### spec.throughputCapacity

`int32` · optional (explicit presence)

Total throughput capacity for the whole file system in MB/s — the
first-generation sizing arm. Exactly one of `throughput_capacity` and
`throughput_capacity_per_ha_pair` must be set.

Valid values: 128, 256, 512, 1024, 2048, 4096.

Use this arm for SINGLE_AZ_1 / MULTI_AZ_1 file systems (single HA pair,
so whole-system and per-pair sizing coincide). For the current generation
(SINGLE_AZ_2 / MULTI_AZ_2), prefer `throughput_capacity_per_ha_pair`,
which carries the second generation's throughput tiers.

### spec.throughputCapacityPerHaPair

`int32` · optional (explicit presence)

Throughput capacity per HA pair in MB/s — the per-pair sizing arm.
Exactly one of `throughput_capacity` and `throughput_capacity_per_ha_pair`
must be set. Total file system throughput = this value × ha_pairs.

Valid values by deployment type:
- SINGLE_AZ_1 / MULTI_AZ_1: 128, 256, 512, 1024, 2048, 4096.
- SINGLE_AZ_2 / MULTI_AZ_2 with 1 HA pair: 384, 768, 1536, 3072, 6144.
- SINGLE_AZ_2 with multiple HA pairs: 1536, 3072, 6144 per pair.

Scales in place on SINGLE_AZ_2 and MULTI_AZ_2. On SINGLE_AZ_1 and
MULTI_AZ_1, increasing this value replaces the file system.

### spec.haPairs

`int32` · optional (explicit presence)

Number of high-availability pairs in the file system. Each HA pair adds
an independent pair of file servers contributing throughput, IOPS, and
up to 512 TiB of storage capacity.

Only SINGLE_AZ_2 supports scale-out (1-12 HA pairs, added in place).
SINGLE_AZ_1, MULTI_AZ_1, and MULTI_AZ_2 are fixed at 1 HA pair.

Default: 1

- default: `1`
- rule: {"int32":{"lte":12,"gte":1}}

### spec.subnetIds

`[]string | valueFrom` · required

Subnet IDs for the file system's network interfaces. Required. ForceNew.

- SINGLE_AZ_1 / SINGLE_AZ_2: exactly one subnet.
- MULTI_AZ_1 / MULTI_AZ_2: exactly two subnets in different availability
  zones (the active and standby file servers).

All compute resources accessing this file system (via NFS, SMB, or iSCSI)
must have network connectivity to these subnets.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1","maxItems":"2"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.preferredSubnetId

`string | valueFrom`

Preferred subnet for the active file server in a multi-AZ deployment.
ForceNew. Required when deployment_type is MULTI_AZ_1 or MULTI_AZ_2 (and
invalid otherwise — single-AZ file systems have only one subnet). Must be
one of the subnets specified in subnet_ids.

In a failover event, the standby file server in the other subnet takes
over automatically. Place the active server in the same AZ as the bulk of
your clients to avoid cross-AZ data charges during normal operation.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom`

Security groups for the file system's network interfaces. ForceNew.
Up to 50. When empty, AWS attaches the VPC's default security group.

Must allow traffic between the file system and its clients:
- TCP port 111 (portmapper)
- TCP port 635 (mountd)
- TCP port 2049 (NFS)
- TCP ports 4045-4046 (NFS lock/status)
- TCP port 445 (SMB)
- TCP port 3260 (iSCSI)
- TCP port 443 (ONTAP REST API)
- TCP port 22 (SSH to the ONTAP CLI, if fsx_admin_password is used)

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"50"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.endpointIpAddressRange

`string`

IP address range for the file system endpoints in a multi-AZ deployment,
in CIDR notation (e.g., "198.19.0.0/24"). ForceNew. AWS assigns floating
IPs from this range so the management, intercluster, and SVM data
endpoints survive failover with unchanged addresses.

Must NOT overlap with any subnet in the VPC (AWS recommends a range
outside the VPC CIDR, such as the 198.19.0.0/16 block); clients reach it
through the route tables in route_table_ids. When omitted, AWS picks an
unused range automatically. Only valid for MULTI_AZ_1 / MULTI_AZ_2.

- rule: endpoint_ip_address_range must be an IPv4 CIDR block (e.g., '198.19.0.0/24')

### spec.routeTableIds

`[]string | valueFrom`

Route tables in which AWS creates and manages routes to the floating
endpoint IP range of a multi-AZ deployment. Specify every route table
associated with the subnets your clients live in; AWS repoints the routes
automatically on failover. Up to 50 route tables. When omitted, AWS uses
the VPC's main route table. Only valid for MULTI_AZ_1 / MULTI_AZ_2.

- references: AwsSubnet (`status.outputs.route_table_id`)
- rule: {"repeated":{"maxItems":"50"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.route_table_id}} -- a bare string does not parse

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key ARN for encryption at rest. ForceNew — the KMS key
cannot be changed after creation. When omitted, the file system uses the
AWS-managed FSx key. All ONTAP file systems are encrypted at rest by
default; this field upgrades to a customer-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.fsxAdminPassword

`string` · sensitive

Password for the ONTAP administrative user ("fsxadmin"). Enables SSH and
REST API access to the file system for advanced administration such as LIF
management, SnapMirror configuration, and aggregate monitoring. Can be
changed after creation.

Length: 8-50 characters. Optional — omit if ONTAP CLI access is not needed.
This value is sensitive and will not be returned in read operations.

### spec.diskIopsConfiguration

`AwsFsxOntapFileSystemDiskIopsConfiguration`

SSD IOPS configuration for the file system. Controls the total provisioned
IOPS. When omitted, AWS uses AUTOMATIC mode which provisions 3 IOPS per GiB
of storage capacity. Use USER_PROVISIONED mode for workloads requiring IOPS
beyond what AUTOMATIC provides. Can be changed after creation.

- rule: mode must be 'AUTOMATIC' or 'USER_PROVISIONED'
- rule: iops can only be set when mode is 'USER_PROVISIONED'

### spec.diskIopsConfiguration.mode

`string` · optional (explicit presence)

IOPS provisioning mode.

- "AUTOMATIC": IOPS scale automatically based on storage capacity.
  Provides 3 IOPS per GiB of storage.
- "USER_PROVISIONED": you specify the exact IOPS via the `iops` field.
  Allows higher performance independent of storage size but at extra cost.

Default: AUTOMATIC

- default: `AUTOMATIC`

### spec.diskIopsConfiguration.iops

`int32`

Total SSD IOPS provisioned. Only valid when mode is "USER_PROVISIONED".

Valid range: 0–2,400,000. The maximum achievable IOPS depends on the number
of HA pairs and their throughput capacity tier.

- rule: {"int32":{"lte":2400000,"gte":0}}

### spec.automaticBackupRetentionDays

`int32` · optional (explicit presence)

Number of days to retain automatic backups. Range: 0-90. Set to 0 to
disable automatic backups. ONTAP's built-in snapshots provide point-in-time
recovery independently of FSx backups, and volume-level backup settings
live on AwsFsxOntapVolume.

Default: 0 (no automatic backups)

- default: `0`
- rule: {"int32":{"lte":90,"gte":0}}

### spec.dailyAutomaticBackupStartTime

`string`

Daily UTC time to start automatic backups, in "HH:MM" format (e.g.,
"05:00"). Only meaningful when automatic backups are enabled; when
omitted AWS chooses a window.

- rule: daily_automatic_backup_start_time must be in 24-hour HH:MM format (e.g., '05:00')

### spec.weeklyMaintenanceStartTime

`string`

Weekly UTC maintenance window in "d:HH:MM" format where d is the day of
the week (1=Monday, 7=Sunday). Example: "7:02:00" for Sunday at 02:00 UTC.
If not specified, AWS chooses a default window.

- rule: weekly_maintenance_start_time must be in d:HH:MM format where d is 1 (Monday) through 7 (Sunday), e.g., '7:02:00'

## Validation Rules

- `deployment_type_valid`: deployment_type must be 'SINGLE_AZ_1', 'SINGLE_AZ_2', 'MULTI_AZ_1', or 'MULTI_AZ_2'
- `storage_type_valid`: storage_type must be 'SSD' — ONTAP file systems support only SSD primary storage (use volume tiering policies for cost tiering)
- `throughput_exactly_one_arm`: exactly one of throughput_capacity and throughput_capacity_per_ha_pair must be set
- `throughput_capacity_values`: throughput_capacity must be one of: 128, 256, 512, 1024, 2048, 4096
- `throughput_per_ha_pair_gen1_values`: throughput_capacity_per_ha_pair must be one of 128, 256, 512, 1024, 2048, 4096 for SINGLE_AZ_1 / MULTI_AZ_1
- `throughput_per_ha_pair_gen2_values`: throughput_capacity_per_ha_pair must be one of 384, 768, 1536, 3072, 6144 for SINGLE_AZ_2 / MULTI_AZ_2 with 1 HA pair
- `throughput_per_ha_pair_scale_out_values`: throughput_capacity_per_ha_pair must be one of 1536, 3072, 6144 when ha_pairs > 1
- `ha_pairs_scale_out_single_az_2_only`: ha_pairs > 1 is only supported for the SINGLE_AZ_2 deployment type (all other deployment types are fixed at 1 HA pair)
- `storage_capacity_min_per_ha_pair`: storage_capacity_gib must be at least 1024 GiB per HA pair
- `storage_capacity_max_per_ha_pair`: storage_capacity_gib must be at most 524288 GiB (512 TiB) per HA pair
- `storage_capacity_gen1_ceiling`: storage_capacity_gib must be at most 196608 GiB (192 TiB) for SINGLE_AZ_1 / MULTI_AZ_1
- `single_az_exactly_one_subnet`: SINGLE_AZ deployment types require exactly one subnet in subnet_ids
- `multi_az_exactly_two_subnets`: MULTI_AZ deployment types require exactly two subnets in subnet_ids (active and standby availability zones)
- `preferred_subnet_multi_az_contract`: preferred_subnet_id is required for MULTI_AZ_1 / MULTI_AZ_2 deployment types and invalid for single-AZ deployment types
- `endpoint_ip_range_requires_multi_az`: endpoint_ip_address_range can only be set for MULTI_AZ_1 or MULTI_AZ_2 deployment types
- `route_tables_require_multi_az`: route_table_ids can only be set for MULTI_AZ_1 or MULTI_AZ_2 deployment types
- `admin_password_length`: fsx_admin_password must be 8-50 characters when provided
- `backup_time_requires_retention`: daily_automatic_backup_start_time requires automatic_backup_retention_days > 0

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsFsxOntapFileSystem, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.file_system_id` | `string` | The ID of the file system (e.g., "fs-0123456789abcdef0"). Primary identifier used by Storage Virtual Machines and other AWS services. |
| `status.outputs.file_system_arn` | `string` | The Amazon Resource Name of the file system. Used in IAM policies for resource-level permissions. |
| `status.outputs.management_dns_name` | `string` | The management endpoint DNS name. Used for SSH (ONTAP CLI) and REST API access to the file system. Connect via: ssh fsxadmin@<management_dns_name> Data access (NFS/SMB/iSCSI) endpoints live on the SVM — see AwsFsxOntapStorageVirtualMachine outputs. An ONTAP file system has no file-system-level data DNS name. |
| `status.outputs.management_ip_addresses` | `[]string` | The management endpoint IP addresses. Alternative to DNS for direct IP access to the ONTAP management interface. |
| `status.outputs.intercluster_dns_name` | `string` | The intercluster endpoint DNS name. Used for NetApp SnapMirror replication between FSx for ONTAP file systems (same or cross-region). |
| `status.outputs.intercluster_ip_addresses` | `[]string` | The intercluster endpoint IP addresses. Used for SnapMirror peering when DNS resolution is not available. |
| `status.outputs.network_interface_ids` | `[]string` | The network interface IDs created for the file system, in order. Single-AZ creates 1 ENI per HA pair; multi-AZ creates 2 ENIs. Useful for security group debugging and network troubleshooting. |
| `status.outputs.vpc_id` | `string` | The VPC ID in which the file system was created. Computed from the subnets. Useful for constructing security group rules and verifying network placement. |
| `status.outputs.owner_id` | `string` | The AWS account ID of the file system owner. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.preferredSubnetId` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.routeTableIds` | AwsSubnet | `status.outputs.route_table_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsFsxOntapStorageVirtualMachine | `spec.fileSystemId` | `status.outputs.file_system_id` |

## See Also

- [Overview](../README.md)
