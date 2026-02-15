# AWS Step Functions

Orchestrate serverless workflows with AWS Step Functions — coordinate Lambda, SQS, SNS, and 200+ AWS services into visual, auditable state machines.

## What It Does

Creates an AWS Step Functions state machine from an Amazon States Language (ASL) definition. Supports both STANDARD workflows (long-running, exactly-once) and EXPRESS workflows (high-volume, short-duration). Includes optional CloudWatch logging, X-Ray tracing, and customer-managed KMS encryption.

## Key Features

- **Visual workflows**: Define multi-step workflows as code, visualize execution in the AWS console
- **Built-in error handling**: Retry logic, catch blocks, and timeout management
- **Two execution modes**: STANDARD for durable workflows, EXPRESS for high-throughput
- **Native YAML authoring**: Write ASL definitions as YAML — serialized to JSON automatically
- **Full observability**: CloudWatch logging with execution data, X-Ray tracing
- **Cross-resource references**: Wire dependencies via `valueFrom` for IAM roles, KMS keys, and log groups

## Common Use Cases

- Order processing pipelines with validation, payment, and fulfillment steps
- ETL/ELT data pipelines coordinating extraction, transformation, and loading
- Event-driven architectures routing events through multi-step processing
- Approval workflows with human-in-the-loop Wait states
- Microservice orchestration coordinating multiple service calls

## Quick Start

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: my-workflow
  org: my-org
  env: dev
  id: my-workflow-dev
spec:
  roleArn:
    value: arn:aws:iam::123456789012:role/StepFunctionsExecRole
  definition:
    StartAt: ProcessData
    States:
      ProcessData:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:process
        End: true
```

## Related Resources

- [AwsLambda](../awslambda/v1/) — Functions invoked by state machine tasks
- [AwsSqsQueue](../awssqsqueue/v1/) — Queues for decoupled processing and dead letter handling
- [AwsSnsTopic](../awssnstopic/v1/) — Pub/sub notifications from workflow steps
- [AwsEventBridgeRule](../awseventbridgerule/v1/) — Event rules that trigger state machine executions
- [AwsHttpApiGateway](../awshttpapigateway/v1/) — HTTP APIs that start workflow executions
- [AwsIamRole](../awsiamrole/v1/) — Execution roles for state machine permissions
