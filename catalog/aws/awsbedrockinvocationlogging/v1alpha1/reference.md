# AwsBedrockInvocationLogging

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockInvocationLoggingSpec defines the desired Bedrock model
invocation logging configuration for one AWS region.

This is a SETTINGS SINGLETON: AWS keeps exactly one invocation
logging configuration per account+region (the resource identity IS
the region), and this component manages it. Deploy at most one
instance per region; two instances targeting the same region fight
over the same configuration object. metadata.name never reaches
AWS - it is Planton-side identity only.

Invocation logging captures every model call in the region - the
full request/response bodies per enabled data type - and delivers
them to CloudWatch Logs, S3, or both. It is the observability
backbone for Bedrock workloads: without it there is no audit trail
of what prompts were sent or what the models returned.

Destroying this component DELETES the logging configuration (the
region reverts to no invocation logging).

## Example

```yaml
# Canonical AwsBedrockInvocationLogging example (hack/dev manifest and
# refgen Example source): full-fidelity delivery -- CloudWatch for
# querying with S3 spillover for oversized payloads, plus the S3
# archive arm. Literal names/ARNs stand in for the composed references
# so the offline `tofu plan` renders the resource.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockInvocationLogging
metadata:
  name: bedrock-logging-us-west-2
  id: bedrock-logging-us-west-2
  org: test-org
  env: dev
spec:
  region: us-west-2
  # Video payloads excluded: keeps CloudWatch volume manageable for
  # text-first workloads while keeping the invocation records.
  videoDataDeliveryEnabled: false
  cloudwatch:
    logGroupName:
      value: /bedrock/invocations
    roleArn:
      value: arn:aws:iam::123456789012:role/bedrock-invocation-logging
    largeDataDeliveryS3:
      bucketName:
        value: bedrock-invocation-large-payloads
      keyPrefix: spillover
  s3:
    bucketName:
      value: bedrock-invocation-archive
    keyPrefix: invocations
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.textDataDeliveryEnabled` | `bool` |  |  |  |
| `spec.imageDataDeliveryEnabled` | `bool` |  |  |  |
| `spec.embeddingDataDeliveryEnabled` | `bool` |  |  |  |
| `spec.videoDataDeliveryEnabled` | `bool` |  |  |  |
| `spec.cloudwatch` | `AwsBedrockInvocationLoggingCloudwatch` |  |  |  |
| `spec.cloudwatch.logGroupName` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_name`) |
| `spec.cloudwatch.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.cloudwatch.largeDataDeliveryS3` | `AwsBedrockInvocationLoggingS3` |  |  |  |
| `spec.cloudwatch.largeDataDeliveryS3.bucketName` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.cloudwatch.largeDataDeliveryS3.keyPrefix` | `string` |  |  |  |
| `spec.s3` | `AwsBedrockInvocationLoggingS3` |  |  |  |
| `spec.s3.bucketName` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.s3.keyPrefix` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region whose Bedrock invocation logging this instance
manages. The region IS the resource identity - one instance per
region, and it is also the provider's import ID.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.textDataDeliveryEnabled

`bool` · optional (explicit presence)

Log the text prompts/completions of model invocations. Unset
means AWS's default: enabled. Set false to exclude text bodies
while keeping the invocation records themselves.

### spec.imageDataDeliveryEnabled

`bool` · optional (explicit presence)

Log image inputs/outputs of model invocations. Unset means AWS's
default: enabled.

### spec.embeddingDataDeliveryEnabled

`bool` · optional (explicit presence)

Log embedding vectors of model invocations. Unset means AWS's
default: enabled.

### spec.videoDataDeliveryEnabled

`bool` · optional (explicit presence)

Log video inputs/outputs of model invocations. Unset means AWS's
default: enabled.

### spec.cloudwatch

`AwsBedrockInvocationLoggingCloudwatch`

Deliver invocation logs to a CloudWatch Logs log group. At least
one of cloudwatch / s3 must be configured; both together are the
full-fidelity posture (CloudWatch for querying, S3 for retention).

### spec.cloudwatch.logGroupName

`string | valueFrom` · required

The CloudWatch Logs log group invocation logs are written to. The
log group must already exist in the same region - Bedrock does
not create it.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_name}} -- a bare string does not parse

### spec.cloudwatch.roleArn

`string | valueFrom` · required

The IAM role Bedrock assumes to write to the log group. The role
must trust "bedrock.amazonaws.com" and carry logs:CreateLogStream
+ logs:PutLogEvents on the log group - AWS validates the
permission chain at apply ("Failed to validate permissions for
log group") and both engines' providers retry through IAM
propagation lag.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.cloudwatch.largeDataDeliveryS3

`AwsBedrockInvocationLoggingS3`

Where CloudWatch delivery spills over for payloads too large for
a log event (CloudWatch caps events at 256 KB; model inputs and
outputs routinely exceed it). Without this, oversized payloads
are truncated out of the CloudWatch stream.

### spec.cloudwatch.largeDataDeliveryS3.bucketName

`string | valueFrom` · required

The S3 bucket invocation logs are written to. The bucket must
already exist in the same region, and its bucket policy must
allow "bedrock.amazonaws.com" to s3:PutObject (scope it with an
aws:SourceAccount condition) - Bedrock writes as its own service
principal, not as an IAM role. Bedrock validates the policy at
configure time by writing zero-byte
"amazon-bedrock-logs-permission-check" objects under every
configured prefix (no invocation needed), and those canaries
OUTLIVE the configuration's delete - plan the bucket's lifecycle
(force-destroy or a cleanup rule) knowing the bucket is never
empty once this configuration has pointed at it.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.cloudwatch.largeDataDeliveryS3.keyPrefix

`string`

Key prefix for delivered objects (for example
"bedrock/invocations"). Empty writes at the bucket root.

- rule: {"string":{"maxLen":"1024"}}

### spec.s3

`AwsBedrockInvocationLoggingS3`

Deliver invocation logs to an S3 bucket. At least one of
cloudwatch / s3 must be configured.

### spec.s3.bucketName

`string | valueFrom` · required

The S3 bucket invocation logs are written to. The bucket must
already exist in the same region, and its bucket policy must
allow "bedrock.amazonaws.com" to s3:PutObject (scope it with an
aws:SourceAccount condition) - Bedrock writes as its own service
principal, not as an IAM role. Bedrock validates the policy at
configure time by writing zero-byte
"amazon-bedrock-logs-permission-check" objects under every
configured prefix (no invocation needed), and those canaries
OUTLIVE the configuration's delete - plan the bucket's lifecycle
(force-destroy or a cleanup rule) knowing the bucket is never
empty once this configuration has pointed at it.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.s3.keyPrefix

`string`

Key prefix for delivered objects (for example
"bedrock/invocations"). Empty writes at the bucket root.

- rule: {"string":{"maxLen":"1024"}}

## Validation Rules

- `spec.delivery_destination_required`: configure at least one delivery destination (cloudwatch and/or s3) - a logging configuration with no destination delivers nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockInvocationLogging, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.configured_region` | `string` | The region whose invocation logging this instance owns - the singleton's identity, and the provider's import ID. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cloudwatch.logGroupName` | AwsCloudwatchLogGroup | `status.outputs.log_group_name` |
| `spec.cloudwatch.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.cloudwatch.largeDataDeliveryS3.bucketName` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.s3.bucketName` | AwsS3Bucket | `status.outputs.bucket_id` |

## See Also

- [Overview](../README.md)
