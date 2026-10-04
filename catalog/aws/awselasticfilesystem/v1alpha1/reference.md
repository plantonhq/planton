# AwsElasticFileSystem

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsElasticFileSystemSpec defines the desired configuration for an AWS Elastic
File System — a fully managed, elastic NFS file system that scales storage
capacity automatically as files are added or removed, with no provisioning or
capacity planning required.

This component bundles the file system with its mount targets, backup policy,
resource policy, replication-overwrite protection, and cross-region/cross-AZ
replication. Mount targets are declared per subnet (one per Availability Zone)
and can pin static IPv4/IPv6 addresses. Access points are a separate,
first-class resource: see the AwsEfsAccessPoint kind, which references this
file system and is itself referenced by Lambda and ECS task definitions.

EFS is the foundation for shared persistent storage across:
- EKS pods via the EFS CSI driver (PersistentVolume backed by file_system_id)
- ECS tasks via EFS volumes in task definitions
- Lambda functions via AwsEfsAccessPoint ARNs
- EC2 instances via direct NFS mount at the dns_name

Key design notes:
- `encrypted`, `kms_key_id`, `performance_mode`, and `availability_zone_name`
  are ForceNew — changing them requires replacing the file system. Plan these
  upfront.
- Mount targets are required (min 1) because an EFS without mount targets
  cannot be mounted. AWS allows at most one mount target per Availability
  Zone; declaring two mount targets in subnets of the same AZ fails at deploy
  time with an API error.
- Security groups must allow NFS traffic (TCP port 2049) from the clients.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticFileSystem
metadata:
  org: example-org
  env: dev
  name: test-efs
  id: test-efs-id
spec:
  region: us-east-1
  encrypted: true
  throughputMode: elastic
  mountTargets:
    - subnetId:
        value: subnet-abc123
    - subnetId:
        value: subnet-def456
  securityGroupIds:
    - value: sg-abc123
  backupEnabled: true
  transitionToIa: AFTER_30_DAYS
  transitionToArchive: AFTER_90_DAYS
  policy:
    Version: "2012-10-17"
    Statement:
      - Sid: DenyUnencryptedTransport
        Effect: Deny
        Principal:
          AWS: "*"
        Action: "*"
        Resource: "*"
        Condition:
          Bool:
            aws:SecureTransport: "false"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.encrypted` | `bool` |  | `true` |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.performanceMode` | `string` |  |  |  |
| `spec.throughputMode` | `string` |  |  |  |
| `spec.provisionedThroughputInMibps` | `double` |  |  |  |
| `spec.availabilityZoneName` | `string` |  |  |  |
| `spec.transitionToIa` | `string` |  |  |  |
| `spec.transitionToArchive` | `string` |  |  |  |
| `spec.transitionToPrimaryStorageClass` | `string` |  |  |  |
| `spec.backupEnabled` | `bool` |  |  |  |
| `spec.replicationOverwriteProtection` | `string` |  |  |  |
| `spec.mountTargets` | `[]AwsElasticFileSystemMountTarget` | yes |  |  |
| `spec.mountTargets[].subnetId` | `string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.mountTargets[].ipAddress` | `string` |  |  |  |
| `spec.mountTargets[].ipAddressType` | `string` |  |  |  |
| `spec.mountTargets[].ipv6Address` | `string` |  |  |  |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.policy` | `object` |  |  |  |
| `spec.bypassPolicyLockoutSafetyCheck` | `bool` |  |  |  |
| `spec.replication` | `AwsElasticFileSystemReplication` |  |  |  |
| `spec.replication.destinationRegion` | `string` |  |  |  |
| `spec.replication.destinationAvailabilityZoneName` | `string` |  |  |  |
| `spec.replication.destinationKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.replication.destinationFileSystemId` | `string \| valueFrom` |  |  | AwsElasticFileSystem (`status.outputs.file_system_id`) |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.encrypted

`bool`

Enable encryption at rest for all data and metadata stored in the file system.
ForceNew — cannot be added after creation. Production environments should
always enable encryption.

- default: `true`

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key for encryption at rest. When omitted, EFS uses the
AWS-managed key `aws/elasticfilesystem`. ForceNew — the KMS key cannot be
changed after creation. Requires `encrypted` to be true.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.performanceMode

`string`

File system performance mode. ForceNew — cannot be changed after creation.
Empty keeps the AWS default ("generalPurpose").

- "generalPurpose" (default): lowest latency, suitable for most workloads.
  Recommended for all new file systems, especially with "elastic" throughput.
- "maxIO": higher aggregate throughput for highly parallelized workloads
  (thousands of EC2 instances). Slightly higher per-operation latency.
  AWS recommends generalPurpose + elastic throughput as a replacement, and
  maxIO cannot be combined with elastic throughput or One Zone storage.

### spec.throughputMode

`string`

Throughput mode controlling how EFS delivers read/write bandwidth.
Empty keeps the AWS default ("bursting").

- "bursting" (default): throughput scales with file system size. 50 MiB/s
  per TiB of Standard storage, with bursts up to 100 MiB/s.
- "provisioned": fixed throughput independent of storage size. Set
  `provisioned_throughput_in_mibps` to specify the exact value.
- "elastic": automatically scales throughput up/down based on workload.
  Recommended for unpredictable or spiky access patterns. Requires the
  generalPurpose performance mode (AWS rejects elastic + maxIO).

Throughput mode is mutable, but AWS enforces a 24-hour cooldown between
throughput-mode changes (and between decreases of provisioned throughput).

### spec.provisionedThroughputInMibps

`double`

Provisioned throughput in MiB/s. Only applicable when `throughput_mode` is
"provisioned". AWS accepts 1.0–3414.0 for generalPurpose and 1.0–1024.0 for
maxIO (server-side limits; higher values require a quota increase).

### spec.availabilityZoneName

`string`

AWS Availability Zone name for One Zone storage classes (e.g., "us-east-1a").
ForceNew — cannot be changed after creation. One Zone storage is ~47% cheaper
than Standard (multi-AZ) but data is stored in a single AZ with no cross-AZ
redundancy. Suitable for dev/test or workloads that tolerate AZ-level failure.

When set, only a single mount target can be created (in a subnet belonging
to this AZ).

### spec.transitionToIa

`string`

Transition files to Infrequent Access (IA) storage after the specified period
of not being accessed. IA storage costs ~92% less than Standard but charges
per-access fees.

Valid values: AFTER_1_DAY, AFTER_7_DAYS, AFTER_14_DAYS, AFTER_30_DAYS,
AFTER_60_DAYS, AFTER_90_DAYS, AFTER_180_DAYS, AFTER_270_DAYS, AFTER_365_DAYS.

### spec.transitionToArchive

`string`

Transition IA files to Archive storage after the specified period. Archive
storage costs ~96% less than Standard. Requires `transition_to_ia` to be set
(files must pass through IA before reaching Archive).

Same valid values as transition_to_ia.

### spec.transitionToPrimaryStorageClass

`string`

Transition files back to Standard storage when accessed from IA or Archive.
This enables automatic "warming" of frequently accessed files.

Only valid value: "AFTER_1_ACCESS". Leave empty to keep files in IA/Archive
even after access.

### spec.backupEnabled

`bool`

Enable automatic daily backups via AWS Backup. AWS recommends enabling
backups for all production file systems. Mutable — can be toggled at any time.

### spec.replicationOverwriteProtection

`string`

Controls whether this file system can be used as the DESTINATION of an EFS
replication configuration. Empty keeps the AWS default ("ENABLED").

- "ENABLED" (default): the file system is protected — it cannot be
  overwritten by a replication from another file system.
- "DISABLED": the file system may be targeted as a replication destination.
  AWS also requires protection to be DISABLED before a replication
  destination file system can be modified or deleted after replication
  stops — set this when adopting an existing file system as a replica.

### spec.mountTargets

`[]AwsElasticFileSystemMountTarget` · required

Mount targets expose the file system as an NFS endpoint inside a VPC.
Required (min 1) — an EFS without mount targets cannot be mounted. Declare
one mount target per Availability Zone (AWS allows at most one per AZ;
two subnets in the same AZ fail at deploy time).

For regional (multi-AZ) file systems, declare one mount target per AZ for
maximum availability and to avoid cross-AZ data charges. For One Zone file
systems, declare exactly one mount target in a subnet belonging to
`availability_zone_name`.

- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: ip_address_type must be 'IPV4_ONLY', 'IPV6_ONLY', or 'DUAL_STACK' when set
- rule: ip_address cannot be set when ip_address_type is 'IPV6_ONLY' (the mount target has no IPv4 address)
- rule: ipv6_address requires ip_address_type to be 'IPV6_ONLY' or 'DUAL_STACK'

### spec.mountTargets[].subnetId

`string | valueFrom` · required

The subnet to place this mount target in. The subnet's Availability Zone
determines which AZ's clients this target serves.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.mountTargets[].ipAddress

`string`

Static IPv4 address for the mount target, from the subnet's IPv4 CIDR.
ForceNew. When omitted, AWS assigns an address automatically. Useful for
static NFS mount configurations that cannot resolve the EFS DNS names.

### spec.mountTargets[].ipAddressType

`string`

Address family for the mount target. Empty keeps the AWS default
("IPV4_ONLY"). ForceNew.

- "IPV4_ONLY" (default): IPv4 address only.
- "IPV6_ONLY": IPv6 address only — requires an IPv6-enabled subnet.
- "DUAL_STACK": both IPv4 and IPv6 addresses.

### spec.mountTargets[].ipv6Address

`string`

Static IPv6 address for the mount target, from the subnet's IPv6 CIDR.
ForceNew. Only valid when `ip_address_type` is "IPV6_ONLY" or "DUAL_STACK".

### spec.securityGroupIds

`[]string | valueFrom`

Security groups applied to ALL mount targets. These must allow inbound NFS
traffic (TCP port 2049) from the clients that will mount the file system.
When omitted, AWS attaches the VPC's default security group.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.policy

`object`

IAM resource policy for the file system. Common uses:
- Enforce encryption in transit (deny unencrypted NFS connections)
- Restrict access to specific IAM principals or VPCs
- Prevent root access from NFS clients

Provide as a JSON object structure. Serialized to JSON by IaC modules.
Consistent with SQS policy, SNS policy, and EventBridge event_pattern.

### spec.bypassPolicyLockoutSafetyCheck

`bool`

Skip AWS's policy-lockout safety check when putting the resource policy.
By default AWS rejects a policy that would lock the requesting principal
out of future PutFileSystemPolicy calls. Only set this when intentionally
deploying such a policy — a locked-out policy can only be fixed by the
account root. Requires `policy` to be set.

### spec.replication

`AwsElasticFileSystemReplication`

Replicate this file system to another region or Availability Zone for
disaster recovery. EFS replication is one-per-file-system and create-time
immutable: changing the destination replaces the replication configuration
(the destination file system itself is never deleted by that replacement).

- rule: at least one of destination_region or destination_availability_zone_name is required

### spec.replication.destinationRegion

`string`

Destination region for the replica (e.g., "us-east-2"). At least one of
`destination_region` or `destination_availability_zone_name` is required.
Same-region replication (with a different AZ) is valid.

### spec.replication.destinationAvailabilityZoneName

`string`

Destination Availability Zone name (e.g., "us-east-2a"). Creates the
replica as a One Zone file system in that AZ — the cheaper DR shape. At
least one of `destination_region` or `destination_availability_zone_name`
is required.

### spec.replication.destinationKmsKeyId

`string | valueFrom`

KMS key for the replica's encryption at rest. When omitted, AWS uses the
AWS-managed key `aws/elasticfilesystem` in the destination region.
Replicas are always encrypted regardless of the source's encryption state.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.replication.destinationFileSystemId

`string | valueFrom`

Replicate into an EXISTING file system instead of having AWS create the
replica. The referenced file system must have
`replication_overwrite_protection` set to "DISABLED". When omitted, AWS
creates a fresh destination file system.

- references: AwsElasticFileSystem (`status.outputs.file_system_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticFileSystem, name: <that resource's name>, fieldPath: status.outputs.file_system_id}} -- a bare string does not parse

## Validation Rules

- `performance_mode_valid`: performance_mode must be 'generalPurpose' or 'maxIO' when set
- `throughput_mode_valid`: throughput_mode must be 'bursting', 'provisioned', or 'elastic' when set
- `elastic_requires_general_purpose`: elastic throughput requires the generalPurpose performance mode (AWS rejects elastic + maxIO)
- `provisioned_throughput_requires_mode`: provisioned_throughput_in_mibps can only be set when throughput_mode is 'provisioned'
- `provisioned_mode_requires_throughput`: provisioned_throughput_in_mibps must be greater than 0 when throughput_mode is 'provisioned'
- `kms_requires_encrypted`: kms_key_id requires encrypted to be true (cannot use a custom KMS key without enabling encryption)
- `archive_requires_ia`: transition_to_archive requires transition_to_ia to be set (files must transition through IA before Archive)
- `transition_to_ia_valid`: transition_to_ia must be one of: AFTER_1_DAY, AFTER_7_DAYS, AFTER_14_DAYS, AFTER_30_DAYS, AFTER_60_DAYS, AFTER_90_DAYS, AFTER_180_DAYS, AFTER_270_DAYS, AFTER_365_DAYS
- `transition_to_archive_valid`: transition_to_archive must be one of: AFTER_1_DAY, AFTER_7_DAYS, AFTER_14_DAYS, AFTER_30_DAYS, AFTER_60_DAYS, AFTER_90_DAYS, AFTER_180_DAYS, AFTER_270_DAYS, AFTER_365_DAYS
- `transition_to_primary_valid`: transition_to_primary_storage_class must be 'AFTER_1_ACCESS' when set
- `replication_overwrite_protection_valid`: replication_overwrite_protection must be 'ENABLED' or 'DISABLED' when set
- `bypass_requires_policy`: bypass_policy_lockout_safety_check requires policy to be set (there is no policy put to bypass the check for)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsElasticFileSystem, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.file_system_id` | `string` | The ID of the file system (e.g., "fs-0123456789abcdef0"). This is the primary identifier used by EKS PersistentVolumes, ECS task definitions, and AwsEfsAccessPoint references. |
| `status.outputs.file_system_arn` | `string` | The Amazon Resource Name of the file system. Used in IAM policies for resource-level permissions. |
| `status.outputs.dns_name` | `string` | The regional DNS name for mounting the file system via NFS (e.g., "fs-0123456789abcdef0.efs.us-east-1.amazonaws.com"). Clients can mount using: `mount -t nfs4 <dns_name>:/ /mnt/efs`. |
| `status.outputs.mount_target_ids` | `map<string, string>` | Map of subnet ID to mount target ID. The keys are the resolved subnet IDs of spec.mount_targets. Use to reference specific mount targets for monitoring or troubleshooting. |
| `status.outputs.mount_target_ips` | `map<string, string>` | Map of subnet ID to the mount target's IPv4 address within that subnet. Useful for static NFS mount configurations or network debugging. Empty values for IPV6_ONLY mount targets. |
| `status.outputs.mount_target_ipv6_addresses` | `map<string, string>` | Map of subnet ID to the mount target's IPv6 address. Populated only for mount targets with ip_address_type "IPV6_ONLY" or "DUAL_STACK". |
| `status.outputs.mount_target_dns_names` | `map<string, string>` | Map of subnet ID to per-AZ mount target DNS name (e.g., "us-east-1a.fs-xxx.efs.us-east-1.amazonaws.com"). AZ-specific DNS names route to the mount target in that AZ, avoiding cross-AZ traffic. |
| `status.outputs.replication_destination_file_system_id` | `string` | The file system ID of the replication destination, when spec.replication is configured and AWS created (or was pointed at) a replica. Empty when replication is not configured. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.mountTargets[].subnetId` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.replication.destinationKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.replication.destinationFileSystemId` | AwsElasticFileSystem | `status.outputs.file_system_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBatchJobDefinition | `spec.container.volumes[].efs.fileSystemId` | `status.outputs.file_system_id` |
| AwsEcsTaskDefinition | `spec.volumes[].efs.fileSystemId` | `status.outputs.file_system_id` |
| AwsEfsAccessPoint | `spec.fileSystemId` | `status.outputs.file_system_id` |
| AwsElasticFileSystem | `spec.replication.destinationFileSystemId` | `status.outputs.file_system_id` |
| AwsSagemakerDomain | `spec.defaultUserSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId` | `status.outputs.file_system_id` |
| AwsSagemakerDomain | `spec.defaultSpaceSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId` | `status.outputs.file_system_id` |
| AwsSagemakerDomain | `spec.userProfiles[].userSettings.customFileSystemConfigs[].efsFileSystemConfig.fileSystemId` | `status.outputs.file_system_id` |
| AwsSagemakerDomain | `spec.spaces[].spaceSettings.customFileSystems[].fileSystemId` | `status.outputs.file_system_id` |

## See Also

- [Overview](../README.md)
