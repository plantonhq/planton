# FIFO Topic with Deduplication

This preset creates a FIFO SNS topic with content-based deduplication and high-throughput mode (per message group). Designed for workflows that require exactly-once, ordered message delivery.

## When to Use

- Financial transaction processing where duplicate messages could cause double-charges
- Order processing pipelines that must maintain sequence
- Event sourcing systems requiring exactly-once, ordered event delivery to SQS FIFO queues
- Any workflow where message ordering and deduplication are critical

## Key Configuration Choices

- **FIFO topic** (`fifoTopic: true`) — guarantees exactly-once delivery and strict ordering; topic name will automatically receive the `.fifo` suffix
- **Content-based deduplication** (`contentBasedDeduplication: true`) — SNS uses a SHA-256 hash of the message body as the deduplication ID, removing the need for publishers to supply explicit deduplication IDs
- **Per-message-group throughput** (`fifoThroughputScope: MessageGroup`) — enables up to 3000 publishes per second per message group (vs 300 per second per topic in standard FIFO mode)
- **SHA256 signatures** (`signatureVersion: 2`)

## Important: FIFO Subscriber Limitations

FIFO topics only support SQS FIFO queue subscriptions. Standard protocols (email, SMS, HTTP/S, Lambda) are not supported for FIFO topics. Ensure your subscribers are SQS FIFO queues.

## Placeholders to Replace

This preset uses a generic `my-fifo-topic` name. Rename `metadata.name` to match your use case (e.g., `payment-events`, `order-processing`).

## Common Additions

- Add `subscriptions` with `protocol: sqs` pointing to FIFO SQS queues
- Add `kmsKeyId` for encryption at rest
- Add `policy` for cross-service publishing permissions

## Related Presets

- **01-standard-topic** — use when strict ordering and exactly-once delivery are not required
- **03-fanout-to-sqs** — demonstrates the fan-out pattern (Standard topic only)
