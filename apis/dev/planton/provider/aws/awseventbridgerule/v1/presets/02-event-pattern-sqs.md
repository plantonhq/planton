# Event Pattern to SQS

This preset creates an event-pattern-based EventBridge rule that routes matching AWS events to an SQS queue with a dead letter queue for failed deliveries. It demonstrates the core event-driven routing pattern — matching events by structure and delivering them to a reliable queue for asynchronous processing.

## When to Use

- Reacting to AWS service events (EC2 state changes, S3 operations, RDS events)
- Building event-driven pipelines where events are queued for processing
- Monitoring infrastructure changes with reliable event capture
- Any pattern where you need to decouple event producers from consumers

## Key Configuration Choices

- **Event pattern** — matches EC2 instance state changes (running/stopped). Replace with your own pattern.
- **SQS target** — events are enqueued for reliable, at-least-once processing.
- **Dead letter queue** — events that fail delivery (e.g., SQS queue is full or permissions are wrong) are routed to a DLQ instead of being lost.
- **Default event bus** — AWS service events arrive on the default bus.

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `<sqs-queue-arn>` | ARN of the SQS queue to receive events | `arn:aws:sqs:us-east-1:123456789012:ec2-events` |
| `<dlq-queue-arn>` | ARN of the SQS dead letter queue | `arn:aws:sqs:us-east-1:123456789012:ec2-events-dlq` |

Alternatively, use `valueFrom` references:

```yaml
arn:
  valueFrom:
    kind: AwsSqsQueue
    name: ec2-events
    fieldPath: status.outputs.queue_arn
```

## Common Additions

- Modify the `eventPattern` to match your specific event types
- Add `retryPolicy` to tune retry behavior
- Add `inputPath: "$.detail"` to deliver only the event detail (not the full envelope)
- Add additional targets for fan-out (e.g., Lambda for real-time processing + SQS for audit)
- Add `sqsConfig.messageGroupId` if targeting a FIFO queue

## Related Presets

- **01-schedule-lambda** — use for time-based (cron/rate) rule triggering
- **03-multi-target-fanout** — use when routing to multiple targets simultaneously
