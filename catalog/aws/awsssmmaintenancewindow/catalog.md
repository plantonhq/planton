# AWS SSM Maintenance Window

Deploys an AWS Systems Manager maintenance window: a recurring change window for your fleet, together with the targets registered to it and the tasks that run inside it. Tasks come in four types — Run Command, Automation runbooks, Lambda invocations, and Step Functions executions — each with priorities, per-task rate controls, and cutoff behavior. Targets and tasks fold into this one component because they cannot outlive their window; the module creates and destroys them together, so ordering is never your problem.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Maintenance Window** — the schedule (cron/rate with timezone and optional day offset), how long it stays open, and the cutoff after which no new task starts. AWS identifies it as `mw-...`.
- **Target Registrations** — one per `targets` entry: instance selections by tag or ID, or resource groups, each getting an AWS-generated target ID (echoed in the `target_ids` output).
- **Task Registrations** — one per `tasks` entry: what runs (document, function, or state machine), at what priority, against which targets, with type-specific invocation parameters (echoed in the `task_ids` output).

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with permissions to manage SSM maintenance windows. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Managed nodes** — instances for the window's targets to select. The window deploys fine before they exist.
- **The task executable** (per task type) — a customer document for RUN_COMMAND/AUTOMATION tasks referencing your own runbook, the Lambda function for LAMBDA tasks, or the state machine for STEP_FUNCTIONS tasks.

## Deploy

### Console

Open the deployment store, find **AWS SSM Maintenance Window**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields covering the schedule, duration and cutoff, target registrations, and tasks. Start from the **Weekly Patch Window** preset in the [Presets](#presets) tab for the classic Sunday-night patch install with rate controls.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSsmMaintenanceWindow
metadata:
  name: weekly-patching
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  schedule: cron(0 2 ? * SUN *)
  scheduleTimezone: America/Los_Angeles
  duration: 4
  cutoff: 1
  description: Sunday 02:00 patching window
  targets:
    - name: prod-instances
      resourceType: INSTANCE
      targets:
        - key: tag:env
          values:
            - prod
  tasks:
    - name: patch-install
      taskType: RUN_COMMAND
      taskArn:
        value: AWS-RunPatchBaseline
      priority: 1
      targets:
        - key: WindowTargetIds
          values:
            - prod-instances
      maxConcurrency: 10%
      maxErrors: 5%
      cutoffBehavior: CANCEL_TASK
      invocation:
        runCommand:
          timeoutSeconds: 3600
          parameters:
            - name: Operation
              values:
                - Install
```

```shell
planton apply -f maintenance-window.yaml
```

This opens a four-hour window Sundays at 02:00 Pacific, registers every instance tagged `env=prod`, and runs a patch install across 10% of the fleet at a time — new starts stop in the final hour, and running installs are cancelled at the cutoff. An Infra Job tracks the provisioning in real time.

### InfraChart

When a task runs a customer runbook deployed in the same chart, wire the task's executable via ValueFromRef:

```yaml
spec:
  region: us-east-1
  schedule: rate(1 day)
  duration: 2
  cutoff: 1
  tasks:
    - name: nightly-cleanup
      taskType: AUTOMATION
      taskArn:
        valueFrom:
          kind: AwsSsmDocument
          name: snapshot-cleanup
          fieldPath: status.outputs.document_name
```

The InfraPipeline resolves the dependency graph, deploys the document first, then registers the window task against it.

## Key Configuration

These are the most important decisions when configuring a maintenance window. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Cutoff arithmetic** — `duration` is 1–24 hours and `cutoff` (0–23) must be strictly less. The cutoff is when NEW task starts stop; anything already running continues or cancels per each task's `cutoffBehavior`, which defaults to CONTINUE_TASK. For disruptive work like patch installs, set CANCEL_TASK so the window's end actually means the end.

**Task priority runs backwards from what you might expect** — 0 is the highest priority and also AWS's default, so an unset priority competes at the front. Tasks sharing a priority run in parallel; give cleanup or verification tasks explicit higher numbers so they wait.

**RUN_COMMAND tasks must be targeted** — AWS rejects an untargeted Run Command task at registration. Untargeted tasks (which run once per window execution regardless of fleet size) are legal only for Automation, Lambda, and Step Functions — pair them with runbooks that manage their own scope. Rate controls (`maxConcurrency`, `maxErrors`) exist only on targeted tasks; AWS rejects them otherwise.

**Reference registered targets by name** — A task targeting `WindowTargetIds` may write the NAME of a target entry declared in the same spec; the module resolves it to the cloud-generated registration ID at deploy. Cross-component consumers read the `target_ids` output instead of guessing IDs.

**A target's name, description, and resourceType force re-registration** — each produces a new target ID. In-spec task references follow automatically through the name-based join; anything external holding the old ID breaks.

**The invocation arm must match the task type** — exactly one of `runCommand`, `automation`, `lambda`, or `stepFunctions`, matching `taskType`. The spec validates this at apply time rather than letting AWS reject it server-side mid-deploy.

**Sensitive invocation payloads** — the Lambda `payload` and Step Functions `input` fields are sensitive: they routinely carry credentials or tokens. Supply a `$secret/...` reference when yours does, never a literal.

**Stage windows dark** — `enabled` defaults to true; an explicit `false` creates the window paused, letting you review targets and tasks before the first execution. Leave `allowUnassociatedTargets` false unless you deliberately want tasks targeting instances the window never registered.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSsmDocument** | `tasks[].taskArn` (RUN_COMMAND / AUTOMATION tasks) | `status.outputs.document_name` |
| **AwsLambda** | `tasks[].taskArn` (LAMBDA tasks) | `status.outputs.function_arn` |
| **AwsStepFunction** | `tasks[].taskArn` (STEP_FUNCTIONS tasks) | `status.outputs.state_machine_arn` |
| **AwsIamRole** | `tasks[].serviceRoleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** | `tasks[].invocation.runCommand.outputS3Bucket` | `status.outputs.bucket_id` |
| **AwsSnsTopic** | `tasks[].invocation.runCommand.notificationConfig.notificationArn` | `status.outputs.topic_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `window_id` | The window's AWS-generated ID (`mw-...`) | Externally registered targets or tasks binding to this window |
| `target_ids` | Target registration IDs keyed by target name | Cross-component task registrations referencing this window's targets via `WindowTargetIds` |

`task_ids` is also echoed — task registration IDs keyed by task name. It exists for import addressing (`{window_id}/{task_id}`) rather than as a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The declared patch window** — Sunday night, four hours, cutoff one hour before close, `AWS-RunPatchBaseline` with `Operation: Install` moving 10% of the fleet at a time. This is patching as an auditable change event rather than a background process — pair it with a continuously scanning AwsSsmAssociation so you know what the window will install before it opens. Start from the **Weekly Patch Window** preset.

**Runbook once per window** — an untargeted AUTOMATION task runs your runbook's `$DEFAULT` version once per execution: snapshot rotation, resource cleanup, drift remediation. The runbook manages its own scope, and releasing a new default document version changes what runs with no window edit — the trade is that scope lives in the document, invisible to the window's target list. Start from the **Automation Runbook Window** preset.

**Mixed-type orchestration** — one window, several tasks at ascending priorities: a Lambda pre-check at priority 0, the Run Command change at 1, a Step Functions verification at 2. The window gives unrelated tools one shared change boundary and one cutoff.

## Works With

- [**AWS SSM Document**](/infra-catalog/aws-ssm-document) — the customer runbook RUN_COMMAND and AUTOMATION tasks execute, wired via `taskArn`
- [**AWS SSM Patch Baseline**](/infra-catalog/aws-ssm-patch-baseline) — governs what a patch-install task approves when the window runs `AWS-RunPatchBaseline`
- [**AWS Lambda**](/infra-catalog/aws-lambda) — the function LAMBDA tasks invoke
- [**AWS Step Functions**](/infra-catalog/aws-step-function) — the state machine STEP_FUNCTIONS tasks execute
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) — receives command lifecycle notifications from Run Command tasks
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — receives Run Command output when the invocation sets an output bucket
