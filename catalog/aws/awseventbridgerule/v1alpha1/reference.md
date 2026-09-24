# AwsEventBridgeRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsEventBridgeRuleSpec defines the desired configuration for an AWS
EventBridge rule with bundled targets.

An EventBridge rule matches incoming events on an event bus and routes them
to one or more targets for processing. Rules can match events by pattern
(JSON event matching) or by schedule (cron/rate expressions).

Targets are bundled with the rule because a rule without targets is
functionally useless — it matches events but does nothing with them.
Each target can independently configure input transformation, retry
policies, and dead letter queues.

Notes:
- The rule name is derived from `metadata.name`.
- Exactly one of `event_pattern` or `schedule_expression` must be set.
- Targets are created as separate Terraform/Pulumi resources but are
  managed as a single unit with the rule.
- Credentials, region, and deployment workflow live outside this spec in
  stack inputs.

## Example

```yaml
# Canonical validated example: a scheduled operations rule fanning out to
# the full typed-target spread -- Redshift Data API, SSM Run Command,
# SageMaker Pipelines, AppSync, and ECS RunTask with task tags (AWS caps a
# rule at 5 targets). Each service-typed target carries the role
# EventBridge assumes to invoke it.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEventBridgeRule
metadata:
  name: test-rule
  org: test-org
  env: dev
  id: test-rule-dev
spec:
  region: us-east-1
  description: Nightly operations fan-out for local development
  scheduleExpression: "rate(1 hour)"
  targets:
    - name: warehouse-refresh
      arn:
        value: arn:aws:redshift:us-east-1:123456789012:cluster:analytics-warehouse
      roleArn:
        value: arn:aws:iam::123456789012:role/events-redshift-data
      redshiftTarget:
        database: analytics
        dbUser: etl_maintenance
        sql: "REFRESH MATERIALIZED VIEW mv_hourly_orders;"
        statementName: hourly-mv-refresh
        withEvent: true
    - name: fleet-log-rotate
      arn:
        value: arn:aws:ssm:us-east-1::document/AWS-RunShellScript
      roleArn:
        value: arn:aws:iam::123456789012:role/events-run-command
      runCommandTargets:
        - key: tag:Environment
          values:
            - production
        - key: tag:Role
          values:
            - app-server
      input: '{"commands":["logrotate -f /etc/logrotate.conf"]}'
    - name: model-retrain
      arn:
        value: arn:aws:sagemaker:us-east-1:123456789012:pipeline/churn-retrain
      roleArn:
        value: arn:aws:iam::123456789012:role/events-sagemaker-pipeline
      sagemakerPipelineTarget:
        pipelineParameterList:
          - name: InputDataS3Uri
            value: s3://training-data/hourly/
    - name: cache-invalidate
      arn:
        value: arn:aws:appsync:us-east-1:123456789012:endpoints/graphql-api/abcdef123456
      roleArn:
        value: arn:aws:iam::123456789012:role/events-appsync-invoke
      appsyncTarget:
        graphqlOperation: "mutation InvalidateCache($ts:String){ invalidateCache(timestamp:$ts){ ok } }"
      inputTransformer:
        inputPaths:
          ts: "$.time"
        inputTemplate: '{"ts": <ts>}'
    - name: batch-compaction-task
      arn:
        value: arn:aws:ecs:us-east-1:123456789012:cluster/ops-cluster
      roleArn:
        value: arn:aws:iam::123456789012:role/events-ecs-run-task
      ecsTarget:
        taskDefinitionArn:
          value: arn:aws:ecs:us-east-1:123456789012:task-definition/compaction:5
        launchType: FARGATE
        networkConfiguration:
          subnets:
            - value: subnet-0a1b2c3d4e5f60718
          securityGroups:
            - value: sg-0a1b2c3d4e5f60718
        tags:
          team: data-platform
          cost-center: warehouse-ops
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.eventBusName` | `string \| valueFrom` |  |  | AwsEventBridgeBus (`status.outputs.bus_name`) |
| `spec.description` | `string` |  |  |  |
| `spec.eventPattern` | `object` |  |  |  |
| `spec.scheduleExpression` | `string` |  |  |  |
| `spec.state` | `string` |  |  |  |
| `spec.roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.targets` | `[]AwsEventBridgeTarget` |  |  |  |
| `spec.targets[].name` | `string` | yes |  |  |
| `spec.targets[].arn` | `string \| valueFrom` | yes |  |  |
| `spec.targets[].roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.targets[].input` | `string` |  |  |  |
| `spec.targets[].inputPath` | `string` |  |  |  |
| `spec.targets[].inputTransformer` | `AwsEventBridgeInputTransformer` |  |  |  |
| `spec.targets[].inputTransformer.inputPaths` | `map<string, string>` |  |  |  |
| `spec.targets[].inputTransformer.inputTemplate` | `string` | yes |  |  |
| `spec.targets[].deadLetterConfig` | `AwsEventBridgeTargetDeadLetterConfig` |  |  |  |
| `spec.targets[].deadLetterConfig.arn` | `string \| valueFrom` | yes |  | AwsSqsQueue (`status.outputs.queue_arn`) |
| `spec.targets[].retryPolicy` | `AwsEventBridgeTargetRetryPolicy` |  |  |  |
| `spec.targets[].retryPolicy.maximumEventAgeInSeconds` | `int32` |  |  |  |
| `spec.targets[].retryPolicy.maximumRetryAttempts` | `int32` |  |  |  |
| `spec.targets[].sqsTarget` | `AwsEventBridgeTargetSqsConfig` |  |  |  |
| `spec.targets[].sqsTarget.messageGroupId` | `string` |  |  |  |
| `spec.targets[].kinesisTarget` | `AwsEventBridgeTargetKinesisConfig` |  |  |  |
| `spec.targets[].kinesisTarget.partitionKeyPath` | `string` |  |  |  |
| `spec.targets[].httpTarget` | `AwsEventBridgeTargetHttpConfig` |  |  |  |
| `spec.targets[].httpTarget.pathParameterValues` | `[]string` |  |  |  |
| `spec.targets[].httpTarget.queryStringParameters` | `map<string, string>` |  |  |  |
| `spec.targets[].httpTarget.headerParameters` | `map<string, string>` |  |  |  |
| `spec.targets[].batchTarget` | `AwsEventBridgeTargetBatchConfig` |  |  |  |
| `spec.targets[].batchTarget.jobDefinition` | `string \| valueFrom` | yes |  | AwsBatchJobDefinition (`status.outputs.job_definition_arn`) |
| `spec.targets[].batchTarget.jobName` | `string` | yes |  |  |
| `spec.targets[].batchTarget.arraySize` | `int32` |  |  |  |
| `spec.targets[].batchTarget.jobAttempts` | `int32` |  |  |  |
| `spec.targets[].ecsTarget` | `AwsEventBridgeTargetEcsConfig` |  |  |  |
| `spec.targets[].ecsTarget.taskDefinitionArn` | `string \| valueFrom` | yes |  | AwsEcsTaskDefinition (`status.outputs.task_definition_arn`) |
| `spec.targets[].ecsTarget.taskCount` | `int32` |  |  |  |
| `spec.targets[].ecsTarget.launchType` | `string` |  |  |  |
| `spec.targets[].ecsTarget.platformVersion` | `string` |  |  |  |
| `spec.targets[].ecsTarget.group` | `string` |  |  |  |
| `spec.targets[].ecsTarget.capacityProviderStrategy` | `[]AwsEventBridgeEcsCapacityProviderStrategy` |  |  |  |
| `spec.targets[].ecsTarget.capacityProviderStrategy[].capacityProvider` | `string` | yes |  |  |
| `spec.targets[].ecsTarget.capacityProviderStrategy[].base` | `int32` |  |  |  |
| `spec.targets[].ecsTarget.capacityProviderStrategy[].weight` | `int32` |  |  |  |
| `spec.targets[].ecsTarget.networkConfiguration` | `AwsEventBridgeEcsNetworkConfiguration` |  |  |  |
| `spec.targets[].ecsTarget.networkConfiguration.subnets` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.targets[].ecsTarget.networkConfiguration.securityGroups` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.targets[].ecsTarget.networkConfiguration.assignPublicIp` | `bool` |  |  |  |
| `spec.targets[].ecsTarget.orderedPlacementStrategy` | `[]AwsEventBridgeEcsPlacementStrategy` |  |  |  |
| `spec.targets[].ecsTarget.orderedPlacementStrategy[].type` | `string` | yes |  |  |
| `spec.targets[].ecsTarget.orderedPlacementStrategy[].field` | `string` |  |  |  |
| `spec.targets[].ecsTarget.placementConstraints` | `[]AwsEventBridgeEcsPlacementConstraint` |  |  |  |
| `spec.targets[].ecsTarget.placementConstraints[].type` | `string` | yes |  |  |
| `spec.targets[].ecsTarget.placementConstraints[].expression` | `string` |  |  |  |
| `spec.targets[].ecsTarget.propagateTags` | `string` |  |  |  |
| `spec.targets[].ecsTarget.enableEcsManagedTags` | `bool` |  |  |  |
| `spec.targets[].ecsTarget.enableExecuteCommand` | `bool` |  |  |  |
| `spec.targets[].ecsTarget.tags` | `map<string, string>` |  |  |  |
| `spec.targets[].redshiftTarget` | `AwsEventBridgeTargetRedshiftConfig` |  |  |  |
| `spec.targets[].redshiftTarget.database` | `string` | yes |  |  |
| `spec.targets[].redshiftTarget.dbUser` | `string` |  |  |  |
| `spec.targets[].redshiftTarget.secretsManagerArn` | `string` |  |  |  |
| `spec.targets[].redshiftTarget.sql` | `string` |  |  |  |
| `spec.targets[].redshiftTarget.statementName` | `string` |  |  |  |
| `spec.targets[].redshiftTarget.withEvent` | `bool` |  |  |  |
| `spec.targets[].runCommandTargets` | `[]AwsEventBridgeTargetRunCommandSelector` |  |  |  |
| `spec.targets[].runCommandTargets[].key` | `string` | yes |  |  |
| `spec.targets[].runCommandTargets[].values` | `[]string` | yes |  |  |
| `spec.targets[].sagemakerPipelineTarget` | `AwsEventBridgeTargetSagemakerPipelineConfig` |  |  |  |
| `spec.targets[].sagemakerPipelineTarget.pipelineParameterList` | `[]AwsEventBridgeSagemakerPipelineParameter` |  |  |  |
| `spec.targets[].sagemakerPipelineTarget.pipelineParameterList[].name` | `string` | yes |  |  |
| `spec.targets[].sagemakerPipelineTarget.pipelineParameterList[].value` | `string` | yes |  |  |
| `spec.targets[].appsyncTarget` | `AwsEventBridgeTargetAppsyncConfig` |  |  |  |
| `spec.targets[].appsyncTarget.graphqlOperation` | `string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.eventBusName

`string | valueFrom`

Name of the event bus to attach this rule to. Defaults to "default" (the
built-in AWS event bus) when not specified. Can reference an
AwsEventBridgeBus resource via `valueFrom`.

Changing this field forces rule replacement (delete + recreate).

Constraint (AWS, not validated here because reference values resolve at
deploy time): schedule_expression is only supported on the DEFAULT bus —
a scheduled rule on a custom bus is rejected by the AWS API. Custom
buses take event-pattern rules only.

- references: AwsEventBridgeBus (`status.outputs.bus_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEventBridgeBus, name: <that resource's name>, fieldPath: status.outputs.bus_name}} -- a bare string does not parse

### spec.description

`string`

Human-readable description of the rule. Maximum 512 characters.

- rule: {"string":{"maxLen":"512"}}

### spec.eventPattern

`object`

JSON event pattern that this rule matches against. Events that match the
pattern are routed to the rule's targets. Expressed as a structured object
in YAML — the IaC module serializes it to JSON. AWS caps the serialized
pattern at 4096 characters.

Mutually exclusive with `schedule_expression`.

Example patterns:
  source: ["aws.ec2"]
  detail-type: ["EC2 Instance State-change Notification"]
  detail:
    state: ["running", "stopped"]

### spec.scheduleExpression

`string`

Schedule expression for time-based rule triggering. Supports cron and rate
expressions.

Mutually exclusive with `event_pattern`.

Examples:
  "rate(5 minutes)"        — fire every 5 minutes
  "rate(1 hour)"           — fire every hour
  "cron(0 12 * * ? *)"     — fire at noon UTC every day
  "cron(0/15 * * * ? *)"   — fire every 15 minutes

- rule: {"string":{"maxLen":"256"}}

### spec.state

`string`

Rule state. Controls whether the rule is actively matching events.
Valid values:
- "ENABLED"  — the rule matches events (IaC default when not set).
- "DISABLED" — the rule exists but matches nothing.
- "ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS" — additionally matches
  read-only CloudTrail management events (AWS API activity like Describe*
  calls), which ENABLED rules never receive. Only meaningful for rules
  whose pattern matches CloudTrail management events.

### spec.roleArn

`string | valueFrom`

IAM role EventBridge assumes when invoking this rule's targets. Used when
a single role should govern the whole rule (targets can also carry their
own per-target `role_arn`, which takes precedence for that target).
Required for schedule rules whose targets need role-based invocation.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.forceDestroy

`bool`

Force rule deletion even when targets are still attached (AWS refuses to
delete a rule that has targets unless forced). Keep the default (false)
so an unexpectedly shared rule fails loudly instead of vanishing under
an out-of-band consumer; the module manages its own targets' teardown
ordering either way.

### spec.targets

`[]AwsEventBridgeTarget`

Targets to invoke when the rule matches an event. At least one target is
required. Each target specifies a destination (Lambda, SQS, SNS, Step
Functions, etc.), optional input transformation, retry policy, dead
letter queue, and an optional service-typed parameter block (SQS,
Kinesis, HTTP/API-destination, Batch, ECS RunTask, Redshift Data API,
SSM Run Command, SageMaker Pipelines, AppSync).

AWS limits: maximum 5 targets per rule (enforced here so the quota
fails at validate time instead of at PutTargets).

- rule: {"repeated":{"maxItems":"5"}}
- rule: only one of input, input_path, or input_transformer may be set
- rule: at most one of sqs_target, kinesis_target, http_target, batch_target, ecs_target, redshift_target, run_command_targets, sagemaker_pipeline_target, or appsync_target may be set

### spec.targets[].name

`string` · required

User-assigned name for this target. Used as the Pulumi resource name
and as the `target_id` in EventBridge. Must be unique within the rule's
targets. Maximum 64 characters, alphanumeric plus hyphen, underscore,
and period.

- rule: {"required":true,"string":{"maxLen":"64","pattern":"^[0-9A-Za-z_.-]+$"}}

### spec.targets[].arn

`string | valueFrom` · required

ARN of the target resource. This is the AWS resource that processes
matched events. Common targets include Lambda functions, SQS queues,
SNS topics, Step Functions state machines, and CloudWatch Log Groups.

No `default_kind` is set because the target resource type varies
(Lambda, SQS, SNS, etc.). Use `valueFrom` to reference specific
Planton resources.

Containment-exempt: a target is where matched events are DELIVERED,
never where the rule lives -- the rule belongs to its event bus. A
target that is itself a container (another event bus, for a
cross-bus forward) would otherwise pull the rule inside the bus it
forwards to; on a diagram the rule stays on its own bus with a line
to each target.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.targets[].roleArn

`string | valueFrom`

IAM role ARN for EventBridge to assume when invoking this target.
Required for targets where EventBridge needs to assume a role:
Step Functions, ECS, Kinesis, Batch, CodeBuild, CodePipeline, and
cross-account event buses.

Not needed for targets that use resource-based policies:
Lambda (function policy), SQS (queue policy), SNS (topic policy).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.targets[].input

`string`

Constant JSON input to pass to the target instead of the matched event.
Maximum 8192 characters. Mutually exclusive with `input_path` and
`input_transformer`.

- rule: {"string":{"maxLen":"8192"}}

### spec.targets[].inputPath

`string`

JSONPath expression to extract a portion of the matched event and pass
to the target. Maximum 256 characters. Mutually exclusive with `input`
and `input_transformer`.

Example: "$.detail" extracts the detail object from the event.

- rule: {"string":{"maxLen":"256"}}

### spec.targets[].inputTransformer

`AwsEventBridgeInputTransformer`

Input transformer to reshape the matched event before passing to the
target. Mutually exclusive with `input` and `input_path`.

- rule: input_paths keys must not start with 'AWS' -- that prefix is reserved by EventBridge

### spec.targets[].inputTransformer.inputPaths

`map<string, string>`

Map of variable names to JSONPath expressions that extract values from
the matched event. Keys become variables available in `input_template`.
Maximum 100 entries. Keys must not start with "AWS" (reserved).

Example:
  instance: "$.detail.instance-id"
  state: "$.detail.state"

- rule: {"map":{"maxPairs":"100"}}

### spec.targets[].inputTransformer.inputTemplate

`string` · required

Template that produces the final input for the target. References
variables from `input_paths` using angle brackets: <variable>.
Maximum 8192 characters.

Example: "Instance <instance> transitioned to <state>"

- rule: {"required":true,"string":{"maxLen":"8192"}}

### spec.targets[].deadLetterConfig

`AwsEventBridgeTargetDeadLetterConfig`

Dead letter queue for events that fail delivery to this target. When
EventBridge cannot deliver an event after all retry attempts, the event
is routed to the specified SQS queue for investigation.

### spec.targets[].deadLetterConfig.arn

`string | valueFrom` · required

ARN of the SQS queue to use as the dead letter queue. The queue must
exist in the same AWS account and region as the rule.

Accepts a direct ARN or a reference to an AwsSqsQueue resource.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

### spec.targets[].retryPolicy

`AwsEventBridgeTargetRetryPolicy`

Retry policy controlling how EventBridge retries failed deliveries to
this target. When not set, EventBridge uses the default policy: retry
for 24 hours with up to 185 attempts using exponential backoff.

### spec.targets[].retryPolicy.maximumEventAgeInSeconds

`int32` · optional (explicit presence)

Maximum time in seconds that EventBridge keeps retrying delivery.
Range: 60 to 86400 (1 minute to 24 hours). Absent: AWS default of
86400 (24 hours).

- rule: {"int32":{"lte":86400,"gte":60}}

### spec.targets[].retryPolicy.maximumRetryAttempts

`int32` · optional (explicit presence)

Maximum number of retry attempts. Range: 0 to 185. Absent: AWS default
of 185. Set to 0 to disable retries (event goes to the DLQ immediately
on failure).

- rule: {"int32":{"lte":185,"gte":0}}

### spec.targets[].sqsTarget

`AwsEventBridgeTargetSqsConfig`

SQS-specific parameters. Required when targeting a FIFO SQS queue
to specify the message group ID.

### spec.targets[].sqsTarget.messageGroupId

`string`

Message group ID for FIFO SQS queues. Required when the target is a
FIFO queue to ensure proper message ordering and deduplication.
Ignored for standard queues.

### spec.targets[].kinesisTarget

`AwsEventBridgeTargetKinesisConfig`

Kinesis-specific parameters. Controls how events map to stream shards.

### spec.targets[].kinesisTarget.partitionKeyPath

`string`

JSONPath expression extracting the partition key from the event (e.g.
"$.detail.customer_id"), determining which shard receives the record.
When unset, EventBridge uses the event ID — an even spread that
sacrifices per-entity ordering. Maximum 256 characters.

- rule: {"string":{"maxLen":"256"}}

### spec.targets[].httpTarget

`AwsEventBridgeTargetHttpConfig`

HTTP/API-destination parameters. Path, query, and header values applied
when the target arn is an EventBridge API destination.

### spec.targets[].httpTarget.pathParameterValues

`[]string`

Values substituted into the API destination's path wildcards ("*"), in
order of appearance in the endpoint URL.

### spec.targets[].httpTarget.queryStringParameters

`map<string, string>`

Query string parameters appended to the invocation URL.

### spec.targets[].httpTarget.headerParameters

`map<string, string>`

HTTP headers added to the request. Keys must be valid header names;
AWS reserves headers starting with "X-Amz" and "X-Amzn".

### spec.targets[].batchTarget

`AwsEventBridgeTargetBatchConfig`

AWS Batch parameters. Required when the target arn is a Batch job queue.

### spec.targets[].batchTarget.jobDefinition

`string | valueFrom` · required

The Batch job definition each matched event submits. Accepts a
reference to an AwsBatchJobDefinition resource (its revision-carrying
ARN output, so a new revision rolls the rule's submissions) or a
literal value -- a bare name or name:revision literal tracks the
name's latest ACTIVE revision on the Batch side instead of pinning
one.

- references: AwsBatchJobDefinition (`status.outputs.job_definition_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBatchJobDefinition, name: <that resource's name>, fieldPath: status.outputs.job_definition_arn}} -- a bare string does not parse

### spec.targets[].batchTarget.jobName

`string` · required

Name assigned to the submitted jobs (visible in the Batch console and
APIs). Maximum 128 characters.

- rule: {"required":true,"string":{"maxLen":"128"}}

### spec.targets[].batchTarget.arraySize

`int32`

Size of an array job. Leave at 0 for a regular (non-array) job.
Range when set: 2-10000.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":10000,"gte":2}}

### spec.targets[].batchTarget.jobAttempts

`int32`

Retry attempts for the submitted job, 1-10. Leave at 0 to use the job
definition's own retry strategy.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":10,"gte":1}}

### spec.targets[].ecsTarget

`AwsEventBridgeTargetEcsConfig`

ECS RunTask parameters. Required when the target arn is an ECS cluster —
the event launches a task from the referenced task definition.

- rule: launch_type must be 'EC2', 'FARGATE', or 'EXTERNAL' when set
- rule: launch_type and capacity_provider_strategy are mutually exclusive; choose one placement mechanism
- rule: platform_version is only valid with launch_type 'FARGATE'
- rule: propagate_tags must be 'TASK_DEFINITION' when set

### spec.targets[].ecsTarget.taskDefinitionArn

`string | valueFrom` · required

The task definition to launch. Accepts a reference to an
AwsEcsTaskDefinition resource (its revision-carrying ARN output, so a
new revision rolls the rule's launches) or a literal ARN.

- references: AwsEcsTaskDefinition (`status.outputs.task_definition_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEcsTaskDefinition, name: <that resource's name>, fieldPath: status.outputs.task_definition_arn}} -- a bare string does not parse

### spec.targets[].ecsTarget.taskCount

`int32`

Number of tasks launched per matched event. Leave at 0 for the AWS
default of 1. Maximum 10.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":10,"gte":1}}

### spec.targets[].ecsTarget.launchType

`string`

Launch type for the task. Valid values: "EC2", "FARGATE", "EXTERNAL".
Leave empty to use the cluster's default capacity provider strategy, or
set `capacity_provider_strategy` for an explicit blend — AWS rejects a
launch type combined with a capacity provider strategy.

### spec.targets[].ecsTarget.platformVersion

`string`

Fargate platform version (e.g. "LATEST", "1.4.0"). Only valid with the
FARGATE launch type.

- rule: {"string":{"maxLen":"1600"}}

### spec.targets[].ecsTarget.group

`string`

Task group name used for placement decisions (defaults to the family
name of the task definition). Maximum 255 characters.

- rule: {"string":{"maxLen":"255"}}

### spec.targets[].ecsTarget.capacityProviderStrategy

`[]AwsEventBridgeEcsCapacityProviderStrategy`

Capacity provider strategy for the launched tasks. Mutually exclusive
with `launch_type` (AWS rejects both together).

### spec.targets[].ecsTarget.capacityProviderStrategy[].capacityProvider

`string` · required

Capacity provider name ("FARGATE", "FARGATE_SPOT", or a cluster-attached
EC2 capacity provider name).

- rule: {"required":true}

### spec.targets[].ecsTarget.capacityProviderStrategy[].base

`int32`

Minimum number of tasks guaranteed to this provider before weights
apply. Only one strategy entry may carry a non-zero base. Range 0-100000.

- rule: {"int32":{"lte":100000,"gte":0}}

### spec.targets[].ecsTarget.capacityProviderStrategy[].weight

`int32`

Relative share of tasks placed on this provider once bases are met.
Range 0-1000.

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.targets[].ecsTarget.networkConfiguration

`AwsEventBridgeEcsNetworkConfiguration`

VPC networking for the task. Required for Fargate (awsvpc network mode);
used with EC2 launch type only when the task definition uses awsvpc mode.

### spec.targets[].ecsTarget.networkConfiguration.subnets

`[]string | valueFrom` · required

Subnets the task's elastic network interface is placed in. Accepts
direct subnet IDs or references to AwsSubnet resources.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.targets[].ecsTarget.networkConfiguration.securityGroups

`[]string | valueFrom`

Security groups attached to the task's network interface. When empty,
AWS uses the VPC's default security group. Accepts direct IDs or
references to AwsSecurityGroup resources.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.targets[].ecsTarget.networkConfiguration.assignPublicIp

`bool`

Assign a public IP to the task's network interface. Only valid for
Fargate tasks in public subnets.

### spec.targets[].ecsTarget.orderedPlacementStrategy

`[]AwsEventBridgeEcsPlacementStrategy`

Placement strategies ordering candidate instances (EC2 launch type).
Evaluated in order; maximum 5 entries.

- rule: {"repeated":{"maxItems":"5"}}
- rule: type must be 'random', 'spread', or 'binpack'

### spec.targets[].ecsTarget.orderedPlacementStrategy[].type

`string` · required

Strategy type. Valid values: "random", "spread", "binpack".

- rule: {"required":true}

### spec.targets[].ecsTarget.orderedPlacementStrategy[].field

`string`

Attribute the strategy applies to — e.g. "instanceId" or
"attribute:ecs.availability-zone" for spread, "cpu" or "memory" for
binpack. Not used by "random".

- rule: {"string":{"maxLen":"255"}}

### spec.targets[].ecsTarget.placementConstraints

`[]AwsEventBridgeEcsPlacementConstraint`

Placement constraints filtering candidate instances (EC2 launch type).
Maximum 10 entries.

- rule: {"repeated":{"maxItems":"10"}}
- rule: type must be 'distinctInstance' or 'memberOf'
- rule: expression is required for 'memberOf' constraints and must be empty for 'distinctInstance'

### spec.targets[].ecsTarget.placementConstraints[].type

`string` · required

Constraint type. Valid values: "distinctInstance", "memberOf".

- rule: {"required":true}

### spec.targets[].ecsTarget.placementConstraints[].expression

`string`

Cluster query language expression for "memberOf" constraints, e.g.
"attribute:ecs.instance-type =~ t2.*". Not used by "distinctInstance".

- rule: {"string":{"maxLen":"2000"}}

### spec.targets[].ecsTarget.propagateTags

`string`

Propagate tags from the task definition to the launched tasks.
The only accepted value is "TASK_DEFINITION".

### spec.targets[].ecsTarget.enableEcsManagedTags

`bool`

Use Amazon ECS managed tags for the launched tasks (the
aws:ecs:clusterName / aws:ecs:serviceName tag pair).

### spec.targets[].ecsTarget.enableExecuteCommand

`bool`

Enable ECS Exec on the launched tasks for interactive debugging.

### spec.targets[].ecsTarget.tags

`map<string, string>`

Tags applied to the ECS TASKS each event launches (cost allocation,
ownership) — distinct from the rule's own resource tags, which the
module manages from metadata. Merged with the task definition's tags
when propagate_tags is set.

### spec.targets[].redshiftTarget

`AwsEventBridgeTargetRedshiftConfig`

Redshift Data API parameters. Required when the target arn is a
Redshift cluster — each matched event runs a SQL statement against it.

### spec.targets[].redshiftTarget.database

`string` · required

The database the statement runs in. Maximum 64 characters.

- rule: {"required":true,"string":{"maxLen":"64"}}

### spec.targets[].redshiftTarget.dbUser

`string`

Authenticate as this database user with temporary credentials
(GetClusterCredentials). Use this OR secrets_manager_arn — AWS
resolves credentials from whichever is provided. Maximum 128
characters.

- rule: {"string":{"maxLen":"128"}}

### spec.targets[].redshiftTarget.secretsManagerArn

`string`

Authenticate with credentials stored in Secrets Manager — the
alternative to db_user temporary credentials. Pass the secret ARN
as a literal, e.g.
"arn:aws:secretsmanager:us-west-2:123456789012:secret:redshift-...".

### spec.targets[].redshiftTarget.sql

`string`

The SQL statement each matched event runs. Maximum 100,000
characters.

- rule: {"string":{"maxLen":"100000"}}

### spec.targets[].redshiftTarget.statementName

`string`

A name for the statement, visible in the Data API's statement
history. Maximum 500 characters.

- rule: {"string":{"maxLen":"500"}}

### spec.targets[].redshiftTarget.withEvent

`bool`

Deliver the matched event to the statement as execution context
(the Data API's WithEvent flag).

### spec.targets[].runCommandTargets

`[]AwsEventBridgeTargetRunCommandSelector`

SSM Run Command instance selectors. Required when the target arn is a
Systems Manager document — each matched event dispatches the command
to the instances the selectors match. Up to 5 selectors, combined
with AND.

- rule: {"repeated":{"maxItems":"5"}}

### spec.targets[].runCommandTargets[].key

`string` · required

The selector key: "InstanceIds" (select by instance id) or
"tag:<tag-key>" (select every instance carrying the tag). Maximum
128 characters.

- rule: {"required":true,"string":{"maxLen":"128"}}

### spec.targets[].runCommandTargets[].values

`[]string` · required

The selector values: instance ids for "InstanceIds", tag values for
a "tag:" key. 1-50 entries, each 1-256 characters.

- rule: {"repeated":{"minItems":"1","maxItems":"50","items":{"string":{"minLen":"1","maxLen":"256"}}}}

### spec.targets[].sagemakerPipelineTarget

`AwsEventBridgeTargetSagemakerPipelineConfig`

SageMaker Pipelines parameters. Set when the target arn is a
SageMaker pipeline — each matched event starts a pipeline execution
with these parameters.

### spec.targets[].sagemakerPipelineTarget.pipelineParameterList

`[]AwsEventBridgeSagemakerPipelineParameter`

Parameters passed to the pipeline execution, up to 200 — each must
be a parameter the pipeline declares.

- rule: {"repeated":{"maxItems":"200"}}

### spec.targets[].sagemakerPipelineTarget.pipelineParameterList[].name

`string` · required

The parameter name, as declared by the pipeline definition.

- rule: {"required":true}

### spec.targets[].sagemakerPipelineTarget.pipelineParameterList[].value

`string` · required

The value passed for this execution.

- rule: {"required":true}

### spec.targets[].appsyncTarget

`AwsEventBridgeTargetAppsyncConfig`

AppSync parameters. Required when the target arn is an AppSync
GraphQL API endpoint — each matched event invokes the operation.

### spec.targets[].appsyncTarget.graphqlOperation

`string` · required

The GraphQL operation (typically a mutation) to invoke, with its
selection set — variables are bound from the (transformed) event
input. Maximum 1,048,576 characters.

- rule: {"required":true,"string":{"maxLen":"1048576"}}

## Validation Rules

- `event_pattern_or_schedule_required`: exactly one of event_pattern or schedule_expression must be set
- `state_valid_values`: state must be 'ENABLED', 'DISABLED', or 'ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS' when set
- `targets_not_empty`: at least one target is required

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEventBridgeRule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.rule_arn` | `string` | The Amazon Resource Name (ARN) of the rule. Used in IAM policies that grant or restrict access to manage this rule, and as a reference in monitoring and alerting configurations. |
| `status.outputs.rule_name` | `string` | The name of the rule. This is the human-readable identifier used in EventBridge API calls. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.eventBusName` | AwsEventBridgeBus | `status.outputs.bus_name` |
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.targets[].roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.targets[].deadLetterConfig.arn` | AwsSqsQueue | `status.outputs.queue_arn` |
| `spec.targets[].batchTarget.jobDefinition` | AwsBatchJobDefinition | `status.outputs.job_definition_arn` |
| `spec.targets[].ecsTarget.taskDefinitionArn` | AwsEcsTaskDefinition | `status.outputs.task_definition_arn` |
| `spec.targets[].ecsTarget.networkConfiguration.subnets` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.targets[].ecsTarget.networkConfiguration.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |

## See Also

- [Overview](../README.md)
