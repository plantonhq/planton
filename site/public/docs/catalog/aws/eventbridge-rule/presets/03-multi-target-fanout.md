---
title: "Multi-Target Fan-Out"
description: "This preset creates an event-pattern-based rule on a custom event bus that routes matching events to two targets simultaneously — a Lambda function with input transformation and retry policy, and an..."
type: "preset"
rank: "03"
presetSlug: "03-multi-target-fanout"
componentSlug: "eventbridge-rule"
componentTitle: "EventBridge Rule"
provider: "aws"
icon: "package"
order: 3
---

# Multi-Target Fan-Out

This preset creates an event-pattern-based rule on a custom event bus that routes matching events to two targets simultaneously — a Lambda function with input transformation and retry policy, and an SQS queue with simple input extraction. It demonstrates the production-grade fan-out pattern with independent reliability per target.

## When to Use

- Event-driven architectures where a single event triggers multiple actions
- Processing pipelines with both real-time (Lambda) and asynchronous (SQS) branches
- Scenarios requiring independent retry and DLQ policies per target
- Custom application events on a dedicated event bus

## Key Configuration Choices

- **Custom event bus** — targets a custom bus (not the default AWS bus). Replace with your bus name or `valueFrom` reference.
- **Two targets** — demonstrates fan-out to Lambda (with transformation) and SQS (with input extraction).
- **Input transformer** — reshapes the event for the Lambda target, extracting only relevant fields.
- **Input path** — extracts only the `detail` object for the SQS target (removes envelope metadata).
- **Retry policy** — Lambda target retries for up to 1 hour with 10 attempts.
- **Dead letter queue** — Lambda target has a DLQ for events that fail after all retries.

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `<custom-bus-name>` | Name of the custom event bus | `order-events` |
| `<lambda-function-arn>` | ARN of the Lambda processing function | `arn:aws:lambda:us-east-1:123456789012:function:order-processor` |
| `<dlq-queue-arn>` | ARN of the dead letter queue | `arn:aws:sqs:us-east-1:123456789012:order-processor-dlq` |
| `<sqs-queue-arn>` | ARN of the analytics SQS queue | `arn:aws:sqs:us-east-1:123456789012:order-analytics` |

Alternatively, use `valueFrom` references for infra-chart composability:

```yaml
eventBusName:
  valueFrom:
    kind: AwsEventBridgeBus
    name: order-events
    fieldPath: status.outputs.bus_name
```

## Common Additions

- Add more targets (up to 5 per rule)
- Add `roleArn` to targets that require assumed-role invocation (Step Functions, ECS)
- Add `sqsConfig.messageGroupId` if targeting a FIFO SQS queue
- Add `state: DISABLED` to create the rule without activating it

## Related Presets

- **01-schedule-lambda** — use for time-based (cron/rate) rule triggering
- **02-event-pattern-sqs** — use for simpler single-target routing
