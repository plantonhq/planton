# AwsSsmMaintenanceWindow

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSsmMaintenanceWindowSpec defines the desired configuration for
one AWS Systems Manager maintenance window: a recurring window of
time plus the folded targets registered with it and the tasks (Run
Command, Automation, Lambda, Step Functions) that execute inside it.

The window's name is metadata.name. Targets and tasks are true
window satellites (their window binding forces replacement at the
provider), folded here as name-keyed collections - both modules
create and destroy them with the window so ordering is never the
author's problem. AWS identifies the window as "mw-..." and each
satellite by its own generated ID (echoed in the target_ids/task_ids
output maps for imports).

## Example

```yaml
# Canonical AwsSsmMaintenanceWindow example (hack/dev manifest and
# refgen Example source): a Sunday-night window with one tag-based
# target and all four task types, so the offline `tofu plan` renders
# every invocation arm (run-command with CloudWatch + SNS delivery,
# automation, lambda, step-functions). Literal ARNs stand in for
# composed references.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSsmMaintenanceWindow
metadata:
  name: sunday-patching
  id: sunday-patching
  org: test-org
  env: dev
spec:
  region: us-west-2
  schedule: cron(0 2 ? * SUN *)
  duration: 4
  cutoff: 1
  description: Sunday 02:00 patching window
  scheduleTimezone: America/Los_Angeles
  allowUnassociatedTargets: false
  targets:
    - name: prod-instances
      resourceType: INSTANCE
      targets:
        - key: tag:env
          values:
            - prod
      description: Production fleet by tag
  tasks:
    - name: patch-scan
      taskType: RUN_COMMAND
      taskArn:
        value: AWS-RunPatchBaseline
      priority: 1
      targets:
        - key: WindowTargetIds
          values:
            - example-window-target-id
      maxConcurrency: 10%
      maxErrors: "1"
      cutoffBehavior: CANCEL_TASK
      invocation:
        runCommand:
          comment: Weekly patch scan
          documentVersion: $DEFAULT
          timeoutSeconds: 600
          parameters:
            - name: Operation
              values:
                - Scan
          outputS3Bucket:
            value: patching-output-bucket
          outputS3KeyPrefix: mw/patch-scan
          cloudwatchConfig:
            cloudwatchLogGroupName: /aws/ssm/sunday-patching
            cloudwatchOutputEnabled: true
          notificationConfig:
            notificationArn:
              value: arn:aws:sns:us-west-2:123456789012:mw-events
            notificationEvents:
              - Success
              - Failed
            notificationType: Command
    - name: restart-runbook
      taskType: AUTOMATION
      taskArn:
        value: AWS-RestartEC2Instance
      priority: 2
      invocation:
        automation:
          documentVersion: $LATEST
          parameters:
            - name: InstanceId
              values:
                - i-0123456789abcdef0
    - name: notify-function
      taskType: LAMBDA
      taskArn:
        value: arn:aws:lambda:us-west-2:123456789012:function:mw-notify
      priority: 3
      invocation:
        lambda:
          payload: '{"source":"maintenance-window"}'
          qualifier: live
    - name: orchestrate-deploy
      taskType: STEP_FUNCTIONS
      taskArn:
        value: arn:aws:states:us-west-2:123456789012:stateMachine:mw-deploy
      priority: 4
      invocation:
        stepFunctions:
          input: '{"window":"sunday-patching"}'
          name: mw-run
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.schedule` | `string` | yes |  |  |
| `spec.duration` | `int32` |  |  |  |
| `spec.cutoff` | `int32` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.enabled` | `bool` |  |  |  |
| `spec.allowUnassociatedTargets` | `bool` |  |  |  |
| `spec.scheduleTimezone` | `string` |  |  |  |
| `spec.scheduleOffset` | `int32` |  |  |  |
| `spec.startDate` | `string` |  |  |  |
| `spec.endDate` | `string` |  |  |  |
| `spec.targets` | `[]AwsSsmMaintenanceWindowTargetEntry` |  |  |  |
| `spec.targets[].name` | `string` | yes |  |  |
| `spec.targets[].resourceType` | `string` |  |  |  |
| `spec.targets[].targets` | `[]AwsSsmMaintenanceWindowTargetSelector` | yes |  |  |
| `spec.targets[].targets[].key` | `string` | yes |  |  |
| `spec.targets[].targets[].values` | `[]string` | yes |  |  |
| `spec.targets[].description` | `string` | yes |  |  |
| `spec.targets[].ownerInformation` | `string` | yes |  |  |
| `spec.tasks` | `[]AwsSsmMaintenanceWindowTaskEntry` |  |  |  |
| `spec.tasks[].name` | `string` | yes |  |  |
| `spec.tasks[].taskType` | `string` |  |  |  |
| `spec.tasks[].taskArn` | `string \| valueFrom` | yes |  |  |
| `spec.tasks[].description` | `string` |  |  |  |
| `spec.tasks[].serviceRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.tasks[].priority` | `int32` |  |  |  |
| `spec.tasks[].maxConcurrency` | `string` |  |  |  |
| `spec.tasks[].maxErrors` | `string` |  |  |  |
| `spec.tasks[].cutoffBehavior` | `string` |  |  |  |
| `spec.tasks[].targets` | `[]AwsSsmMaintenanceWindowTargetSelector` |  |  |  |
| `spec.tasks[].targets[].key` | `string` | yes |  |  |
| `spec.tasks[].targets[].values` | `[]string` | yes |  |  |
| `spec.tasks[].invocation` | `AwsSsmMaintenanceWindowTaskInvocation` |  |  |  |
| `spec.tasks[].invocation.runCommand` | `AwsSsmMaintenanceWindowRunCommandInvocation` |  |  |  |
| `spec.tasks[].invocation.runCommand.comment` | `string` |  |  |  |
| `spec.tasks[].invocation.runCommand.documentHash` | `string` |  |  |  |
| `spec.tasks[].invocation.runCommand.documentHashType` | `string` |  |  |  |
| `spec.tasks[].invocation.runCommand.documentVersion` | `string` |  |  |  |
| `spec.tasks[].invocation.runCommand.outputS3Bucket` | `string \| valueFrom` |  |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.tasks[].invocation.runCommand.outputS3KeyPrefix` | `string` |  |  |  |
| `spec.tasks[].invocation.runCommand.serviceRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.tasks[].invocation.runCommand.timeoutSeconds` | `int32` |  |  |  |
| `spec.tasks[].invocation.runCommand.parameters` | `[]AwsSsmMaintenanceWindowTaskParameter` |  |  |  |
| `spec.tasks[].invocation.runCommand.parameters[].name` | `string` | yes |  |  |
| `spec.tasks[].invocation.runCommand.parameters[].values` | `[]string` | yes |  |  |
| `spec.tasks[].invocation.runCommand.cloudwatchConfig` | `AwsSsmMaintenanceWindowCloudwatchConfig` |  |  |  |
| `spec.tasks[].invocation.runCommand.cloudwatchConfig.cloudwatchLogGroupName` | `string` |  |  |  |
| `spec.tasks[].invocation.runCommand.cloudwatchConfig.cloudwatchOutputEnabled` | `bool` |  |  |  |
| `spec.tasks[].invocation.runCommand.notificationConfig` | `AwsSsmMaintenanceWindowNotificationConfig` |  |  |  |
| `spec.tasks[].invocation.runCommand.notificationConfig.notificationArn` | `string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.tasks[].invocation.runCommand.notificationConfig.notificationEvents` | `[]string` |  |  |  |
| `spec.tasks[].invocation.runCommand.notificationConfig.notificationType` | `string` |  |  |  |
| `spec.tasks[].invocation.automation` | `AwsSsmMaintenanceWindowAutomationInvocation` |  |  |  |
| `spec.tasks[].invocation.automation.documentVersion` | `string` |  |  |  |
| `spec.tasks[].invocation.automation.parameters` | `[]AwsSsmMaintenanceWindowTaskParameter` |  |  |  |
| `spec.tasks[].invocation.automation.parameters[].name` | `string` | yes |  |  |
| `spec.tasks[].invocation.automation.parameters[].values` | `[]string` | yes |  |  |
| `spec.tasks[].invocation.lambda` | `AwsSsmMaintenanceWindowLambdaInvocation` |  |  |  |
| `spec.tasks[].invocation.lambda.clientContext` | `string` | yes |  |  |
| `spec.tasks[].invocation.lambda.payload` | `string` (sensitive) |  |  |  |
| `spec.tasks[].invocation.lambda.qualifier` | `string` | yes |  |  |
| `spec.tasks[].invocation.stepFunctions` | `AwsSsmMaintenanceWindowStepFunctionsInvocation` |  |  |  |
| `spec.tasks[].invocation.stepFunctions.input` | `string` (sensitive) |  |  |  |
| `spec.tasks[].invocation.stepFunctions.name` | `string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the maintenance window lives in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.schedule

`string` · required

When the window opens, as a cron or rate expression evaluated in
schedule_timezone (e.g. "cron(0 2 ? * SUN *)" for Sundays 02:00).

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.duration

`int32`

How long the window stays open, in hours (1-24).

- rule: {"int32":{"lte":24,"gte":1}}

### spec.cutoff

`int32`

Hours before the window closes when Systems Manager stops
STARTING new tasks (0-23; running tasks keep going or stop per
each task's cutoff_behavior). Must be less than duration.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.description

`string`

Human description of the window (1-128 characters).

- rule: {"string":{"maxLen":"128"}}

### spec.enabled

`bool` · optional (explicit presence)

Whether the window is enabled. Unset = enabled (the provider
default is true) - set an explicit false to create the window
paused.

### spec.allowUnassociatedTargets

`bool`

Allow tasks to run on targets that were never registered with the
window (register-less task targeting). Unset = false: every task
target must be a registered window target.

### spec.scheduleTimezone

`string`

IANA timezone the schedule is evaluated in (e.g.
"America/Los_Angeles"). Unset = UTC.

### spec.scheduleOffset

`int32`

Days to wait after the scheduled cron time before opening the
window (1-6) - e.g. offset 2 on "every Tuesday" opens Thursdays.
Cron schedules only.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":6,"gte":1}}

### spec.startDate

`string`

ISO 8601 timestamp when the window becomes active (e.g.
"2026-09-01T00:00:00-07:00"). Unset = active immediately.

### spec.endDate

`string`

ISO 8601 timestamp when the window stops occurring. Unset = never
expires.

### spec.targets

`[]AwsSsmMaintenanceWindowTargetEntry`

The targets registered with this window, keyed by name. Tasks
reference registered targets by ID (via the target_ids output) or
with the "WindowTargetIds" key.

### spec.targets[].name

`string` · required

Target name, 3-128 characters of letters, digits, hyphens,
underscores, and periods. The for_each key on both engines and
the target_ids output key.

- rule: {"string":{"minLen":"3","maxLen":"128","pattern":"^[0-9A-Za-z_.-]{3,128}$"}}

### spec.targets[].resourceType

`string`

What kind of resources the selectors below name: managed instances
or resource groups.

- rule: {"string":{"in":["INSTANCE","RESOURCE_GROUP"]}}

### spec.targets[].targets

`[]AwsSsmMaintenanceWindowTargetSelector` · required

The selectors (up to 5): instance IDs, tag matches, or
resource-group filters per the SSM targeting grammar.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}

### spec.targets[].targets[].key

`string` · required

The selector key: "InstanceIds", "tag:<key>",
"resource-groups:Name", or "resource-groups:ResourceTypeFilters".

- rule: {"string":{"minLen":"1","maxLen":"163"}}

### spec.targets[].targets[].values

`[]string` · required

The values for the key (instance IDs, the tag value, the resource
group name).

- rule: {"repeated":{"minItems":"1","maxItems":"50","items":{"string":{"minLen":"1"}}}}

### spec.targets[].description

`string` · required

Human description of the target set (3-128 characters). Changing
it replaces the registration.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"3","maxLen":"128"}}

### spec.targets[].ownerInformation

`string` · required

Free-form information the window owner records about the target
(surfaced in task run details).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"128"}}

### spec.tasks

`[]AwsSsmMaintenanceWindowTaskEntry`

The tasks registered with this window, keyed by name.

- rule: invocation must set exactly one of run_command, automation, lambda, and step_functions
- rule: invocation.run_command requires task_type RUN_COMMAND
- rule: invocation.automation requires task_type AUTOMATION
- rule: invocation.lambda requires task_type LAMBDA
- rule: invocation.step_functions requires task_type STEP_FUNCTIONS

### spec.tasks[].name

`string` · required

Task name, 3-128 characters of letters, digits, hyphens,
underscores, and periods. The for_each key on both engines and
the task_ids output key.

- rule: {"string":{"minLen":"3","maxLen":"128","pattern":"^[0-9A-Za-z_.-]{3,128}$"}}

### spec.tasks[].taskType

`string`

What the task runs. Changing it replaces the task registration.

- rule: {"string":{"in":["RUN_COMMAND","AUTOMATION","LAMBDA","STEP_FUNCTIONS"]}}

### spec.tasks[].taskArn

`string | valueFrom` · required

What executes: the document name for RUN_COMMAND/AUTOMATION (an
AWS-managed name as a literal, or an AwsSsmDocument's
status.outputs.document_name by reference), the function ARN for
LAMBDA (an AwsLambda's status.outputs.function_arn), or the state
machine ARN for STEP_FUNCTIONS (an AwsStepFunction's
status.outputs.state_machine_arn). No default kind - the target
kind follows task_type.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.tasks[].description

`string`

Human description of the task (1-128 characters).

- rule: {"string":{"maxLen":"128"}}

### spec.tasks[].serviceRoleArn

`string | valueFrom`

IAM role Systems Manager assumes to run the task. Unset = AWS's
service-linked role for Systems Manager.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.tasks[].priority

`int32`

Execution priority inside the window: 0 is the highest and also
AWS's default; tasks sharing a priority run in parallel.

- rule: {"int32":{"gte":0}}

### spec.tasks[].maxConcurrency

`string`

Maximum targets the task runs on concurrently: an absolute count
("10") or a percentage ("10%"). Only meaningful on a task WITH
targets - AWS rejects rate controls on untargeted tasks, and both
engines omit them when targets are absent.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([1-9][0-9]*|[1-9][0-9]%|[1-9]%|100%)$"}}

### spec.tasks[].maxErrors

`string`

Failures after which the task stops scheduling new target runs:
an absolute count ("1", "0") or a percentage ("10%"). Same
targets-required contract as max_concurrency.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([1-9][0-9]*|[0]|[1-9][0-9]%|[0-9]%|100%)$"}}

### spec.tasks[].cutoffBehavior

`string`

What happens to a RUNNING task instance when the window's cutoff
arrives: keep going (CONTINUE_TASK) or attempt cancellation
(CANCEL_TASK). Unset = CONTINUE_TASK (AWS's default).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CONTINUE_TASK","CANCEL_TASK"]}}

### spec.tasks[].targets

`[]AwsSsmMaintenanceWindowTargetSelector`

What the task runs on: registered window targets (key
"WindowTargetIds") or ad-hoc selectors when the window allows
unassociated targets. A "WindowTargetIds" value may be the NAME of
a target entry declared in this spec - the modules resolve it to
the registration's cloud-generated ID at deploy (the name-based
join); values naming no in-spec entry pass through unchanged for
externally registered target IDs. Unset = an untargeted task (no
rate controls) - legal ONLY for Automation, Lambda, and Step
Functions tasks: AWS rejects an untargeted RUN_COMMAND task
server-side ("you must specify at least one resource as the
target"), so RUN_COMMAND tasks must carry instance IDs or window
target IDs.

- rule: {"repeated":{"maxItems":"5"}}

### spec.tasks[].targets[].key

`string` · required

The selector key: "InstanceIds", "tag:<key>",
"resource-groups:Name", or "resource-groups:ResourceTypeFilters".

- rule: {"string":{"minLen":"1","maxLen":"163"}}

### spec.tasks[].targets[].values

`[]string` · required

The values for the key (instance IDs, the tag value, the resource
group name).

- rule: {"repeated":{"minItems":"1","maxItems":"50","items":{"string":{"minLen":"1"}}}}

### spec.tasks[].invocation

`AwsSsmMaintenanceWindowTaskInvocation`

Type-specific invocation parameters. Unset = the task runs with
the document's/function's own defaults.

### spec.tasks[].invocation.runCommand

`AwsSsmMaintenanceWindowRunCommandInvocation`

RUN_COMMAND parameters.

### spec.tasks[].invocation.runCommand.comment

`string`

Operator comment attached to every command (up to 100 characters).

- rule: {"string":{"maxLen":"100"}}

### spec.tasks[].invocation.runCommand.documentHash

`string`

Pin the document to a content hash so a changed document is
rejected instead of silently run.

- rule: {"string":{"maxLen":"256"}}

### spec.tasks[].invocation.runCommand.documentHashType

`string`

The hash algorithm document_hash uses. Sha1 exists for legacy
documents only - AWS deprecates SHA-1.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Sha256","Sha1"]}}

### spec.tasks[].invocation.runCommand.documentVersion

`string`

The document version to run: "$LATEST", "$DEFAULT", or a concrete
version number.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([$]LATEST|[$]DEFAULT|[1-9][0-9]*)$"}}

### spec.tasks[].invocation.runCommand.outputS3Bucket

`string | valueFrom`

S3 bucket receiving command output.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.tasks[].invocation.runCommand.outputS3KeyPrefix

`string`

Key prefix inside the output bucket.

- rule: {"string":{"maxLen":"500"}}

### spec.tasks[].invocation.runCommand.serviceRoleArn

`string | valueFrom`

IAM role for the Run Command service interactions (distinct from
the task-level role). Unset = the task-level role or AWS's
service-linked role.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.tasks[].invocation.runCommand.timeoutSeconds

`int32`

Seconds each target has to report the command started before AWS
marks it failed (30-2592000).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":2592000,"gte":30}}

### spec.tasks[].invocation.runCommand.parameters

`[]AwsSsmMaintenanceWindowTaskParameter`

Values for the document's input parameters.

### spec.tasks[].invocation.runCommand.parameters[].name

`string` · required

The parameter's name as the document declares it.

- rule: {"string":{"minLen":"1"}}

### spec.tasks[].invocation.runCommand.parameters[].values

`[]string` · required

The parameter's values (list-typed parameters take several).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.tasks[].invocation.runCommand.cloudwatchConfig

`AwsSsmMaintenanceWindowCloudwatchConfig`

CloudWatch Logs delivery for command output.

### spec.tasks[].invocation.runCommand.cloudwatchConfig.cloudwatchLogGroupName

`string`

The log group receiving output. Unset with output enabled = AWS's
default group for the document.

### spec.tasks[].invocation.runCommand.cloudwatchConfig.cloudwatchOutputEnabled

`bool`

Whether CloudWatch output delivery is on.

### spec.tasks[].invocation.runCommand.notificationConfig

`AwsSsmMaintenanceWindowNotificationConfig`

SNS notifications about command lifecycle events.

### spec.tasks[].invocation.runCommand.notificationConfig.notificationArn

`string | valueFrom`

The SNS topic notified.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.tasks[].invocation.runCommand.notificationConfig.notificationEvents

`[]string`

Which lifecycle events notify.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["All","InProgress","Success","TimedOut","Cancelled","Failed"]}}}}

### spec.tasks[].invocation.runCommand.notificationConfig.notificationType

`string`

Notification granularity: one per command, or one per target
invocation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Command","Invocation"]}}

### spec.tasks[].invocation.automation

`AwsSsmMaintenanceWindowAutomationInvocation`

AUTOMATION parameters.

### spec.tasks[].invocation.automation.documentVersion

`string`

The runbook version to run: "$LATEST", "$DEFAULT", or a concrete
version number.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([$]LATEST|[$]DEFAULT|[1-9][0-9]*)$"}}

### spec.tasks[].invocation.automation.parameters

`[]AwsSsmMaintenanceWindowTaskParameter`

Values for the runbook's input parameters.

### spec.tasks[].invocation.automation.parameters[].name

`string` · required

The parameter's name as the document declares it.

- rule: {"string":{"minLen":"1"}}

### spec.tasks[].invocation.automation.parameters[].values

`[]string` · required

The parameter's values (list-typed parameters take several).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.tasks[].invocation.lambda

`AwsSsmMaintenanceWindowLambdaInvocation`

LAMBDA parameters.

### spec.tasks[].invocation.lambda.clientContext

`string` · required

Base64-encoded client context passed to the function.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"8000"}}

### spec.tasks[].invocation.lambda.payload

`string` · sensitive

JSON payload the function is invoked with (up to 4096 bytes).
Sensitive at the provider - payloads routinely carry credentials
or tokens; supply a managed-secret reference when this one does.

- rule: {"string":{"maxLen":"4096"}}

### spec.tasks[].invocation.lambda.qualifier

`string` · required

Function alias or version qualifier to invoke.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"128"}}

### spec.tasks[].invocation.stepFunctions

`AwsSsmMaintenanceWindowStepFunctionsInvocation`

STEP_FUNCTIONS parameters.

### spec.tasks[].invocation.stepFunctions.input

`string` · sensitive

JSON input for the execution (up to 4096 bytes). Sensitive at the
provider - inputs routinely carry credentials or tokens; supply a
managed-secret reference when this one does.

- rule: {"string":{"maxLen":"4096"}}

### spec.tasks[].invocation.stepFunctions.name

`string` · required

Name for the execution.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"80"}}

## Validation Rules

- `spec.target_names_unique`: targets entries must have unique names
- `spec.task_names_unique`: tasks entries must have unique names
- `spec.cutoff_below_duration`: cutoff must be less than duration (AWS's window arithmetic)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSsmMaintenanceWindow, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.window_id` | `string` | The window's AWS-generated ID ("mw-..." - also the provider's import ID). |
| `status.outputs.target_ids` | `map<string, string>` | AWS-generated target registration IDs keyed by target name (each registration imports as "{window_id}/{target_id}"; tasks reference these with the "WindowTargetIds" key). |
| `status.outputs.task_ids` | `map<string, string>` | AWS-generated task registration IDs keyed by task name (each registration imports as "{window_id}/{task_id}"). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.tasks[].serviceRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.tasks[].invocation.runCommand.outputS3Bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.tasks[].invocation.runCommand.serviceRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.tasks[].invocation.runCommand.notificationConfig.notificationArn` | AwsSnsTopic | `status.outputs.topic_arn` |

## See Also

- [Overview](../README.md)
