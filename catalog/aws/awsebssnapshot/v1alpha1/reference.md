# AwsEbsSnapshot

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsEbsSnapshotSpec defines one EBS snapshot as a three-way source
union:

The VOLUME arm (volume_id set) snapshots a live volume in this
region.

The COPY arm (copy_from set) copies an existing snapshot - same
region or cross-region - optionally re-encrypting under a different
key. AWS's import for copied snapshots does not exist at the
provider, so a copy is create-only surface.

The IMPORT arm (import_from set) builds the snapshot from a disk
image (VMDK/VHD/RAW) staged in S3 or served over a signed URL,
through the VM Import/Export service role. Also create-only surface
at the provider.

Shared across all three arms: archive tiering (storage_tier +
restore dials), fast snapshot restore per availability zone, and
createVolumePermission grants to other accounts - all managed
in-line as part of the snapshot.

## Example

```yaml
# Canonical AwsEbsSnapshot example (hack/dev manifest and refgen
# Example source): a snapshot of a live volume with fast restore in
# one zone.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEbsSnapshot
metadata:
  name: app-data-backup
  id: app-data-backup
  org: test-org
  env: dev
spec:
  region: us-west-2
  volumeId:
    value: vol-0123456789abcdef0 # replace with your volume id
  description: nightly app-data backup
  fastRestoreAvailabilityZones:
    - us-west-2a
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.volumeId` | `string \| valueFrom` |  |  | AwsEbsVolume (`status.outputs.volume_id`) |
| `spec.copyFrom` | `AwsEbsSnapshotCopyFrom` |  |  |  |
| `spec.copyFrom.sourceSnapshotId` | `string \| valueFrom` | yes |  | AwsEbsSnapshot (`status.outputs.snapshot_id`) |
| `spec.copyFrom.sourceRegion` | `string` | yes |  |  |
| `spec.copyFrom.encrypted` | `bool` |  |  |  |
| `spec.copyFrom.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.copyFrom.completionDurationMinutes` | `int64` |  |  |  |
| `spec.importFrom` | `AwsEbsSnapshotImportFrom` |  |  |  |
| `spec.importFrom.diskContainer` | `AwsEbsSnapshotDiskContainer` | yes |  |  |
| `spec.importFrom.diskContainer.format` | `string` |  |  |  |
| `spec.importFrom.diskContainer.description` | `string` |  |  |  |
| `spec.importFrom.diskContainer.url` | `string` |  |  |  |
| `spec.importFrom.diskContainer.s3Bucket` | `string` |  |  |  |
| `spec.importFrom.diskContainer.s3Key` | `string` |  |  |  |
| `spec.importFrom.roleName` | `string` |  |  |  |
| `spec.importFrom.encrypted` | `bool` |  |  |  |
| `spec.importFrom.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.description` | `string` |  |  |  |
| `spec.storageTier` | `string` |  |  |  |
| `spec.permanentRestore` | `bool` |  |  |  |
| `spec.temporaryRestoreDays` | `int64` |  |  |  |
| `spec.fastRestoreAvailabilityZones` | `[]string` |  |  |  |
| `spec.shareWithAccountIds` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the snapshot lives in. Example: "us-east-1". For
the copy arm this is the DESTINATION region.

- rule: {"string":{"minLen":"1"}}

### spec.volumeId

`string | valueFrom`

The VOLUME arm: snapshot this volume. Reference an AwsEbsVolume
volume_id output or pass a literal vol-... id. Fixed for life.

- references: AwsEbsVolume (`status.outputs.volume_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEbsVolume, name: <that resource's name>, fieldPath: status.outputs.volume_id}} -- a bare string does not parse

### spec.copyFrom

`AwsEbsSnapshotCopyFrom`

The COPY arm: copy an existing snapshot into this region.

- rule: kms_key_id requires encrypted=true - an unencrypted copy has no key

### spec.copyFrom.sourceSnapshotId

`string | valueFrom` · required

The snapshot to copy. Reference another AwsEbsSnapshot's
snapshot_id output or pass a literal snap-... id.

- references: AwsEbsSnapshot (`status.outputs.snapshot_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEbsSnapshot, name: <that resource's name>, fieldPath: status.outputs.snapshot_id}} -- a bare string does not parse

### spec.copyFrom.sourceRegion

`string` · required

The region the source snapshot lives in - same-region copies
re-encrypt or duplicate; cross-region copies replicate. Example:
"us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.copyFrom.encrypted

`bool`

Encrypt the copy. Copying is the ONLY way to encrypt an
unencrypted snapshot (or rotate its key) - snapshots never
re-encrypt in place.

### spec.copyFrom.kmsKeyId

`string | valueFrom`

The KMS key for the copy. Unset with encrypted=true means the
AWS-managed aws/ebs key. Reference an AwsKmsKey key_arn output or
pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.copyFrom.completionDurationMinutes

`int64`

Time-box the copy: AWS completes it within this many minutes
(multiples of 15, up to 2880). Only valid for same-region copies;
faster completion bills at a premium.

- rule: completion_duration_minutes must be a multiple of 15, at most 2880

### spec.importFrom

`AwsEbsSnapshotImportFrom`

The IMPORT arm: build the snapshot from a disk image.

- rule: kms_key_id requires encrypted=true - an unencrypted import has no key

### spec.importFrom.diskContainer

`AwsEbsSnapshotDiskContainer` · required

Where the disk image lives.

- rule: {"required":true}
- rule: configure exactly one of url and s3_bucket/s3_key - the image is fetched from one place
- rule: s3_bucket and s3_key are set together

### spec.importFrom.diskContainer.format

`string`

The image format.

- rule: {"string":{"in":["VMDK","VHD","RAW"]}}

### spec.importFrom.diskContainer.description

`string`

What this image is. Carried onto the import task.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.importFrom.diskContainer.url

`string`

A URL the image downloads from (S3 signed URLs included). Use
s3_bucket/s3_key for plain S3 objects.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.importFrom.diskContainer.s3Bucket

`string`

The S3 bucket holding the image. The import role (vmimport by
default) must be able to read it.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.importFrom.diskContainer.s3Key

`string`

The S3 object key of the image within s3_bucket.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.importFrom.roleName

`string`

The IAM role name the VM Import/Export service assumes to read
the image bucket. Unset means AWS's conventional "vmimport" role
- which must exist with the documented trust and S3 policy
before any import runs.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.importFrom.encrypted

`bool`

Encrypt the imported snapshot at rest.

### spec.importFrom.kmsKeyId

`string | valueFrom`

The KMS key for the import. Unset with encrypted=true means the
AWS-managed aws/ebs key. Reference an AwsKmsKey key_arn output or
pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.description

`string`

What this snapshot captures. Shown in every snapshot listing;
fixed for life (AWS has no description update for snapshots).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.storageTier

`string`

The storage tier. "standard" is instant-restore online storage;
"archive" is ~75% cheaper cold storage with a 24-90 day minimum
and a temporary/permanent restore step before use.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["standard","archive"]}}

### spec.permanentRestore

`bool`

Restore an archived snapshot permanently back to standard access.

### spec.temporaryRestoreDays

`int64`

Restore an archived snapshot temporarily for this many days
(1-180), after which it re-archives itself.

- rule: temporary_restore_days must be between 1 and 180

### spec.fastRestoreAvailabilityZones

`[]string`

Enable fast snapshot restore in these availability zones: volumes
created from the snapshot in a listed zone deliver full
performance instantly instead of lazy-loading blocks. Billed per
zone-hour while enabled - the priciest dial on this kind; enable
deliberately.

- rule: fast_restore_availability_zones must be unique

### spec.shareWithAccountIds

`[]string`

Grant these AWS accounts createVolumePermission - they can create
volumes from this snapshot (sharing without copying). Encrypted
snapshots additionally need the KMS key shared with the same
accounts.

- rule: share_with_account_ids must be unique
- rule: {"repeated":{"items":{"string":{"pattern":"^[0-9]{12}$"}}}}

## Validation Rules

- `spec.exactly_one_source`: configure exactly one of volume_id (snapshot a volume), copy_from (copy a snapshot), or import_from (import a disk image)
- `spec.temporary_restore_excludes_permanent`: temporary_restore_days and permanent_restore are mutually exclusive - a restore is either time-boxed or permanent
- `spec.restore_dials_require_archive`: permanent_restore and temporary_restore_days apply only when storage_tier is archive

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEbsSnapshot, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.snapshot_id` | `string` | The snapshot's id (snap-...) - what volume restores, copies, and permission grants reference, and the provider's import ID (volume arm only; copies and imports are create-only at the provider). |
| `status.outputs.snapshot_arn` | `string` | The snapshot's ARN. |
| `status.outputs.owner_id` | `string` | The AWS account that owns the snapshot. |
| `status.outputs.volume_size_gb` | `string` | The size (GiB) of the volume the snapshot captures - for imports, the size AWS derived from the disk image. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.volumeId` | AwsEbsVolume | `status.outputs.volume_id` |
| `spec.copyFrom.sourceSnapshotId` | AwsEbsSnapshot | `status.outputs.snapshot_id` |
| `spec.copyFrom.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.importFrom.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsEbsSnapshot | `spec.copyFrom.sourceSnapshotId` | `status.outputs.snapshot_id` |
| AwsEbsVolume | `spec.snapshotId` | `status.outputs.snapshot_id` |

## See Also

- [Overview](../README.md)
