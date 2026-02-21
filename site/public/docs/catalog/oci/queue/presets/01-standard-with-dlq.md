---
title: "Standard Queue with Dead Letter Queue"
description: "This preset creates an OCI Queue configured for typical microservice-to-microservice asynchronous messaging. Messages are retained for 7 days, consumers have a 30-second visibility window to process..."
type: "preset"
rank: "01"
presetSlug: "01-standard-with-dlq"
componentSlug: "queue"
componentTitle: "Queue"
provider: "oci"
icon: "package"
order: 1
---

# Standard Queue with Dead Letter Queue

This preset creates an OCI Queue configured for typical microservice-to-microservice asynchronous messaging. Messages are retained for 7 days, consumers have a 30-second visibility window to process each message, and undeliverable messages are routed to a dead letter queue after 5 failed delivery attempts. Oracle-managed encryption protects message content at rest.

## When to Use

- Decoupling microservices with asynchronous request/response or event-driven patterns
- Task queues for background job processing (image resizing, email sending, report generation)
- Buffering bursty workloads to protect downstream services from overload
- Any messaging use case where standard message sizes (up to 128 KB) are sufficient

## Key Configuration Choices

- **7-day retention** (`retentionInSeconds: 604800`) -- messages remain in the queue for up to 7 days if not consumed. This is OCI's default and provides a generous safety window for consumer outages. Note: retention is ForceNew -- changing it recreates the queue.
- **30-second visibility timeout** (`visibilityInSeconds: 30`) -- after a consumer receives a message, it becomes invisible to other consumers for 30 seconds. The consumer must delete the message within this window or it becomes visible again for redelivery. Increase for long-running tasks; decrease for fast processors.
- **30-second polling timeout** (`timeoutInSeconds: 30`) -- GetMessages calls block for up to 30 seconds waiting for new messages (long polling). This reduces empty-response API calls and cost.
- **Dead letter queue with 5 retries** (`deadLetterQueueDeliveryCount: 5`) -- messages that fail delivery 5 times are moved to the queue's built-in DLQ. This prevents poison messages from blocking consumers indefinitely. Set to 0 to disable the DLQ.
- **Oracle-managed encryption** -- no `customEncryptionKeyId` is set, so OCI encrypts message content at rest using Oracle-managed keys. This is sufficient for most workloads and avoids KMS key management overhead.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the queue will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |

## Related Presets

- **02-encrypted-large-messages** -- use instead when messages exceed 128 KB, customer-managed encryption is required, or partitioned consumption via consumer groups is needed
