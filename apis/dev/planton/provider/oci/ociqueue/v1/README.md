# OciQueue

## Overview

OciQueue is an Planton component that deploys an OCI Queue. It provides a single declarative manifest to create a fully managed, serverless message queue with configurable delivery semantics, encryption, and consumption patterns.

## Purpose

OCI Queue is a serverless messaging service for asynchronous communication between decoupled services. It provides at-least-once delivery with configurable visibility timeouts, dead-letter queue support for handling poison messages, and optional capabilities for large messages and consumer groups. This component provisions the queue; producers and consumers interact with it via the messages endpoint exported as a stack output.

## Key Features

- **Dead-letter queue** — configurable delivery count before messages are moved to DLQ, preventing infinite reprocessing of failing messages.
- **Customer-managed encryption** — optional KMS key for encrypting message content at rest.
- **Large message support** — optional capability enabling messages up to 512 KB (standard limit is lower).
- **Consumer groups** — optional capability for partitioned consumption patterns, enabling multiple consumer groups to process messages independently.
- **Configurable timeouts** — visibility timeout (how long a consumed message stays invisible), polling timeout (long-polling duration), and retention period.
- **Foreign key references** — `compartmentId` and `customEncryptionKeyId` support `valueFrom` for infra-chart composability.

## Constraints

- `retentionInSeconds` is ForceNew — changing it forces queue recreation.
- All other fields are updatable after creation.
- Capabilities (`isLargeMessagesEnabled`, `consumerGroupConfig`) are additive — once enabled, they cannot be removed.
- `channelConsumptionLimit` is an integer percentage; consult OCI documentation for the valid range.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Development message bus | Minimal queue with defaults |
| Order processing pipeline | DLQ enabled, extended visibility timeout, KMS encryption |
| Large payload ingestion | `isLargeMessagesEnabled: true` for up to 512 KB messages |
| Multi-consumer fan-out | Consumer groups with separate DLQ counts per group |
| High-security data pipeline | KMS encryption + compartment isolation |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **KMS encryption** — customer-managed keys for message content encryption at rest.
- **Dead-letter queue** — automatic poison message isolation after configurable delivery attempts.
- **Messages endpoint** — output provides the URL for producing and consuming messages.
