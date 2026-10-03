# AWS EventBridge Rule

Deploys an EventBridge rule with bundled targets that matches events by pattern or schedule and routes them to one or more downstream services. Each target independently supports input transformation, retry policies, dead letter queues, and service-typed parameters (SQS FIFO message groups, Kinesis partition keys, API-destination path/query/header values, Batch job submission, ECS RunTask launches with task tags, Redshift Data API statements, SSM Run Command dispatch, SageMaker pipeline executions, AppSync GraphQL operations). The event bus, invocation roles, target ARNs, and per-target dead letter queues all accept ValueFromRef wiring, so a rule composes with the resources on both sides of the routing edge in one InfraChart.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EventBridge Rule** -- an event rule attached to the specified bus (or the default bus), configured with either a JSON event pattern for event matching or a schedule expression (cron/rate) for time-based triggering
- **EventBridge Targets** -- one target resource per entry in the `targets` list, each wired to the rule with its own ARN, optional IAM role, input transformation, retry policy, and dead letter configuration
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A custom event bus** (optional) -- required when attaching the rule to a custom bus rather than the default AWS event bus. Provide the bus name directly or reference an AwsEventBridgeBus Infra Component via ValueFromRef.
- **Target resources** -- at least one target (Lambda function, SQS queue, SNS topic, Step Functions state machine, etc.) must exist. Provide the target ARN directly or via ValueFromRef.
- **An IAM role** (optional) -- required for targets where EventBridge must assume a role (Step Functions, ECS, Kinesis, Batch). Not needed for Lambda, SQS, or SNS targets that use resource-based policies. Provide the ARN directly or reference an AwsIamRole Infra Component via ValueFromRef.
- **An SQS dead letter queue** (optional) -- required when configuring per-target dead letter queues. Provide the ARN directly or reference an AwsSqsQueue Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **AWS EventBridge Rule**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Scheduled Lambda** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEventBridgeRule
metadata:
  name: hourly-cleanup
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  scheduleExpression: "rate(1 hour)"
  targets:
    - name: cleanup-function
      arn:
        value: "arn:aws:lambda:us-west-2:123456789012:function:cleanup"
```

```shell
planton apply -f event-rule.yaml
```

This creates a schedule-based rule on the default event bus that triggers a Lambda function every hour. No input transformation, retry policy, or dead letter queue is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the rule to a custom event bus and target resources deployed in the same InfraPipeline:

```yaml
spec:
  eventBusName:
    valueFrom:
      kind: AwsEventBridgeBus
      name: order-events
      fieldPath: status.outputs.bus_name
  targets:
    - name: process-order
      arn:
        value: "arn:aws:lambda:us-west-2:123456789012:function:process-order"
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: eventbridge-invoke-role
          fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the event bus and IAM role first, then provisions the rule with the resolved values.

## Key Configuration

These are the most important decisions when configuring an EventBridge rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Event pattern vs. schedule** -- Use `eventPattern` to match events by structure (source, detail-type, detail fields) for reactive event routing. Use `scheduleExpression` with a rate or cron expression for time-based triggering. Exactly one must be set; they are mutually exclusive. Note: AWS supports schedule expressions only on the DEFAULT event bus -- a scheduled rule on a custom bus is rejected at deploy time.

**Target input transformation** -- By default, the full event is passed to each target. Use `input` for a constant JSON payload, `inputPath` to extract a specific portion (e.g., `$.detail`), or `inputTransformer` to reshape the event using JSONPath extraction and a template. The three options are mutually exclusive per target.

**Service-typed target parameters** -- A target invoking SQS FIFO, Kinesis, an API destination, Batch, ECS, Redshift, SSM Run Command, SageMaker Pipelines, or AppSync carries its service's parameter block (at most one per target): `sqsTarget.messageGroupId` for FIFO ordering, `kinesisTarget.partitionKeyPath` for shard routing, `httpTarget` path/query/header values, `batchTarget` naming the job definition and job each event submits (the target arn is the JOB QUEUE), `ecsTarget` describing the RunTask launch (the target arn is the CLUSTER; task definition, launch type XOR capacity-provider strategy, VPC networking, placement, and per-task `tags` live in the block), `redshiftTarget` running a Data API statement (the target arn is the CLUSTER; database plus `dbUser` or Secrets Manager credentials), `runCommandTargets` selecting instances by id or tag for an SSM document dispatch (the target arn is the DOCUMENT), `sagemakerPipelineTarget` starting a pipeline execution with parameters, and `appsyncTarget` invoking a GraphQL operation (the target arn is the API's endpoint ARN).

**Invocation identity** -- Targets on resource-based-policy services (Lambda, SQS, SNS) need no role. Step Functions, ECS, Kinesis, Batch, Redshift, SSM Run Command, SageMaker Pipelines, AppSync, and cross-account bus targets require a role EventBridge assumes: set the rule-level `roleArn` to govern the whole rule, or a per-target `roleArn` (which takes precedence).

**Retry policy and dead letter queue** -- By default, EventBridge retries failed deliveries for up to 24 hours with 185 attempts. Customize `retryPolicy` to shorten the window for latency-sensitive targets (0 retries sends failures straight to the DLQ). Configure `deadLetterConfig` on each target to capture events that exhaust all retries, preventing silent event loss.

**Rule state** -- Defaults to `ENABLED`. Set `state: DISABLED` to create the rule without activating it, useful for pre-deploying rules that should only activate during a rollout. `ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS` additionally matches read-only CloudTrail management events (Describe*/List* API activity) that ENABLED rules never receive.

**Teardown posture** -- AWS refuses to delete a rule that still has targets. `forceDestroy` defaults to false so an unexpectedly shared rule fails loudly instead of vanishing under an out-of-band consumer; the module manages its own targets' teardown ordering either way.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEventBridgeBus** (optional) | `eventBusName` | `status.outputs.bus_name` |
| **AwsIamRole** (optional) | `roleArn`, `targets[*].roleArn` | `status.outputs.role_arn` |
| **AwsSqsQueue** (optional) | `targets[*].deadLetterConfig.arn` | `status.outputs.queue_arn` |
| **AwsBatchJobDefinition** (optional) | `targets[*].batchTarget.jobDefinition` | `status.outputs.job_definition_arn` |
| **AwsEcsTaskDefinition** (optional) | `targets[*].ecsTarget.taskDefinitionArn` | `status.outputs.task_definition_arn` |
| **AwsSubnet** (optional) | `targets[*].ecsTarget.networkConfiguration.subnets` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `targets[*].ecsTarget.networkConfiguration.securityGroups` | `status.outputs.security_group_id` |

The target `arn` itself carries no default kind (the destination varies) — reference any resource's ARN output via `valueFrom`, e.g. an AwsLambda's `function_arn`, an AwsSqsQueue's `queue_arn`, an AwsKinesisStream's `stream_arn`, an AwsStepFunction's `state_machine_arn`, an AwsEcsCluster's `cluster_arn`, or an AwsBatchJobQueue's `job_queue_arn`.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `rule_arn` | Amazon Resource Name of the rule | IAM policies, monitoring and alerting configurations |
| `rule_name` | Name of the rule | EventBridge API calls, operational dashboards |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Scheduled Lambda** -- Rate-based schedule rule triggering a Lambda function on a recurring interval. The serverless replacement for cron jobs -- no servers to manage, no crontab files to maintain. Start from the **Scheduled Lambda** preset.

**Event pattern to SQS** -- Pattern-based rule matching AWS service events (EC2 state changes, S3 operations) and routing them to an SQS queue with a dead letter queue for failed deliveries. The core event-driven routing pattern for decoupling producers and consumers. Start from the **Event Pattern to SQS** preset.

**Multi-target fanout** -- Single rule routing matched events to multiple targets simultaneously (Lambda for real-time processing, SQS for audit trail, SNS for notifications). Each target has independent retry and DLQ configuration. Start from the **Multi-Target Fan-Out** preset.

## Works With

- [**AWS EventBridge Bus**](/infra-catalog/aws-event-bridge-bus) -- provides the custom event bus to attach the rule to
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the invocation role for targets that require assumed credentials
- [**AWS SQS Queue**](/infra-catalog/aws-sqs-queue) -- provides dead letter queues for targets that need failed event capture