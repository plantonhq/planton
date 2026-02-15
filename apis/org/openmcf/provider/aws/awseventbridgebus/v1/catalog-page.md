# AWS EventBridge Bus

Deploys an AWS EventBridge custom event bus with optional KMS encryption, dead letter queue routing for undeliverable events, and configurable logging. EventBridge is the backbone of event-driven architectures on AWS, enabling decoupled communication between microservices, serverless functions, and SaaS integrations.

## What Gets Created

When you deploy an AwsEventBridgeBus resource, OpenMCF provisions:

- **EventBridge Custom Bus** — a custom event bus with the specified name, description, and encryption configuration
- **Dead Letter Config** — created only when `deadLetterConfig` is provided, routes undeliverable events to an SQS queue for investigation
- **Log Config** — created only when `logConfig` is provided, sends event delivery logs to CloudWatch Logs

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An SQS queue** if you plan to use dead letter queue routing — the queue must exist in the same account and region
- **A KMS key** if you plan to use customer-managed encryption — the key must exist and allow EventBridge to use it

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
  description: Custom event bus for application events
```

Deploy:

```shell
openmcf apply -f bus.yaml
```

This creates a custom EventBridge bus with AWS-managed encryption and all other defaults.

## Configuration Reference

### Core

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description (max 512 chars). |
| `kmsKeyIdentifier` | `StringValueOrRef` | AWS-owned key | KMS key for event encryption. Can reference AwsKmsKey via `valueFrom`. |
| `eventSourceName` | `string` | — | Partner event source name. Immutable. Bus name must match. |

### Dead Letter Queue

| Field | Type | Description |
|-------|------|-------------|
| `deadLetterConfig.arn` | `StringValueOrRef` | SQS queue ARN for undeliverable events. Can reference AwsSqsQueue via `valueFrom`. Required when `deadLetterConfig` is set. |

### Logging

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `logConfig.level` | `string` | — | `"OFF"`, `"ERROR"`, `"INFO"`, or `"TRACE"`. Required when `logConfig` is set. |
| `logConfig.includeDetail` | `string` | `"NONE"` | `"NONE"` or `"FULL"`. Whether to include event detail in logs. |

## Examples

### Production Bus with Encryption and DLQ

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: payment-events
spec:
  description: Payment processing event bus
  kmsKeyIdentifier:
    valueFrom:
      kind: AwsKmsKey
      name: payment-key
      fieldPath: status.outputs.key_arn
  deadLetterConfig:
    arn:
      valueFrom:
        kind: AwsSqsQueue
        name: payment-bus-dlq
        fieldPath: status.outputs.queue_arn
  logConfig:
    level: ERROR
```

### Development Bus with Trace Logging

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeBus
metadata:
  name: dev-events
spec:
  description: Development bus with verbose logging
  logConfig:
    level: TRACE
    includeDetail: FULL
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bus_name` | `string` | Event bus name — primary identifier for API calls and rule configurations |
| `bus_arn` | `string` | Event bus ARN — used in IAM policies and cross-service references |

## Related Components

- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides a customer-managed encryption key for event encryption
- [AwsSqsQueue](/docs/catalog/aws/awssqsqueue) — provides a dead letter queue for undeliverable events
- [AwsEventBridgeRule](/docs/catalog/aws/awseventbridgerule) — attaches rules to this bus for event routing
- [AwsLambda](/docs/catalog/aws/awslambda) — common target for EventBridge rules
- [AwsSnsTopic](/docs/catalog/aws/awssnstopic) — fan-out target for EventBridge rules
