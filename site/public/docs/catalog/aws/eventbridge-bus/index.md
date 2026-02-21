---
title: "EventBridge Bus"
description: "EventBridge Bus deployment documentation"
icon: "package"
order: 100
componentName: "awseventbridgebus"
---

# AWS EventBridge Bus

Deploys an AWS EventBridge custom event bus with optional KMS encryption, dead letter queue routing for undeliverable events, and configurable CloudWatch logging. Custom buses isolate event traffic from the default bus, enabling fine-grained access control and independent dead-letter queue routing for event-driven architectures.

## What Gets Created

When you deploy an AwsEventBridgeBus resource, OpenMCF provisions:

- **EventBridge Custom Event Bus** — an `aws_cloudwatch_event_bus` resource named after `metadata.name`, with optional description, KMS encryption, and AWS resource tags for organization, environment, and resource tracking
- **Dead Letter Config** — configured only when `deadLetterConfig` is provided, routes events that fail delivery to any rule target on this bus to the specified SQS queue
- **Log Config** — configured only when `logConfig` is provided, sends event delivery logs to CloudWatch Logs at the specified verbosity level

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An SQS queue** if using dead letter queue routing — the queue must exist in the same account and region as the event bus
- **A KMS key** if using customer-managed encryption — the key must grant EventBridge permission to encrypt and decrypt
- **A partner event source** if creating a partner bus — the source must already exist in the account

## Quick Start

Create a file `bus.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: my-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEventBridgeBus.my-events
spec:
  region: us-east-1
  description: Custom event bus for application events
```

Deploy:

```shell
openmcf apply -f bus.yaml
```

This creates a custom EventBridge bus with AWS-managed encryption and no dead letter queue or logging.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the EventBridge bus will be created (e.g., `us-east-1`, `eu-west-1`). | Required; non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description of the event bus. Maximum 512 characters. |
| `kmsKeyIdentifier` | `StringValueOrRef` | AWS-owned key | KMS key identifier for encrypting events on this bus. Accepts a key ARN, key ID, key alias, or key alias ARN. Can reference `AwsKmsKey` via `valueFrom`. |
| `eventSourceName` | `string` | — | Partner event source name for SaaS integrations (e.g., Datadog, PagerDuty). Must match the pattern `aws.partner/{partner}/{...}` and `metadata.name` must match this value. Immutable — changing it forces bus replacement. |
| `deadLetterConfig.arn` | `StringValueOrRef` | — | ARN of the SQS queue to use as the dead letter queue. Required when `deadLetterConfig` is set. The queue must exist in the same account and region. Can reference `AwsSqsQueue` via `valueFrom`. |
| `logConfig.level` | `string` | — | Logging verbosity. One of `OFF`, `ERROR`, `INFO`, `TRACE`. Required when `logConfig` is set. |
| `logConfig.includeDetail` | `string` | `NONE` | Whether to include full event detail in log entries. One of `NONE`, `FULL`. |

## Examples

### Production Bus with Encryption and DLQ

A bus with customer-managed KMS encryption, dead letter queue for undeliverable events, and error-level logging:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: payment-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEventBridgeBus.payment-events
spec:
  region: us-east-1
  description: Payment processing event bus
  kmsKeyIdentifier: arn:aws:kms:us-east-1:123456789012:key/abcd-1234-efgh-5678
  deadLetterConfig:
    arn: arn:aws:sqs:us-east-1:123456789012:payment-bus-dlq
  logConfig:
    level: ERROR
```

### Development Bus with Trace Logging

Verbose logging with full event detail for debugging event routing during development:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: dev-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEventBridgeBus.dev-events
spec:
  region: us-east-1
  description: Development bus with verbose logging
  logConfig:
    level: TRACE
    includeDetail: FULL
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding ARNs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: order-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEventBridgeBus.order-events
spec:
  region: us-east-1
  description: Order processing event bus with referenced resources
  kmsKeyIdentifier:
    valueFrom:
      kind: AwsKmsKey
      name: order-key
      field: status.outputs.key_arn
  deadLetterConfig:
    arn:
      valueFrom:
        kind: AwsSqsQueue
        name: order-bus-dlq
        field: status.outputs.queue_arn
  logConfig:
    level: INFO
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bus_name` | `string` | Event bus name — primary identifier used in EventBridge API calls and rule configurations |
| `bus_arn` | `string` | Event bus ARN — used in IAM policies, cross-account event delivery, and resource references |

## Related Components

- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides a customer-managed encryption key for event encryption
- [AwsSqsQueue](/docs/catalog/aws/sqs-queue) — provides a dead letter queue for undeliverable events
- [AwsEventBridgeRule](/docs/catalog/aws/eventbridge-rule) — attaches rules to this bus for event routing
- [AwsLambda](/docs/catalog/aws/lambda) — common target for EventBridge rules
- [AwsSnsTopic](/docs/catalog/aws/sns-topic) — fan-out target for EventBridge rules
