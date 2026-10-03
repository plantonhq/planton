# AwsCloudwatchCompositeAlarm

A **CloudWatch composite alarm** combines the states of OTHER alarms (metric alarms or composite alarms) with a boolean rule expression, and takes action only when the combined expression is true. It has no metrics, periods, or thresholds of its own — it re-evaluates its rule whenever any referenced alarm changes state.

## When to Use

- **Suppress alert storms** — Page once for a shared-cause outage (`ALARM("db") AND ALARM("api")`) instead of once per symptom.
- **Dependency-aware paging** — Page on a service's alarm only when its upstream dependency is healthy (`ALARM("api") AND NOT ALARM("upstream-db")`), so the team that can act gets paged.
- **Maintenance suppression** — Pair with `actionsSuppressor` to silence actions while a designated maintenance-flag alarm is in ALARM; state transitions still record, only paging is withheld.
- **Incident-level signals** — Build one "the product is down" alarm from independent service-level alarms.

## When NOT to Use

- To watch a metric directly — use `AwsCloudwatchAlarm`; the composite only combines existing alarms.
- For notification routing/dedup logic beyond boolean state combination — that belongs in the paging tool (PagerDuty, Opsgenie) the SNS action feeds.

## Prerequisites

- One or more existing alarms (`AwsCloudwatchAlarm` or other composites) in the same account and region — the rule references them by NAME.
- (Optional) SNS topics for actions.

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | — | AWS region (must match the referenced alarms). |
| `alarmRule` | string | Yes | — | Boolean expression over other alarms' states: `ALARM("name")`, `OK("name")`, `INSUFFICIENT_DATA("name")` combined with `AND`/`OR`/`NOT` and parentheses; `TRUE`/`FALSE` constants allowed. Max 10240 chars. |
| `alarmDescription` | string | No | — | What this composite represents and what to do when it fires. Max 1024 chars. |
| `actionsEnabled` | bool | No | `true` (AWS default) | Whether actions execute on state transitions. **ForceNew on this resource** — changing it replaces the alarm. |
| `alarmActions[]` | StringValueOrRef | No | — | Actions on ALARM (SNS topics, SSM OpsItems — not Auto Scaling/EC2 actions). Max 5. |
| `okActions[]` | StringValueOrRef | No | — | Actions on OK. Max 5. |
| `insufficientDataActions[]` | StringValueOrRef | No | — | Actions on INSUFFICIENT_DATA. Max 5. |
| `actionsSuppressor.alarm` | StringValueOrRef | With suppressor | — | The alarm whose ALARM state suppresses actions, by alarm NAME (reference an AwsCloudwatchAlarm's `alarm_name` output). |
| `actionsSuppressor.waitPeriod` | int32 | No | 0 | Seconds to wait for the suppressor to enter ALARM after this composite transitions, before acting. |
| `actionsSuppressor.extensionPeriod` | int32 | No | 0 | Seconds actions stay suppressed after the suppressor leaves ALARM. |

## Outputs

| Output | Description |
|--------|-------------|
| `alarm_arn` | ARN of the composite alarm. |
| `alarm_name` | Name of the composite alarm — the join key parent composites use in their rule expressions. |

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchCompositeAlarm
metadata:
  name: shared-cause
spec:
  region: us-west-2
  alarmRule: 'ALARM("cpu-high") AND ALARM("error-rate-high")'
  alarmActions:
    - value: arn:aws:sns:us-west-2:123456789012:ops-critical
```

## Composing Alarm Names

The rule addresses alarms by their CloudWatch names. `AwsCloudwatchAlarm` resources name their alarms after `metadata.name` and export it as the `alarm_name` output — reference that output when wiring the suppressor, and use the known names inside `alarmRule` text.

## What Is Deliberately Omitted (v1)

- **Rule-expression parsing/validation** — The rule grammar is validated by the CloudWatch API at deploy time; the spec validates length only. Alarm references inside the expression compose by name, not by `StringValueOrRef` (the same class as metric-math expression variables).
- **Tags** — Identity tags derive from `metadata`; custom user tags are a platform-wide concern.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
