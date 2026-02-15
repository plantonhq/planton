# AWS SQS Queue

Deploys an AWS SQS queue — Standard or FIFO — with server-side encryption, dead letter queue routing, IAM access policies, and fine-grained delivery tuning. SQS is the foundational message queuing service for decoupling microservices, buffering requests, and building event-driven architectures on AWS.

## What Gets Created

When you deploy an AwsSqsQueue resource, OpenMCF provisions:

- **SQS Queue** — a Standard or FIFO queue with the specified delivery settings (visibility timeout, retention, delay, long polling), encryption configuration, and access policy
- **Redrive Policy** — created only when `deadLetterConfig` is provided, configures automatic routing of failed messages to a dead letter queue after the specified number of receive attempts

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A dead letter queue** (another AwsSqsQueue) if you plan to use dead letter queue routing — both queues must be the same type and in the same account/region

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
  sqsManagedSseEnabled: true
```

Deploy:

```shell
openmcf apply -f queue.yaml
```

This creates a Standard SQS queue with SQS-managed encryption and all other AWS defaults.

## Configuration Reference

### Queue Type

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `fifoQueue` | `bool` | `false` | Create a FIFO queue (exactly-once, ordered). Cannot be changed after creation. |

### Delivery Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `visibilityTimeoutSeconds` | `int32` | AWS: 30 | Time a received message is hidden (0–43200). |
| `messageRetentionSeconds` | `int32` | AWS: 345600 | How long SQS retains messages (60–1209600). |
| `maxMessageSizeBytes` | `int32` | AWS: 262144 | Maximum message size in bytes (1024–1048576). |
| `delaySeconds` | `int32` | AWS: 0 | Delay before new messages are visible (0–900). |
| `receiveWaitTimeSeconds` | `int32` | AWS: 0 | Long polling wait time (0–20). |

### FIFO Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `contentBasedDeduplication` | `bool` | `false` | SHA-256 body hash as dedup ID. FIFO only. |
| `deduplicationScope` | `string` | — | `"messageGroup"` or `"queue"`. FIFO only. |
| `fifoThroughputLimit` | `string` | — | `"perMessageGroupId"` or `"perQueue"`. FIFO only. |

### Dead Letter Queue

| Field | Type | Description |
|-------|------|-------------|
| `deadLetterConfig.targetArn` | `StringValueOrRef` | DLQ queue ARN. Can reference another AwsSqsQueue via `valueFrom`. Required when `deadLetterConfig` is set. |
| `deadLetterConfig.maxReceiveCount` | `int32` | Receive attempts before routing to DLQ (1–1000). Required. |

### Encryption

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sqsManagedSseEnabled` | `bool` | `false` | SQS-managed encryption. Mutually exclusive with `kmsKeyId`. |
| `kmsKeyId` | `StringValueOrRef` | — | Customer-managed KMS key. Can reference AwsKmsKey via `valueFrom`. |
| `kmsDataKeyReusePeriodSeconds` | `int32` | AWS: 300 | KMS data key reuse window (60–86400). Only with `kmsKeyId`. |

### Access Control

| Field | Type | Description |
|-------|------|-------------|
| `policy` | `Struct` | IAM access policy document (JSON structure in YAML). |

## Examples

### FIFO Queue with DLQ

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: payment-events
spec:
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

### Long-Polling Standard Queue

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: task-queue
spec:
  sqsManagedSseEnabled: true
  receiveWaitTimeSeconds: 20
  visibilityTimeoutSeconds: 60
  messageRetentionSeconds: 604800
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `queue_url` | `string` | SQS queue URL — primary identifier for API calls |
| `queue_arn` | `string` | Queue ARN — used in IAM policies and cross-service references |
| `queue_name` | `string` | Queue name (includes `.fifo` suffix for FIFO queues) |

## Related Components

- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides a customer-managed encryption key for SSE-KMS
- [AwsLambda](/docs/catalog/aws/awslambda) — event source mapping from SQS to Lambda for serverless processing
- [AwsSnsTopic](/docs/catalog/aws/awssnstopic) — fan-out messages to multiple SQS queues via SNS subscriptions
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — provides roles for services that need to interact with this queue
