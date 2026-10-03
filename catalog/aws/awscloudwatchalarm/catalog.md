# AWS CloudWatch Alarm

Deploys a CloudWatch alarm that watches a single metric, a metric math expression, or a PromQL query against an Amazon Managed Prometheus workspace, and triggers actions on breach. The three evaluation modes are mutually exclusive -- simple metric for single-metric thresholds, metric queries for computed expressions and anomaly detection, PromQL for Prometheus-native alerting -- and M-of-N evaluation windows guard every mode against paging on transient spikes.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CloudWatch Metric Alarm** -- a metric alarm configured with the specified comparison operator, evaluation periods, threshold (static or anomaly detection band), and optional M-of-N datapoints-to-alarm evaluation
- **Alarm Actions** -- created only when `alarmActions` entries are provided; executes the specified ARNs (typically SNS topics) when the alarm transitions to ALARM state
- **OK Actions** -- created only when `okActions` entries are provided; executes the specified ARNs when the alarm transitions to OK state
- **Insufficient Data Actions** -- created only when `insufficientDataActions` entries are provided; executes the specified ARNs when the alarm transitions to INSUFFICIENT_DATA state
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An SNS topic** (optional) -- required when configuring alarm, OK, or insufficient data actions. Provide the topic ARN directly or reference an AwsSnsTopic Infra Component via ValueFromRef.
- **A CloudWatch metric** -- the alarm monitors an existing metric in CloudWatch. Ensure the metric is being published by the target AWS service or custom application before creating the alarm.

## Deploy

### Console

Open the deployment store, find **AWS CloudWatch Alarm**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **CPU Utilization Alarm** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchAlarm
metadata:
  name: high-cpu-alarm
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  comparisonOperator: GreaterThanOrEqualToThreshold
  evaluationPeriods: 3
  datapointsToAlarm: 2
  threshold: 80
  metricName: CPUUtilization
  namespace: "AWS/EC2"
  period: 300
  statistic: Average
  treatMissingData: breaching
  alarmActions:
    - value: "arn:aws:sns:us-west-2:123456789012:alerts"
```

```shell
planton apply -f cloudwatch-alarm.yaml
```

This creates a CloudWatch alarm monitoring EC2 CPU utilization with 2-of-3 M-of-N evaluation. No OK or insufficient data actions are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the alarm to an SNS topic deployed in the same InfraPipeline:

```yaml
spec:
  alarmActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: alerts-topic
        fieldPath: status.outputs.topic_arn
  okActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: recovery-topic
        fieldPath: status.outputs.topic_arn
```

The InfraPipeline resolves the dependency graph, deploys the SNS topics first, then provisions the alarm with the resolved ARNs.

## Key Configuration

These are the most important decisions when configuring a CloudWatch alarm. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Simple metric vs. metric queries vs. PromQL** -- Use simple metric mode (`metricName`, `namespace`, `period`, `statistic`, optional `unit`) for single-metric alarms like CPU utilization or queue depth. Use `metricQueries` for metric math (error rate = errors / total * 100), anomaly detection (ANOMALY_DETECTION_BAND), or multi-metric evaluations -- exactly one query carries `returnData: true` as the alarmed signal. Use `evaluationCriteria.promqlCriteria` to alarm on a PromQL query against an Amazon Managed Prometheus workspace; the query carries its own comparison, so the threshold fields do not apply, and `evaluationInterval` plus the pending/recovery periods control cadence and flap-guarding. The three modes are mutually exclusive.

**M-of-N evaluation** -- Set `datapointsToAlarm` lower than `evaluationPeriods` to require only M of N periods to breach before triggering. For example, `evaluationPeriods: 5` with `datapointsToAlarm: 3` means 3 of the last 5 periods must breach. This reduces false positives from transient spikes.

**Missing data treatment** -- Controls alarm behavior when data points are absent. Use `notBreaching` for intermittent metrics (low-traffic services) to avoid false alarms during idle periods. Use `breaching` for metrics that must always report (heartbeat checks). Defaults to `missing` which preserves the current alarm state.

**Comparison operator** -- Standard operators (`GreaterThanThreshold`, `LessThanOrEqualToThreshold`, etc.) for static thresholds. Anomaly detection operators (`LessThanLowerOrGreaterThanUpperThreshold`) for dynamic thresholds using `thresholdMetricId`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSnsTopic** (optional) | `alarmActions` | `status.outputs.topic_arn` |
| **AwsSnsTopic** (optional) | `okActions` | `status.outputs.topic_arn` |
| **AwsSnsTopic** (optional) | `insufficientDataActions` | `status.outputs.topic_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `alarm_arn` | Amazon Resource Name of the metric alarm | Composite alarms, dashboards, operational tooling |
| `alarm_name` | Name of the metric alarm (unique per account/region) | CloudWatch API calls, CLI operations, dashboard widgets |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CPU utilization alarm** -- Single-metric alarm on EC2 CPUUtilization with 2-of-3 M-of-N evaluation and SNS notification. The standard monitoring pattern for compute workloads. Start from the **CPU Utilization Alarm** preset.

**Error rate metric math** -- Three-query metric math alarm computing error rate as a percentage (errors / requests * 100). Suitable for ALB, API Gateway, and any ratio-based alerting where raw counts are misleading. Start from the **Error Rate Metric Math Alarm** preset.

**Production multi-action alarm** -- Full lifecycle alerting with separate SNS topics for alarm, recovery, and insufficient data state transitions. Recommended for any production workload where knowing about recovery is as important as knowing about the incident. Start from the **Production Multi-Action Alarm** preset.

**PromQL Prometheus alarm** -- Alarms on a PromQL query evaluated against an Amazon Managed Prometheus workspace, with pending/recovery flap guards and an explicit evaluation interval. The bridge between Prometheus-native monitoring and CloudWatch's alarm-state and SNS action machinery. Start from the **PromQL Alarm on a Prometheus Workspace** preset.

## Works With

- [**AWS SNS Topic**](/infra-catalog/aws-sns-topic) -- provides notification targets for alarm, OK, and insufficient data state transitions