# AwsSqsQueue

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsSqsQueueSpec defines the desired configuration for an AWS SQS queue.

SQS provides two queue types:
- Standard queues offer maximum throughput, best-effort ordering, and at-least-once delivery.
- FIFO queues guarantee exactly-once processing and strict message ordering within each
  message group, at the cost of lower throughput.

Notes:
- Set `fifo_queue` to true to create a FIFO queue. This cannot be changed after creation.
- FIFO queue names must end with `.fifo`; the IaC modules append this suffix automatically
  when `fifo_queue` is true and the metadata name does not already include it.
- Encryption at rest: AWS enables SQS-managed SSE (SSE-SQS) on new queues by
  default. Set `sqs_managed_sse_enabled` explicitly to pin it on or off, or
  supply a customer-managed KMS key via `kms_key_id` (mutually exclusive with
  setting `sqs_managed_sse_enabled` at all).
- Dead letter queue configuration allows routing failed messages to a separate queue
  for investigation and reprocessing.
- Credentials, region, and deployment workflow live outside this spec in IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSqsQueue
metadata:
  name: test-queue
  org: test-org
  env: dev
  id: test-queue-dev
  annotations:
    planton.dev/provisioner: pulumi
spec:
  region: us-west-2
  sqsManagedSseEnabled: true
  messageRetentionSeconds: 345600
  visibilityTimeoutSeconds: 30
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.fifoQueue` | `bool` |  |  |  |
| `spec.visibilityTimeoutSeconds` | `int32` |  |  |  |
| `spec.messageRetentionSeconds` | `int32` |  |  |  |
| `spec.maxMessageSizeBytes` | `int32` |  |  |  |
| `spec.delaySeconds` | `int32` |  |  |  |
| `spec.receiveWaitTimeSeconds` | `int32` |  |  |  |
| `spec.contentBasedDeduplication` | `bool` |  |  |  |
| `spec.deduplicationScope` | `string` |  |  |  |
| `spec.fifoThroughputLimit` | `string` |  |  |  |
| `spec.deadLetterConfig` | `AwsSqsQueueDeadLetterConfig` |  |  |  |
| `spec.deadLetterConfig.targetArn` | `string \| valueFrom` | yes |  | AwsSqsQueue (`status.outputs.queue_arn`) |
| `spec.deadLetterConfig.maxReceiveCount` | `int32` |  |  |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.kmsDataKeyReusePeriodSeconds` | `int32` |  |  |  |
| `spec.sqsManagedSseEnabled` | `bool` |  |  |  |
| `spec.policy` | `object` |  |  |  |
| `spec.redriveAllowPolicy` | `AwsSqsQueueRedriveAllowPolicy` |  |  |  |
| `spec.redriveAllowPolicy.redrivePermission` | `string` | yes |  |  |
| `spec.redriveAllowPolicy.sourceQueueArns` | `[]string \| valueFrom` |  |  | AwsSqsQueue (`status.outputs.queue_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.fifoQueue

`bool`

Whether to create a FIFO queue. Standard queues are created when false.
FIFO queues guarantee exactly-once processing and strict ordering within
each message group. This setting cannot be changed after queue creation.

### spec.visibilityTimeoutSeconds

`int32`

Time in seconds that a received message is hidden from subsequent receive
requests. After the timeout expires the message becomes visible again unless
it was deleted. Range: 0–43200 (0s to 12h). AWS default: 30.

- rule: {"int32":{"lte":43200,"gte":0}}

### spec.messageRetentionSeconds

`int32`

Duration in seconds that SQS retains a message. After the retention period
expires SQS deletes the message regardless of whether it was consumed.
Range: 60–1209600 (1 min to 14 days). AWS default: 345600 (4 days).
Leave at 0 to use the AWS default.

### spec.maxMessageSizeBytes

`int32`

Maximum size of a message body in bytes. Messages exceeding this limit are
rejected by SQS. Range: 1024–1048576 (1 KB to 1 MB). AWS default: 262144 (256 KB).
Leave at 0 to use the AWS default.

### spec.delaySeconds

`int32`

Delay in seconds before a newly sent message becomes visible in the queue.
Useful for implementing delayed processing patterns.
Range: 0–900 (0s to 15 min). AWS default: 0.

- rule: {"int32":{"lte":900,"gte":0}}

### spec.receiveWaitTimeSeconds

`int32`

Wait time in seconds for the ReceiveMessage API call. A value greater than
0 enables long polling, which reduces the number of empty responses and
lowers cost. Range: 0–20. AWS default: 0 (short polling).

- rule: {"int32":{"lte":20,"gte":0}}

### spec.contentBasedDeduplication

`bool`

Enable content-based deduplication for FIFO queues. When enabled SQS uses
a SHA-256 hash of the message body as the deduplication ID, removing the
need for the producer to supply an explicit deduplication ID.
Only valid when `fifo_queue` is true.

### spec.deduplicationScope

`string`

Deduplication scope for FIFO queues. Controls whether deduplication is
applied per message group or across the entire queue.
Valid values: "messageGroup", "queue". Only valid when `fifo_queue` is true.

### spec.fifoThroughputLimit

`string`

Throughput limit for FIFO queues. Controls whether throughput quota applies
per message group ID or per queue. Set to "perMessageGroupId" to enable
high throughput mode for FIFO queues.
Valid values: "perMessageGroupId", "perQueue". Only valid when `fifo_queue` is true.

### spec.deadLetterConfig

`AwsSqsQueueDeadLetterConfig`

Dead letter queue configuration. When a message is received more than
`max_receive_count` times without being deleted, SQS moves it to the
specified target queue for investigation and reprocessing.

### spec.deadLetterConfig.targetArn

`string | valueFrom` · required

ARN of the target dead letter queue. Accepts a direct ARN or a reference
to another AwsSqsQueue resource. Both queues must be the same type
(both Standard or both FIFO) and reside in the same AWS account and region.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

### spec.deadLetterConfig.maxReceiveCount

`int32`

Number of times a message can be received before being moved to the dead
letter queue. Must be at least 1. Common values: 3–5 for transient errors,
1 for poison pill detection. Range: 1–1000.

- rule: {"int32":{"lte":1000,"gte":1}}

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key for server-side encryption. When set SQS encrypts
message bodies using this key. Accepts a direct KMS key ID/ARN or a
reference to an AwsKmsKey resource. Mutually exclusive with
`sqs_managed_sse_enabled`.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.kmsDataKeyReusePeriodSeconds

`int32`

Duration in seconds that SQS reuses a data encryption key before calling
KMS again. Higher values reduce KMS costs but increase the window for key
reuse. Range: 60–86400 (1 min to 24h). AWS default: 300 (5 min).
Only relevant when `kms_key_id` is set.

### spec.sqsManagedSseEnabled

`bool` · optional (explicit presence)

SQS-managed server-side encryption (SSE-SQS). SQS manages the encryption
key automatically at no additional cost. AWS enables SSE-SQS on newly
created queues by default, so this field is PRESENCE-typed:
- unset  — keep AWS's default (new queues come up encrypted with SSE-SQS)
- true   — pin SSE-SQS on explicitly
- false  — explicitly disable server-side encryption (an unencrypted queue)

Setting this field at all (either value) conflicts with `kms_key_id`:
the provider rejects the combination on config PRESENCE, not value, so
an explicit `false` alongside a KMS key is still an error — leave this
field unset when using customer-managed KMS encryption.

### spec.policy

`object`

IAM access policy for the queue. Controls which AWS principals can perform
actions on this queue (e.g., SendMessage, ReceiveMessage). Expressed as a
standard IAM policy document structure. Common use cases include granting
SNS topics permission to publish to this queue or allowing cross-account
access.

### spec.redriveAllowPolicy

`AwsSqsQueueRedriveAllowPolicy`

Controls which source queues are allowed to use THIS queue as their
dead-letter queue. This is the permission side of the dead-letter
relationship: `dead_letter_config` on a source queue points at a DLQ,
while `redrive_allow_policy` on the DLQ itself restricts who may point
at it. When unset, AWS allows all source queues in the account (the
"allowAll" behavior). Locking a shared DLQ down with "byQueue" prevents
unrelated workloads from silently routing their failures into it.

- rule: redrive_permission must be 'allowAll', 'denyAll', or 'byQueue'
- rule: source_queue_arns is required (1-10 entries) with redrive_permission 'byQueue' and must be empty for 'allowAll' or 'denyAll'

### spec.redriveAllowPolicy.redrivePermission

`string` · required

The redrive permission mode.
Valid values:
- "allowAll": any source queue in the same account and region may use this
  queue as its DLQ (AWS's default behavior when no policy is set).
- "denyAll": no source queue may use this queue as a DLQ. Use this to
  protect a queue that must never receive redriven messages.
- "byQueue": only the queues listed in `source_queue_arns` may use this
  queue as their DLQ. The recommended mode for shared/central DLQs.

- rule: {"required":true}

### spec.redriveAllowPolicy.sourceQueueArns

`[]string | valueFrom`

ARNs of the source queues permitted to use this queue as their dead-letter
queue. Each entry accepts a direct ARN or a reference to another
AwsSqsQueue resource. Only valid (and required) when `redrive_permission`
is "byQueue". AWS caps the list at 10 source queues; to allow more than
10, use "allowAll" instead.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: {"repeated":{"maxItems":"10"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

## Validation Rules

- `content_based_deduplication_requires_fifo`: content_based_deduplication can only be enabled on FIFO queues (fifo_queue must be true)
- `deduplication_scope_requires_fifo`: deduplication_scope is only valid for FIFO queues and must be 'messageGroup' or 'queue'
- `fifo_throughput_limit_requires_fifo`: fifo_throughput_limit is only valid for FIFO queues and must be 'perMessageGroupId' or 'perQueue'
- `encryption_mutual_exclusion`: kms_key_id and sqs_managed_sse_enabled are mutually exclusive (even sqs_managed_sse_enabled: false conflicts with a KMS key); set one or neither
- `kms_data_key_reuse_requires_kms_key`: kms_data_key_reuse_period_seconds requires kms_key_id to be set
- `message_retention_range`: message_retention_seconds must be between 60 and 1209600 (1 min to 14 days) when set
- `max_message_size_range`: max_message_size_bytes must be between 1024 and 1048576 (1 KB to 1 MB) when set
- `kms_data_key_reuse_range`: kms_data_key_reuse_period_seconds must be between 60 and 86400 when set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSqsQueue, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.queue_url` | `string` | The URL of the SQS queue. This is the primary identifier used in the SQS API for sending, receiving, and deleting messages. |
| `status.outputs.queue_arn` | `string` | The Amazon Resource Name (ARN) of the queue. Used for IAM policies, cross-service permissions, and as a target reference in other resources (e.g., dead letter queue targets, SNS subscription endpoints). |
| `status.outputs.queue_name` | `string` | The name of the SQS queue. For FIFO queues this includes the `.fifo` suffix. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.deadLetterConfig.targetArn` | AwsSqsQueue | `status.outputs.queue_arn` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.redriveAllowPolicy.sourceQueueArns` | AwsSqsQueue | `status.outputs.queue_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsEventBridgeBus | `spec.deadLetterConfig.arn` | `status.outputs.queue_arn` |
| AwsEventBridgePipe | `spec.sourceParameters.kinesis.deadLetterQueueArn` | `status.outputs.queue_arn` |
| AwsEventBridgePipe | `spec.sourceParameters.dynamodb.deadLetterQueueArn` | `status.outputs.queue_arn` |
| AwsEventBridgeRule | `spec.targets[].deadLetterConfig.arn` | `status.outputs.queue_arn` |
| AwsEventBridgeScheduler | `spec.target.arn` | `status.outputs.queue_arn` |
| AwsEventBridgeScheduler | `spec.target.deadLetterQueueArn` | `status.outputs.queue_arn` |
| AwsLambda | `spec.deadLetterTargetArn` | `status.outputs.queue_arn` |
| AwsLambda | `spec.asyncInvokeConfig.onSuccessDestinationArn` | `status.outputs.queue_arn` |
| AwsLambda | `spec.asyncInvokeConfig.onFailureDestinationArn` | `status.outputs.queue_arn` |
| AwsLambdaEventSourceMapping | `spec.eventSourceArn` | `status.outputs.queue_arn` |
| AwsLambdaEventSourceMapping | `spec.onFailureDestinationArn` | `status.outputs.queue_arn` |
| AwsS3Bucket | `spec.notification.queues[].queueArn` | `status.outputs.queue_arn` |
| AwsSnsSubscription | `spec.deadLetterConfig.deadLetterTargetArn` | `status.outputs.queue_arn` |
| AwsSqsQueue | `spec.deadLetterConfig.targetArn` | `status.outputs.queue_arn` |
| AwsSqsQueue | `spec.redriveAllowPolicy.sourceQueueArns` | `status.outputs.queue_arn` |

## See Also

- [Overview](../README.md)
