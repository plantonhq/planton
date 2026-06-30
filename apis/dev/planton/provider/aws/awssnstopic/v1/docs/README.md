# AwsSnsTopic — Research Documentation

## Overview

Amazon Simple Notification Service (SNS) is a fully managed pub/sub messaging service that enables decoupling of microservices, distributed systems, and serverless applications. SNS supports fan-out delivery to multiple subscribers simultaneously, including SQS queues, Lambda functions, HTTP/S endpoints, email, SMS, and mobile push notifications.

## Architecture

SNS operates as a publish-subscribe system. Publishers send messages to a topic; SNS delivers copies to all subscribed endpoints. The decoupling between publishers and subscribers enables event-driven architectures where new consumers can be added without modifying publishers.

### Topic Types

**Standard topics** provide at-least-once delivery with best-effort ordering. They support nearly unlimited throughput and all subscriber protocol types. Suitable for the vast majority of pub/sub workloads.

**FIFO topics** guarantee strict ordering and exactly-once message delivery, but only to SQS FIFO queue subscribers. They support up to 300 publishes per second (or 3000 per second with high-throughput mode per message group). FIFO topics enforce message deduplication via explicit dedup IDs or content-based hashing.

### Message Filtering

SNS supports server-side message filtering at the subscription level. A filter policy is a JSON document that defines which messages the subscription receives based on message attributes or body content. Messages that do not match the filter are not delivered, reducing unnecessary processing at the subscriber. Filter policy scope determines whether filtering applies to message attributes (default, more performant) or the message body (flexible but heavier).

### Subscription Dead Letter Queues

Each subscription can have its own dead letter queue (DLQ). When SNS fails to deliver a message after all retry attempts (as defined by the delivery policy), the message is sent to the specified SQS queue. This is distinct from an SQS queue's own DLQ — the subscription DLQ handles SNS-to-subscriber delivery failures, while the SQS DLQ handles consumer processing failures.

### Encryption

SNS supports server-side encryption using AWS KMS customer-managed keys. Unlike SQS (which offers both SQS-managed SSE and KMS encryption), SNS only supports KMS-based encryption. When encryption is enabled, all messages are encrypted at rest. Subscribers must have permission to decrypt using the KMS key.

## Design Decisions

### Why Bundle Subscriptions with the Topic

Subscriptions are bundled inline with the topic (as a repeated field in the spec) because:
1. Subscriptions cannot exist without a topic — they are tightly coupled.
2. Most users define subscriptions at the same time as the topic.
3. Bundling enables the `subscription_arns` output map, allowing downstream resources to reference specific subscription ARNs via `valueFrom`.
4. The alternative (a separate AwsSnsSubscription resource kind) would create unnecessary resource sprawl for a resource that has no independent lifecycle.

### Why `name` Field on Subscriptions

The `name` field is a user-assigned key that serves three purposes:
1. **Output map key**: The `subscription_arns` output uses `name` as the map key, enabling `status.outputs.subscription_arns.{name}` references.
2. **Pulumi resource naming**: Each subscription's Pulumi resource name is derived from the topic name and subscription name, ensuring stable resource identity across updates.
3. **YAML readability**: Named subscriptions are more readable than array-indexed ones.

This pattern follows the AwsS3ObjectSet precedent, which uses object keys as map keys for output tracking.

### Why StringValueOrRef for Subscription Endpoints (Polymorphic)

The `endpoint` field uses `StringValueOrRef` without a `default_kind` because the target resource type varies by protocol:
- `sqs` → AwsSqsQueue ARN
- `lambda` → AwsLambda function ARN
- `https` → URL string
- `email` → email address
- `firehose` → Firehose stream ARN

Setting a single `default_kind` would be misleading since most protocols don't target the same resource type.

### Why google.protobuf.Struct for Policy and Filter Policy

Both the access policy and subscription filter policies use `google.protobuf.Struct`. This provides a native YAML authoring experience — users write the JSON structure directly in YAML without escaping. The middleware and IaC modules serialize the struct to JSON when passing to the AWS API.

### Why String for delivery_policy

The delivery policy is a specialized JSON format (HTTP retry configuration with specific structure for healthyRetryPolicy, sicklyRetryPolicy, throttlePolicy). It's not a general-purpose JSON document like an IAM policy. A plain string is more appropriate because:
1. Most users copy the delivery policy from AWS documentation or examples.
2. The structure is too specialized to benefit from `google.protobuf.Struct` authoring.
3. It's used by a small minority of SNS users (HTTP/S subscriptions with custom retry behavior).

### Deliberately Omitted for v1

- **Per-protocol delivery status logging**: Success/failure feedback role ARNs and sample rates for application, firehose, http, lambda, and sqs protocols. This is 15 fields (5 protocols × 3 fields each) used by less than 20% of SNS users. Adding these would significantly expand the spec for a feature that can be added in v2 without breaking changes.
- **archive_policy**: Message archiving for FIFO topics, enabling message replay. Relatively new AWS feature (2023+) with niche adoption. Can be added later.
- **Subscription delivery_policy**: Per-subscription HTTP/S retry configuration. The topic-level `delivery_policy` covers the common case.
- **Subscription replay_policy**: Message replay from FIFO topic archives. Requires `archive_policy` which is also omitted.
- **Subscription confirmation_timeout_in_minutes / endpoint_auto_confirms**: HTTP/S confirmation behavior. Niche. The IaC modules cannot wait for email/SMS confirmation.

## Terraform Provider Reference

The primary Terraform resources are:
- `aws_sns_topic` — the topic itself with all configuration
- `aws_sns_topic_subscription` — each subscription as a separate resource
- `aws_sns_topic_policy` — standalone policy resource (we use the inline `policy` attribute instead)

Key Terraform attributes:
- `fifo_topic` (ForceNew) — topic type cannot be changed after creation
- `name` (ForceNew) — topic name is immutable; FIFO topics must end with `.fifo`
- `content_based_deduplication` requires `fifo_topic = true`
- `fifo_throughput_scope` valid values: `"Topic"`, `"MessageGroup"` (note: different naming from SQS's `fifo_throughput_limit`)
- Subscription `protocol` and `endpoint` are ForceNew — changing them recreates the subscription

## Pulumi Resource Reference

The Pulumi resources are:
- `sns.Topic` from `pulumi-aws/sdk/v7/go/aws/sns` — all topic configuration
- `sns.TopicSubscription` from the same package — one per subscription

Input properties map directly to Terraform attributes with camelCase naming. The topic ARN is the primary output used for IAM policies and cross-service references.

## Infra Chart Composability

The AwsSnsTopic resource is designed for composition in infra charts:

**As a publisher target**: Other resources (EventBridge rules, CloudWatch alarms, S3 event notifications) reference `topic_arn` to publish events.

**As a fan-out hub**: Subscriptions use `valueFrom` to reference AwsSqsQueue, AwsLambda, and other resource outputs, creating dependency edges in the infra chart DAG.

**Subscription-level references**: The `subscription_arns` map enables fine-grained downstream references to specific subscriptions, not just the topic.
