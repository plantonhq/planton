# AwsS3DirectoryBucket

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsS3DirectoryBucketSpec defines one S3 directory bucket (S3
Express One Zone): single-digit-millisecond object storage that
lives in exactly ONE availability zone (or Local Zone / Dedicated
Local Zone), co-located with the compute that hammers it.

AWS mandates the full bucket name be
"{base}--{zone_id}--x-s3" - a value nobody should hand-assemble, so
BOTH modules derive it from metadata.name + zone_id and export the
result as the bucket_name output. Everything on this spec is fixed
for life except force_destroy: a directory bucket is replaced, not
edited.

Directory buckets authenticate via the S3 Express session API and
use a directory-bucket-scoped policy surface; they never appear in
general-purpose bucket listings.

## Example

```yaml
# Canonical AwsS3DirectoryBucket example (hack/dev manifest and
# refgen Example source): a scratch bucket co-located with compute in
# one zone. The modules derive the full bucket name
# "scratch-data--use1-az4--x-s3".
apiVersion: aws.planton.dev/v1alpha1
kind: AwsS3DirectoryBucket
metadata:
  name: scratch-data
  id: scratch-data
  org: test-org
  env: dev
spec:
  region: us-east-1
  zoneId: use1-az4
  forceDestroy: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.zoneId` | `string` | yes |  |  |
| `spec.zoneType` | `string` |  |  |  |
| `spec.dataRedundancy` | `string` |  |  |  |
| `spec.forceDestroy` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the bucket's zone belongs to. Example:
"us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.zoneId

`string` · required

The zone the bucket lives in - an availability zone ID (never the
letter name: "use1-az4", not "us-east-1a") or a Local Zone ID
("usw2-lax1-az1"). Fixed for life; the zone choice IS the
performance model, so co-locate it with the compute.

- rule: {"string":{"minLen":"1","pattern":"^[a-z0-9][a-z0-9-]*$"}}

### spec.zoneType

`string`

What kind of zone zone_id names. Unset means AvailabilityZone.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AvailabilityZone","LocalZone"]}}

### spec.dataRedundancy

`string`

The redundancy class. Unset lets AWS derive it from zone_type
(SingleAvailabilityZone for AZs, SingleLocalZone for Local
Zones) - the only valid pairing anyway.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SingleAvailabilityZone","SingleLocalZone"]}}

### spec.forceDestroy

`bool`

Let destroy succeed on a non-empty bucket by deleting every
object first. Off by default: destroying data-holding buckets
should hurt. Config-only at AWS - imports never round-trip it.

## Validation Rules

- `spec.data_redundancy_matches_zone_type`: data_redundancy must match zone_type: SingleAvailabilityZone for AvailabilityZone, SingleLocalZone for LocalZone (or leave it unset - AWS derives it)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsS3DirectoryBucket, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.bucket_name` | `string` | The bucket's FULL name ("{base}--{zone_id}--x-s3") - derived by the modules, what S3 Express clients address, and the provider's import ID. |
| `status.outputs.bucket_arn` | `string` | The bucket's ARN. |

## See Also

- [Overview](../README.md)
