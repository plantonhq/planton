# AwsEventBridgeScheduler

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsEventBridgeSchedulerSpec defines one EventBridge Scheduler
schedule: cron for the cloud - a cron, rate, or one-time expression
invoking one target under an execution role, with optional flexible
time windows, retries, and a dead-letter queue.

The schedule's name in AWS is metadata.name. Its GROUP is a
name-and-tags container (the provider's own update path is
tags-only): create one here (group) or join one that exists
elsewhere by name (group_name); unset means AWS's "default" group.
The group is fixed for life - changing it replaces the schedule.

The execution role is assumed by the Scheduler service - its trust
policy must allow scheduler.amazonaws.com, and IAM propagation makes
a first deploy with a fresh role eventually consistent (the provider
retries for up to two minutes).

## Example

```yaml
# Canonical AwsEventBridgeScheduler example (hack/dev manifest and
# refgen Example source): a five-minute rate schedule delivering a
# static payload to an SQS queue, in an owned group, with bounded
# retries.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEventBridgeScheduler
metadata:
  name: queue-heartbeat
  id: queue-heartbeat
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Heartbeat message to the jobs queue every five minutes
  group:
    name: batch-jobs
  scheduleExpression: rate(5 minutes)
  flexibleTimeWindow:
    mode: "OFF"
  target:
    arn:
      value: arn:aws:sqs:us-west-2:123456789012:jobs
    roleArn:
      value: arn:aws:iam::123456789012:role/scheduler-exec
    input: '{"type": "heartbeat"}'
    retryPolicy:
      maximumEventAgeInSeconds: 3600
      maximumRetryAttempts: 3
    sqsParameters: {}
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.group` | `AwsEventBridgeScheduleGroup` |  |  |  |
| `spec.group.name` | `string` | yes |  |  |
| `spec.groupName` | `string` |  |  |  |
| `spec.scheduleExpression` | `string` | yes |  |  |
| `spec.scheduleExpressionTimezone` | `string` |  |  |  |
| `spec.startDate` | `string` |  |  |  |
| `spec.endDate` | `string` |  |  |  |
| `spec.state` | `string` |  |  |  |
| `spec.actionAfterCompletion` | `string` |  |  |  |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.flexibleTimeWindow` | `AwsEventBridgeScheduleFlexibleTimeWindow` | yes |  |  |
| `spec.flexibleTimeWindow.mode` | `string` |  |  |  |
| `spec.flexibleTimeWindow.maximumWindowInMinutes` | `int32` |  |  |  |
| `spec.target` | `AwsEventBridgeScheduleTarget` | yes |  |  |
| `spec.target.arn` | `string \| valueFrom` | yes |  |  |
| `spec.target.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.target.input` | `string` |  |  |  |
| `spec.target.deadLetterQueueArn` | `string \| valueFrom` |  |  | AwsSqsQueue (`status.outputs.queue_arn`) |
| `spec.target.retryPolicy` | `AwsEventBridgeScheduleRetryPolicy` |  |  |  |
| `spec.target.retryPolicy.maximumEventAgeInSeconds` | `int32` |  |  |  |
| `spec.target.retryPolicy.maximumRetryAttempts` | `int32` |  |  |  |
| `spec.target.ecsParameters` | `AwsEventBridgeScheduleEcsParameters` |  |  |  |
| `spec.target.ecsParameters.taskDefinitionArn` | `string \| valueFrom` | yes |  | AwsEcsTaskDefinition (`status.outputs.task_definition_arn`) |
| `spec.target.ecsParameters.taskCount` | `int32` |  |  |  |
| `spec.target.ecsParameters.launchType` | `string` |  |  |  |
| `spec.target.ecsParameters.capacityProviderStrategy` | `[]AwsEventBridgeScheduleCapacityProviderStrategy` |  |  |  |
| `spec.target.ecsParameters.capacityProviderStrategy[].capacityProvider` | `string` | yes |  |  |
| `spec.target.ecsParameters.capacityProviderStrategy[].base` | `int32` |  |  |  |
| `spec.target.ecsParameters.capacityProviderStrategy[].weight` | `int32` |  |  |  |
| `spec.target.ecsParameters.networkConfiguration` | `AwsEventBridgeScheduleNetworkConfiguration` |  |  |  |
| `spec.target.ecsParameters.networkConfiguration.subnets` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.target.ecsParameters.networkConfiguration.securityGroups` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.target.ecsParameters.networkConfiguration.assignPublicIp` | `bool` |  |  |  |
| `spec.target.ecsParameters.group` | `string` |  |  |  |
| `spec.target.ecsParameters.platformVersion` | `string` |  |  |  |
| `spec.target.ecsParameters.placementConstraints` | `[]AwsEventBridgeSchedulePlacementConstraint` |  |  |  |
| `spec.target.ecsParameters.placementConstraints[].type` | `string` |  |  |  |
| `spec.target.ecsParameters.placementConstraints[].expression` | `string` |  |  |  |
| `spec.target.ecsParameters.placementStrategy` | `[]AwsEventBridgeSchedulePlacementStrategy` |  |  |  |
| `spec.target.ecsParameters.placementStrategy[].type` | `string` |  |  |  |
| `spec.target.ecsParameters.placementStrategy[].field` | `string` |  |  |  |
| `spec.target.ecsParameters.propagateTags` | `string` |  |  |  |
| `spec.target.ecsParameters.tags` | `map<string, string>` |  |  |  |
| `spec.target.ecsParameters.enableEcsManagedTags` | `bool` |  |  |  |
| `spec.target.ecsParameters.enableExecuteCommand` | `bool` |  |  |  |
| `spec.target.ecsParameters.referenceId` | `string` |  |  |  |
| `spec.target.eventbridgeParameters` | `AwsEventBridgeScheduleEventBridgeParameters` |  |  |  |
| `spec.target.eventbridgeParameters.detailType` | `string` | yes |  |  |
| `spec.target.eventbridgeParameters.source` | `string` | yes |  |  |
| `spec.target.kinesisParameters` | `AwsEventBridgeScheduleKinesisParameters` |  |  |  |
| `spec.target.kinesisParameters.partitionKey` | `string` | yes |  |  |
| `spec.target.sagemakerPipelineParameters` | `AwsEventBridgeScheduleSageMakerPipelineParameters` |  |  |  |
| `spec.target.sagemakerPipelineParameters.pipelineParameters` | `[]AwsEventBridgeScheduleSageMakerPipelineParameter` |  |  |  |
| `spec.target.sagemakerPipelineParameters.pipelineParameters[].name` | `string` | yes |  |  |
| `spec.target.sagemakerPipelineParameters.pipelineParameters[].value` | `string` | yes |  |  |
| `spec.target.sqsParameters` | `AwsEventBridgeScheduleSqsParameters` |  |  |  |
| `spec.target.sqsParameters.messageGroupId` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the schedule lives in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

What this schedule triggers and why.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512"}}

### spec.group

`AwsEventBridgeScheduleGroup`

An owned schedule group, created by this instance (carrying the
standard identity tags). One group serves many schedules - other
instances join it via group_name.

### spec.group.name

`string` · required

The group's name in AWS (its identity - renaming replaces it, and
with it the schedule). Up to 64 characters: letters, digits, dot,
dash, underscore.

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_.-]+$"}}

### spec.groupName

`string`

An existing group to join by name (another instance's owned group,
or any pre-existing one). Unset (with no owned group) uses AWS's
"default" group. Fixed for life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"64"}}

### spec.scheduleExpression

`string` · required

When the schedule fires: "cron(0 12 * * ? *)" (noon daily),
"rate(5 minutes)", or "at(2026-12-31T23:59:00)" (one-time).

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.scheduleExpressionTimezone

`string`

The IANA timezone the cron expression evaluates in. Example:
"America/Los_Angeles". Unset keeps AWS's default (UTC).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"50"}}

### spec.startDate

`string`

Don't fire before this instant (RFC3339). Example:
"2026-09-01T00:00:00Z". Unset starts immediately.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.endDate

`string`

Don't fire after this instant (RFC3339). Unset never expires.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.state

`string`

Pause the schedule without deleting it. Unset keeps AWS's default
(ENABLED).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENABLED","DISABLED"]}}

### spec.actionAfterCompletion

`string`

What happens after a one-time ("at(...)") schedule completes:
DELETE removes the schedule itself. Unset keeps AWS's default
(NONE). NOTE with DELETE, AWS deletes the schedule out from under
IaC state - the next deploy recreates it; use DELETE only for
fire-and-forget schedules.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","DELETE"]}}

### spec.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key that encrypts the schedule's target input
(unset uses AWS-owned keys). Reference an AwsKmsKey key_arn output
or pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.flexibleTimeWindow

`AwsEventBridgeScheduleFlexibleTimeWindow` · required

The invocation-time flexibility window. AWS requires an explicit
choice: OFF fires exactly on schedule, FLEXIBLE lets AWS spread
invocations within a window (smoothing load spikes across a fleet
of schedules).

- rule: {"required":true}
- rule: mode FLEXIBLE requires maximum_window_in_minutes; mode OFF forbids it

### spec.flexibleTimeWindow.mode

`string`

OFF fires exactly on schedule; FLEXIBLE fires anywhere inside the
window starting at the scheduled time.

- rule: {"string":{"in":["OFF","FLEXIBLE"]}}

### spec.flexibleTimeWindow.maximumWindowInMinutes

`int32` · optional (explicit presence)

The window size in minutes (1-1440). Only with mode FLEXIBLE.

- rule: {"int32":{"lte":1440,"gte":1}}

### spec.target

`AwsEventBridgeScheduleTarget` · required

What the schedule invokes.

- rule: {"required":true}
- rule: configure at most one service parameter block (ecs, eventbridge, kinesis, sagemaker_pipeline, sqs) - it must match the target arn's service

### spec.target.arn

`string | valueFrom` · required

The target's ARN: a Lambda function, SQS queue, ECS cluster,
Kinesis stream, Step Functions state machine, EventBridge bus,
SageMaker pipeline, API destination, or any of the universal-
target API ARNs. No single kind dominates, so references here
carry NO default kind - in manifests, a valueFrom on this field
must state its kind explicitly.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.target.roleArn

`string | valueFrom` · required

The role Scheduler assumes to invoke the target. Its trust policy
must allow scheduler.amazonaws.com. Reference an AwsIamRole
role_arn output or pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.target.input

`string`

The static payload delivered on every invocation (for most
targets, a JSON document; templating placeholders like
<aws.scheduler.scheduled-time> are substituted by AWS). Unset
sends the target service's default event.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.target.deadLetterQueueArn

`string | valueFrom`

The SQS queue that receives events Scheduler could not deliver
after exhausting retries. Reference an AwsSqsQueue queue_arn
output or pass a literal ARN.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

### spec.target.retryPolicy

`AwsEventBridgeScheduleRetryPolicy`

How long and how often Scheduler retries failed invocations.
Unset keeps AWS's defaults (24h event age, 185 attempts).

### spec.target.retryPolicy.maximumEventAgeInSeconds

`int32` · optional (explicit presence)

Give up on events older than this (60-86400 seconds). Unset keeps
AWS's default (86400). Presence-typed so boundaries are
expressible.

- rule: {"int32":{"lte":86400,"gte":60}}

### spec.target.retryPolicy.maximumRetryAttempts

`int32` · optional (explicit presence)

Retry at most this many times (0-185). Unset keeps AWS's default
(185). Presence-typed so 0 (no retries) is expressible.

- rule: {"int32":{"lte":185,"gte":0}}

### spec.target.ecsParameters

`AwsEventBridgeScheduleEcsParameters`

ECS RunTask parameters (target arn: the ECS cluster).

### spec.target.ecsParameters.taskDefinitionArn

`string | valueFrom` · required

The task definition to run (with or without revision - without
pins to the latest ACTIVE). Reference an AwsEcsTaskDefinition
task_definition_arn output or pass a literal ARN.

- references: AwsEcsTaskDefinition (`status.outputs.task_definition_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEcsTaskDefinition, name: <that resource's name>, fieldPath: status.outputs.task_definition_arn}} -- a bare string does not parse

### spec.target.ecsParameters.taskCount

`int32` · optional (explicit presence)

How many tasks each invocation launches (0-10). Unset keeps AWS's
default (1). Presence-typed so 0 is expressible.

- rule: {"int32":{"lte":10,"gte":0}}

### spec.target.ecsParameters.launchType

`string`

FARGATE, EC2, or EXTERNAL. Unset lets the cluster's default
capacity providers decide (mutually exclusive with
capacity_provider_strategy at AWS).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["EC2","FARGATE","EXTERNAL"]}}

### spec.target.ecsParameters.capacityProviderStrategy

`[]AwsEventBridgeScheduleCapacityProviderStrategy`

Capacity provider strategy entries (at most 6). Mutually exclusive
with launch_type at AWS.

- rule: {"repeated":{"maxItems":"6"}}

### spec.target.ecsParameters.capacityProviderStrategy[].capacityProvider

`string` · required

The capacity provider's name (e.g. "FARGATE", "FARGATE_SPOT", or a
custom provider).

- rule: {"string":{"minLen":"1","maxLen":"255"}}

### spec.target.ecsParameters.capacityProviderStrategy[].base

`int32`

How many tasks run on this provider before the weights apply
(0-100000).

- rule: {"int32":{"lte":100000,"gte":0}}

### spec.target.ecsParameters.capacityProviderStrategy[].weight

`int32`

The relative share of remaining tasks on this provider (0-1000).

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.target.ecsParameters.networkConfiguration

`AwsEventBridgeScheduleNetworkConfiguration`

VPC networking for awsvpc-mode tasks (required for Fargate).

### spec.target.ecsParameters.networkConfiguration.subnets

`[]string | valueFrom` · required

The subnets tasks launch into. Reference AwsSubnet subnet_id
outputs or pass literal subnet-... ids.

Containment-exempt: these subnets place the ECS TASKS the schedule
launches, not the schedule itself, which is a managed service outside
the network. On a diagram the schedule stands beside the VPC with a
line in -- the verdict an EventBridge rule's ECS target subnets
already carry.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.target.ecsParameters.networkConfiguration.securityGroups

`[]string | valueFrom`

The security groups attached to tasks. Unset uses the VPC's
default. Reference AwsSecurityGroup security_group_id outputs or
pass literal sg-... ids.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.target.ecsParameters.networkConfiguration.assignPublicIp

`bool`

Assign public IPs to tasks (public-subnet Fargate tasks that pull
images over the internet need this).

### spec.target.ecsParameters.group

`string`

The ECS task group name (up to 255 characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.target.ecsParameters.platformVersion

`string`

The Fargate platform version (e.g. "LATEST", "1.4.0"). Fargate
only.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.target.ecsParameters.placementConstraints

`[]AwsEventBridgeSchedulePlacementConstraint`

Placement constraints (at most 10). EC2 launch type only.

- rule: {"repeated":{"maxItems":"10"}}

### spec.target.ecsParameters.placementConstraints[].type

`string`

"distinctInstance" (spread tasks across instances) or "memberOf"
(cluster-query-language expression).

- rule: {"string":{"in":["distinctInstance","memberOf"]}}

### spec.target.ecsParameters.placementConstraints[].expression

`string`

The cluster query language expression, for type memberOf.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"2000"}}

### spec.target.ecsParameters.placementStrategy

`[]AwsEventBridgeSchedulePlacementStrategy`

Placement strategies (at most 5). EC2 launch type only.

- rule: {"repeated":{"maxItems":"5"}}

### spec.target.ecsParameters.placementStrategy[].type

`string`

"random", "spread", or "binpack".

- rule: {"string":{"in":["random","spread","binpack"]}}

### spec.target.ecsParameters.placementStrategy[].field

`string`

The field the strategy applies to (e.g. "instanceId" or
"attribute:ecs.availability-zone" for spread, "cpu"/"memory" for
binpack).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.target.ecsParameters.propagateTags

`string`

Propagate tags from the task definition to launched tasks. The
only value AWS accepts is "TASK_DEFINITION"; unset propagates
nothing.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["TASK_DEFINITION"]}}

### spec.target.ecsParameters.tags

`map<string, string>`

Tags applied to every launched task (in addition to propagated
ones).

### spec.target.ecsParameters.enableEcsManagedTags

`bool`

Use ECS managed tags on launched tasks.

### spec.target.ecsParameters.enableExecuteCommand

`bool`

Enable ECS Exec on launched tasks (interactive debugging).

### spec.target.ecsParameters.referenceId

`string`

The reference id attached to each RunTask call.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.target.eventbridgeParameters

`AwsEventBridgeScheduleEventBridgeParameters`

EventBridge PutEvents parameters (target arn: the event bus).

### spec.target.eventbridgeParameters.detailType

`string` · required

The event's detail-type (what rules match on).

- rule: {"string":{"minLen":"1","maxLen":"128"}}

### spec.target.eventbridgeParameters.source

`string` · required

The event's source (e.g. "com.acme.billing").

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.target.kinesisParameters

`AwsEventBridgeScheduleKinesisParameters`

Kinesis PutRecord parameters (target arn: the stream).

### spec.target.kinesisParameters.partitionKey

`string` · required

The partition key (decides the shard; up to 256 characters).

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.target.sagemakerPipelineParameters

`AwsEventBridgeScheduleSageMakerPipelineParameters`

SageMaker StartPipelineExecution parameters (target arn: the
pipeline).

- rule: pipeline_parameters entries must have unique names

### spec.target.sagemakerPipelineParameters.pipelineParameters

`[]AwsEventBridgeScheduleSageMakerPipelineParameter`

Name/value parameters passed to the pipeline execution (at most
200).

- rule: {"repeated":{"maxItems":"200"}}

### spec.target.sagemakerPipelineParameters.pipelineParameters[].name

`string` · required

The parameter's name.

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.target.sagemakerPipelineParameters.pipelineParameters[].value

`string` · required

The parameter's value.

- rule: {"string":{"minLen":"1","maxLen":"1024"}}

### spec.target.sqsParameters

`AwsEventBridgeScheduleSqsParameters`

SQS SendMessage parameters (target arn: the queue).

### spec.target.sqsParameters.messageGroupId

`string`

The message group id - REQUIRED by AWS for FIFO queues, forbidden
for standard queues.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128"}}

## Validation Rules

- `spec.group_own_xor_existing`: set at most one of group (create the group here) and group_name (join an existing group) - unset both for AWS's default group

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEventBridgeScheduler, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.schedule_arn` | `string` | The schedule's ARN. |
| `status.outputs.group_name` | `string` | The group the schedule lives in - the owned group's name, the joined group_name, or "default". With the schedule name (metadata.name), this forms the provider's "{group}/{name}" import ID. |
| `status.outputs.group_arn` | `string` | The owned group's ARN. Empty when the instance owns no group. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.target.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.target.deadLetterQueueArn` | AwsSqsQueue | `status.outputs.queue_arn` |
| `spec.target.ecsParameters.taskDefinitionArn` | AwsEcsTaskDefinition | `status.outputs.task_definition_arn` |
| `spec.target.ecsParameters.networkConfiguration.subnets` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.target.ecsParameters.networkConfiguration.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |

## See Also

- [Overview](../README.md)
