# AwsStepFunction Examples

## 1. Minimal Pass State (Hello World)

The simplest possible state machine with a single Pass state.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: hello-world
  org: my-org
  env: dev
  id: hello-world-dev
spec:
  region: us-east-1
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

## 2. Single Lambda Task

A state machine that invokes a single Lambda function.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: process-order
  org: my-org
  env: staging
  id: process-order-staging
spec:
  region: us-east-1
  type: STANDARD
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: sfn-exec-role
      fieldPath: status.outputs.role_arn
  definition:
    StartAt: ProcessOrder
    States:
      ProcessOrder:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:process-order
        End: true
```

## 3. Multi-Step Pipeline with Error Handling

A workflow with sequential steps, retry logic, and error catching.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: order-pipeline
  org: my-org
  env: prod
  id: order-pipeline-prod
spec:
  region: us-east-1
  type: STANDARD
  description: Multi-step order processing with retries and error handling
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
        Resource: arn:aws:lambda:us-east-1:123456789012:function:validate
        Next: ProcessPayment
        Catch:
          - ErrorEquals: ["ValidationError"]
            Next: RejectOrder
      ProcessPayment:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:charge
        Retry:
          - ErrorEquals: ["States.TaskFailed"]
            IntervalSeconds: 3
            MaxAttempts: 3
            BackoffRate: 2
        Next: FulfillOrder
        Catch:
          - ErrorEquals: ["States.ALL"]
            Next: RefundAndNotify
      FulfillOrder:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:fulfill
        End: true
      RejectOrder:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:reject
        End: true
      RefundAndNotify:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:refund
        End: true
```

## 4. EXPRESS Workflow for Event Processing

A high-throughput Express workflow for processing events from EventBridge.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: event-processor
  org: my-org
  env: prod
  id: event-processor-prod
spec:
  region: us-east-1
  type: EXPRESS
  description: High-throughput event processing pipeline
  roleArn:
    value: arn:aws:iam::123456789012:role/ExpressWorkflowRole
  definition:
    StartAt: EnrichEvent
    States:
      EnrichEvent:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:enrich
        Next: RouteEvent
      RouteEvent:
        Type: Choice
        Choices:
          - Variable: "$.eventType"
            StringEquals: order
            Next: ProcessOrder
          - Variable: "$.eventType"
            StringEquals: notification
            Next: SendNotification
        Default: ArchiveEvent
      ProcessOrder:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:process-order
        End: true
      SendNotification:
        Type: Task
        Resource: arn:aws:states:::sns:publish
        Parameters:
          TopicArn: arn:aws:sns:us-east-1:123456789012:notifications
          Message.$: "$.message"
        End: true
      ArchiveEvent:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:archive
        End: true
  logging:
    level: ERROR
    logDestination:
      value: arn:aws:logs:us-east-1:123456789012:log-group:/aws/stepfunctions/event-processor
```

## 5. Full Production Configuration

A production-ready workflow with all observability and security features.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: data-pipeline
  org: my-org
  env: prod
  id: data-pipeline-prod
spec:
  region: us-east-1
  type: STANDARD
  description: Production data pipeline with full observability and encryption
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: pipeline-exec-role
      fieldPath: status.outputs.role_arn
  definition:
    StartAt: ExtractData
    States:
      ExtractData:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:extract
        Next: TransformData
        Retry:
          - ErrorEquals: ["States.TaskFailed"]
            IntervalSeconds: 5
            MaxAttempts: 3
            BackoffRate: 2
      TransformData:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:transform
        Next: LoadData
      LoadData:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:load
        End: true
  tracingEnabled: true
  logging:
    level: ALL
    includeExecutionData: true
    logDestination:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: sfn-logs
        fieldPath: status.outputs.log_group_arn
  encryption:
    kmsKeyId:
      valueFrom:
        kind: AwsKmsKey
        name: platform-key
        fieldPath: status.outputs.key_arn
    kmsDataKeyReusePeriodSeconds: 600
```

## 6. Parallel Processing

A workflow that processes multiple branches in parallel.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: parallel-processor
  org: my-org
  env: dev
  id: parallel-processor-dev
spec:
  region: us-east-1
  roleArn:
    value: arn:aws:iam::123456789012:role/StepFunctionsExecRole
  definition:
    StartAt: FanOut
    States:
      FanOut:
        Type: Parallel
        Branches:
          - StartAt: SendEmail
            States:
              SendEmail:
                Type: Task
                Resource: arn:aws:lambda:us-east-1:123456789012:function:send-email
                End: true
          - StartAt: SendSMS
            States:
              SendSMS:
                Type: Task
                Resource: arn:aws:lambda:us-east-1:123456789012:function:send-sms
                End: true
          - StartAt: UpdateDatabase
            States:
              UpdateDatabase:
                Type: Task
                Resource: arn:aws:lambda:us-east-1:123456789012:function:update-db
                End: true
        Next: Aggregate
      Aggregate:
        Type: Pass
        End: true
```

## 7. Infra Chart Reference Pattern (valueFrom)

This example shows how AwsStepFunction references other resources in an infra chart context, using `valueFrom` to wire dependencies.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsStepFunction
metadata:
  name: "{{ values.env }}-workflow"
  org: "{{ values.org }}"
  env: "{{ values.env }}"
  id: "{{ values.env }}-workflow"
spec:
  region: us-east-1
  type: STANDARD
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: "{{ values.env }}-sfn-role"
      fieldPath: status.outputs.role_arn
  definition:
    StartAt: InvokeLambda
    States:
      InvokeLambda:
        Type: Task
        Resource: arn:aws:lambda:us-east-1:123456789012:function:{{ values.function_name }}
        End: true
  tracingEnabled: true
  logging:
    level: ERROR
    logDestination:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: "{{ values.env }}-sfn-logs"
        fieldPath: status.outputs.log_group_arn
```
