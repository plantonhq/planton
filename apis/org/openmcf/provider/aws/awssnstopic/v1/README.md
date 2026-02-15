# AwsSnsTopic

The **AwsSnsTopic** resource provides a standardized way to provision and manage AWS SNS topics through OpenMCF. It supports both Standard and FIFO topic types with bundled subscriptions, KMS encryption, IAM access policies, message filtering, and subscription dead letter queues.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)

- **fifo_topic**: Whether to create a FIFO topic (exactly-once, ordered delivery to SQS FIFO queues) or a Standard topic (maximum throughput, best-effort ordering). Cannot be changed after creation.
- **kms_key_id**: Customer-managed KMS key for server-side encryption. Accepts a literal ARN or a `valueFrom` reference to an AwsKmsKey. Unlike SQS, SNS has no "managed SSE" option — encryption requires an explicit KMS key.
- **subscriptions**: Inline subscriptions defining protocol, endpoint, and optional filter policies. Each subscription's `name` is used as a key in the `subscription_arns` output map.

### FIFO-Specific (When fifo_topic = true)

- **content_based_deduplication**: Use SHA-256 of message body as deduplication ID. Removes the need for publishers to supply explicit deduplication IDs.
- **fifo_throughput_scope**: `"Topic"` or `"MessageGroup"`. Controls whether throughput quota applies per topic or per message group.

### Display and Identity

- **display_name**: Human-readable display name for the topic. Used as the "from" label in SMS messages.

### Access Control

- **policy**: IAM access policy document expressed as a JSON structure. Controls which AWS principals can perform actions on this topic.

### Delivery Configuration

- **delivery_policy**: HTTP/HTTPS delivery retry policy expressed as a JSON string. Most users do not need to customize this.

### Observability

- **tracing_config**: AWS X-Ray tracing. Valid values: `"Active"`, `"PassThrough"`.
- **signature_version**: SNS message signature version. `1` (SHA1) or `2` (SHA256). SHA256 is recommended for new topics.

## Subscription Fields

Each entry in `subscriptions` supports:

- **name** (required): User-assigned key for the subscription. Used in the `subscription_arns` output map.
- **protocol** (required): Delivery protocol — `"sqs"`, `"lambda"`, `"http"`, `"https"`, `"email"`, `"email-json"`, `"sms"`, `"firehose"`, `"application"`.
- **endpoint** (required): Target for delivery. Accepts a literal value or a `valueFrom` reference (e.g., an AwsSqsQueue's queue ARN).
- **filter_policy**: Message filter expressed as a JSON structure. Only matching messages are delivered.
- **filter_policy_scope**: `"MessageAttributes"` (default) or `"MessageBody"`. Controls where the filter is applied.
- **raw_message_delivery**: When true, delivers raw message without JSON envelope. Supported for SQS, HTTP/S, and Firehose.
- **redrive_config**: Dead letter queue for subscription delivery failures.
  - **dead_letter_target_arn**: SQS queue ARN for failed deliveries. Accepts a `valueFrom` reference to an AwsSqsQueue.
- **subscription_role_arn**: IAM role for Firehose delivery. Required when protocol is `"firehose"`.

## Stack Outputs

After provisioning, the AwsSnsTopic resource provides the following outputs:

- **topic_arn**: The Amazon Resource Name (ARN) of the SNS topic — used in IAM policies, cross-service permissions, and as a target reference in other resources.
- **topic_name**: The name of the SNS topic. Includes the `.fifo` suffix for FIFO topics.
- **subscription_arns**: Map of subscription name to subscription ARN. Downstream resources can reference specific subscriptions via `status.outputs.subscription_arns.{name}`.

## How It Works

When you define an AwsSnsTopic resource, OpenMCF:

1. **Creates Topic**: Provisions a Standard or FIFO SNS topic with the specified configuration.
2. **Configures Encryption**: Enables KMS encryption when `kms_key_id` is provided.
3. **Applies Access Policy**: Attaches the IAM access policy to the topic when `policy` is provided.
4. **Creates Subscriptions**: Provisions each inline subscription with its protocol, endpoint, filter policy, and DLQ configuration.
5. **Applies Tags**: Tags the topic with OpenMCF metadata (organization, environment, resource kind, resource ID).

## Use Cases

### Fan-Out to Multiple SQS Queues
Create an SNS topic with multiple SQS subscriptions. Each subscription can have its own filter policy so different queues receive different subsets of messages.

### Event-Driven Lambda Invocation
Subscribe a Lambda function to process every message published to the topic. Use filter policies to invoke Lambda only for specific event types.

### Alarm Notifications
Use email or SMS subscriptions to send operational alerts. Combine with CloudWatch Alarm actions that publish to the topic ARN.

### Cross-Service Event Distribution
Use an IAM access policy to allow EventBridge, S3, or other AWS services to publish events to the topic. Subscribers receive events without direct coupling to the publisher.

### FIFO Message Ordering
Use a FIFO topic with SQS FIFO queue subscriptions for workflows that require strict message ordering and exactly-once delivery, such as financial transaction processing.

## Important Notes

### Standard vs FIFO Topics
- **Standard**: Nearly unlimited throughput, best-effort ordering, at-least-once delivery to all subscriber types.
- **FIFO**: Strict ordering and exactly-once delivery, but only to SQS FIFO queue subscribers. 300 publishes per second (3000 with high-throughput mode).
- Topic type cannot be changed after creation.

### Encryption
- Unlike SQS, SNS does not offer a "managed SSE" option. Encryption at rest requires an explicit KMS key (`kms_key_id`).
- When encryption is enabled, subscribers must have permission to decrypt using the KMS key.

### Subscription Confirmation
- SQS, Lambda, and Firehose subscriptions are confirmed automatically.
- Email, SMS, HTTP, and HTTPS subscriptions require the endpoint owner to confirm. The subscription ARN will show "PendingConfirmation" until confirmed.

### Subscription DLQ
- The subscription `redrive_config` handles SNS-to-subscriber delivery failures (e.g., the SQS queue is deleted, the Lambda function errors). This is separate from any DLQ configured on the SQS queue itself.

## Validation Rules

The API enforces several validations:

1. **FIFO-only fields**: `content_based_deduplication` and `fifo_throughput_scope` require `fifo_topic` to be true.
2. **Throughput scope values**: Must be `"Topic"` or `"MessageGroup"` when set.
3. **Signature version**: Must be `1` or `2` when set.
4. **Tracing config**: Must be `"Active"` or `"PassThrough"` when set.
5. **Subscription protocol**: Must be one of the 9 valid SNS subscription protocols.
6. **Filter policy scope**: Must be `"MessageAttributes"` or `"MessageBody"` when set; requires `filter_policy`.
7. **Firehose role**: `subscription_role_arn` is required when protocol is `"firehose"`.
8. **Redrive config**: `dead_letter_target_arn` is required when `redrive_config` is set.

## References

- [Amazon SNS Developer Guide](https://docs.aws.amazon.com/sns/latest/dg/welcome.html)
- [SNS FIFO Topics](https://docs.aws.amazon.com/sns/latest/dg/sns-fifo-topics.html)
- [SNS Message Filtering](https://docs.aws.amazon.com/sns/latest/dg/sns-message-filtering.html)
- [SNS Subscription Dead Letter Queues](https://docs.aws.amazon.com/sns/latest/dg/sns-dead-letter-queues.html)
- [SNS Server-Side Encryption](https://docs.aws.amazon.com/sns/latest/dg/sns-server-side-encryption.html)
