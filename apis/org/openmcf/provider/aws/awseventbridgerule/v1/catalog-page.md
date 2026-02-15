# AWS EventBridge Rule

Deploys an AWS EventBridge rule with bundled targets for event-driven routing or scheduled execution. Rules match incoming events by pattern or fire on a schedule, then route matched events to one or more targets (Lambda, SQS, SNS, Step Functions, and more) with optional input transformation, retry policies, and dead letter queues.

## What Gets Created

When you deploy an AwsEventBridgeRule resource, OpenMCF provisions:

- **EventBridge Rule** — a rule with event pattern matching or schedule expression, attached to the specified event bus
- **EventBridge Targets** — one target resource per entry in `targets`, linked to the rule with optional input transformation, retry policy, and DLQ

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An event bus** if targeting a custom bus — use AwsEventBridgeBus to create one, or omit `eventBusName` to use the default AWS bus
- **Target resources** must exist — Lambda functions, SQS queues, SNS topics, etc. that the rule routes events to
- **An IAM role** if targeting Step Functions, ECS, Kinesis, or Batch — EventBridge needs a role to invoke these targets
- **An SQS queue** if using dead letter queues — for catching events that fail delivery

## Quick Start

Create a file `rule.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: hourly-cleanup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEventBridgeRule.hourly-cleanup
spec:
  description: Trigger cleanup function every hour
  scheduleExpression: "rate(1 hour)"
  targets:
    - name: cleanup-function
      arn:
        value: arn:aws:lambda:us-east-1:123456789012:function:cleanup
```

Deploy:

```shell
openmcf apply -f rule.yaml
```

This creates a scheduled rule on the default event bus that triggers a Lambda function every hour.

## Configuration Reference

### Rule Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `eventBusName` | `StringValueOrRef` | `"default"` | Event bus name. Can reference AwsEventBridgeBus via `valueFrom`. |
| `description` | `string` | — | Human-readable description (max 512 chars). |
| `eventPattern` | `Struct` | — | JSON event pattern for matching. Mutually exclusive with `scheduleExpression`. |
| `scheduleExpression` | `string` | — | Cron or rate expression. Mutually exclusive with `eventPattern`. |
| `state` | `string` | `"ENABLED"` | `"ENABLED"` or `"DISABLED"`. |

### Target Configuration

| Field | Type | Description |
|-------|------|-------------|
| `targets[].name` | `string` | Target name (max 64 chars, `[0-9A-Za-z_.-]+`). Required. |
| `targets[].arn` | `StringValueOrRef` | Target resource ARN. Required. |
| `targets[].roleArn` | `StringValueOrRef` | IAM role for target invocation. Can reference AwsIamRole via `valueFrom`. |
| `targets[].input` | `string` | Constant JSON input (max 8192 chars). |
| `targets[].inputPath` | `string` | JSONPath extraction (max 256 chars). |
| `targets[].inputTransformer` | `object` | Template-based transformation. |
| `targets[].deadLetterConfig.arn` | `StringValueOrRef` | SQS DLQ ARN. Can reference AwsSqsQueue via `valueFrom`. |
| `targets[].retryPolicy.maximumEventAgeInSeconds` | `int32` | Max retry duration (60-86400). |
| `targets[].retryPolicy.maximumRetryAttempts` | `int32` | Max retry count (0-185). |
| `targets[].sqsConfig.messageGroupId` | `string` | FIFO SQS message group ID. |

## Examples

### Event Pattern with Fan-Out

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: order-router
spec:
  eventBusName:
    valueFrom:
      kind: AwsEventBridgeBus
      name: order-events
  description: Route order events to processor and analytics
  eventPattern:
    source:
      - com.myapp.orders
    detail-type:
      - OrderCreated
  targets:
    - name: processor
      arn:
        valueFrom:
          kind: AwsLambda
          name: order-processor
          fieldPath: status.outputs.function_arn
      deadLetterConfig:
        arn:
          valueFrom:
            kind: AwsSqsQueue
            name: order-dlq
            fieldPath: status.outputs.queue_arn
    - name: analytics
      arn:
        valueFrom:
          kind: AwsSqsQueue
          name: order-analytics
          fieldPath: status.outputs.queue_arn
      inputPath: "$.detail"
```

### Scheduled Cron Job

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: nightly-batch
spec:
  description: Run batch processing at 2 AM UTC daily
  scheduleExpression: "cron(0 2 * * ? *)"
  targets:
    - name: batch-processor
      arn:
        valueFrom:
          kind: AwsLambda
          name: batch-processor
          fieldPath: status.outputs.function_arn
      input: '{"mode": "full", "dryRun": false}'
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `rule_arn` | `string` | Rule ARN — used in IAM policies and monitoring |
| `rule_name` | `string` | Rule name — used in EventBridge API calls |

## Related Components

- [AwsEventBridgeBus](/docs/catalog/aws/awseventbridgebus) — create a custom event bus for this rule to attach to
- [AwsLambda](/docs/catalog/aws/awslambda) — common target for event processing and scheduled tasks
- [AwsSqsQueue](/docs/catalog/aws/awssqsqueue) — message queue target and dead letter queue
- [AwsSnsTopic](/docs/catalog/aws/awssnstopic) — fan-out notification target
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — IAM role for targets requiring assumed-role invocation
