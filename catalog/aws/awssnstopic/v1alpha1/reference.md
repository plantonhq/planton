# AwsSnsTopic

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsSnsTopicSpec defines the desired configuration for an AWS SNS topic.

SNS provides two topic types:
- Standard topics offer maximum throughput, best-effort ordering, and at-least-once delivery
  to a wide variety of subscriber protocols (SQS, Lambda, HTTP/S, email, SMS).
- FIFO topics guarantee strict ordering and exactly-once message delivery to SQS FIFO
  subscribers, at the cost of lower throughput.

Notes:
- Set `fifo_topic` to true to create a FIFO topic. This cannot be changed after creation.
- FIFO topic names must end with `.fifo`; the IaC modules append this suffix automatically
  when `fifo_topic` is true and the metadata name does not already include it.
- Encryption at rest is supported via a customer-managed KMS key.
- Subscriptions are first-class AwsSnsSubscription resources that reference this topic's
  `topic_arn` output. A topic owns its identity, policy, and delivery posture; each
  subscription owns its own protocol, endpoint, filtering, and redrive lifecycle.
- Credentials, region, and deployment workflow live outside this spec in IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSnsTopic
metadata:
  name: test-topic
  org: test-org
  env: dev
  id: test-topic-dev
  annotations:
    planton.dev/provisioner: pulumi
spec:
  region: us-west-2
  signatureVersion: 2
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.fifoTopic` | `bool` |  |  |  |
| `spec.contentBasedDeduplication` | `bool` |  |  |  |
| `spec.fifoThroughputScope` | `string` |  |  |  |
| `spec.archivePolicy` | `object` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.policy` | `object` |  |  |  |
| `spec.dataProtectionPolicy` | `object` |  |  |  |
| `spec.deliveryPolicy` | `string` |  |  |  |
| `spec.deliveryFeedback` | `AwsSnsTopicDeliveryFeedback` |  |  |  |
| `spec.deliveryFeedback.application` | `AwsSnsTopicProtocolFeedback` |  |  |  |
| `spec.deliveryFeedback.application.successFeedbackRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.deliveryFeedback.application.failureFeedbackRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.deliveryFeedback.application.successFeedbackSampleRate` | `int32` |  |  |  |
| `spec.deliveryFeedback.firehose` | `AwsSnsTopicProtocolFeedback` |  |  |  |
| `spec.deliveryFeedback.firehose.successFeedbackRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.deliveryFeedback.firehose.failureFeedbackRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.deliveryFeedback.firehose.successFeedbackSampleRate` | `int32` |  |  |  |
| `spec.deliveryFeedback.http` | `AwsSnsTopicProtocolFeedback` |  |  |  |
| `spec.deliveryFeedback.http.successFeedbackRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.deliveryFeedback.http.failureFeedbackRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.deliveryFeedback.http.successFeedbackSampleRate` | `int32` |  |  |  |
| `spec.deliveryFeedback.lambda` | `AwsSnsTopicProtocolFeedback` |  |  |  |
| `spec.deliveryFeedback.lambda.successFeedbackRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.deliveryFeedback.lambda.failureFeedbackRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.deliveryFeedback.lambda.successFeedbackSampleRate` | `int32` |  |  |  |
| `spec.deliveryFeedback.sqs` | `AwsSnsTopicProtocolFeedback` |  |  |  |
| `spec.deliveryFeedback.sqs.successFeedbackRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.deliveryFeedback.sqs.failureFeedbackRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.deliveryFeedback.sqs.successFeedbackSampleRate` | `int32` |  |  |  |
| `spec.tracingConfig` | `string` |  |  |  |
| `spec.signatureVersion` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.fifoTopic

`bool`

Whether to create a FIFO topic. Standard topics are created when false.
FIFO topics guarantee strict ordering and exactly-once delivery to SQS FIFO
queue subscribers. This setting cannot be changed after topic creation.

### spec.contentBasedDeduplication

`bool`

Enable content-based deduplication for FIFO topics. When enabled, SNS uses
a SHA-256 hash of the message body as the deduplication ID, removing the
need for the publisher to supply an explicit deduplication ID.
Only valid when `fifo_topic` is true.

### spec.fifoThroughputScope

`string`

Throughput scope for FIFO topics. Controls whether the throughput quota
applies per topic or per message group. "MessageGroup" enables high
throughput mode (each message group gets its own quota); it pairs with
SQS FIFO subscribers configured with `fifo_throughput_limit:
"perMessageGroupId"` for an end-to-end high-throughput FIFO pipeline.
Valid values: "Topic", "MessageGroup". Only valid when `fifo_topic` is true.

### spec.archivePolicy

`object`

Message archive policy for FIFO topics. When set, SNS retains published
messages for the configured window so subscriptions can replay them (each
AwsSnsSubscription opts into replay via its own `replay_policy`). Expressed
as the SNS archive policy JSON document, e.g.
{"MessageRetentionPeriod": 30} for a 30-day archive. Only valid when
`fifo_topic` is true — AWS does not archive standard topics. The
`beginning_archive_time` output reports when the archive became active.

### spec.displayName

`string`

Human-readable display name for the topic. Used as the "from" label in SMS
messages and as a readable identifier in the AWS console. Maximum 256
characters for Standard topics, 10 characters for SMS display names.

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key for server-side encryption. When set, SNS encrypts
message bodies using this key. Accepts a direct KMS key ID/ARN or a
reference to an AwsKmsKey resource. When not set, SNS does not encrypt
messages at rest (unlike SQS, SNS has no "managed SSE" option — encryption
requires an explicit KMS key).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.policy

`object`

IAM access policy for the topic. Controls which AWS principals can perform
actions on this topic (e.g., Publish, Subscribe). Expressed as a standard
IAM policy document structure. Common use cases include granting EventBridge
permission to publish, allowing cross-account subscriptions, or restricting
publishing to specific IAM roles. Note: an SNS topic always carries a
policy — when this field is unset AWS applies its default owner-only
policy, and removing a previously set policy reverts to that default
rather than leaving the topic policy-less.

### spec.dataProtectionPolicy

`object`

Data protection policy for the topic. Detects and optionally audits,
masks, or blocks sensitive data (PII/PHI such as names, addresses, card
numbers) flowing through the topic. Expressed as the SNS data protection
policy JSON document (Name/Description/Version/Statement with
DataIdentifier selectors and Audit/Deidentify/Deny operations). Only
supported on standard topics — AWS rejects data protection policies on
FIFO topics.

### spec.deliveryPolicy

`string`

HTTP/HTTPS delivery retry policy for the topic. Expressed as a JSON string
matching the SNS delivery policy format. Controls retry backoff, max retries,
and throttle behavior for HTTP/S subscriptions. Most users do not need to
customize this. When not set, AWS applies its default delivery policy.

### spec.deliveryFeedback

`AwsSnsTopicDeliveryFeedback`

Per-protocol delivery status logging. Each configured protocol block makes
SNS write delivery success/failure log entries to CloudWatch Logs using the
supplied IAM roles. Configure only the protocols this topic actually
delivers to — each block is independent.

### spec.deliveryFeedback.application

`AwsSnsTopicProtocolFeedback`

Delivery status logging for mobile platform application endpoints.

- rule: success_feedback_sample_rate requires success_feedback_role to be set

### spec.deliveryFeedback.application.successFeedbackRole

`string | valueFrom`

IAM role SNS assumes to log SUCCESSFUL deliveries for this protocol.
Accepts a direct role ARN or a reference to an AwsIamRole resource.
Required for success logging; the sample rate controls what fraction of
successes are logged.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.deliveryFeedback.application.failureFeedbackRole

`string | valueFrom`

IAM role SNS assumes to log FAILED deliveries for this protocol. Accepts
a direct role ARN or a reference to an AwsIamRole resource. Failures are
always logged when this role is set (there is no failure sample rate).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.deliveryFeedback.application.successFeedbackSampleRate

`int32`

Percentage (0-100) of successful deliveries to log. Leave at 0 to let AWS
apply its default. Only meaningful when `success_feedback_role` is set.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.deliveryFeedback.firehose

`AwsSnsTopicProtocolFeedback`

Delivery status logging for Kinesis Data Firehose delivery streams.

- rule: success_feedback_sample_rate requires success_feedback_role to be set

### spec.deliveryFeedback.firehose.successFeedbackRole

`string | valueFrom`

IAM role SNS assumes to log SUCCESSFUL deliveries for this protocol.
Accepts a direct role ARN or a reference to an AwsIamRole resource.
Required for success logging; the sample rate controls what fraction of
successes are logged.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.deliveryFeedback.firehose.failureFeedbackRole

`string | valueFrom`

IAM role SNS assumes to log FAILED deliveries for this protocol. Accepts
a direct role ARN or a reference to an AwsIamRole resource. Failures are
always logged when this role is set (there is no failure sample rate).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.deliveryFeedback.firehose.successFeedbackSampleRate

`int32`

Percentage (0-100) of successful deliveries to log. Leave at 0 to let AWS
apply its default. Only meaningful when `success_feedback_role` is set.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.deliveryFeedback.http

`AwsSnsTopicProtocolFeedback`

Delivery status logging for HTTP/HTTPS endpoints.

- rule: success_feedback_sample_rate requires success_feedback_role to be set

### spec.deliveryFeedback.http.successFeedbackRole

`string | valueFrom`

IAM role SNS assumes to log SUCCESSFUL deliveries for this protocol.
Accepts a direct role ARN or a reference to an AwsIamRole resource.
Required for success logging; the sample rate controls what fraction of
successes are logged.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.deliveryFeedback.http.failureFeedbackRole

`string | valueFrom`

IAM role SNS assumes to log FAILED deliveries for this protocol. Accepts
a direct role ARN or a reference to an AwsIamRole resource. Failures are
always logged when this role is set (there is no failure sample rate).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.deliveryFeedback.http.successFeedbackSampleRate

`int32`

Percentage (0-100) of successful deliveries to log. Leave at 0 to let AWS
apply its default. Only meaningful when `success_feedback_role` is set.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.deliveryFeedback.lambda

`AwsSnsTopicProtocolFeedback`

Delivery status logging for Lambda function endpoints.

- rule: success_feedback_sample_rate requires success_feedback_role to be set

### spec.deliveryFeedback.lambda.successFeedbackRole

`string | valueFrom`

IAM role SNS assumes to log SUCCESSFUL deliveries for this protocol.
Accepts a direct role ARN or a reference to an AwsIamRole resource.
Required for success logging; the sample rate controls what fraction of
successes are logged.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.deliveryFeedback.lambda.failureFeedbackRole

`string | valueFrom`

IAM role SNS assumes to log FAILED deliveries for this protocol. Accepts
a direct role ARN or a reference to an AwsIamRole resource. Failures are
always logged when this role is set (there is no failure sample rate).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.deliveryFeedback.lambda.successFeedbackSampleRate

`int32`

Percentage (0-100) of successful deliveries to log. Leave at 0 to let AWS
apply its default. Only meaningful when `success_feedback_role` is set.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.deliveryFeedback.sqs

`AwsSnsTopicProtocolFeedback`

Delivery status logging for SQS queue endpoints.

- rule: success_feedback_sample_rate requires success_feedback_role to be set

### spec.deliveryFeedback.sqs.successFeedbackRole

`string | valueFrom`

IAM role SNS assumes to log SUCCESSFUL deliveries for this protocol.
Accepts a direct role ARN or a reference to an AwsIamRole resource.
Required for success logging; the sample rate controls what fraction of
successes are logged.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.deliveryFeedback.sqs.failureFeedbackRole

`string | valueFrom`

IAM role SNS assumes to log FAILED deliveries for this protocol. Accepts
a direct role ARN or a reference to an AwsIamRole resource. Failures are
always logged when this role is set (there is no failure sample rate).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.deliveryFeedback.sqs.successFeedbackSampleRate

`int32`

Percentage (0-100) of successful deliveries to log. Leave at 0 to let AWS
apply its default. Only meaningful when `success_feedback_role` is set.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.tracingConfig

`string`

AWS X-Ray tracing configuration. When set to "Active", SNS publishes trace
data for messages. When set to "PassThrough", SNS passes through the trace
header but does not sample. Leave empty to use the AWS default (PassThrough).
Valid values: "Active", "PassThrough".

### spec.signatureVersion

`int32`

SNS message signature version. Version 1 uses SHA1, version 2 uses SHA256.
SHA256 (version 2) is recommended for new topics. Leave at 0 to use the
AWS default (version 1).

## Validation Rules

- `content_based_deduplication_requires_fifo`: content_based_deduplication can only be enabled on FIFO topics (fifo_topic must be true)
- `fifo_throughput_scope_requires_fifo`: fifo_throughput_scope is only valid for FIFO topics and must be 'Topic' or 'MessageGroup'
- `archive_policy_requires_fifo`: archive_policy is only supported on FIFO topics (fifo_topic must be true)
- `data_protection_policy_standard_only`: data_protection_policy is only supported on standard topics (not FIFO)
- `signature_version_valid`: signature_version must be 1 (SHA1) or 2 (SHA256) when set
- `tracing_config_valid`: tracing_config must be 'Active' or 'PassThrough' when set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSnsTopic, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.topic_arn` | `string` | The Amazon Resource Name (ARN) of the SNS topic. This is the primary identifier used for IAM policies, cross-service permissions, subscription wiring (AwsSnsSubscription.topic_arn), and as a target reference in other resources (e.g., EventBridge rule targets, CloudWatch alarm actions). |
| `status.outputs.topic_name` | `string` | The name of the SNS topic. For FIFO topics this includes the `.fifo` suffix. |
| `status.outputs.owner` | `string` | The AWS account ID that owns the topic. Useful for composing cross-account policies without hardcoding the account number. |
| `status.outputs.beginning_archive_time` | `string` | When `archive_policy` is active on a FIFO topic, the timestamp from which archived messages are available for replay (ISO 8601). Empty when message archiving is not enabled. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.deliveryFeedback.application.successFeedbackRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.deliveryFeedback.application.failureFeedbackRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.deliveryFeedback.firehose.successFeedbackRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.deliveryFeedback.firehose.failureFeedbackRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.deliveryFeedback.http.successFeedbackRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.deliveryFeedback.http.failureFeedbackRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.deliveryFeedback.lambda.successFeedbackRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.deliveryFeedback.lambda.failureFeedbackRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.deliveryFeedback.sqs.successFeedbackRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.deliveryFeedback.sqs.failureFeedbackRole` | AwsIamRole | `status.outputs.role_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAutoScalingGroup | `spec.lifecycleHooks[].notificationTargetArn` | `status.outputs.topic_arn` |
| AwsAutoScalingGroup | `spec.notifications.topic` | `status.outputs.topic_arn` |
| AwsBackupVault | `spec.standard.notifications.snsTopicArn` | `status.outputs.topic_arn` |
| AwsBudget | `spec.notifications[].subscriberSnsTopicArns` | `status.outputs.topic_arn` |
| AwsBudget | `spec.actions[].subscribers[].address` | `status.outputs.topic_arn` |
| AwsCloudTrail | `spec.snsTopicName` | `status.outputs.topic_name` |
| AwsCloudwatchAlarm | `spec.alarmActions` | `status.outputs.topic_arn` |
| AwsCloudwatchAlarm | `spec.okActions` | `status.outputs.topic_arn` |
| AwsCloudwatchAlarm | `spec.insufficientDataActions` | `status.outputs.topic_arn` |
| AwsCloudwatchCompositeAlarm | `spec.alarmActions` | `status.outputs.topic_arn` |
| AwsCloudwatchCompositeAlarm | `spec.okActions` | `status.outputs.topic_arn` |
| AwsCloudwatchCompositeAlarm | `spec.insufficientDataActions` | `status.outputs.topic_arn` |
| AwsConfigRecorder | `spec.deliveryChannel.snsTopicArn` | `status.outputs.topic_arn` |
| AwsCostAnomalyMonitor | `spec.subscriptions[].subscribers[].address` | `status.outputs.topic_arn` |
| AwsEventBridgeScheduler | `spec.target.arn` | `status.outputs.topic_arn` |
| AwsMemcachedElasticache | `spec.notificationTopicArn` | `status.outputs.topic_arn` |
| AwsMemorydbCluster | `spec.snsTopicArn` | `status.outputs.topic_arn` |
| AwsRedisElasticache | `spec.notificationTopicArn` | `status.outputs.topic_arn` |
| AwsS3Bucket | `spec.notification.topics[].topicArn` | `status.outputs.topic_arn` |
| AwsSagemakerEndpoint | `spec.asyncInference.successTopicArn` | `status.outputs.topic_arn` |
| AwsSagemakerEndpoint | `spec.asyncInference.errorTopicArn` | `status.outputs.topic_arn` |
| AwsSesConfigurationSet | `spec.eventDestinations[].snsTopic` | `status.outputs.topic_arn` |
| AwsSnsSubscription | `spec.topicArn` | `status.outputs.topic_arn` |
| AwsSsmMaintenanceWindow | `spec.tasks[].invocation.runCommand.notificationConfig.notificationArn` | `status.outputs.topic_arn` |

## See Also

- [Overview](../README.md)
