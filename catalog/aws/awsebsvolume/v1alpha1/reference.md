# AwsEbsVolume

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsEbsVolumeSpec defines one standalone EBS volume as a
create-XOR-copy source union:

The CREATE arm (the default) provisions a fresh volume in an
availability zone - empty (sized) or restored from a snapshot.

The COPY arm (copy_from set) clones an existing volume. AWS creates
the copy in the SOURCE volume's availability zone and inherits its
encryption posture - which is why the create-arm placement and
encryption fields are forbidden alongside copy_from. Size, type,
iops, and throughput may be overridden on the copy.

Attachments are managed in-line as part of the volume: each entry
attaches the volume to one instance at one device name. More than
one attachment requires multi_attach_enabled (io1/io2 only) - AWS
rejects a second attachment on a regular volume.

## Example

```yaml
# Canonical AwsEbsVolume example (hack/dev manifest and refgen
# Example source): an encrypted gp3 data volume attached to one
# instance.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEbsVolume
metadata:
  name: app-data
  id: app-data
  org: test-org
  env: dev
spec:
  region: us-west-2
  availabilityZone: us-west-2a
  type: gp3
  sizeGb: 100
  iops: 3000
  throughputMibps: 125
  encrypted: true
  attachments:
    - deviceName: /dev/sdf
      instanceId:
        value: i-0123456789abcdef0 # replace with your instance id
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.availabilityZone` | `string` |  |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.sizeGb` | `int64` |  |  |  |
| `spec.snapshotId` | `string \| valueFrom` |  |  | AwsEbsSnapshot (`status.outputs.snapshot_id`) |
| `spec.encrypted` | `bool` |  |  |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.iops` | `int64` |  |  |  |
| `spec.throughputMibps` | `int64` |  |  |  |
| `spec.multiAttachEnabled` | `bool` |  |  |  |
| `spec.finalSnapshot` | `bool` |  |  |  |
| `spec.volumeInitializationRate` | `int64` |  |  |  |
| `spec.copyFrom` | `AwsEbsVolumeCopyFrom` |  |  |  |
| `spec.copyFrom.sourceVolumeId` | `string \| valueFrom` | yes |  | AwsEbsVolume (`status.outputs.volume_id`) |
| `spec.attachments` | `[]AwsEbsVolumeAttachment` |  |  |  |
| `spec.attachments[].deviceName` | `string` | yes |  |  |
| `spec.attachments[].instanceId` | `string \| valueFrom` | yes |  | AwsEc2Instance (`status.outputs.instance_id`) |
| `spec.attachments[].forceDetach` | `bool` |  |  |  |
| `spec.attachments[].skipDestroy` | `bool` |  |  |  |
| `spec.attachments[].stopInstanceBeforeDetaching` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the volume lives in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.availabilityZone

`string`

The availability zone for a FRESH volume. Example: "us-east-1a".
Fixed for life - a volume never moves zones (snapshot + restore
is the move). Forbidden with copy_from (the copy lands in the
source's zone).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.type

`string`

The volume type. gp3 is the current-generation default and the
only type with a throughput dial; io1/io2 are provisioned-IOPS
(and the only multi-attach-capable types); sc1/st1 are throughput
HDDs; standard is previous-generation magnetic.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["gp2","gp3","io1","io2","sc1","st1","standard"]}}

### spec.sizeGb

`int64`

The volume size in GiB. Optional when snapshot_id or copy_from
sets the size floor - AWS uses the source size when unset, and a
larger value grows the volume at create.

- rule: {"int64":{"gte":"0"}}

### spec.snapshotId

`string | valueFrom`

Restore the fresh volume from this snapshot. Reference an
AwsEbsSnapshot snapshot_id output or pass a literal snap-... id.
Forbidden with copy_from.

- references: AwsEbsSnapshot (`status.outputs.snapshot_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEbsSnapshot, name: <that resource's name>, fieldPath: status.outputs.snapshot_id}} -- a bare string does not parse

### spec.encrypted

`bool`

Encrypt the fresh volume at rest. Fixed for life. Snapshots of an
encrypted volume (and volumes restored from them) stay encrypted.
Forbidden with copy_from (copies inherit the source posture).

### spec.kmsKeyId

`string | valueFrom`

The KMS key for encryption. Unset with encrypted=true means the
AWS-managed aws/ebs key. Reference an AwsKmsKey key_arn output or
pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.iops

`int64`

Provisioned IOPS. Required for io1/io2 (no default exists);
optional for gp3 (defaults to 3000); forbidden elsewhere.

- rule: {"int64":{"gte":"0"}}

### spec.throughputMibps

`int64`

Throughput in MiB/s - gp3 only. 0 means the gp3 default (125).

- rule: throughput_mibps must be between 125 and 2000

### spec.multiAttachEnabled

`bool`

Let the volume attach to multiple instances at once (io1/io2
only). Fixed for life. Multi-attach volumes need a
cluster-aware filesystem - a regular ext4/xfs mount on two
instances corrupts data.

### spec.finalSnapshot

`bool`

Snapshot the volume automatically before destroy - the safety net
for teardown-with-data. The snapshot is AWS-side only (not
tracked by this resource) and is never visible in volume reads,
so imports do not round-trip it.

### spec.volumeInitializationRate

`int64`

The rate (MiB/s, 100-300) at which AWS eagerly hydrates the
volume from its snapshot instead of lazy-loading blocks on first
read. Only valid with snapshot_id.

- rule: volume_initialization_rate must be between 100 and 300 MiB/s

### spec.copyFrom

`AwsEbsVolumeCopyFrom`

The COPY arm: clone an existing volume instead of creating fresh.

### spec.copyFrom.sourceVolumeId

`string | valueFrom` · required

The volume to copy. Reference another AwsEbsVolume's volume_id
output or pass a literal vol-... id. Fixed for life.

- references: AwsEbsVolume (`status.outputs.volume_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEbsVolume, name: <that resource's name>, fieldPath: status.outputs.volume_id}} -- a bare string does not parse

### spec.attachments

`[]AwsEbsVolumeAttachment`

Attach the volume to instances, in-line. Each entry is one
instance + device name pair. More than one entry requires
multi_attach_enabled.

### spec.attachments[].deviceName

`string` · required

The device name the instance sees. Example: "/dev/sdf" (Linux
exposes it as /dev/xvdf on Xen instances; Nitro instances get an
NVMe name regardless). Fixed for life of the attachment.

- rule: {"string":{"minLen":"1","pattern":"^/dev/[a-z0-9/]+$"}}

### spec.attachments[].instanceId

`string | valueFrom` · required

The instance to attach to. Reference an AwsEc2Instance
instance_id output or pass a literal i-... id. Fixed for life of
the attachment.

- references: AwsEc2Instance (`status.outputs.instance_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEc2Instance, name: <that resource's name>, fieldPath: status.outputs.instance_id}} -- a bare string does not parse

### spec.attachments[].forceDetach

`bool`

Force the detach on removal - the filesystem gets no chance to
flush. A last resort for hung instances; data loss is possible.

### spec.attachments[].skipDestroy

`bool`

Leave the volume attached at AWS when this entry (or the whole
resource) is destroyed - the attachment merely leaves
management. The volume itself cannot be deleted while attached.

### spec.attachments[].stopInstanceBeforeDetaching

`bool`

Stop the instance before detaching, then leave it stopped. The
clean-detach option for root-adjacent data volumes that cannot be
unmounted live.

## Validation Rules

- `spec.create_requires_availability_zone`: availability_zone is required when creating a fresh volume (only a copy_from volume inherits its placement)
- `spec.create_requires_size_or_snapshot`: a fresh volume needs size_gb, snapshot_id, or both (a snapshot fixes the minimum size); only copy_from volumes inherit size
- `spec.copy_forbids_create_arm_fields`: copy_from inherits the source volume's zone, encryption, and snapshot lineage - availability_zone, snapshot_id, encrypted, kms_key_id, multi_attach_enabled, final_snapshot, and volume_initialization_rate cannot be set on a copy
- `spec.iops_requires_iops_capable_type`: iops is configurable only on gp3, io1, and io2 volumes
- `spec.io_types_require_iops`: io1 and io2 volumes require an explicit iops value
- `spec.throughput_is_gp3_only`: throughput_mibps is configurable only on gp3 volumes
- `spec.multi_attach_requires_io_type`: multi_attach_enabled requires an io1 or io2 volume
- `spec.many_attachments_require_multi_attach`: attaching to more than one instance requires multi_attach_enabled (io1/io2 only)
- `spec.init_rate_requires_snapshot`: volume_initialization_rate applies only to volumes created from a snapshot

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEbsVolume, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.volume_id` | `string` | The volume's id (vol-...) - what attachments, snapshots, and copies reference, and the provider's import ID. |
| `status.outputs.volume_arn` | `string` | The volume's ARN. |
| `status.outputs.availability_zone` | `string` | The availability zone the volume actually lives in - notably useful for copies, which inherit the source's zone. |
| `status.outputs.size_gb` | `string` | The volume's actual size in GiB - the snapshot's size when size_gb was left unset. |
| `status.outputs.create_time` | `string` | When AWS created the volume (RFC3339). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.snapshotId` | AwsEbsSnapshot | `status.outputs.snapshot_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.copyFrom.sourceVolumeId` | AwsEbsVolume | `status.outputs.volume_id` |
| `spec.attachments[].instanceId` | AwsEc2Instance | `status.outputs.instance_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsEbsSnapshot | `spec.volumeId` | `status.outputs.volume_id` |
| AwsEbsVolume | `spec.copyFrom.sourceVolumeId` | `status.outputs.volume_id` |

## See Also

- [Overview](../README.md)
