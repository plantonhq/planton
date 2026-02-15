# AwsSqsQueue — Research Documentation

## Overview

Amazon Simple Queue Service (SQS) is a fully managed message queuing service that enables decoupling and scaling of microservices, distributed systems, and serverless applications. SQS eliminates the complexity and overhead associated with managing and operating message-oriented middleware.

## Architecture

SQS operates as a distributed system with messages stored redundantly across multiple servers and data centers within an AWS region. Producers send messages to a queue; consumers poll the queue, process messages, and delete them after successful processing.

### Queue Types

**Standard queues** provide at-least-once delivery with best-effort ordering. They support a nearly unlimited number of API calls per second and are suitable for the vast majority of messaging workloads.

**FIFO queues** guarantee exactly-once processing and strict first-in-first-out ordering within each message group. They support up to 300 messages per second per API action (or 3000 per second with high-throughput mode enabled per message group).

### Dead Letter Queues

A dead letter queue (DLQ) receives messages that cannot be processed successfully after a configured number of attempts. The redrive policy specifies the source queue, the DLQ, and the maximum number of receives before a message is moved. DLQs are critical for isolating problematic messages without blocking the main queue.

### Encryption

SQS supports two encryption modes:
- **SSE-SQS**: SQS manages the encryption key. Zero additional cost. Sufficient for most compliance requirements.
- **SSE-KMS**: Uses a customer-managed AWS KMS key. Provides key rotation control, CloudTrail audit logging of key usage, and cross-account key sharing. Costs include KMS API calls.

## Design Decisions

### Why StringValueOrRef for kms_key_id

The KMS key ID field uses `StringValueOrRef` to enable infra-chart composability. In a typical infra chart, a KMS key is created as a separate resource and its ARN is wired into downstream resources. The `valueFrom` reference creates a dependency edge in the deployment DAG, ensuring the KMS key is provisioned before the queue.

### Why StringValueOrRef for dead_letter_config.target_arn

The DLQ target ARN uses `StringValueOrRef` to enable the common pattern of defining both the main queue and its DLQ in the same infra chart. The DLQ is deployed first (as it has no dependencies), and the main queue's `deadLetterConfig.targetArn` references the DLQ's output ARN via `valueFrom`.

### Why google.protobuf.Struct for policy

The IAM access policy is expressed as a `google.protobuf.Struct` rather than a plain JSON string. This provides a native YAML authoring experience — users write the policy structure directly in their YAML manifests without JSON escaping. The OpenMCF middleware handles serialization to JSON when passing to the IaC layer.

### Why string + CEL for FIFO fields (not proto enums)

Fields like `deduplication_scope` and `fifo_throughput_limit` use plain strings with CEL `in` validation rather than protobuf enums. This keeps the values provider-authentic (matching the exact AWS API strings) and avoids the prefix conventions required by proto enums (e.g., `DEDUPLICATION_SCOPE_MESSAGE_GROUP` vs the natural `messageGroup`).

### Deliberately Omitted for v1

- **redrive_allow_policy**: Controls which queues can use this queue as a DLQ. This is a niche use case affecting less than 20% of SQS deployments. Can be added in v2 if demand emerges.
- **Custom queue names**: OpenMCF derives the queue name from `metadata.name`. The FIFO `.fifo` suffix is appended automatically.

## Terraform Provider Reference

The primary Terraform resource is `aws_sqs_queue` from the `hashicorp/aws` provider. Separate resources exist for `aws_sqs_queue_policy`, `aws_sqs_queue_redrive_policy`, and `aws_sqs_queue_redrive_allow_policy`, but the main resource supports inline `policy` and `redrive_policy` attributes. This component uses the inline approach for simplicity.

Key Terraform attributes:
- `fifo_queue` (ForceNew) — queue type cannot be changed after creation
- `name` (ForceNew) — queue name is immutable
- `kms_master_key_id` conflicts with `sqs_managed_sse_enabled`
- `content_based_deduplication` requires `fifo_queue = true`
- `max_message_size` range: 1024–1048576 (up to 1 MB per recent AWS expansion)

## Pulumi Resource Reference

The Pulumi resource is `sqs.Queue` from `pulumi-aws/sdk/v7/go/aws/sqs`. Input properties map directly to Terraform attributes with camelCase naming. The queue URL is the primary output used for SQS API operations, while the ARN is used for IAM policies and cross-service references.
