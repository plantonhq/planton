# AwsSqsQueue

The **AwsSqsQueue** resource provides a standardized way to provision and manage AWS SQS queues through Planton. It supports both Standard and FIFO queue types with dead letter queue routing, server-side encryption (SSE-SQS or customer-managed KMS), IAM access policies, and fine-grained delivery tuning.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)

- **fifo_queue**: Whether to create a FIFO queue (exactly-once, ordered) or a Standard queue (best-effort ordering, at-least-once). Cannot be changed after creation.
- **sqs_managed_sse_enabled**: Enable SQS-managed server-side encryption at zero cost. Mutually exclusive with `kms_key_id`.
- **dead_letter_config**: Route messages that fail processing beyond a threshold to a separate dead letter queue.
  - **target_arn**: ARN of the dead letter queue. Accepts a literal value or a `valueFrom` reference to another AwsSqsQueue.
  - **max_receive_count**: Number of receive attempts before a message is sent to the DLQ (1–1000).

### Delivery Tuning (Common Adjustments)

- **visibility_timeout_seconds**: How long a received message is hidden from other consumers (0–43200). AWS default: 30. Set to slightly longer than your consumer's processing time.
- **message_retention_seconds**: How long SQS retains undelivered messages (60–1209600). AWS default: 345600 (4 days).
- **max_message_size_bytes**: Maximum message body size in bytes (1024–1048576). AWS default: 262144 (256 KB).
- **delay_seconds**: Delay before newly sent messages become visible (0–900). AWS default: 0.
- **receive_wait_time_seconds**: Long polling wait time (0–20). Set >0 to reduce empty responses and cost.

### FIFO-Specific (When fifo_queue = true)

- **content_based_deduplication**: Use SHA-256 of message body as deduplication ID. Removes the need for producers to supply explicit deduplication IDs.
- **deduplication_scope**: `"messageGroup"` or `"queue"`. Controls deduplication granularity.
- **fifo_throughput_limit**: `"perMessageGroupId"` or `"perQueue"`. Set to `"perMessageGroupId"` for high-throughput FIFO mode.

### Encryption

- **kms_key_id**: Customer-managed KMS key for encryption. Accepts a literal ARN or a `valueFrom` reference to an AwsKmsKey. Mutually exclusive with `sqs_managed_sse_enabled`.
- **kms_data_key_reuse_period_seconds**: KMS data key reuse window (60–86400). AWS default: 300. Only relevant when `kms_key_id` is set.

### Access Control

- **policy**: IAM access policy document expressed as a JSON structure. Controls which AWS principals can perform actions on this queue (e.g., granting SNS publish permission, cross-account access).

## Stack Outputs

After provisioning, the AwsSqsQueue resource provides the following outputs:

- **queue_url**: The URL of the SQS queue — the primary identifier for API calls (SendMessage, ReceiveMessage, DeleteMessage).
- **queue_arn**: The Amazon Resource Name (ARN) of the queue — used in IAM policies, cross-service permissions, and as a target reference in other resources.
- **queue_name**: The name of the SQS queue. Includes the `.fifo` suffix for FIFO queues.

## How It Works

When you define an AwsSqsQueue resource, Planton:

1. **Creates Queue**: Provisions a Standard or FIFO SQS queue with the specified delivery settings.
2. **Configures Encryption**: Enables SQS-managed SSE or customer-managed KMS encryption.
3. **Sets Up Dead Letter Queue**: Configures the redrive policy to route failed messages to a DLQ when `dead_letter_config` is provided.
4. **Applies Access Policy**: Attaches the IAM access policy to the queue when `policy` is provided.
5. **Applies Tags**: Tags the queue with Planton metadata (organization, environment, resource kind, resource ID).

The resource uses Pulumi or Terraform under the hood depending on your stack configuration.

## Use Cases

### Message Decoupling Between Microservices
Use a Standard queue to decouple producers and consumers. The producer sends messages at its own rate; the consumer processes them asynchronously.

### Exactly-Once Ordered Processing
Use a FIFO queue with content-based deduplication for workflows that require strict message ordering and exactly-once delivery, such as financial transaction processing.

### Dead Letter Queue for Poison Messages
Configure a DLQ with `max_receive_count: 3` to isolate messages that consistently fail processing, preventing them from blocking the queue.

### Long Polling for Cost Efficiency
Set `receive_wait_time_seconds: 20` to enable long polling, significantly reducing the number of empty ReceiveMessage responses and lowering SQS costs.

### Event-Driven Architecture with SNS Fan-Out
Use an SQS queue as an SNS subscription endpoint. Set a queue `policy` that grants the SNS topic permission to publish messages. Multiple queues can subscribe to the same topic for fan-out.

## Important Notes

### Standard vs FIFO Queues
- **Standard**: Nearly unlimited throughput, best-effort ordering, at-least-once delivery. Suitable for most messaging workloads.
- **FIFO**: 300 TPS per API action (3000 with high-throughput mode), strict ordering within message groups, exactly-once processing. Required for order-sensitive workflows.
- Queue type (`fifo_queue`) cannot be changed after creation.

### Dead Letter Queue Pairing
- Both the source queue and the DLQ must be the same type (both Standard or both FIFO).
- Both must reside in the same AWS account and region.
- Set `max_receive_count` based on the nature of failures: use 1 for poison pill detection, 3–5 for transient error recovery.

### Encryption Choices
- **SSE-SQS** (`sqs_managed_sse_enabled: true`): Zero-cost encryption managed by SQS. Sufficient for most compliance requirements.
- **SSE-KMS** (`kms_key_id`): Use when you need customer-managed key rotation, CloudTrail auditing of key usage, or cross-account key sharing.
- The two options are mutually exclusive.

### Visibility Timeout Design
- Set the visibility timeout to at least as long as your consumer's maximum processing time plus a safety margin.
- If consumers frequently exceed the timeout, messages will be delivered more than once — design consumers to be idempotent.

## Validation Rules

The API enforces several validations:

1. **FIFO-only fields**: `content_based_deduplication`, `deduplication_scope`, and `fifo_throughput_limit` require `fifo_queue` to be true.
2. **Deduplication scope values**: Must be `"messageGroup"` or `"queue"` when set.
3. **Throughput limit values**: Must be `"perMessageGroupId"` or `"perQueue"` when set.
4. **Encryption mutual exclusion**: Cannot set both `kms_key_id` and `sqs_managed_sse_enabled`.
5. **KMS data key reuse**: Requires `kms_key_id` to be set; range 60–86400 when set.
6. **Message retention range**: 60–1209600 when set (0 uses AWS default).
7. **Max message size range**: 1024–1048576 when set (0 uses AWS default).
8. **Dead letter config**: `target_arn` is required; `max_receive_count` must be 1–1000.

## References

- [Amazon SQS Developer Guide](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/welcome.html)
- [Standard vs FIFO Queues](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-queue-types.html)
- [Dead-Letter Queues](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-dead-letter-queues.html)
- [Server-Side Encryption](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-server-side-encryption.html)
- [SQS Best Practices](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-best-practices.html)
