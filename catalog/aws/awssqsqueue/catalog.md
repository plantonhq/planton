# AWS SQS Queue

Deploys an SQS queue (Standard or FIFO) with configurable delivery settings, dead letter queue routing, encryption (SQS-managed SSE or customer-managed KMS), and IAM access policies. The KMS key, the dead letter queue target, and the redrive-allow source list all accept ValueFromRef wiring, so a queue composes directly with AwsKmsKey and other AwsSqsQueue resources in the same InfraChart. The queue type is a one-way door: `fifoQueue` cannot be changed after creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SQS Queue** -- a Standard or FIFO queue named from your manifest's `metadata.name`, with configurable visibility timeout, message retention, delay, and long polling settings
- **FIFO Configuration** -- created only when `fifoQueue` is `true`; enables content-based deduplication, per-message-group deduplication scope, and high-throughput mode. FIFO queue names automatically receive the `.fifo` suffix
- **Dead Letter Queue Redrive Policy** -- configured only when `deadLetterConfig` is provided; routes messages that exceed `maxReceiveCount` receive attempts to a target DLQ
- **Redrive Allow Policy** -- attached only when `redriveAllowPolicy` is provided; restricts which source queues may use this queue as their dead letter queue (`allowAll`, `denyAll`, or `byQueue` with an explicit list of up to 10 source queue ARNs)
- **Server-Side Encryption** -- AWS encrypts new queues with SSE-SQS by default; `sqsManagedSseEnabled` pins that on (`true`) or explicitly opts out (`false`), or provide `kmsKeyId` for SSE-KMS (mutually exclusive with setting `sqsManagedSseEnabled` at all)
- **IAM Access Policy** -- attached only when `policy` is provided; controls which AWS principals can send, receive, or manage messages
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A KMS key** (optional) -- required only when using KMS encryption instead of SQS-managed SSE. Provide the key ID or ARN directly, or reference an AwsKmsKey Infra Component via ValueFromRef.
- **A dead letter queue** (optional) -- must exist in the same AWS account and region, and be the same type (Standard or FIFO) as the source queue. Provide the ARN directly or reference another AwsSqsQueue Infra Component.

## Deploy

### Console

Open the deployment store, find **AWS SQS Queue**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Queue** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSqsQueue
metadata:
  name: order-events
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  sqsManagedSseEnabled: true
  receiveWaitTimeSeconds: 20
  visibilityTimeoutSeconds: 30
```

```shell
planton apply -f sqs-queue.yaml
```

This creates a Standard queue with SQS-managed encryption, long polling enabled (20s wait), and a 30-second visibility timeout. No dead letter queue or FIFO settings are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When using KMS encryption or dead letter queues, use ValueFromRef to wire dependencies deployed in the same InfraPipeline:

```yaml
spec:
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: messaging-key
      fieldPath: status.outputs.key_arn
  deadLetterConfig:
    targetArn:
      valueFrom:
        kind: AwsSqsQueue
        name: order-events-dlq
        fieldPath: status.outputs.queue_arn
    maxReceiveCount: 3
```

The InfraPipeline resolves the dependency graph, deploys the KMS key and DLQ first, then provisions the queue with the resolved values.

## Key Configuration

These are the most important decisions when configuring an SQS queue. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Standard vs. FIFO** -- Standard queues offer maximum throughput with best-effort ordering and at-least-once delivery. FIFO queues guarantee exactly-once processing and strict message ordering within each message group, at the cost of lower throughput. Set `fifoQueue: true` for FIFO. This cannot be changed after creation.

**Encryption method** -- Leave `sqsManagedSseEnabled` unset to keep AWS's default (new queues come up SSE-SQS encrypted), set it `true` to pin zero-cost SQS-managed encryption explicitly, set it `false` to opt out of default encryption entirely, or provide `kmsKeyId` for customer-managed KMS encryption with CloudTrail audit trails and key rotation control. Setting `sqsManagedSseEnabled` (either value) and `kmsKeyId` are mutually exclusive.

**Dead letter queue** -- Configure `deadLetterConfig` with a `targetArn` and `maxReceiveCount` to route poison messages to a separate queue for investigation. Both queues must be the same type (Standard or FIFO) and in the same account and region.

**Shared DLQ protection** -- When this queue serves as a dead letter queue for other queues, set `redriveAllowPolicy` with `redrivePermission: byQueue` and list the permitted `sourceQueueArns` (1-10 entries, each a literal ARN or a ValueFromRef to another AwsSqsQueue). Leaving the policy unset keeps AWS's default: any queue in the account may dead-letter here.

**Access policy** -- Provide `policy` (a standard IAM policy document) to let SNS topics publish to the queue, receive S3 or EventBridge notifications, or grant cross-account consumers access. Leave it unset to keep the queue private to the owning account.

**Long polling** -- Set `receiveWaitTimeSeconds` to 1-20 to reduce empty responses and lower costs. The default (0) uses short polling, which returns immediately even when no messages are available.

**FIFO high throughput** -- For FIFO queues processing independent message groups, set `fifoThroughputLimit: perMessageGroupId` and `deduplicationScope: messageGroup` to unlock higher per-queue throughput.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsSqsQueue** (optional) | `deadLetterConfig.targetArn` | `status.outputs.queue_arn` |
| **AwsSqsQueue** (optional) | `redriveAllowPolicy.sourceQueueArns` | `status.outputs.queue_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `queue_url` | SQS queue URL | Application SDK calls (SendMessage, ReceiveMessage) |
| `queue_arn` | Amazon Resource Name | IAM policies, SNS subscription endpoints, dead letter queue targets |
| `queue_name` | Queue name (includes `.fifo` suffix for FIFO queues) | Application configuration, CloudWatch alarm dimensions |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard queue with long polling** -- SQS-managed encryption, 20-second long polling, and 30-second visibility timeout. The baseline configuration for most workloads. Start from the **Standard Queue** preset.

**FIFO queue with deduplication** -- Content-based deduplication, per-message-group scope, high-throughput mode, and a dead letter queue for poison messages. Suitable for order processing and financial transaction pipelines. Start from the **FIFO Queue with Deduplication** preset.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for KMS-based message encryption
- [**AWS SQS Queue**](/infra-catalog/aws-sqs-queue) -- provides a dead letter queue target for failed message routing