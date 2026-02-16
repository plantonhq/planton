---
title: "Scheduled Lambda"
description: "This preset creates a scheduled EventBridge rule that triggers a Lambda function on a recurring schedule. It is the serverless replacement for traditional cron jobs — no servers to manage, no crontab..."
type: "preset"
rank: "01"
presetSlug: "01-schedule-lambda"
componentSlug: "eventbridge-rule"
componentTitle: "EventBridge Rule"
provider: "aws"
icon: "package"
order: 1
---

# Scheduled Lambda

This preset creates a scheduled EventBridge rule that triggers a Lambda function on a recurring schedule. It is the serverless replacement for traditional cron jobs — no servers to manage, no crontab files to maintain.

## When to Use

- Periodic data cleanup, archival, or maintenance tasks
- Recurring report generation
- Health checks and monitoring probes
- Any task that needs to run on a fixed interval or cron schedule

## Key Configuration Choices

- **Rate expression** (`rate(1 hour)`) — fires every hour. Change to `rate(5 minutes)`, `rate(1 day)`, or a cron expression for different schedules.
- **Default event bus** — schedule rules always use the default bus (schedules are AWS-managed events).
- **No input transformation** — the Lambda receives the standard EventBridge scheduled event payload.

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `<lambda-function-arn>` | ARN of the Lambda function to invoke | `arn:aws:lambda:us-east-1:123456789012:function:cleanup` |

Alternatively, use a `valueFrom` reference:

```yaml
arn:
  valueFrom:
    kind: AwsLambda
    name: cleanup-function
    fieldPath: status.outputs.function_arn
```

## Common Additions

- Add `input` with a constant JSON payload if the Lambda needs parameters
- Add `deadLetterConfig` with an SQS queue to catch failed invocations
- Add `retryPolicy` to tune retry behavior for critical scheduled tasks
- Change to a cron expression (e.g., `"cron(0 2 * * ? *)"` for 2 AM daily)

## Related Presets

- **02-event-pattern-sqs** — use when routing events by pattern rather than schedule
- **03-multi-target-fanout** — use when routing to multiple targets simultaneously
