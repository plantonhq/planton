# FIFO Queue with Deduplication

This preset creates a FIFO SQS queue with content-based deduplication, high-throughput mode, and a dead letter queue for failed messages. Designed for workflows that require exactly-once processing and strict ordering within message groups.

## When to Use

- Financial transaction processing where duplicate messages could cause double-charges
- Order processing pipelines that must maintain sequence
- Event sourcing systems requiring exactly-once, ordered event delivery
- Any workflow where message ordering and deduplication are critical

## Key Configuration Choices

- **FIFO queue** (`fifoQueue: true`) — guarantees exactly-once processing and strict ordering within each message group; queue name will automatically receive the `.fifo` suffix
- **Content-based deduplication** (`contentBasedDeduplication: true`) — SQS uses a SHA-256 hash of the message body as the deduplication ID, removing the need for producers to supply explicit deduplication IDs
- **Per-message-group deduplication** (`deduplicationScope: messageGroup`) — deduplication is scoped to each message group, allowing different groups to have identical messages
- **High-throughput FIFO** (`fifoThroughputLimit: perMessageGroupId`) — enables up to 3000 messages per second per message group (vs 300 TPS per queue in standard FIFO mode)
- **Dead letter queue** (`deadLetterConfig`) — routes messages that fail processing 3 times to a DLQ for investigation
- **SQS-managed SSE** — zero-cost encryption at rest

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `<dlq-queue-arn>` | ARN of a FIFO dead letter queue (must also be FIFO) | `arn:aws:sqs:us-east-1:123456789012:my-fifo-queue-dlq.fifo` |

Alternatively, replace the `targetArn.value` with a `valueFrom` reference to another AwsSqsQueue resource:

```yaml
targetArn:
  valueFrom:
    kind: AwsSqsQueue
    name: my-fifo-queue-dlq
    fieldPath: status.outputs.queue_arn
```

## Common Additions

- Increase `visibilityTimeoutSeconds` for consumers with longer processing times
- Set `messageRetentionSeconds` for retention beyond the 4-day default
- Add `policy` for cross-service access (e.g., SNS FIFO topic publishing to this queue)

## Related Presets

- **01-standard-queue** — use when strict ordering and exactly-once delivery are not required
