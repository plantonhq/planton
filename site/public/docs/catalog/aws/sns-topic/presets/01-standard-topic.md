---
title: "Standard Topic"
description: "This preset creates a Standard SNS topic with SHA256 message signatures. It is the simplest starting point for a notification or event distribution topic."
type: "preset"
rank: "01"
presetSlug: "01-standard-topic"
componentSlug: "sns-topic"
componentTitle: "SNS Topic"
provider: "aws"
icon: "package"
order: 1
---

# Standard Topic

This preset creates a Standard SNS topic with SHA256 message signatures. It is the simplest starting point for a notification or event distribution topic.

## When to Use

- Fan-out notifications to multiple subscribers (SQS, Lambda, email, HTTP/S)
- Event distribution across microservices
- Alert routing (CloudWatch → SNS → email/PagerDuty)
- Any pub/sub workload where strict ordering is not required

## Key Configuration Choices

- **SHA256 signatures** (`signatureVersion: 2`) — recommended for all new topics; provides stronger cryptographic signing than the legacy SHA1 default
- All other settings use AWS defaults (no encryption, no delivery policy override, PassThrough tracing)

## Placeholders to Replace

This preset uses a generic `my-topic` name. Rename `metadata.name` to match your use case (e.g., `order-events`, `system-alerts`, `image-processing`).

## Common Additions

- Add `kmsKeyId` with a `valueFrom` reference to an AwsKmsKey for encryption at rest
- Add `subscriptions` to define SQS, Lambda, or email delivery targets
- Add `policy` to grant other AWS services (EventBridge, S3) permission to publish
- Set `displayName` for topics that send SMS notifications
- Set `tracingConfig: Active` for X-Ray distributed tracing

## Related Presets

- **02-fifo-with-deduplication** — use when you need exactly-once delivery and strict ordering
- **03-fanout-to-sqs** — demonstrates the fan-out pattern with SQS subscriptions and filtering
