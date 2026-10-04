# AwsConfigRecorder

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsConfigRecorderSpec defines the desired AWS Config recording
posture for one region: the configuration recorder, its S3/SNS
delivery channel, and the retention window for recorded
configuration items.

This is a REGION SINGLETON: AWS allows exactly one configuration
recorder and one delivery channel per region, and both are named
"default" by AWS convention - the names are not an identity you
choose, so metadata.name never reaches AWS. Deploy at most one
instance per region; two instances targeting the same region fight
over the same recorder.

The recorder needs an IAM role that trusts "config.amazonaws.com"
(the AWS-managed policy AWS_ConfigRole is the canonical grant), and
the delivery bucket needs a bucket POLICY granting the Config
service principal s3:PutObject and s3:GetBucketAcl - AWS rejects
the channel without it ("insufficient delivery policy").

Destroying this component is a REAL delete: the recorder is
stopped, then recorder, channel, and retention configuration are
removed (already-recorded configuration items stay queryable until
their retention lapses). Teardown ordering matters - AWS refuses to
delete a delivery channel while its recorder is running - and both
engines' modules stop the recorder first.

## Example

```yaml
# Canonical AwsConfigRecorder example (hack/dev manifest and refgen
# Example source): the region's recording posture -- a scoped inclusion
# recorder delivering to S3 with a retention window. Literal
# ARNs/names stand in for composed references so the offline `tofu
# plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsConfigRecorder
metadata:
  name: config-recording-us-west-2
  id: config-recording-us-west-2
  org: test-org
  env: dev
spec:
  region: us-west-2
  roleArn:
    value: arn:aws:iam::123456789012:role/config-recorder
  recordingGroup:
    allSupported: false
    resourceTypes:
      - AWS::S3::Bucket
      - AWS::EC2::SecurityGroup
    recordingStrategy: INCLUSION_BY_RESOURCE_TYPES
  recordingMode:
    recordingFrequency: CONTINUOUS
    override:
      description: Daily snapshots for noisy compute types
      recordingFrequency: DAILY
      resourceTypes:
        - AWS::EC2::Instance
  deliveryChannel:
    s3BucketName:
      value: my-config-history
    s3KeyPrefix: config
    snapshotDeliveryFrequency: TwentyFour_Hours
  retentionPeriodInDays: 365
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.recordingEnabled` | `bool` |  |  |  |
| `spec.recordingGroup` | `AwsConfigRecorderRecordingGroup` |  |  |  |
| `spec.recordingGroup.allSupported` | `bool` |  |  |  |
| `spec.recordingGroup.includeGlobalResourceTypes` | `bool` |  |  |  |
| `spec.recordingGroup.resourceTypes` | `[]string` |  |  |  |
| `spec.recordingGroup.exclusionByResourceTypes` | `[]string` |  |  |  |
| `spec.recordingGroup.recordingStrategy` | `string` |  |  |  |
| `spec.recordingMode` | `AwsConfigRecorderRecordingMode` |  |  |  |
| `spec.recordingMode.recordingFrequency` | `string` |  |  |  |
| `spec.recordingMode.override` | `AwsConfigRecorderRecordingModeOverride` |  |  |  |
| `spec.recordingMode.override.description` | `string` |  |  |  |
| `spec.recordingMode.override.recordingFrequency` | `string` |  |  |  |
| `spec.recordingMode.override.resourceTypes` | `[]string` | yes |  |  |
| `spec.deliveryChannel` | `AwsConfigRecorderDeliveryChannel` |  |  |  |
| `spec.deliveryChannel.s3BucketName` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.deliveryChannel.s3KeyPrefix` | `string` |  |  |  |
| `spec.deliveryChannel.s3KmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.deliveryChannel.snsTopicArn` | `string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.deliveryChannel.snapshotDeliveryFrequency` | `string` |  |  |  |
| `spec.retentionPeriodInDays` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region whose Config recording this instance manages. The
region IS the resource identity - one instance per region.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.roleArn

`string | valueFrom` · required

The IAM role AWS Config assumes to read resource configurations
and write to the delivery channel. Must trust
"config.amazonaws.com"; AWS_ConfigRole is the canonical managed
policy, plus write access to the delivery bucket.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.recordingEnabled

`bool` · optional (explicit presence)

Whether the recorder is RUNNING after apply. Unset = recording on
(the reason this component exists). Set false to keep the
recorder configured but stopped (e.g. pausing per-item recording
costs during a migration) - stopping does not lose already
recorded history.

### spec.recordingGroup

`AwsConfigRecorderRecordingGroup`

What the recorder records. Unset = record every supported
resource type including global ones (the AWS default posture).

- rule: recording_strategy ALL_SUPPORTED_RESOURCE_TYPES requires all_supported (and vice versa when a strategy is set)
- rule: exclusion_by_resource_types requires all_supported = false, an empty resource_types, and the EXCLUSION_BY_RESOURCE_TYPES strategy
- rule: resource_types requires all_supported = false, an empty exclusion list, and the INCLUSION_BY_RESOURCE_TYPES strategy

### spec.recordingGroup.allSupported

`bool` · optional (explicit presence)

Record ALL supported resource types (the default posture). Unset =
enabled. Must be explicitly false to use resource_types or
exclusion_by_resource_types.

### spec.recordingGroup.includeGlobalResourceTypes

`bool` · optional (explicit presence)

Also record global resource types (IAM users, roles, policies).
Record them in exactly ONE region - global resources recorded in
every region multiply configuration items and cost.

### spec.recordingGroup.resourceTypes

`[]string`

Record ONLY these resource types (e.g. "AWS::EC2::Instance",
"AWS::S3::Bucket"). Requires all_supported = false and the
INCLUSION_BY_RESOURCE_TYPES strategy.

- rule: {"repeated":{"unique":true}}

### spec.recordingGroup.exclusionByResourceTypes

`[]string`

Record everything EXCEPT these resource types. Requires
all_supported = false, the EXCLUSION_BY_RESOURCE_TYPES strategy,
and an empty resource_types.

- rule: {"repeated":{"unique":true}}

### spec.recordingGroup.recordingStrategy

`string`

The recording strategy AWS should apply. Unset = AWS derives it
(ALL_SUPPORTED_RESOURCE_TYPES when all_supported is on).

- rule: {"string":{"in":["","ALL_SUPPORTED_RESOURCE_TYPES","INCLUSION_BY_RESOURCE_TYPES","EXCLUSION_BY_RESOURCE_TYPES"]}}

### spec.recordingMode

`AwsConfigRecorderRecordingMode`

How often the recorder records: continuous change-driven items or
daily snapshots (a cost lever), with per-type overrides.

### spec.recordingMode.recordingFrequency

`string`

The default frequency for every recorded type. Unset =
CONTINUOUS.

- rule: {"string":{"in":["","CONTINUOUS","DAILY"]}}

### spec.recordingMode.override

`AwsConfigRecorderRecordingModeOverride`

One override for specific resource types (AWS accepts at most
one override block).

### spec.recordingMode.override.description

`string`

Why this override exists (shown in the AWS console).

### spec.recordingMode.override.recordingFrequency

`string`

The frequency for the listed types.

- rule: {"string":{"in":["CONTINUOUS","DAILY"]}}

### spec.recordingMode.override.resourceTypes

`[]string` · required

The resource types this override applies to.

- rule: {"repeated":{"minItems":"1"}}

### spec.deliveryChannel

`AwsConfigRecorderDeliveryChannel`

Where recorded configuration items are delivered. REQUIRED while
the recorder runs - AWS refuses to start a recorder without a
delivery channel.

### spec.deliveryChannel.s3BucketName

`string | valueFrom` · required

The S3 bucket that receives configuration history and snapshot
files. Needs the Config service-principal bucket policy (see the
spec comment).

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.deliveryChannel.s3KeyPrefix

`string`

Key prefix for delivered files inside the bucket.

### spec.deliveryChannel.s3KmsKeyArn

`string | valueFrom`

KMS key that encrypts delivered files (the key policy must let
AWS Config use it). Unset = SSE-S3.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.deliveryChannel.snsTopicArn

`string | valueFrom`

SNS topic notified when configuration history files are
delivered. Unset = no notifications.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.deliveryChannel.snapshotDeliveryFrequency

`string`

How often AWS Config delivers a full configuration SNAPSHOT (the
per-change history files deliver continuously regardless). Unset =
the AWS default.

- rule: {"string":{"in":["","One_Hour","Three_Hours","Six_Hours","Twelve_Hours","TwentyFour_Hours"]}}

### spec.retentionPeriodInDays

`int32`

How many days AWS Config keeps recorded configuration items
(30-2557). Unset = the AWS default (2557 days / 7 years) and the
retention configuration object is not managed.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":2557,"gte":30}}

## Validation Rules

- `running_recorder_needs_channel`: delivery_channel is required unless recording_enabled is explicitly false - AWS refuses to start a recorder without one

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsConfigRecorder, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.recorder_name` | `string` | The recorder's name - AWS's regional singleton convention ("default"); also the provider's import ID for the recorder. |
| `status.outputs.delivery_channel_name` | `string` | The delivery channel's name ("default"); set only when spec.delivery_channel is configured. |
| `status.outputs.recording_enabled` | `bool` | Whether the recorder is running after apply (the folded recorder-status truth, echoed for chart consumers). |
| `status.outputs.region` | `string` | The AWS region the recording posture lives in. Config's singletons are addressed by REGION + the literal name "default", so any consumer (or verifier) reaching the recorder off the ambient region needs this alongside recorder_name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.deliveryChannel.s3BucketName` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.deliveryChannel.s3KmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.deliveryChannel.snsTopicArn` | AwsSnsTopic | `status.outputs.topic_arn` |

## See Also

- [Overview](../README.md)
