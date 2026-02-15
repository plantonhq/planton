---
title: "Standard Queue"
description: "This preset creates a Standard SQS queue with SQS-managed encryption and long polling enabled. It is the fastest way to get a production-safe message queue running with sensible defaults."
type: "preset"
rank: "01"
presetSlug: "01-standard-queue"
componentSlug: "sqs-queue"
componentTitle: "SQS Queue"
provider: "aws"
icon: "package"
order: 1
---

# Standard Queue

This preset creates a Standard SQS queue with SQS-managed encryption and long polling enabled. It is the fastest way to get a production-safe message queue running with sensible defaults.

## When to Use

- General-purpose message decoupling between microservices
- Background job queues and task processing
- Buffering writes ahead of a database or external API
- Any messaging workload where strict ordering is not required

## Key Configuration Choices

- **SQS-managed SSE** (`sqsManagedSseEnabled: true`) — zero-cost encryption at rest managed by SQS; no KMS key management required
- **Long polling** (`receiveWaitTimeSeconds: 20`) — reduces empty ReceiveMessage responses by waiting up to 20 seconds for messages to arrive, lowering SQS costs and unnecessary API calls
- **Default visibility timeout** (`visibilityTimeoutSeconds: 30`) — 30 seconds is sufficient for most lightweight consumers; increase if your consumer takes longer to process messages

## Placeholders to Replace

This preset uses a generic `my-queue` name. Rename `metadata.name` to match your use case (e.g., `order-events`, `notification-handler`, `task-queue`).

## Common Additions

- Add `deadLetterConfig` with a separate DLQ queue to isolate poison messages
- Increase `visibilityTimeoutSeconds` if your consumer processing exceeds 30 seconds
- Set `messageRetentionSeconds` to 1209600 (14 days) for queues that need longer retention
- Add `policy` to grant other AWS services (SNS, EventBridge) permission to send messages

## Related Presets

- **02-fifo-with-deduplication** — use when you need exactly-once processing and strict message ordering
