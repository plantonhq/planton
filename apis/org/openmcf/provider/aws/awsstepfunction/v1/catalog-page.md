# AWS Step Functions

Deploys an AWS Step Functions state machine with an Amazon States Language (ASL) workflow definition, configurable execution type (STANDARD or EXPRESS), optional CloudWatch Logs logging, X-Ray tracing, and customer-managed KMS encryption.

## What Gets Created

When you deploy an AwsStepFunction resource, OpenMCF provisions:

- **Step Functions State Machine** — an `aws_sfn_state_machine` resource with the provided ASL definition serialized to JSON, assigned the specified IAM execution role, and tagged with OpenMCF resource metadata
- **Tracing Configuration** — enabled only when `tracingEnabled` is `true`, sends trace data to AWS X-Ray for visualizing request flows
- **Logging Configuration** — configured only when `logging.level` is set to a value other than `OFF`, delivers execution history events to the specified CloudWatch Logs log group (the module automatically appends `:*` to the log group ARN if missing)
- **Encryption Configuration** — configured only when the `encryption` block is provided, uses a customer-managed KMS key for encrypting state machine data, execution history, and input/output payloads

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An IAM execution role** with a trust policy for `states.amazonaws.com` and policies granting access to all services invoked by the workflow (e.g., `lambda:InvokeFunction`, `sqs:SendMessage`, `sns:Publish`)
- **A CloudWatch Logs log group** if enabling execution logging
- **A customer-managed KMS key** if enabling encryption (must be a symmetric encryption key in the same region)
- **X-Ray permissions** on the execution role (`xray:PutTraceSegments`, `xray:PutTelemetryRecords`) if enabling tracing

## Quick Start

Create a file `step-function.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: my-step-function
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsStepFunction.my-step-function
spec:
  roleArn: arn:aws:iam::123456789012:role/step-functions-exec
  definition:
    StartAt: Hello
    States:
      Hello:
        Type: Pass
        Result: Hello, World!
        End: true
```

Deploy:

```shell
openmcf apply -f step-function.yaml
```

This creates a STANDARD state machine with a single Pass state that returns a static result.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `definition` | `object` | State machine definition in Amazon States Language (ASL). Write as native YAML; the module serializes it to JSON for the AWS API. ASL key casing (`StartAt`, `States`, `Type`, `Resource`) is preserved. | Required. Max 1 MB after JSON serialization. |
| `roleArn` | `string` | IAM execution role ARN. The role must trust `states.amazonaws.com` and grant access to all services the workflow invokes. Can reference an AwsIamRole resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | `string` | `"STANDARD"` | State machine type. `STANDARD` for long-running workflows (up to 1 year, exactly-once). `EXPRESS` for high-volume short-duration workflows (up to 5 minutes, at-most-once). Cannot be changed after creation (forces replacement). Must be `STANDARD` or `EXPRESS`. |
| `description` | `string` | `""` | Free-form description visible in the AWS Console. |
| `tracingEnabled` | `bool` | `false` | Enables AWS X-Ray tracing. Requires `xray:PutTraceSegments` and `xray:PutTelemetryRecords` permissions on the execution role. |
| `logging.level` | `string` | `"OFF"` | Logging level for execution history events. `ALL` logs every event type, `ERROR` logs errors only, `FATAL` logs fatal errors only, `OFF` disables logging. |
| `logging.includeExecutionData` | `bool` | `false` | When `true`, includes full JSON payloads passed between states in log entries. Increases log volume and may expose sensitive data. |
| `logging.logDestination` | `string` | — | CloudWatch Logs log group ARN for log delivery. Required when `logging.level` is not `OFF`. The module appends `:*` automatically if not present. |
| `encryption.kmsKeyId` | `string` | — | Customer-managed KMS key ARN for encrypting state machine data, execution history, and payloads. Must be a symmetric key in the same region. Can reference an AwsKmsKey resource via `valueFrom`. Required when the `encryption` block is present. |
| `encryption.kmsDataKeyReusePeriodSeconds` | `int` | `300` (AWS default) | Duration in seconds Step Functions reuses a data key before calling `GenerateDataKey` again. Range: 60–900. |

## Examples

### Standard Workflow with Lambda Integration

A workflow that invokes a Lambda function and handles success or failure:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: order-processor
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsStepFunction.order-processor
spec:
  roleArn: arn:aws:iam::123456789012:role/step-functions-exec
  definition:
    StartAt: ProcessOrder
    States:
      ProcessOrder:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:process-order
        Next: OrderComplete
        Catch:
          - ErrorEquals:
              - States.ALL
            Next: OrderFailed
      OrderComplete:
        Type: Succeed
      OrderFailed:
        Type: Fail
        Error: OrderProcessingError
        Cause: The order could not be processed.
```

### Express Workflow for Event Processing

A high-throughput EXPRESS state machine for processing events with logging enabled:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: event-ingest
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsStepFunction.event-ingest
spec:
  type: EXPRESS
  roleArn: arn:aws:iam::123456789012:role/step-functions-express-exec
  tracingEnabled: true
  logging:
    level: ERROR
    includeExecutionData: false
    logDestination: arn:aws:logs:us-east-1:123456789012:log-group:/aws/stepfunctions/event-ingest
  definition:
    StartAt: ValidateEvent
    States:
      ValidateEvent:
        Type: Choice
        Choices:
          - Variable: $.eventType
            StringEquals: order
            Next: RouteToOrders
        Default: DropEvent
      RouteToOrders:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:route-orders
        End: true
      DropEvent:
        Type: Succeed
```

### Encrypted Workflow with Full Observability

Production configuration with customer-managed KMS encryption, X-Ray tracing, and verbose logging:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: payment-workflow
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsStepFunction.payment-workflow
spec:
  roleArn: arn:aws:iam::123456789012:role/payment-sfn-exec
  description: Processes payment transactions with PCI-compliant encryption.
  tracingEnabled: true
  logging:
    level: ALL
    includeExecutionData: true
    logDestination: arn:aws:logs:us-east-1:123456789012:log-group:/aws/stepfunctions/payment-workflow
  encryption:
    kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/abcd-1234-efgh-5678
    kmsDataKeyReusePeriodSeconds: 60
  definition:
    StartAt: AuthorizePayment
    States:
      AuthorizePayment:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:authorize-payment
        Next: CapturePayment
      CapturePayment:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:capture-payment
        End: true
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding ARNs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: ref-step-function
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsStepFunction.ref-step-function
spec:
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: sfn-exec-role
      field: status.outputs.role_arn
  encryption:
    kmsKeyId:
      valueFrom:
        kind: AwsKmsKey
        name: sfn-encryption-key
        field: status.outputs.key_arn
    kmsDataKeyReusePeriodSeconds: 300
  definition:
    StartAt: InvokeService
    States:
      InvokeService:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:my-service
        End: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `state_machine_arn` | `string` | ARN of the created Step Functions state machine, used for invoking executions and cross-service references (EventBridge targets, API Gateway integrations, IAM policies) |
| `state_machine_name` | `string` | Name of the state machine, useful for dashboards, monitoring, and human-readable log references |

## Related Components

- [AwsIamRole](/docs/catalog/aws/awsiamrole) — provides the execution role assumed by the state machine
- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides the customer-managed encryption key for state machine data
- [AwsLambda](/docs/catalog/aws/awslambda) — common Task state target invoked by Step Functions workflows
