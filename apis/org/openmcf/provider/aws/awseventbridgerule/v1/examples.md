# AwsEventBridgeRule examples

## Scheduled Lambda (Cron Job)

A rule that triggers a Lambda function every hour. The simplest and most common EventBridge pattern — replacing traditional cron jobs with serverless scheduling.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: hourly-cleanup
  labels:
    app: maintenance
spec:
  description: Trigger cleanup function every hour
  scheduleExpression: "rate(1 hour)"
  targets:
    - name: cleanup-function
      arn:
        valueFrom:
          kind: AwsLambda
          name: data-cleanup
          fieldPath: status.outputs.function_arn
```

## EC2 State Change to SQS

A rule that matches EC2 instance state changes and routes them to an SQS queue for asynchronous processing. Demonstrates event pattern matching with a structured filter.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: ec2-state-monitor
spec:
  description: Route EC2 state changes to processing queue
  eventPattern:
    source:
      - aws.ec2
    detail-type:
      - "EC2 Instance State-change Notification"
    detail:
      state:
        - running
        - stopped
        - terminated
  targets:
    - name: state-queue
      arn:
        valueFrom:
          kind: AwsSqsQueue
          name: ec2-events
          fieldPath: status.outputs.queue_arn
      deadLetterConfig:
        arn:
          valueFrom:
            kind: AwsSqsQueue
            name: ec2-events-dlq
            fieldPath: status.outputs.queue_arn
```

## Custom Bus Rule with Input Transformer

A rule on a custom event bus that transforms order events before delivering them to a Lambda function. Demonstrates the input transformer pattern.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: order-processor
spec:
  eventBusName:
    valueFrom:
      kind: AwsEventBridgeBus
      name: order-events
      fieldPath: status.outputs.bus_name
  description: Process new order events
  eventPattern:
    source:
      - com.myapp.orders
    detail-type:
      - OrderCreated
  targets:
    - name: order-lambda
      arn:
        valueFrom:
          kind: AwsLambda
          name: order-processor
          fieldPath: status.outputs.function_arn
      inputTransformer:
        inputPaths:
          orderId: "$.detail.order_id"
          customerId: "$.detail.customer_id"
          total: "$.detail.total"
        inputTemplate: '{"orderId": <orderId>, "customerId": <customerId>, "total": <total>, "source": "eventbridge"}'
```

## Fan-Out to Multiple Targets with DLQ

A rule that routes events to both a Lambda function and an SQS queue simultaneously. Each target has its own retry policy and dead letter queue for independent reliability.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: payment-router
spec:
  eventBusName:
    valueFrom:
      kind: AwsEventBridgeBus
      name: payment-events
      fieldPath: status.outputs.bus_name
  description: Route payment events to processing and analytics
  eventPattern:
    source:
      - com.myapp.payments
    detail-type:
      - PaymentCompleted
  targets:
    - name: payment-processor
      arn:
        valueFrom:
          kind: AwsLambda
          name: payment-processor
          fieldPath: status.outputs.function_arn
      retryPolicy:
        maximumEventAgeInSeconds: 3600
        maximumRetryAttempts: 10
      deadLetterConfig:
        arn:
          valueFrom:
            kind: AwsSqsQueue
            name: payment-processor-dlq
            fieldPath: status.outputs.queue_arn
    - name: analytics-queue
      arn:
        valueFrom:
          kind: AwsSqsQueue
          name: payment-analytics
          fieldPath: status.outputs.queue_arn
      inputPath: "$.detail"
```

## Step Functions Orchestration

A rule that triggers a Step Functions state machine, requiring an IAM role for invocation.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: order-workflow
spec:
  description: Start order fulfillment workflow on new orders
  eventPattern:
    source:
      - com.myapp.orders
    detail-type:
      - OrderCreated
  targets:
    - name: fulfillment-workflow
      arn:
        value: arn:aws:states:us-east-1:123456789012:stateMachine:order-fulfillment
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: eb-invoke-sfn
          fieldPath: status.outputs.role_arn
```

## Daily Cron with Constant Input

A scheduled rule that fires at midnight UTC with a constant JSON payload.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: daily-report
spec:
  description: Generate daily report at midnight UTC
  scheduleExpression: "cron(0 0 * * ? *)"
  targets:
    - name: report-generator
      arn:
        valueFrom:
          kind: AwsLambda
          name: report-generator
          fieldPath: status.outputs.function_arn
      input: '{"reportType": "daily", "format": "pdf"}'
```

## SQS FIFO Target with Message Group

A rule targeting a FIFO SQS queue with a message group ID for ordered processing.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: ordered-events
spec:
  description: Route events to FIFO queue with ordering
  eventPattern:
    source:
      - com.myapp.transactions
  targets:
    - name: fifo-queue
      arn:
        valueFrom:
          kind: AwsSqsQueue
          name: transaction-queue
          fieldPath: status.outputs.queue_arn
      sqsConfig:
        messageGroupId: "transactions"
```

## CLI flows

Validate manifest:
```bash
openmcf validate --manifest ./manifest.yaml | cat
```

Pulumi deploy:
```bash
openmcf pulumi update --manifest ./manifest.yaml --stack myorg/infra/dev --module-dir ./apis/org/openmcf/provider/aws/awseventbridgerule/v1/iac/pulumi | cat
```

Terraform deploy:
```bash
openmcf tofu apply --manifest ./manifest.yaml --auto-approve | cat
```

> Note: Provider credentials are supplied via stack input, not in the spec.
