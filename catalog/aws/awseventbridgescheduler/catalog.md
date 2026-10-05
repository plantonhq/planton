# AWS EventBridge Scheduler

Deploys an EventBridge Scheduler schedule — serverless cron that fires a Lambda, queues a message, runs an ECS task, or starts a pipeline on a cron, rate, or one-time expression, with timezones, retries, and a dead-letter queue built in. The schedule invokes one target under an execution role; optionally it creates or joins a schedule group, the name-and-tags container schedules live in. Schedules are stateless, so pausing, editing, and even replacing them costs nothing but the momentary gap.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EventBridge Scheduler schedule** — named after `metadata.name`, firing on `scheduleExpression` (cron, rate, or one-time `at(...)`) in its timezone, within its start/end window, in `ENABLED` or `DISABLED` state, with an exact or flexible invocation window
- **Target wiring** — the target ARN and the `scheduler.amazonaws.com`-trusting execution role, an optional static input payload, the retry policy, the dead-letter queue, and at most one service parameter block (ECS RunTask, EventBridge PutEvents, Kinesis PutRecord, SageMaker StartPipelineExecution, SQS SendMessage) matching the target's service
- **Schedule group** — created only when `group` is set: an owned group carrying the standard identity tags (the schedule itself is untaggable at AWS). Other instances join it via `groupName`; unset both and the schedule lives in AWS's `default` group

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with EventBridge Scheduler permissions and `iam:PassRole` on the execution role. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **The target resource** — the schedule invokes an existing ARN; deploy it first or in the same InfraChart.
- **An execution role trusting `scheduler.amazonaws.com`** — with permission to invoke the target. Goes in `target.roleArn`. A freshly created role takes up to two minutes to propagate; the provider retries exactly that error.
- **For a dead-letter queue** — an SQS queue you actually monitor, wired via `target.deadLetterQueueArn`.

## Deploy

### Console

Open the deployment store, find **AWS EventBridge Scheduler**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the schedule: expression and timezone, the flexible time window, the target and its execution role, and retry/DLQ policy. Start from the **Nightly Cron to Lambda** preset in the [Presets](#presets) tab for the workhorse shape, or the **Scheduled Fargate Task** preset for recurring container jobs.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEventBridgeScheduler
metadata:
  name: nightly-report
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  description: Nightly report generation at 2am Eastern
  scheduleExpression: cron(0 2 * * ? *)
  scheduleExpressionTimezone: America/New_York
  flexibleTimeWindow:
    mode: "OFF"
  target:
    arn:
      valueFrom:
        kind: AwsLambda
        name: report-generator
        fieldPath: status.outputs.function_arn
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: scheduler-exec
        fieldPath: status.outputs.role_arn
    input: '{"report": "daily-summary"}'
    retryPolicy:
      maximumEventAgeInSeconds: 3600
      maximumRetryAttempts: 3
    deadLetterQueueArn:
      valueFrom:
        kind: AwsSqsQueue
        name: scheduler-dlq
        fieldPath: status.outputs.queue_arn
```

```shell
planton apply -f eventbridge-scheduler.yaml
```

This creates an enabled schedule that invokes the referenced Lambda at 2 AM Eastern every night (daylight saving handled by AWS), retrying failures for at most an hour and three attempts before dead-lettering. An Infra Job tracks the provisioning in real time.

### InfraChart

When the schedule deploys alongside its target and role in one chart, wire the edges via ValueFromRef — `target.arn` carries no default kind, so its valueFrom states the kind explicitly:

```yaml
spec:
  region: us-east-1
  scheduleExpression: rate(15 minutes)
  flexibleTimeWindow:
    mode: FLEXIBLE
    maximumWindowInMinutes: 5
  target:
    arn:
      valueFrom:
        kind: AwsSqsQueue
        name: refresh-jobs
        fieldPath: status.outputs.queue_arn
    roleArn:
      valueFrom:
        kind: AwsIamRole
        name: scheduler-exec
        fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the queue and role first, then creates the schedule against them.

## Key Configuration

These are the most important decisions when configuring a schedule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Timezones belong on the schedule, not in your head** — `cron()` expressions evaluate in `scheduleExpressionTimezone`; set it to the business timezone (`America/New_York`) and daylight-saving shifts are AWS's problem. Unset means UTC, and a "2 AM" job silently drifting an hour twice a year is the classic symptom.

**The flexible time window is a required choice** — AWS makes you pick: `OFF` fires exactly on schedule; `FLEXIBLE` with a window lets AWS spread invocations inside it. A fleet of `rate(15 minutes)` schedules with a few minutes of flex stops stampeding a shared cluster or API at the same instant. Quote `"OFF"` in YAML — unquoted, it parses as a boolean.

**Retries need somewhere to go** — AWS's default retry policy (24-hour event age, 185 attempts) can hammer a broken target for a day. Bound it with `retryPolicy.maximumRetryAttempts` and give exhausted events a `deadLetterQueueArn` pointing at a queue someone monitors — before the first failure, not after.

**The group is a filing cabinet, not a feature** — Groups exist to tag and bulk-delete schedules; they have no runtime behavior. Own one group per team or chart (it carries the identity tags) and join it from other schedules via `groupName`. The group binding is fixed for life: moving a schedule replaces it — cheap for stateless schedules, but plan for the momentary gap.

**DELETE-after-completion deletes out from under IaC** — `actionAfterCompletion: DELETE` makes AWS remove a completed one-time `at(...)` schedule itself; the next deploy finds it missing and recreates it. Use DELETE only for fire-and-forget schedules nothing re-deploys, and NONE (the default) when IaC owns the lifecycle.

**Pause with state, not deletion** — `state: DISABLED` stops firing while keeping the schedule, its history, and its wiring; flip it back to ENABLED when ready. AWS's default is ENABLED — a freshly applied schedule starts firing on its expression immediately.

## Outputs and Dependencies

### What This Kind Consumes

`target.arn` ranges over many services (Lambda, SQS, ECS, Kinesis, Step Functions, event buses, SageMaker, API destinations), so it carries no default kind — a valueFrom on it states its kind explicitly. The common wirings:

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `target.roleArn` | `status.outputs.role_arn` |
| **AwsLambda** | `target.arn` | `status.outputs.function_arn` |
| **AwsSqsQueue** | `target.arn` / `target.deadLetterQueueArn` | `status.outputs.queue_arn` |
| **AwsEcsTaskDefinition** | `target.ecsParameters.taskDefinitionArn` | `status.outputs.task_definition_arn` |
| **AwsKmsKey** | `kmsKeyArn` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `schedule_arn` | The schedule's ARN | IAM policies scoping who may update or delete the schedule |
| `group_arn` | The owned group's ARN; empty when the instance owns no group | IAM policies for group-scoped schedule administration |

`group_name` is also echoed — the group the schedule lives in (`default` when none is configured). With the schedule name it forms the provider's `{group}/{name}` import ID rather than a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Timezone-aware nightly cron to Lambda** — A business-timezone cron with bounded retries and a monitored DLQ: the replacement for the cron box that existed to run one script. Start from the **Nightly Cron to Lambda** preset.

**Recurring Fargate task with a flexible window** — A rate schedule launching a task every N minutes, joined to an existing group by name, with a few minutes of flex so a fleet of these never stampedes the cluster at once. The trade against a long-running service: cold task startup per run, in exchange for paying only while the task runs. Start from the **Scheduled Fargate Task** preset.

**One-time scheduled action** — An `at(...)` expression for a future cutover or delayed job. Keep `actionAfterCompletion: NONE` when IaC owns the schedule; reach for DELETE only when nothing will ever re-apply the manifest.

**Schedule into an event bus** — Target an EventBridge bus with `eventbridgeParameters.detailType` and `source` set, letting rules fan one clock tick out to many consumers instead of maintaining one schedule per consumer.

## Works With

- [**AWS Lambda**](/infra-catalog/aws-lambda) — the most common target; the schedule delivers `input` as the event payload
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the scheduler-trusting execution role wired via `target.roleArn`
- [**AWS SQS Queue**](/infra-catalog/aws-sqs-queue) — message target and the dead-letter queue for exhausted retries
- [**AWS ECS Task Definition**](/infra-catalog/aws-ecs-task-definition) — the task each invocation launches via `ecsParameters`
- [**AWS EventBridge Bus**](/infra-catalog/aws-event-bridge-bus) — target for fanning scheduled events out through rules
- [**AWS EventBridge API Destination**](/infra-catalog/aws-event-bridge-api-destination) — authenticated HTTP endpoint a schedule can invoke
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption of the target input via `kmsKeyArn`
