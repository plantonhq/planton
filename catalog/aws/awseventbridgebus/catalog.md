# AWS EventBridge Bus

Deploys a custom EventBridge event bus with optional customer-managed KMS encryption, dead letter queue for undeliverable events, CloudWatch Logs delivery logging, replay-able event archives, and a resource-based access policy for cross-account event publishing. Custom buses isolate application event traffic from the default AWS event bus, enabling fine-grained access control and dedicated event routing. The bus-level and per-archive KMS keys and the dead letter queue accept ValueFromRef wiring, so the bus composes directly with AwsKmsKey and AwsSqsQueue resources in the same InfraChart; AwsEventBridgeRule resources attach to the bus by referencing its `bus_name` output.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EventBridge Event Bus** -- a custom event bus named from your manifest's `metadata.name`, with optional description and partner event source configuration
- **KMS Encryption** -- configured only when `kmsKeyIdentifier` is provided; encrypts events at rest with a customer-managed key instead of the default AWS-owned key
- **Dead Letter Configuration** -- configured only when `deadLetterConfig` is provided; routes events that fail delivery to any rule target to the specified SQS queue
- **Log Configuration** -- configured only when `logConfig` is provided; sends event delivery logs to CloudWatch Logs at the specified verbosity level
- **Resource Policy** -- configured only when `resourcePolicy` is provided; attaches an IAM policy document to the bus granting other accounts, organizations, or roles permission to put events
- **Event Archives** -- one archive per `archives` entry, recorded from this bus: replay-able event history (EventBridge StartReplay) with optional pattern scoping, retention window, and per-archive KMS encryption
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A KMS key** (optional) -- required only when using customer-managed encryption. Provide the key ARN directly or reference an AwsKmsKey Infra Component via ValueFromRef.
- **An SQS queue** (optional) -- required when configuring a bus-level dead letter queue. The queue must exist in the same account and region. Provide the ARN directly or reference an AwsSqsQueue Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **AWS EventBridge Bus**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Simple Custom Bus** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEventBridgeBus
metadata:
  name: order-events
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: Custom event bus for order domain events
```

```shell
planton apply -f event-bus.yaml
```

This creates a custom event bus with AWS-managed encryption. No dead letter queue or logging is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the event bus to a KMS key and SQS dead letter queue deployed in the same InfraPipeline:

```yaml
spec:
  kmsKeyIdentifier:
    valueFrom:
      kind: AwsKmsKey
      name: events-encryption-key
      fieldPath: status.outputs.key_arn
  deadLetterConfig:
    arn:
      valueFrom:
        kind: AwsSqsQueue
        name: events-dlq
        fieldPath: status.outputs.queue_arn
```

The InfraPipeline resolves the dependency graph, deploys the KMS key and SQS queue first, then provisions the event bus with the resolved values.

## Key Configuration

These are the most important decisions when configuring an EventBridge bus. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**KMS encryption** -- By default, events are encrypted with an AWS-owned key at no additional cost. Provide `kmsKeyIdentifier` with a customer-managed KMS key when you need key rotation control, CloudTrail audit logging of event data access, or cross-account key sharing. Required for compliance-sensitive environments.

**Dead letter queue** -- Configure `deadLetterConfig` with an SQS queue ARN to capture events that fail delivery to any rule target on this bus. Without a DLQ, failed events are silently dropped. Essential for production workloads where event loss is unacceptable.

**Logging** -- Set `logConfig.level` to control event delivery logging. Use `ERROR` for production (logs delivery failures only) or `TRACE` for debugging (logs all events including matched/unmatched). Set `logConfig.includeDetail` to `FULL` to include complete event payloads in log entries, or `NONE` to reduce log volume.

**Partner event source** -- Set `eventSourceName` only when creating a bus for a SaaS partner integration (Datadog, Zendesk, PagerDuty). The bus name must match the partner source name exactly. This field is immutable after creation.

**Cross-account access** -- By default only the owning account can put events onto the bus. Set `resourcePolicy` to a standard IAM policy document to grant other accounts, an entire AWS Organization, or specific roles `events:PutEvents` permission — the mechanism behind hub-and-spoke event architectures where workload accounts publish into a central event hub.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `kmsKeyIdentifier` | `status.outputs.key_arn` |
| **AwsSqsQueue** (optional) | `deadLetterConfig.arn` | `status.outputs.queue_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `bus_name` | Name of the event bus | AwsEventBridgeRule `eventBusName`, PutEvents API calls |
| `bus_arn` | Amazon Resource Name of the event bus | IAM policies, cross-account event delivery |
| `archives` | Per-archive `name`/`arn` pairs for the archives declared in `spec.archives` | StartReplay operations and IAM policies scoped to an archive |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Simple custom bus** -- Minimal event bus with a description and AWS-managed encryption. The fastest way to isolate application events from the default bus for development or low-security workloads. Start from the **Simple Custom Bus** preset.

**Production encrypted bus** -- Customer-managed KMS encryption, SQS dead letter queue, and error-level logging. The recommended configuration for production event-driven architectures where event loss and compliance are concerns. Start from the **Production Encrypted Bus** preset.

**Partner event bus** -- Bus configured for a SaaS partner integration via EventBridge partner event sources. Use when receiving events from third-party services like Datadog or PagerDuty. Start from the **Partner Event Bus** preset.

## Works With

- [**AWS EventBridge Rule**](/infra-catalog/aws-event-bridge-rule) -- attaches to this bus and routes matching events to Lambda, SQS, SNS, and Step Functions targets
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for encrypting events at rest
- [**AWS SQS Queue**](/infra-catalog/aws-sqs-queue) -- provides a dead letter queue for events that fail delivery