---
title: "EventBridge Rule"
description: "EventBridge Rule deployment documentation"
icon: "package"
order: 100
componentName: "awseventbridgerule"
---

# AWS EventBridge Rule

Deploys an AWS EventBridge rule with bundled targets for event-driven routing or scheduled execution. Rules match incoming events by pattern or fire on a cron/rate schedule, then route matched events to one or more targets (Lambda, SQS, SNS, Step Functions, etc.) with optional input transformation, retry policies, and dead letter queues.

## What Gets Created

When you deploy an AwsEventBridgeRule resource, OpenMCF provisions:

- **EventBridge Rule** — an `aws_cloudwatch_event_rule` attached to the specified event bus (or the default bus), configured with either an event pattern or a schedule expression
- **EventBridge Targets** — one `aws_cloudwatch_event_target` per entry in `targets`, linked to the rule with optional input transformation, retry policy, and dead letter queue configuration

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An event bus** if targeting a custom bus — use AwsEventBridgeBus to create one, or omit `eventBusName` to use the default AWS bus
- **Target resources** must exist — Lambda functions, SQS queues, SNS topics, Step Functions state machines, etc.
- **An IAM role** if targeting Step Functions, ECS, Kinesis, Batch, CodeBuild, or CodePipeline — EventBridge needs a role to invoke these targets
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
  region: us-east-1
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

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the EventBridge rule will be created (e.g., `us-east-1`). | Required; non-empty |
| `eventPattern` | `object` | JSON event pattern for matching incoming events. Expressed as a structured object in YAML. Mutually exclusive with `scheduleExpression`. | Exactly one of `eventPattern` or `scheduleExpression` must be set |
| `scheduleExpression` | `string` | Cron or rate expression for time-based triggering (e.g., `rate(5 minutes)`, `cron(0 12 * * ? *)`). Mutually exclusive with `eventPattern`. | Exactly one of `eventPattern` or `scheduleExpression` must be set; max 256 chars |
| `targets` | `AwsEventBridgeTarget[]` | Targets to invoke when the rule matches an event. | Minimum 1 item required; AWS limit 5 per rule |
| `targets[].name` | `string` | Unique name for the target within the rule. Used as the EventBridge `target_id`. | Required; max 64 chars; pattern `[0-9A-Za-z_.-]+` |
| `targets[].arn` | `StringValueOrRef` | ARN of the target resource (Lambda, SQS, SNS, Step Functions, etc.). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `eventBusName` | `StringValueOrRef` | `"default"` | Event bus to attach the rule to. Changing this forces rule replacement. Can reference AwsEventBridgeBus via `valueFrom`. |
| `description` | `string` | — | Human-readable description. Max 512 characters. |
| `state` | `string` | `"ENABLED"` | Rule state. Valid values: `"ENABLED"`, `"DISABLED"`. |
| `targets[].roleArn` | `StringValueOrRef` | — | IAM role ARN for EventBridge to assume when invoking this target. Required for Step Functions, ECS, Kinesis, Batch, CodeBuild, and CodePipeline targets. Can reference AwsIamRole via `valueFrom`. |
| `targets[].input` | `string` | — | Constant JSON input passed to the target instead of the matched event. Max 8192 chars. Mutually exclusive with `inputPath` and `inputTransformer`. |
| `targets[].inputPath` | `string` | — | JSONPath expression to extract a portion of the matched event (e.g., `$.detail`). Max 256 chars. Mutually exclusive with `input` and `inputTransformer`. |
| `targets[].inputTransformer.inputPaths` | `map<string, string>` | — | Map of variable names to JSONPath expressions that extract values from the event. Max 100 entries. |
| `targets[].inputTransformer.inputTemplate` | `string` | — | Template producing the final input. References variables from `inputPaths` with `<variable>`. Max 8192 chars. Required when `inputTransformer` is set. |
| `targets[].deadLetterConfig.arn` | `StringValueOrRef` | — | SQS queue ARN for events that fail delivery after all retries. Queue must be in the same account and region. Can reference AwsSqsQueue via `valueFrom`. |
| `targets[].retryPolicy.maximumEventAgeInSeconds` | `int32` | `86400` | Max time in seconds EventBridge keeps retrying delivery. Range: 60–86400. |
| `targets[].retryPolicy.maximumRetryAttempts` | `int32` | `185` | Max number of retry attempts. Range: 0–185. Set to 0 to send failures directly to DLQ. |
| `targets[].sqsConfig.messageGroupId` | `string` | — | Message group ID for FIFO SQS queues. Required when targeting a FIFO queue. |

## Examples

### Event Pattern with Fan-Out

Route order events from a custom bus to a Lambda processor and an SQS analytics queue:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: order-router
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEventBridgeRule.order-router
spec:
  region: us-east-1
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

### Scheduled Cron Job with Custom Retry

Run a batch processor at 2 AM UTC daily with a constant JSON payload and reduced retry window:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: nightly-batch
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEventBridgeRule.nightly-batch
spec:
  region: us-east-1
  description: Run batch processing at 2 AM UTC daily
  scheduleExpression: "cron(0 2 * * ? *)"
  targets:
    - name: batch-processor
      arn:
        value: arn:aws:lambda:us-east-1:123456789012:function:batch-processor
      input: '{"mode": "full", "dryRun": false}'
      retryPolicy:
        maximumEventAgeInSeconds: 3600
        maximumRetryAttempts: 3
      deadLetterConfig:
        arn:
          value: arn:aws:sqs:us-east-1:123456789012:batch-dlq
```

### Input Transformer with Step Functions Target

Match EC2 state-change events, reshape the payload, and invoke a Step Functions state machine:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: ec2-state-handler
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEventBridgeRule.ec2-state-handler
spec:
  region: us-east-1
  description: Handle EC2 instance state changes
  eventPattern:
    source:
      - aws.ec2
    detail-type:
      - EC2 Instance State-change Notification
    detail:
      state:
        - stopped
        - terminated
  targets:
    - name: remediation-workflow
      arn:
        value: arn:aws:states:us-east-1:123456789012:stateMachine:ec2-remediation
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: eventbridge-sfn-role
          fieldPath: status.outputs.role_arn
      inputTransformer:
        inputPaths:
          instance: "$.detail.instance-id"
          state: "$.detail.state"
          time: "$.time"
        inputTemplate: '{"instanceId": "<instance>", "newState": "<state>", "eventTime": "<time>"}'
```

### Disabled Rule for Staging

A rule created in disabled state for pre-production testing:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEventBridgeRule
metadata:
  name: staging-order-router
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsEventBridgeRule.staging-order-router
spec:
  region: us-east-1
  description: Order routing rule (disabled for staging validation)
  state: DISABLED
  eventPattern:
    source:
      - com.myapp.orders
    detail-type:
      - OrderCreated
      - OrderUpdated
  targets:
    - name: order-handler
      arn:
        value: arn:aws:lambda:us-east-1:123456789012:function:staging-order-handler
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `rule_arn` | `string` | ARN of the EventBridge rule, used in IAM policies and monitoring configurations |
| `rule_name` | `string` | Name of the EventBridge rule, used in EventBridge API calls |

## Related Components

- [AwsEventBridgeBus](/docs/catalog/aws/eventbridge-bus) — custom event bus for this rule to attach to
- [AwsLambda](/docs/catalog/aws/lambda) — common target for event processing and scheduled tasks
- [AwsSqsQueue](/docs/catalog/aws/sqs-queue) — message queue target and dead letter queue for failed deliveries
- [AwsSnsTopic](/docs/catalog/aws/sns-topic) — fan-out notification target
- [AwsIamRole](/docs/catalog/aws/iam-role) — IAM role for targets requiring assumed-role invocation (Step Functions, ECS, Kinesis, Batch)
