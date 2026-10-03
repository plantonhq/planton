# AWS SNS Topic

Deploys an SNS topic (Standard or FIFO) with KMS encryption, IAM access and data protection policies, FIFO message archiving, per-protocol delivery status logging, and X-Ray tracing. The topic owns its identity, policies, and delivery posture; subscriptions are first-class AwsSnsSubscription resources that reference this topic's ARN — each owning its own protocol, endpoint, filtering, and dead-letter lifecycle. The KMS key and the delivery-logging IAM roles accept ValueFromRef wiring, so an encrypted topic composes directly with AwsKmsKey and AwsIamRole resources in the same InfraChart. The topic type is a one-way door: `fifoTopic` cannot be changed after creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SNS Topic** -- a Standard or FIFO topic named from your manifest's `metadata.name`, with configurable display name, signature version, and tracing
- **FIFO Configuration** -- created only when `fifoTopic` is `true`; enables content-based deduplication, per-message-group throughput scope, and message archiving. FIFO topic names automatically receive the `.fifo` suffix
- **Message Archive** -- created only when `archivePolicy` is set on a FIFO topic; retains published messages so AwsSnsSubscription resources can replay history
- **KMS Encryption** -- configured only when `kmsKeyId` is provided; encrypts message bodies at rest using a customer-managed key
- **IAM Access Policy** -- attached only when `policy` is provided; controls which AWS principals can publish or subscribe
- **Data Protection Policy** -- attached only when `dataProtectionPolicy` is provided on a Standard topic; audits, masks, or blocks sensitive data (PII/PHI) flowing through the topic
- **Delivery Status Logging** -- configured per protocol in `deliveryFeedback`; SNS writes per-delivery success/failure entries to CloudWatch Logs using the supplied IAM roles
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A KMS key** (optional) -- required when encrypting messages at rest. Unlike SQS, SNS has no managed SSE option -- encryption requires an explicit KMS key. Provide the key ID or ARN directly, or reference an AwsKmsKey Infra Component via ValueFromRef.
- **IAM logging roles** (optional) -- required for delivery status logging. Each role needs CloudWatch Logs write permissions (`logs:CreateLogGroup`, `logs:CreateLogStream`, `logs:PutLogEvents`) and a trust policy allowing `sns.amazonaws.com` to assume it.

## Deploy

### Console

Open the deployment store, find **AWS SNS Topic**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Topic** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSnsTopic
metadata:
  name: order-events
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  signatureVersion: 2
```

```shell
planton apply -f sns-topic.yaml
```

This creates a Standard topic with SHA-256 message signatures. No encryption, custom policies, or delivery logging are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the topic to a KMS key deployed in the same InfraPipeline:

```yaml
spec:
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: messaging-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the KMS key first, then provisions the SNS topic with the resolved values. Subscriptions join the same chart as AwsSnsSubscription resources referencing this topic's `topic_arn` output.

## Key Configuration

These are the most important decisions when configuring an SNS topic. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Standard vs. FIFO** -- Standard topics deliver messages to SQS, Lambda, HTTP/S, email, SMS, and Firehose subscribers with best-effort ordering. FIFO topics guarantee strict ordering and exactly-once delivery but only support SQS subscribers. Set `fifoTopic: true` for FIFO. This cannot be changed after creation.

**Subscriptions live on their own resource** -- Attach consumers by creating AwsSnsSubscription resources that reference this topic's `topic_arn` output. Each subscription owns its own protocol, endpoint, message filtering, raw delivery, dead-letter queue, and replay configuration -- a topic can have many, including subscriptions owned by other teams.

**Message archiving (FIFO)** -- Set `archivePolicy` (e.g. `{"MessageRetentionPeriod": 30}`) to retain published messages. New subscriptions can then replay history via their own `replayPolicy` before receiving live traffic -- the mechanism for backfilling consumers added after launch. The `beginning_archive_time` output reports when the archive became active.

**Encryption** -- Provide `kmsKeyId` to encrypt message bodies at rest. SNS does not offer a managed SSE option -- encryption always requires an explicit KMS key. Consider the additional KMS API call costs when publishing at high volume, and grant every subscriber `kms:Decrypt` on the key.

**Data protection (Standard only)** -- Set `dataProtectionPolicy` to audit, mask, or block sensitive data (names, card numbers, health identifiers) in message payloads. AWS rejects data protection policies on FIFO topics.

**Delivery status logging** -- Configure `deliveryFeedback` per protocol (SQS, Lambda, HTTP/S, Firehose, mobile push) to log delivery successes and failures to CloudWatch. Setting a failure role is the only way to see silent delivery drops -- such as an SQS queue whose resource policy is missing the SNS grant.

**Signature version** -- Set `signatureVersion: 2` (SHA-256) for new topics. Version 1 (SHA-1) is the AWS default but SHA-256 is recommended for stronger message authentication.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsIamRole** (optional) | `deliveryFeedback.<protocol>.successFeedbackRole` | `status.outputs.role_arn` |
| **AwsIamRole** (optional) | `deliveryFeedback.<protocol>.failureFeedbackRole` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `topic_arn` | Amazon Resource Name of the topic | AwsSnsSubscription `topicArn`, EventBridge rule targets, CloudWatch alarm actions, IAM policies |
| `topic_name` | Topic name (includes `.fifo` suffix for FIFO topics) | Application configuration, CloudWatch alarm dimensions |
| `owner` | AWS account ID that owns the topic | Cross-account subscription policies |
| `beginning_archive_time` | When the FIFO message archive became active | Choosing a valid replay starting point on subscriptions |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard topic** -- SHA-256 signatures with AWS-default delivery behavior. The starting point for pub/sub fan-out; attach consumers as AwsSnsSubscription resources. Start from the **Standard Topic** preset.

**FIFO topic with deduplication** -- Content-based deduplication and per-message-group throughput scope. Suitable for ordered event processing where downstream consumers are SQS FIFO queues. Start from the **FIFO Topic with Deduplication** preset.

**FIFO topic with archive** -- Strict ordering plus a message archive for consumer backfill: subscriptions added later replay history from the archive before going live. Start from the **FIFO Topic with Message Archive** preset.

## Works With

- [**AWS SNS Subscription**](/infra-catalog/aws-sns-subscription) -- delivers this topic's messages to an SQS queue, Lambda function, HTTP/S endpoint, or other target
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for message encryption at rest
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the CloudWatch logging roles for delivery status feedback
