# AWS CloudWatch Composite Alarm

Deploys a CloudWatch composite alarm — a boolean rule over the states of OTHER alarms (metric alarms from [AwsCloudwatchAlarm](/infra-catalog/aws-cloudwatch-alarm), or other composites) that acts only when the combined expression is true. It has no metrics, periods, or thresholds of its own; it re-evaluates whenever any referenced alarm changes state. Composites are the standard way to suppress alert storms (one page for a shared-cause outage instead of one per symptom), express dependencies (page the team that can actually act), and gate paging on maintenance windows via the actions suppressor.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CloudWatch Composite Alarm** -- the rule expression, evaluated against the referenced alarms' states in the same account and region
- **State-Transition Actions** -- up to 5 actions each for ALARM, OK, and INSUFFICIENT_DATA transitions (SNS topics and SSM OpsItems; Auto Scaling and EC2 actions are metric-alarm-only)
- **Actions Suppressor Binding** -- optional: a designated alarm whose ALARM state withholds this composite's actions, with wait/extension windows
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Constituent alarms** -- the [AwsCloudwatchAlarm](/infra-catalog/aws-cloudwatch-alarm) resources the rule references, deployed first so their `alarm_name` outputs exist. The notification path usually starts with an [AwsSnsTopic](/infra-catalog/aws-sns-topic).

### AWS Account

- **Same account and region** -- the rule addresses alarms by bare name; every referenced alarm must live in the composite's own account and region.
- **Alarm names are the contract** -- renaming a constituent alarm breaks the rule silently (it evaluates INSUFFICIENT_DATA); reference alarm names through resource outputs to keep wiring rename-proof.

## Deploy

### Console

Open the deployment store, find **AWS CloudWatch Composite Alarm**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Shared-Cause Outage Page** preset in the [Presets](#presets) tab to pre-populate the storm-suppression shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchCompositeAlarm
metadata:
  name: orders-shared-cause
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  alarmRule: ALARM("orders-db-cpu-high") AND ALARM("orders-api-error-rate-high")
  alarmDescription: Orders stack shared-cause outage — page platform on-call.
  alarmActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: oncall-pager
        fieldPath: status.outputs.topic_arn
  okActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: oncall-pager
        fieldPath: status.outputs.topic_arn
```

```shell
planton apply -f composite-alarm.yaml
```

This pages once when the database drags the API down — instead of once per symptom. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the metric alarms and SNS topic deploy first, then the composite that references them:

```yaml
# A maintenance-suppressed composite:
spec:
  alarmRule: ALARM("orders-db-cpu-high") AND ALARM("orders-api-error-rate-high")
  actionsSuppressor:
    alarm:
      valueFrom:
        kind: AwsCloudwatchAlarm
        name: orders-maintenance-flag
        fieldPath: status.outputs.alarm_name
    waitPeriod: 60
    extensionPeriod: 120
```

## Key Configuration

These are the most important decisions when configuring a composite alarm. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The rule is the resource** -- `ALARM("name")`, `OK("name")`, and `INSUFFICIENT_DATA("name")` state functions combined with AND / OR / NOT and parentheses (max 10,240 characters). `ALARM(a) AND ALARM(b)` suppresses storms; `ALARM(svc) AND OK(upstream)` expresses dependencies; the constants TRUE/FALSE let you exercise wiring before the real rule lands.

**The actions switch is create-time-only** -- unlike metric alarms, changing `actions_enabled` on a composite replaces the alarm (same name, new ARN). Leave it unset to keep AWS's default (on) without recording an opinion; for planned quiet windows use the suppressor instead.

**The suppressor is the maintenance mechanism** -- while the designated alarm is in ALARM, this composite's actions are withheld; transitions still record. The wait period bounds how long the composite waits for the suppressor to fire after its own transition; the extension period keeps suppression up while things settle after the window closes.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSnsTopic** (optional) | `alarmActions` | `status.outputs.topic_arn` |
| **AwsSnsTopic** (optional) | `okActions` | `status.outputs.topic_arn` |
| **AwsSnsTopic** (optional) | `insufficientDataActions` | `status.outputs.topic_arn` |
| **AwsCloudwatchAlarm** (optional) | `actionsSuppressor.alarm` | `status.outputs.alarm_name` |

The rule expression itself composes alarm names as plain strings — typically taken from constituent alarms' `alarm_name` outputs.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `alarm_name` | The composite's name (unique per account and region) | Parent composites' rule expressions; another composite's suppressor; a Route 53 health check's CloudWatch arm |
| `alarm_arn` | Amazon Resource Name of the composite alarm | IAM policies; EventBridge alarm-state-change patterns |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Shared-cause outage** -- one composite over the database and API alarms; one page instead of five. Start from the **Shared-Cause Outage Page** preset.

**Maintenance-suppressed paging** -- the same composite with an actions suppressor bound to a maintenance-flag alarm the deploy pipeline flips. Start from the **Maintenance-Suppressed Paging** preset.

## Works With

- [**AWS CloudWatch Alarm**](/infra-catalog/aws-cloudwatch-alarm) -- the constituent alarms the rule composes, and the suppressor flag
- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) -- where ALARM / OK / INSUFFICIENT_DATA notifications go
- [**AWS Route 53 Health Check**](/infra-catalog/aws-route53-health-check) -- a CLOUDWATCH_METRIC health check can mirror an alarm's state into DNS failover
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- where the metrics feeding the constituent alarms often originate
