# AWS SNS Topic

Deploys an AWS SNS topic — Standard or FIFO — with inline subscriptions, optional KMS encryption, IAM access policies, message filtering, and subscription dead letter queues. The component handles FIFO naming conventions automatically and exports a subscription ARN map for downstream wiring.

## What Gets Created

When you deploy an AwsSnsTopic resource, OpenMCF provisions:

- **SNS Topic** — an `aws_sns_topic` resource configured as Standard or FIFO, with the specified encryption, access policy, delivery policy, tracing, and signature settings
- **SNS Subscriptions** — one `aws_sns_topic_subscription` per entry in `subscriptions`, each with its protocol, endpoint, optional filter policy, raw message delivery flag, optional dead letter queue, and optional Firehose role

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **Target endpoints** must exist before subscribing (SQS queues, Lambda functions, HTTP endpoints, etc.) — use `valueFrom` references to ensure correct deployment ordering
- **KMS key** if encryption is desired — SNS has no managed SSE option, encryption requires an explicit customer-managed KMS key

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

### Required Fields

All top-level spec fields are optional. A topic can be created with an empty `spec` to get a Standard topic with all AWS defaults.

When entries are added to `subscriptions`, each entry requires:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `subscriptions[].name` | `string` | Unique name for this subscription. Used as the key in the `subscription_arns` output map. | Required |
| `subscriptions[].protocol` | `string` | Delivery protocol: `sqs`, `lambda`, `http`, `https`, `email`, `email-json`, `sms`, `firehose`, `application`. | Required. Must be one of the listed values. |
| `subscriptions[].endpoint` | `StringValueOrRef` | Target for message delivery. Format depends on protocol (ARN, URL, email, phone number). | Required. Can reference any resource via `valueFrom`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `fifoTopic` | `bool` | `false` | Create a FIFO topic with strict ordering and exactly-once delivery. Cannot be changed after creation. FIFO topics automatically get a `.fifo` name suffix. |
| `contentBasedDeduplication` | `bool` | `false` | Use SHA-256 hash of the message body as the deduplication ID. Only valid when `fifoTopic` is `true`. |
| `fifoThroughputScope` | `string` | — | Throughput quota scope: `"Topic"` or `"MessageGroup"`. Only valid when `fifoTopic` is `true`. |
| `displayName` | `string` | — | Human-readable name shown in the AWS console and as the SMS "from" label. Max 256 characters for Standard topics. |
| `kmsKeyId` | `StringValueOrRef` | — | Customer-managed KMS key ID or ARN for server-side encryption. Can reference AwsKmsKey via `valueFrom`. |
| `policy` | `Struct` | — | IAM access policy document controlling which principals can Publish, Subscribe, etc. Expressed as a JSON structure in YAML. |
| `deliveryPolicy` | `string` | — | HTTP/S delivery retry policy JSON. Controls retry backoff, max retries, and throttle behavior. |
| `tracingConfig` | `string` | AWS: PassThrough | X-Ray tracing mode: `"Active"` or `"PassThrough"`. |
| `signatureVersion` | `int32` | AWS: 1 | Message signature hash: `1` (SHA1) or `2` (SHA256). SHA256 recommended for new topics. |
| `subscriptions` | `AwsSnsTopicSubscription[]` | `[]` | Inline subscriptions delivered with the topic. |
| `subscriptions[].filterPolicy` | `Struct` | — | Message filter selecting which messages this subscription receives. JSON structure in YAML. |
| `subscriptions[].filterPolicyScope` | `string` | `MessageAttributes` | Filter evaluation target: `"MessageAttributes"` or `"MessageBody"`. Requires `filterPolicy` to be set. |
| `subscriptions[].rawMessageDelivery` | `bool` | `false` | Deliver the raw message without the SNS JSON envelope. Supported for SQS, HTTP/S, and Firehose protocols. |
| `subscriptions[].redriveConfig.deadLetterTargetArn` | `StringValueOrRef` | — | SQS queue ARN for failed delivery attempts. Must be in the same account and region. Can reference AwsSqsQueue via `valueFrom`. |
| `subscriptions[].subscriptionRoleArn` | `StringValueOrRef` | — | IAM role ARN granting SNS permission to write to a Firehose delivery stream. Required when protocol is `firehose`. Can reference AwsIamRole via `valueFrom`. |

## Examples

### Fan-Out to SQS with Filtering

Route order events to separate queues based on message attributes:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: order-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSnsTopic.order-events
spec:
  signatureVersion: 2
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

### FIFO Topic with Deduplication

Ordered, exactly-once delivery for payment processing:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: payment-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSnsTopic.payment-events
spec:
  fifoTopic: true
  contentBasedDeduplication: true
  fifoThroughputScope: MessageGroup
  displayName: Payments
  subscriptions:
    - name: ledger
      protocol: sqs
      endpoint:
        valueFrom:
          kind: AwsSqsQueue
          name: ledger-queue
          fieldPath: status.outputs.queue_arn
      rawMessageDelivery: true
```

### Encrypted Topic with Lambda and DLQ

KMS-encrypted topic with a Lambda subscriber and a dead letter queue for failed deliveries:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: audit-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSnsTopic.audit-events
spec:
  displayName: Audit Events
  signatureVersion: 2
  tracingConfig: Active
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: audit-key
      fieldPath: status.outputs.key_arn
  subscriptions:
    - name: processor
      protocol: lambda
      endpoint:
        valueFrom:
          kind: AwsLambda
          name: audit-processor
          fieldPath: status.outputs.function_arn
      filterPolicy:
        severity:
          - critical
          - high
      filterPolicyScope: MessageAttributes
      redriveConfig:
        deadLetterTargetArn:
          valueFrom:
            kind: AwsSqsQueue
            name: audit-dlq
            fieldPath: status.outputs.queue_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `topic_arn` | `string` | ARN of the SNS topic — used in IAM policies, EventBridge targets, and cross-service references |
| `topic_name` | `string` | Name of the topic (includes `.fifo` suffix for FIFO topics) |
| `subscription_arns` | `map<string, string>` | Map of subscription name to subscription ARN — reference specific subscriptions downstream via `status.outputs.subscription_arns.{name}` |

## Related Components

- [AwsSqsQueue](/docs/catalog/aws/awssqsqueue) — common subscription target for fan-out and decoupling patterns
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides a customer-managed encryption key for topic encryption
- [AwsLambda](/docs/catalog/aws/awslambda) — subscribe Lambda functions for serverless event processing
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — provides roles for Firehose subscriptions and cross-account access
- [AwsEventBridgeBus](/docs/catalog/aws/awseventbridgebus) — routes EventBridge events to SNS topics as targets
