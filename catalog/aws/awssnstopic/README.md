# AwsSnsTopic

The **AwsSnsTopic** resource provides a standardized way to provision and manage AWS SNS topics through Planton. It supports Standard and FIFO topic types with KMS encryption, IAM access and data-protection policies, FIFO message archiving, per-protocol delivery-status logging, and X-Ray tracing. Consumers attach through first-class `AwsSnsSubscription` resources that reference this topic's `topic_arn` output.

## Spec highlights

- **region** (required): AWS region for the topic.
- **fifo_topic**: Create a FIFO topic (strict ordering, exactly-once to SQS FIFO subscribers). Immutable; the `.fifo` name suffix is appended automatically.
- **content_based_deduplication / fifo_throughput_scope**: FIFO delivery semantics; `MessageGroup` scope enables high-throughput mode.
- **archive_policy**: FIFO-only message retention for subscription replay, e.g. `{"MessageRetentionPeriod": 30}`.
- **display_name**: Human-readable label (SMS "from" name).
- **kms_key_id**: Customer-managed KMS key for encryption at rest (references `AwsKmsKey`).
- **policy**: Resource-based access policy; AWS keeps an owner-only default when unset.
- **data_protection_policy**: Standard-only PII/PHI detection (audit/mask/deny), materialized as its own AWS resource.
- **delivery_policy**: Topic-level HTTP/S retry policy JSON.
- **delivery_feedback**: Per-protocol (application/firehose/http/lambda/sqs) delivery-status logging with success/failure IAM roles and a success sample rate.
- **tracing_config / signature_version**: X-Ray tracing and SHA1/SHA256 message signing.

## Outputs

- **topic_arn**: Topic ARN — the reference target for `AwsSnsSubscription.topic_arn`, EventBridge targets, and CloudWatch alarm actions.
- **topic_name**: Topic name (with `.fifo` suffix for FIFO topics).
- **owner**: AWS account ID owning the topic.
- **beginning_archive_time**: Start of the replayable archive window (FIFO archive only).

## How the modules deploy it

1. **Derives the name** from `metadata.name`, appending `.fifo` for FIFO topics.
2. **Creates the topic** with delivery, encryption, policy, logging, and tracing settings; identity tags match across both engines.
3. **Attaches the data protection policy** as its own resource when configured.

## Common patterns

- **Fan-out**: one topic, many `AwsSnsSubscription` resources targeting different SQS queues, each with its own filter policy.
- **Operational alerts**: CloudWatch alarms publish to the topic; email/SMS subscriptions notify humans, an SQS subscription feeds automation.
- **Ordered event streams**: a FIFO topic with `archive_policy` feeding SQS FIFO queues, with replay-based backfill for new consumers.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
