# AwsStepFunction

Deploy and manage an AWS Step Functions state machine using Planton.

## Overview

Step Functions orchestrates distributed workflows by coordinating AWS services (Lambda, SQS, SNS, DynamoDB, ECS, and more) into serverless state machines defined in Amazon States Language (ASL). This component creates a single state machine with optional logging, tracing, and encryption.

## When to Use

- **Workflow orchestration**: Coordinate multiple Lambda functions, SQS queues, or other AWS services into a reliable, visual workflow.
- **Event-driven pipelines**: Process events through multi-step pipelines with error handling and retry logic.
- **Long-running processes**: Use STANDARD type for workflows up to 1 year with full execution history.
- **High-volume processing**: Use EXPRESS type for short-duration, high-throughput event processing.

## State Machine Types

| Feature | STANDARD | EXPRESS |
|---------|----------|---------|
| Max duration | 1 year | 5 minutes |
| Execution semantics | Exactly-once | At-most-once |
| Execution history | Full (viewable in console) | CloudWatch Logs only |
| Pricing | Per state transition | Per execution + duration |
| Use case | Long-running, auditable | High-volume, short-lived |

## Prerequisites

- An IAM execution role with a trust policy for `states.amazonaws.com` and policies granting access to all services invoked by the workflow.
- (Optional) A CloudWatch Log Group for execution logging.
- (Optional) A KMS key for customer-managed encryption.

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsStepFunction
metadata:
  name: hello-workflow
  org: my-org
  env: dev
  id: hello-workflow-dev
spec:
  roleArn:
    value: arn:aws:iam::123456789012:role/StepFunctionsExecRole
  definition:
    StartAt: Hello
    States:
      Hello:
        Type: Pass
        Result: Hello, World!
        End: true
```

## Production-Ready Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsStepFunction
metadata:
  name: order-processor
  org: my-org
  env: prod
  id: order-processor-prod
spec:
  type: STANDARD
  publish: true
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: sfn-exec-role
      fieldPath: status.outputs.role_arn
  definition:
    StartAt: ValidateOrder
    States:
      ValidateOrder:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:validate-order
        Next: ProcessPayment
        Catch:
          - ErrorEquals: ["States.ALL"]
            Next: HandleError
      ProcessPayment:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:process-payment
        Next: FulfillOrder
        Retry:
          - ErrorEquals: ["States.TaskFailed"]
            IntervalSeconds: 3
            MaxAttempts: 3
            BackoffRate: 2
      FulfillOrder:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:fulfill-order
        End: true
      HandleError:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:handle-error
        End: true
  tracingEnabled: true
  logging:
    level: ALL
    includeExecutionData: true
    logDestination:
      value: arn:aws:logs:us-east-1:123456789012:log-group:/aws/stepfunctions/order-processor
  encryption:
    kmsKeyId:
      valueFrom:
        kind: AwsKmsKey
        name: platform-key
        fieldPath: status.outputs.key_arn
    kmsDataKeyReusePeriodSeconds: 600
```

## Spec Reference

### Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | No | `STANDARD` or `EXPRESS`. Defaults to `STANDARD`. Cannot be changed after creation. |
| `definition` | Struct | Yes | ASL workflow definition as native YAML. Serialized to JSON by the IaC module. |
| `roleArn` | StringValueOrRef | Yes | IAM execution role ARN. Must trust `states.amazonaws.com`. |
| `publish` | bool | No | Publish an immutable version on create and on every configuration change. The latest version ARN is exported in outputs. Default: false. |
| `aliases` | AwsStepFunctionAlias[] | No | Named aliases routing 100% of traffic to the version this deployment published (requires `publish: true`). Alias ARNs are exported keyed by name. |
| `tracingEnabled` | optional bool | No | Enable AWS X-Ray tracing. Tri-state: unset keeps the AWS default (off); an explicit `false` is what turns tracing OFF on a machine that had it on. |
| `logging` | AwsStepFunctionLoggingConfig | No | Execution history logging configuration. Keep the block with `level: OFF` to explicitly turn logging off on a machine that had it on -- removing the block reverts nothing. |
| `encryption` | AwsStepFunctionEncryptionConfig | No | Customer-managed KMS encryption. One-way in practice: removing the block does not revert to AWS-owned keys. |

### AwsStepFunctionAlias

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Alias name (1-80 chars of `[0-9A-Za-z_-]`); keys the provider resource. |
| `description` | string | No | Human-readable description (up to 256 chars). |

### AwsStepFunctionLoggingConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `level` | string | Yes (when block present) | `ALL`, `ERROR`, `FATAL`, or `OFF`. |
| `includeExecutionData` | bool | No | Include input/output data in logs. Default: false. |
| `logDestination` | StringValueOrRef | Conditional | CloudWatch Log Group ARN. Required when level is not `OFF`. The `:*` suffix is auto-appended. |

### AwsStepFunctionEncryptionConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `kmsKeyId` | StringValueOrRef | Yes (when block present) | Customer-managed KMS key ARN. |
| `kmsDataKeyReusePeriodSeconds` | int32 | No | Data key reuse period. Range: 60-900. Default: 300 (AWS default). |

## Outputs

| Output | Description |
|--------|-------------|
| `state_machine_arn` | ARN of the state machine. Used by EventBridge targets, API Gateway integrations, and IAM policies. |
| `state_machine_name` | Name of the state machine. Useful for dashboards and monitoring. |
| `state_machine_version_arn` | ARN of the most recently published version (empty unless `publish` is true). Pin consumers here for immutable snapshots. |
| `revision_id` | Revision identifier of the current definition; changes on every definition or configuration update. |
| `status` | Lifecycle status reported by AWS (e.g. `ACTIVE`). |
| `creation_date` | RFC3339 timestamp of state machine creation. |

## Infra Chart Role

Step Functions serves as the orchestration layer in event-driven and serverless API infra charts. It coordinates Lambda functions, SQS queues, SNS topics, and other AWS services into reliable, visual workflows with built-in error handling and retry logic.

## Deliberately Omitted

- **Description**: The `CreateStateMachine` API has no description input (the AWS console derives one from the definition's `Comment` field), so a spec field would be silently dropped. Document the workflow in the ASL `Comment` instead.
- **Aliases** (`aws_sfn_alias`): Weighted traffic routing between published versions -- a deployment-orchestration surface that composes later through the exported `state_machine_version_arn`. Deferred until pulled by real demand.
- **Activities** (`aws_sfn_activity`): The legacy worker-polling integration pattern; modern workflows use direct service integrations or Lambda. Deferred.
- **Name prefix**: Planton derives resource names from metadata.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
