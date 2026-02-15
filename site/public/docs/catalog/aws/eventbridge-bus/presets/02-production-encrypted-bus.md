---
title: "Production Encrypted Bus"
description: "This preset creates a production-grade EventBridge custom event bus with customer-managed KMS encryption, a dead letter queue for undeliverable events, and error-level logging. Designed for workloads..."
type: "preset"
rank: "02"
presetSlug: "02-production-encrypted-bus"
componentSlug: "eventbridge-bus"
componentTitle: "EventBridge Bus"
provider: "aws"
icon: "package"
order: 2
---

# Production Encrypted Bus

This preset creates a production-grade EventBridge custom event bus with customer-managed KMS encryption, a dead letter queue for undeliverable events, and error-level logging. Designed for workloads that require compliance-grade encryption, zero event loss, and operational observability.

## When to Use

- Production event-driven architectures where event loss is unacceptable
- Compliance environments requiring customer-managed encryption keys (SOC 2, HIPAA, PCI DSS)
- Systems that need audit trails for event delivery failures
- Any high-value event bus where operational visibility is critical

## Key Configuration Choices

- **KMS encryption** (`kmsKeyIdentifier`) — customer-managed key for event encryption at rest; provides key rotation control, CloudTrail audit logging, and cross-account key sharing
- **Dead letter queue** (`deadLetterConfig`) — routes events that fail delivery to any rule target to an SQS queue for investigation and reprocessing
- **Error logging** (`logConfig.level: ERROR`) — logs delivery failures to CloudWatch Logs without the volume overhead of logging all events
- **Exclude event detail** (`logConfig.includeDetail: NONE`) — reduces log volume by excluding full event payloads from log entries

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `<kms-key-arn>` | ARN of a KMS key for event encryption | `arn:aws:kms:us-east-1:123456789012:key/mrk-abc123` |
| `<dlq-queue-arn>` | ARN of an SQS queue for the dead letter queue | `arn:aws:sqs:us-east-1:123456789012:my-events-dlq` |

Alternatively, replace the literal values with `valueFrom` references to other OpenMCF resources:

```yaml
kmsKeyIdentifier:
  valueFrom:
    kind: AwsKmsKey
    name: my-key
    fieldPath: status.outputs.key_arn
deadLetterConfig:
  arn:
    valueFrom:
      kind: AwsSqsQueue
      name: my-events-dlq
      fieldPath: status.outputs.queue_arn
```

## Common Additions

- Increase `logConfig.level` to `TRACE` with `includeDetail: FULL` for debugging event routing during incident investigation
- Add EventBridge rules (AwsEventBridgeRule) to route events from this bus to Lambda, SQS, or Step Functions targets

## Related Presets

- **01-simple-custom-bus** — use for development/staging where encryption and DLQ are not required
- **03-partner-event-bus** — use when integrating with a SaaS partner via EventBridge
