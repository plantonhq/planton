# AWS SQS Queue

Deploys an AWS SQS queue — Standard or FIFO — with optional server-side encryption, dead letter queue routing, IAM access policies, and delivery tuning. The module automatically appends the `.fifo` suffix to the queue name when `fifoQueue` is true and the metadata name does not already include it.

## What Gets Created

When you deploy an AwsSqsQueue resource, OpenMCF provisions:

- **SQS Queue** — a Standard or FIFO `aws_sqs_queue` resource with the specified delivery settings, encryption configuration, and access policy
- **Redrive Policy** — configured on the queue only when `deadLetterConfig` is provided, routes messages to a dead letter queue after the specified number of receive attempts
- **AWS Tags** — resource, organization, environment, resource kind, and resource ID tags applied to the queue

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A dead letter queue** (another AwsSqsQueue) if using dead letter queue routing — both queues must be the same type (both Standard or both FIFO) and reside in the same account and region
- **A KMS key** if using customer-managed encryption instead of SQS-managed SSE

## Quick Start

Create a file `queue.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: my-queue
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSqsQueue.my-queue
spec:
  region: us-east-1
  sqsManagedSseEnabled: true
```

Deploy:

```shell
openmcf apply -f queue.yaml
```

This creates a Standard SQS queue with SQS-managed encryption and all other settings at AWS defaults.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | The AWS region where the SQS queue will be created. | Required |

All other configuration is optional with AWS defaults.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `fifoQueue` | `bool` | `false` | Create a FIFO queue with exactly-once processing and strict ordering. Cannot be changed after creation. FIFO queue names must end with `.fifo`; the module appends this suffix automatically. |
| `visibilityTimeoutSeconds` | `int32` | AWS: 30 | Time in seconds a received message is hidden from subsequent receive requests (0–43200). |
| `messageRetentionSeconds` | `int32` | AWS: 345600 | Duration in seconds SQS retains a message before deleting it (60–1209600). Leave at 0 for AWS default of 4 days. |
| `maxMessageSizeBytes` | `int32` | AWS: 262144 | Maximum message body size in bytes (1024–1048576). Leave at 0 for AWS default of 256 KB. |
| `delaySeconds` | `int32` | AWS: 0 | Delay in seconds before newly sent messages become visible (0–900). |
| `receiveWaitTimeSeconds` | `int32` | AWS: 0 | Long polling wait time in seconds for the ReceiveMessage API (0–20). Values greater than 0 enable long polling. |
| `contentBasedDeduplication` | `bool` | `false` | Use SHA-256 hash of the message body as the deduplication ID. Only valid when `fifoQueue` is `true`. |
| `deduplicationScope` | `string` | — | Deduplication scope: `"messageGroup"` or `"queue"`. Only valid when `fifoQueue` is `true`. |
| `fifoThroughputLimit` | `string` | — | Throughput limit: `"perMessageGroupId"` (high throughput mode) or `"perQueue"`. Only valid when `fifoQueue` is `true`. |
| `deadLetterConfig.targetArn` | `StringValueOrRef` | — | ARN of the dead letter queue. Can reference AwsSqsQueue via `valueFrom`. Required when `deadLetterConfig` is set. Both queues must be the same type. |
| `deadLetterConfig.maxReceiveCount` | `int32` | — | Number of receive attempts before routing to the DLQ (1–1000). Required when `deadLetterConfig` is set. |
| `kmsKeyId` | `StringValueOrRef` | — | Customer-managed KMS key ID or ARN for server-side encryption. Can reference AwsKmsKey via `valueFrom`. Mutually exclusive with `sqsManagedSseEnabled`. |
| `kmsDataKeyReusePeriodSeconds` | `int32` | AWS: 300 | Duration in seconds SQS reuses a data key before calling KMS again (60–86400). Only relevant when `kmsKeyId` is set. |
| `sqsManagedSseEnabled` | `bool` | `false` | Enable SQS-managed server-side encryption (SSE-SQS). Mutually exclusive with `kmsKeyId`. |
| `policy` | `Struct` | — | IAM access policy document controlling which principals can perform actions on this queue. Expressed as a JSON structure in YAML. |

## Examples

### FIFO Queue with Dead Letter Queue

A FIFO queue for payment processing with content-based deduplication and a dead letter queue for failed messages:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: payment-events
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSqsQueue.payment-events
spec:
  region: us-east-1
  fifoQueue: true
  contentBasedDeduplication: true
  sqsManagedSseEnabled: true
  visibilityTimeoutSeconds: 120
  deadLetterConfig:
    targetArn:
      valueFrom:
        kind: AwsSqsQueue
        name: payment-events-dlq
        fieldPath: status.outputs.queue_arn
    maxReceiveCount: 3
```

### Long-Polling Standard Queue with Extended Retention

A Standard queue configured for long polling to reduce empty responses, with a 7-day retention period:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: task-queue
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSqsQueue.task-queue
spec:
  region: us-east-1
  sqsManagedSseEnabled: true
  receiveWaitTimeSeconds: 20
  visibilityTimeoutSeconds: 60
  messageRetentionSeconds: 604800
```

### KMS-Encrypted Queue with Cross-Account Access Policy

A queue encrypted with a customer-managed KMS key and an IAM policy granting an SNS topic permission to publish messages:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: notifications-queue
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSqsQueue.notifications-queue
spec:
  region: us-east-1
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: sqs-encryption-key
      fieldPath: status.outputs.key_arn
  kmsDataKeyReusePeriodSeconds: 600
  visibilityTimeoutSeconds: 300
  messageRetentionSeconds: 1209600
  policy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: sns.amazonaws.com
        Action: sqs:SendMessage
        Resource: "*"
        Condition:
          ArnEquals:
            aws:SourceArn: arn:aws:sns:us-east-1:123456789012:order-events
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `queue_url` | `string` | SQS queue URL — primary identifier for API calls (SendMessage, ReceiveMessage, DeleteMessage) |
| `queue_arn` | `string` | Queue ARN — used in IAM policies, cross-service permissions, and dead letter queue target references |
| `queue_name` | `string` | Queue name (includes `.fifo` suffix for FIFO queues) |

## Related Components

- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides a customer-managed encryption key for SSE-KMS
- [AwsLambda](/docs/catalog/aws/awslambda) — event source mapping from SQS to Lambda for serverless processing
- [AwsSnsTopic](/docs/catalog/aws/awssnstopic) — fan-out messages to multiple SQS queues via SNS subscriptions
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — provides roles for services that need to interact with this queue
