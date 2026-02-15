# AwsSnsTopic examples

## Minimal Standard Topic

A standard topic with no encryption and no subscriptions. The simplest possible configuration.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: order-events
  labels:
    app: shop
spec: {}
```

## Standard Topic with KMS Encryption

A standard topic encrypted with a customer-managed KMS key.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: audit-events
spec:
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: audit-key
      fieldPath: status.outputs.key_arn
  signatureVersion: 2
```

## Fan-Out to SQS Queues with Filtering

A standard topic with two SQS subscriptions. Each subscription uses a filter policy so different queues receive different event types.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: order-notifications
spec:
  subscriptions:
    - name: fulfillment-queue
      protocol: sqs
      endpoint:
        valueFrom:
          kind: AwsSqsQueue
          name: fulfillment-events
          fieldPath: status.outputs.queue_arn
      filterPolicy:
        event_type:
          - order_placed
          - order_shipped
      filterPolicyScope: MessageAttributes
      rawMessageDelivery: true
    - name: billing-queue
      protocol: sqs
      endpoint:
        valueFrom:
          kind: AwsSqsQueue
          name: billing-events
          fieldPath: status.outputs.queue_arn
      filterPolicy:
        event_type:
          - payment_received
          - refund_issued
      filterPolicyScope: MessageAttributes
      rawMessageDelivery: true
```

## Lambda Subscription with DLQ

A topic that invokes a Lambda function for processing. Failed deliveries are routed to a dead letter SQS queue.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: image-processing
spec:
  tracingConfig: Active
  subscriptions:
    - name: resize-processor
      protocol: lambda
      endpoint:
        valueFrom:
          kind: AwsLambda
          name: image-resizer
          fieldPath: status.outputs.function_arn
      redriveConfig:
        deadLetterTargetArn:
          valueFrom:
            kind: AwsSqsQueue
            name: image-processing-dlq
            fieldPath: status.outputs.queue_arn
```

## FIFO Topic with SQS FIFO Subscription

A FIFO topic with content-based deduplication and high-throughput mode, delivering to a FIFO SQS queue.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: payment-events
spec:
  fifoTopic: true
  contentBasedDeduplication: true
  fifoThroughputScope: MessageGroup
  subscriptions:
    - name: payment-processor
      protocol: sqs
      endpoint:
        valueFrom:
          kind: AwsSqsQueue
          name: payment-processing
          fieldPath: status.outputs.queue_arn
      rawMessageDelivery: true
```

## Multi-Protocol Subscriptions

A topic with subscriptions spanning multiple protocols: SQS for processing, email for alerts, and HTTPS for a webhook.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: system-alerts
spec:
  displayName: SystemAlerts
  subscriptions:
    - name: alert-queue
      protocol: sqs
      endpoint:
        valueFrom:
          kind: AwsSqsQueue
          name: alert-handler
          fieldPath: status.outputs.queue_arn
      rawMessageDelivery: true
    - name: ops-email
      protocol: email
      endpoint:
        value: ops-team@example.com
    - name: pagerduty-webhook
      protocol: https
      endpoint:
        value: https://events.pagerduty.com/integration/abc123/enqueue

```

## Topic with IAM Access Policy

A topic with an access policy that allows EventBridge to publish events and restricts subscription to a specific account.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSnsTopic
metadata:
  name: cross-service-events
spec:
  policy:
    Version: "2012-10-17"
    Statement:
      - Sid: AllowEventBridgePublish
        Effect: Allow
        Principal:
          Service: events.amazonaws.com
        Action: sns:Publish
        Resource: "*"
      - Sid: AllowAccountSubscribe
        Effect: Allow
        Principal:
          AWS: "arn:aws:iam::123456789012:root"
        Action:
          - sns:Subscribe
          - sns:Receive
        Resource: "*"
```

## CLI flows

Validate manifest:
```bash
openmcf validate --manifest ./manifest.yaml | cat
```

Pulumi deploy:
```bash
openmcf pulumi update --manifest ./manifest.yaml --stack myorg/infra/dev --module-dir ./apis/org/openmcf/provider/aws/awssnstopic/v1/iac/pulumi | cat
```

Terraform deploy:
```bash
openmcf tofu apply --manifest ./manifest.yaml --auto-approve | cat
```

> Note: Provider credentials are supplied via stack input, not in the spec.
