# AWS SNS Topic

Deploys an AWS SNS topic — Standard or FIFO — with bundled subscriptions, KMS encryption, IAM access policies, message filtering, and subscription dead letter queues. SNS is the foundational pub/sub service for fan-out notifications, event-driven architectures, and cross-service messaging on AWS.

## What Gets Created

When you deploy an AwsSnsTopic resource, OpenMCF provisions:

- **SNS Topic** — a Standard or FIFO topic with the specified encryption, access policy, delivery policy, and observability configuration
- **SNS Subscriptions** — one per entry in the `subscriptions` list, each with its protocol, endpoint, filter policy, raw message delivery setting, and optional dead letter queue

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **Target endpoints** must exist before subscribing (SQS queues, Lambda functions, etc.) — use `valueFrom` references to ensure correct deployment ordering
- **KMS key** if encryption is desired — SNS requires an explicit KMS key (no managed SSE option)

## Quick Start

Create a file `topic.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: my-notifications
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSnsTopic.my-notifications
spec:
  signatureVersion: 2
```

Deploy:

```shell
openmcf apply -f topic.yaml
```

This creates a Standard SNS topic with SHA256 message signatures and all other AWS defaults.

## Configuration Reference

### Topic Type

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `fifoTopic` | `bool` | `false` | Create a FIFO topic (ordered, exactly-once). Cannot be changed after creation. |

### FIFO Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `contentBasedDeduplication` | `bool` | `false` | SHA-256 body hash as dedup ID. FIFO only. |
| `fifoThroughputScope` | `string` | — | `"Topic"` or `"MessageGroup"`. FIFO only. |

### Display

| Field | Type | Description |
|-------|------|-------------|
| `displayName` | `string` | Human-readable name for SMS and console display. |

### Encryption

| Field | Type | Description |
|-------|------|-------------|
| `kmsKeyId` | `StringValueOrRef` | Customer-managed KMS key. Can reference AwsKmsKey via `valueFrom`. |

### Access Control

| Field | Type | Description |
|-------|------|-------------|
| `policy` | `Struct` | IAM access policy document (JSON structure in YAML). |

### Delivery

| Field | Type | Description |
|-------|------|-------------|
| `deliveryPolicy` | `string` | HTTP/S retry policy JSON. |

### Observability

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `tracingConfig` | `string` | AWS: PassThrough | `"Active"` or `"PassThrough"`. |
| `signatureVersion` | `int32` | AWS: 1 | `1` (SHA1) or `2` (SHA256). |

### Subscriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Key for `subscription_arns` output map. |
| `protocol` | `string` | Yes | `sqs`, `lambda`, `http`, `https`, `email`, `email-json`, `sms`, `firehose`, `application`. |
| `endpoint` | `StringValueOrRef` | Yes | Target for delivery. Can reference other resources via `valueFrom`. |
| `filterPolicy` | `Struct` | No | Message filter (JSON structure in YAML). |
| `filterPolicyScope` | `string` | No | `"MessageAttributes"` or `"MessageBody"`. |
| `rawMessageDelivery` | `bool` | No | Deliver raw message without JSON envelope. |
| `redriveConfig.deadLetterTargetArn` | `StringValueOrRef` | No | SQS DLQ for delivery failures. |
| `subscriptionRoleArn` | `StringValueOrRef` | No | IAM role for Firehose. Required when protocol is `firehose`. |

## Examples

### Fan-Out to SQS with Filtering

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: order-events
spec:
  subscriptions:
    - name: fulfillment
      protocol: sqs
      endpoint:
        valueFrom:
          kind: AwsSqsQueue
          name: fulfillment-queue
          fieldPath: status.outputs.queue_arn
      filterPolicy:
        event_type:
          - order_placed
      rawMessageDelivery: true
    - name: analytics
      protocol: sqs
      endpoint:
        valueFrom:
          kind: AwsSqsQueue
          name: analytics-queue
          fieldPath: status.outputs.queue_arn
      rawMessageDelivery: true
```

### FIFO Topic

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: payment-events
spec:
  fifoTopic: true
  contentBasedDeduplication: true
  fifoThroughputScope: MessageGroup
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `topic_arn` | `string` | SNS topic ARN — used in IAM policies and cross-service references |
| `topic_name` | `string` | Topic name (includes `.fifo` suffix for FIFO topics) |
| `subscription_arns` | `map<string, string>` | Subscription name to ARN — reference specific subscriptions downstream |

## Related Components

- [AwsSqsQueue](/docs/catalog/aws/awssqsqueue) — common subscription target for fan-out and decoupling patterns
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides a customer-managed encryption key for topic encryption
- [AwsLambda](/docs/catalog/aws/awslambda) — subscribe Lambda functions for serverless event processing
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — provides roles for Firehose subscriptions and cross-account access
- [AwsEventBridgeRule](/docs/catalog/aws/awseventbridgerule) — routes EventBridge events to SNS topics as targets
