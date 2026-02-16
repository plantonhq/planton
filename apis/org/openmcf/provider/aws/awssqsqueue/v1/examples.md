# AwsSqsQueue examples

## Minimal Standard Queue

A standard queue with SQS-managed encryption and all other AWS defaults.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: order-events
  labels:
    app: shop
spec:
  sqsManagedSseEnabled: true
```

## Standard Queue with Dead Letter Queue

A standard queue that routes failed messages to a DLQ after 5 receive attempts. Both queues use SSE-SQS encryption.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: order-events-dlq
spec:
  sqsManagedSseEnabled: true
  messageRetentionSeconds: 1209600  # 14 days — keep DLQ messages longer for investigation
---
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: order-events
spec:
  sqsManagedSseEnabled: true
  visibilityTimeoutSeconds: 60
  receiveWaitTimeSeconds: 20
  deadLetterConfig:
    targetArn:
      valueFrom:
        kind: AwsSqsQueue
        name: order-events-dlq
        fieldPath: status.outputs.queue_arn
    maxReceiveCount: 5
```

## FIFO Queue with Content-Based Deduplication

A FIFO queue that uses message body hashing for deduplication. Ideal for financial transactions or any workflow requiring exactly-once ordered processing.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: payment-processing
spec:
  fifoQueue: true
  contentBasedDeduplication: true
  deduplicationScope: messageGroup
  fifoThroughputLimit: perMessageGroupId
  sqsManagedSseEnabled: true
  visibilityTimeoutSeconds: 120
```

## Production Queue with KMS Encryption

A standard queue encrypted with a customer-managed KMS key, long polling enabled, and custom retention.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: audit-log-ingest
spec:
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: audit-key
      fieldPath: status.outputs.key_arn
  kmsDataKeyReusePeriodSeconds: 3600
  messageRetentionSeconds: 1209600
  maxMessageSizeBytes: 262144
  receiveWaitTimeSeconds: 20
  visibilityTimeoutSeconds: 300
```

## Queue with IAM Access Policy (SNS Fan-Out)

A standard queue that allows an SNS topic to publish messages to it.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSqsQueue
metadata:
  name: notification-handler
spec:
  sqsManagedSseEnabled: true
  receiveWaitTimeSeconds: 10
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
            aws:SourceArn: arn:aws:sns:us-east-1:123456789012:order-notifications
```

## CLI flows

Validate manifest:
```bash
openmcf validate --manifest ./manifest.yaml | cat
```

Pulumi deploy:
```bash
openmcf pulumi update --manifest ./manifest.yaml --stack myorg/infra/dev --module-dir ./apis/org/openmcf/provider/aws/awssqsqueue/v1/iac/pulumi | cat
```

Terraform deploy:
```bash
openmcf tofu apply --manifest ./manifest.yaml --auto-approve | cat
```

> Note: Provider credentials are supplied via stack input, not in the spec.
