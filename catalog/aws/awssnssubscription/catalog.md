# AWS SNS Subscription

Deploys the delivery edge that wires an SNS topic to a target endpoint — an SQS queue, a Lambda function, an HTTP/S webhook, an email address, an SMS number, a Kinesis Data Firehose stream, or a mobile push endpoint. The subscription is its own node in the resource graph, not a setting of the topic: a topic can have many subscriptions, each owning its own protocol, endpoint, message filtering, raw delivery, dead-letter queue, and archived-message replay. The topic, the endpoint, the dead-letter queue, and the Firehose delivery role all accept ValueFromRef wiring, so a subscription composes directly with the resources on both ends of the delivery edge. Region, topic, protocol, and endpoint are create-time immutable — a different target is a different subscription.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SNS Subscription** -- attaches the chosen endpoint to the topic using the configured protocol. SQS, Lambda, Firehose, and mobile push endpoints confirm automatically; HTTP/S endpoints go through a confirmation handshake and email recipients confirm with a manual click
- **Message Filter** -- attached only when `filterPolicy` is provided; SNS evaluates it against message attributes (default) or the message body and delivers only matching messages
- **Dead-Letter Redrive** -- configured only when `deadLetterConfig` names an SQS queue; messages SNS could not deliver after all retries are moved there instead of being lost
- **Delivery Policy Override** -- attached only when `deliveryPolicy` is provided; overrides the topic-level HTTP/S retry policy for this subscription only
- **Archive Replay** -- configured only when `replayPolicy` is set; the subscription replays the FIFO topic's message archive from the chosen starting point before receiving live traffic

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An SNS topic** -- the topic this subscription delivers from. Reference an AwsSnsTopic Infra Component, or paste a literal topic ARN — including a cross-account topic that has granted this account `sns:Subscribe`.
- **The endpoint's own permission grant** -- SNS delivery is a two-sided contract. For `sqs`, the QUEUE's resource policy must grant `sqs:SendMessage` to `sns.amazonaws.com` (scoped by `aws:SourceArn` to the topic) — creating the subscription succeeds without it, but every delivery is silently dropped. For `lambda`, the function needs an invoke permission for `sns.amazonaws.com` (the AwsLambda `invokePermissions` fold).
- **A Firehose delivery role** (firehose only) -- an IAM role granting SNS permission to write to the delivery stream.

## Deploy

### Console

Open the deployment store, find **AWS SNS Subscription**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **SQS Fan-out Consumer** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSnsSubscription
metadata:
  name: fulfillment-events
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  topicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: order-events
      fieldPath: status.outputs.topic_arn
  protocol: sqs
  endpoint:
    valueFrom:
      kind: AwsSqsQueue
      name: fulfillment-queue
      fieldPath: status.outputs.queue_arn
  rawMessageDelivery: true
```

```shell
planton apply -f sns-subscription.yaml
```

This subscribes the referenced queue to the referenced topic with raw delivery — the consumer reads message bodies directly, without the SNS JSON envelope. To subscribe to a topic owned by another account, replace the `valueFrom` block with `value: <literal topic ARN>`. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the subscription to a topic and queue deployed in the same InfraPipeline:

```yaml
spec:
  topicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: order-events
      fieldPath: status.outputs.topic_arn
  endpoint:
    valueFrom:
      kind: AwsSqsQueue
      name: fulfillment-queue
      fieldPath: status.outputs.queue_arn
```

The InfraPipeline resolves the dependency graph, deploys the topic and queue first, then provisions the subscription with the resolved ARNs.

## Key Configuration

These are the most important decisions when configuring an SNS subscription. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Protocol and endpoint are the identity** -- The protocol decides how SNS delivers and what the endpoint must contain: `sqs` (queue ARN), `lambda` (function ARN), `http`/`https` (URL), `email`/`email-json` (address), `sms` (E.164 number), `firehose` (delivery stream ARN, requires `subscriptionRoleArn`), `application` (mobile platform endpoint ARN). Region, topic, protocol, and endpoint cannot change after creation — a different target is a different subscription.

**Message filtering** -- Set `filterPolicy` so this consumer receives only the messages it processes; everything else is filtered out before delivery. `filterPolicyScope` chooses whether the filter reads message attributes (default) or the message body. Filtering at the subscription beats discarding in the consumer — you do not pay for deliveries you never wanted.

**Raw message delivery** -- When `rawMessageDelivery` is `true` (SQS, HTTP/S, and Firehose only), SNS delivers the message body as-is instead of wrapping it in the JSON envelope containing MessageId, TopicArn, and Timestamp. Enable it when consumers expect the original payload.

**Dead-letter queue** -- Name an SQS queue in `deadLetterConfig` to catch messages SNS could not deliver after all retries. This is the subscription's own delivery DLQ — separate from any DLQ the endpoint has. The queue must live in the topic's account and region, and its policy must allow `sns.amazonaws.com` to SendMessage.

**Archived-message replay (FIFO)** -- When the topic is a FIFO topic with an `archivePolicy`, set `replayPolicy` (e.g. `{"PointType": "Timestamp", "StartingPoint": "2026-07-01T00:00:00Z"}`) to backfill this consumer from the archive before it receives live traffic — the mechanism for adding a consumer after launch without losing history. The topic's `beginning_archive_time` output reports the earliest valid starting point.

**HTTP/S confirmation handshake** -- HTTP/S endpoints must answer the SubscriptionConfirmation callback before deliveries start. Set `endpointAutoConfirms: true` when the endpoint confirms on its own, and `confirmationTimeoutMinutes` to bound how long deployment waits. A pending subscription delivers nothing — the `pending_confirmation` output shows the handshake state.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSnsTopic** | `topicArn` | `status.outputs.topic_arn` |
| **AwsSqsQueue** (sqs protocol) | `endpoint` | `status.outputs.queue_arn` |
| **AwsLambda** (lambda protocol) | `endpoint` | `status.outputs.function_arn` |
| **AwsKinesisFirehose** (firehose protocol) | `endpoint` | `status.outputs.delivery_stream_arn` |
| **AwsSqsQueue** (optional) | `deadLetterConfig.deadLetterTargetArn` | `status.outputs.queue_arn` |
| **AwsIamRole** (firehose protocol) | `subscriptionRoleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `subscription_arn` | Amazon Resource Name of the subscription | IAM policies for `sns:Unsubscribe` / `sns:SetSubscriptionAttributes` |
| `owner_id` | AWS account ID that owns the subscription | Cross-account auditing |
| `pending_confirmation` | Whether the endpoint has yet to confirm (HTTP/S, email) | Answering "why is my endpoint not receiving messages?" |
| `confirmation_was_authenticated` | Whether the confirmation request was signed | Security auditing of the handshake |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**SQS fan-out** -- The durable pub/sub backbone: topic to queue with raw delivery and a message filter so each consumer receives only its event types. Start from the **SQS Fan-out Consumer** preset.

**HTTPS webhook** -- Deliver events to an external system over HTTPS with a delivery-retry override and a DLQ for failed posts. Start from the **HTTPS Webhook with Dead-Letter Queue** preset.

## Works With

- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) -- the topic this subscription delivers from; its FIFO archive powers `replayPolicy` backfill
- [**AWS SQS Queue**](/infra-catalog/aws-sqs-queue) -- the most common delivery target, and the dead-letter queue for failed deliveries
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- invoke a function per message; pair with the function's `invokePermissions` for the SNS grant
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the delivery role SNS assumes for Firehose subscriptions
