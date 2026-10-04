# AwsCloudwatchCompositeAlarm

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsCloudwatchCompositeAlarmSpec defines the desired configuration for an
AWS CloudWatch composite alarm.

A composite alarm combines the states of OTHER alarms (metric alarms or
composite alarms) with a boolean rule expression, and takes action only
when the combined expression is true. This is the standard way to:

- **Suppress alert storms** — one composite over "database ALARM AND
  api ALARM" pages once for a shared-cause outage instead of once per
  symptom.
- **Express dependencies** — page on a service's alarm only when its
  upstream dependency is healthy, so the team that can act gets paged.
- **Gate actions on maintenance** — pair with `actions_suppressor` to
  silence actions while a designated suppressor alarm (e.g. a maintenance
  flag) is in ALARM.

The composite alarm evaluates the rule whenever any referenced alarm
changes state — it has no metrics, periods, or thresholds of its own.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchCompositeAlarm
metadata:
  name: test-composite-alarm
  org: test-org
  env: dev
  id: test-composite-alarm-dev
spec:
  region: us-west-2
  alarmRule: 'ALARM("test-cpu-alarm") AND ALARM("test-error-alarm")'
  alarmDescription: "Test composite: both CPU and error alarms breaching"
  alarmActions:
    - value: arn:aws:sns:us-west-2:123456789012:ops-alerts
  actionsSuppressor:
    alarm:
      value: test-maintenance-alarm
    waitPeriod: 60
    extensionPeriod: 120
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.alarmRule` | `string` | yes |  |  |
| `spec.alarmDescription` | `string` |  |  |  |
| `spec.actionsEnabled` | `bool` |  |  |  |
| `spec.alarmActions` | `[]string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.okActions` | `[]string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.insufficientDataActions` | `[]string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.actionsSuppressor` | `AwsCloudwatchCompositeAlarmActionsSuppressor` |  |  |  |
| `spec.actionsSuppressor.alarm` | `string \| valueFrom` | yes |  | AwsCloudwatchAlarm (`status.outputs.alarm_name`) |
| `spec.actionsSuppressor.waitPeriod` | `int32` |  |  |  |
| `spec.actionsSuppressor.extensionPeriod` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.alarmRule

`string` · required

Boolean rule expression over the states of other alarms in the same
account and region. Alarms are addressed BY NAME inside state functions:

  ALARM("my-metric-alarm")                 — true while that alarm is in ALARM
  OK("my-metric-alarm")                    — true while it is in OK
  INSUFFICIENT_DATA("my-metric-alarm")     — true while it lacks data

Functions combine with AND / OR / NOT and parentheses; the constants
TRUE and FALSE are also valid (useful when testing a new composite).

Example: "ALARM(\"cpu-high\") AND ALARM(\"error-rate-high\")"

Compose alarm names from AwsCloudwatchAlarm resources via their exported
`alarm_name` output. Maximum 10240 characters.

- rule: {"required":true,"string":{"maxLen":"10240"}}

### spec.alarmDescription

`string`

Human-readable description of what this composite alarm represents and
what to do when it fires. Maximum 1024 characters.

- rule: {"string":{"maxLen":"1024"}}

### spec.actionsEnabled

`bool` · optional (explicit presence)

Whether actions execute when the composite alarm changes state. When
unset, AWS defaults to true. Unlike metric alarms, this flag is
create-time-only on composite alarms (ForceNew): changing it replaces
the alarm.

### spec.alarmActions

`[]string | valueFrom`

Actions to execute when the composite alarm transitions to ALARM state.
Composite alarms support SNS topic ARNs and Systems Manager OpsItem
actions (not Auto Scaling or EC2 actions — those belong on metric
alarms). Maximum 5 actions.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.okActions

`[]string | valueFrom`

Actions to execute when the composite alarm transitions to OK state.
Maximum 5 actions.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.insufficientDataActions

`[]string | valueFrom`

Actions to execute when the composite alarm transitions to
INSUFFICIENT_DATA state. Maximum 5 actions.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.actionsSuppressor

`AwsCloudwatchCompositeAlarmActionsSuppressor`

Suppresses this composite alarm's actions while a designated suppressor
alarm is in ALARM state — the mechanism for maintenance windows and
deploy freezes. State transitions still happen and are recorded; only
the actions are withheld.

### spec.actionsSuppressor.alarm

`string | valueFrom` · required

The alarm that suppresses actions while it is in ALARM state, addressed
by alarm NAME (the CloudWatch API contract). Reference an
AwsCloudwatchAlarm's exported `alarm_name` output, or provide a literal
alarm name.

- references: AwsCloudwatchAlarm (`status.outputs.alarm_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchAlarm, name: <that resource's name>, fieldPath: status.outputs.alarm_name}} -- a bare string does not parse

### spec.actionsSuppressor.waitPeriod

`int32`

Maximum time (in seconds) the composite alarm waits for the suppressor
alarm to enter ALARM state after the composite itself transitions,
before concluding the suppressor is not firing and executing actions.

AWS requires this field whenever a suppressor is configured (the
PutCompositeAlarm contract) — both engines always send it with the
suppressor. 0 means the composite acts immediately unless the suppressor
is already in ALARM.

- rule: {"int32":{"gte":0}}

### spec.actionsSuppressor.extensionPeriod

`int32`

Maximum time (in seconds) actions remain suppressed AFTER the suppressor
alarm leaves ALARM state — a grace window that avoids acting on
transitions that occur while the suppression is winding down.

AWS requires this field whenever a suppressor is configured (the
PutCompositeAlarm contract) — both engines always send it with the
suppressor. 0 ends suppression the moment the suppressor recovers.

- rule: {"int32":{"gte":0}}

## Validation Rules

- `alarm_actions_max_5`: maximum 5 alarm_actions allowed
- `ok_actions_max_5`: maximum 5 ok_actions allowed
- `insufficient_data_actions_max_5`: maximum 5 insufficient_data_actions allowed

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudwatchCompositeAlarm, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.alarm_arn` | `string` | The Amazon Resource Name (ARN) of the composite alarm. |
| `status.outputs.alarm_name` | `string` | The name of the composite alarm. Unique within the AWS account and region; this is how other composite alarms reference it in their rule expressions. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.alarmActions` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.okActions` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.insufficientDataActions` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.actionsSuppressor.alarm` | AwsCloudwatchAlarm | `status.outputs.alarm_name` |

## See Also

- [Overview](../README.md)
