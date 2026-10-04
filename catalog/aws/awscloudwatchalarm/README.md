# AwsCloudwatchAlarm

A **CloudWatch metric alarm** watches a metric (or a computed signal) and performs one or more actions when the value crosses a threshold for a sustained number of evaluation periods. It is the fundamental building block for AWS-native monitoring, driving notifications, auto-scaling decisions, and automated remediation.

The alarm supports three metric definition modes (exactly one per alarm): **simple metric** (a single metric + statistic), **metric queries** (metric math, anomaly detection bands, multi-metric and cross-account evaluation), and **PromQL** (a query against an Amazon Managed Service for Prometheus workspace, where the query itself expresses the condition).

## When to Use

- **Threshold-based alerting** — Monitor a standard AWS metric (CPU, memory, error count, queue depth) and notify an SNS topic when it breaches a static threshold.
- **Anomaly detection** — Use machine-learning-based anomaly detection bands instead of hard-coded thresholds for metrics with seasonal or unpredictable patterns.
- **Metric math alerting** — Combine multiple metrics into a single alarm (e.g., error rate = 5xx errors / total requests * 100) using metric math expressions.
- **Prometheus-native alerting** — Alarm on PromQL queries against a Prometheus workspace with Prometheus-style pending/recovery semantics, while keeping CloudWatch's action routing.
- **M-of-N evaluation** — Reduce false positives by requiring only M breaching data points within the last N evaluation periods before transitioning to ALARM state.
- **Auto-scaling triggers** — Drive EC2 Auto Scaling policies based on metric thresholds (though most users prefer Target Tracking policies for scaling).

## When NOT to Use

- For **combining the states of multiple alarms** with boolean logic — use `AwsCloudwatchCompositeAlarm`, which pages once for shared-cause outages and supports maintenance suppression.
- For **dashboard widgets** or visual monitoring — use CloudWatch Dashboards (separate resource).
- For **third-party monitoring** (Datadog, Grafana Cloud, PagerDuty native) where CloudWatch is not the primary alerting plane.
- For **log-based alerting** — attach a metric filter on the `AwsCloudwatchLogGroup` to extract a metric from log events, then alarm on that metric here.

## Prerequisites

- An AWS account and region configured in your Planton IaC input.
- (Optional) One or more SNS topics if you want alarm state transitions to trigger notifications. Reference them via `value` (literal ARN) or `valueFrom` (AwsSnsTopic resource).
- (Optional) A metric already being published — CloudWatch will enter INSUFFICIENT_DATA state if the metric does not yet exist.

## Quick Start

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchAlarm
metadata:
  name: high-cpu
  org: acme
  env: prod
  id: high-cpu-prod
spec:
  comparisonOperator: GreaterThanThreshold
  evaluationPeriods: 3
  threshold: 80.0
  metricName: CPUUtilization
  namespace: AWS/EC2
  period: 300
  statistic: Average
  dimensions:
    InstanceId: i-0abc123def456789a
  alarmActions:
    - value: arn:aws:sns:us-east-1:123456789012:ops-alerts
```

## Spec Reference

### Alarm Evaluation

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `comparisonOperator` | string | Conditional | — | Arithmetic operation to compare the statistic against the threshold. Required for simple-metric and metric-query alarms; must be omitted for PromQL alarms. **Standard:** `GreaterThanOrEqualToThreshold`, `GreaterThanThreshold`, `LessThanThreshold`, `LessThanOrEqualToThreshold`. **Anomaly detection:** `LessThanLowerOrGreaterThanUpperThreshold`, `LessThanLowerThreshold`, `GreaterThanUpperThreshold`. |
| `evaluationPeriods` | int32 | Conditional | — | Number of consecutive periods over which the metric is compared. Required (>= 1) for simple-metric and metric-query alarms; must be omitted for PromQL alarms. Combined with `datapointsToAlarm` for M-of-N evaluation. |
| `datapointsToAlarm` | int32 | No | Same as `evaluationPeriods` | Number of data points within the evaluation window that must breach to trigger ALARM. Must be <= `evaluationPeriods`. Use for M-of-N patterns to reduce false positives. |
| `threshold` | double | Conditional | — | Static threshold value. Required for static threshold alarms. Mutually exclusive with `thresholdMetricId`. |
| `thresholdMetricId` | string | Conditional | — | ID of the `ANOMALY_DETECTION_BAND` function in `metricQueries`. Used for anomaly detection alarms. Mutually exclusive with `threshold`. Max 255 chars. |
| `treatMissingData` | string | No | `"missing"` | How missing data points are treated during evaluation. Valid values: `missing`, `notBreaching`, `breaching`, `ignore`. |
| `actionsEnabled` | bool | No | `true` (AWS default) | Whether actions execute during state transitions. Unset lets AWS default to enabled; an explicit `false` suppresses actions during maintenance or tuning while the alarm keeps evaluating. |

### Simple Metric Mode

Use these fields for single-metric alarms. **Mutually exclusive** with `metricQueries`.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `metricName` | string | Conditional | — | CloudWatch metric name (e.g., `CPUUtilization`, `5XXError`). When set, `namespace`, `period`, and one of `statistic`/`extendedStatistic` are required. Max 255 chars. |
| `namespace` | string | Conditional | — | Metric namespace (e.g., `AWS/EC2`, `AWS/ECS`, `AWS/ApplicationELB`). Required when `metricName` is set. Max 255 chars. |
| `period` | int32 | Conditional | — | Period in seconds for statistic aggregation. Valid values: `10`, `20`, `30`, or any multiple of 60. High-resolution metrics support 10/20/30s periods. Required when `metricName` is set. |
| `statistic` | string | Conditional | — | Standard statistic: `SampleCount`, `Average`, `Sum`, `Minimum`, `Maximum`. Mutually exclusive with `extendedStatistic`. |
| `extendedStatistic` | string | Conditional | — | Percentile or extended statistic (e.g., `p95`, `p99`, `p99.9`, `IQM`, `TM(10%:90%)`). Mutually exclusive with `statistic`. |
| `dimensions` | map\<string, string\> | No | — | Key-value pairs that identify the specific metric stream. Example: `{"InstanceId": "i-0abc..."}`. |
| `unit` | string | No | — | Metric unit filter. When set, only data points matching this unit are evaluated. Most alarms omit this. |

### Metric Query Mode

Use `metricQueries` for metric math, anomaly detection, or multi-metric alarms. **Mutually exclusive** with simple metric fields and `evaluationCriteria`.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `metricQueries` | repeated MetricQuery | Conditional | — | Up to 20 metric queries. Exactly one must set `returnData: true`; each query sets exactly one of `expression`/`metric`. See [Nested Messages](#nested-messages) below. |

### PromQL Mode

Use `evaluationCriteria` to alarm on a PromQL query against an Amazon Managed Service for Prometheus workspace. **Mutually exclusive** with the other two modes — and because the query itself expresses the alarm condition, the threshold fields (`comparisonOperator`, `evaluationPeriods`, `threshold`, `datapointsToAlarm`, and the simple-metric fields) must be omitted.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `evaluationCriteria.promqlCriteria.query` | string | **Yes** | — | The PromQL query, producing a boolean-style firing signal (e.g. `avg(rate(node_cpu_seconds_total{mode!="idle"}[5m])) > 0.8`). Max 10000 chars. |
| `evaluationCriteria.promqlCriteria.pendingPeriod` | int32 | No | AWS default | Seconds the query must continuously fire before ALARM — the PromQL analog of Prometheus `for:`. 0–86400. |
| `evaluationCriteria.promqlCriteria.recoveryPeriod` | int32 | No | AWS default | Seconds the query must continuously be quiet before returning to OK (dampens flapping). 0–86400. |
| `evaluationInterval` | int32 | No | AWS default | How often (seconds) the query runs. Valid: `10`, `20`, `30`, or any multiple of 60. Only valid with `evaluationCriteria`. |

### Actions

Each action field accepts up to 5 entries. Each entry is a `StringValueOrRef` with `default_kind = AwsSnsTopic`.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `alarmActions` | repeated StringValueOrRef | No | — | Actions for ALARM state transitions. Typically SNS topic ARNs. Max 5. |
| `okActions` | repeated StringValueOrRef | No | — | Actions for OK state transitions. Max 5. |
| `insufficientDataActions` | repeated StringValueOrRef | No | — | Actions for INSUFFICIENT_DATA state transitions. Max 5. |

### Description and Percentile Handling

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `alarmDescription` | string | No | — | Human-readable description. Include what the alarm monitors, expected thresholds, and remediation steps. Max 1024 chars. |
| `evaluateLowSampleCountPercentiles` | string | No | — | Behavior during low-sample-count periods for percentile statistics. `evaluate` (always evaluate) or `ignore` (maintain current state). |

## Nested Messages

### AwsCloudwatchAlarmMetricQuery

A single query within metric math mode. Each query is either a raw metric retrieval or an expression that combines other queries.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `id` | string | **Yes** | — | Unique identifier, used as a variable name in expressions. Must start with a lowercase letter; valid chars: `[a-z0-9_]`. Max 255 chars. |
| `expression` | string | Conditional | — | Metric math expression (e.g., `m1/m2*100`) or Metrics Insights query. Mutually exclusive with `metric`. Max 1024 chars. |
| `metric` | MetricQueryMetric | Conditional | — | Raw metric definition. Mutually exclusive with `expression`. |
| `label` | string | No | — | Human-readable label displayed in the CloudWatch console. |
| `period` | int32 | No | Inherited | Override period in seconds for this query. Valid: `1`, `5`, `10`, `20`, `30`, or any multiple of 60. |
| `returnData` | bool | No | `false` | Whether this query's result is the alarm's evaluation signal. Exactly one query must set this to `true`. |
| `accountId` | string | No | — | AWS account ID for cross-account metric monitoring. Max 255 chars. |

### AwsCloudwatchAlarmMetricQueryMetric

Raw CloudWatch metric definition within a metric query.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `metricName` | string | **Yes** | — | CloudWatch metric name (e.g., `CPUUtilization`, `5XXError`). Max 255 chars. |
| `namespace` | string | No | — | Metric namespace (e.g., `AWS/EC2`, `AWS/ApplicationELB`). Optional in the CloudWatch API but virtually always needed to address a metric unambiguously. Max 255 chars. |
| `period` | int32 | **Yes** | — | Period in seconds. High-resolution: `1`, `5`, `10`, `20`, `30`. Standard: multiples of 60. |
| `stat` | string | **Yes** | — | Statistic to apply. Standard (`Average`, `Sum`, etc.) or extended (`p95`, `p99.9`, `IQM`). |
| `dimensions` | map\<string, string\> | No | — | Dimensions to narrow the metric stream. |
| `unit` | string | No | — | Unit filter for data point selection. |

## Outputs

| Output | Description |
|--------|-------------|
| `alarm_arn` | The Amazon Resource Name (ARN) of the metric alarm. Use to reference this alarm in composite alarms, dashboards, and operational tooling. |
| `alarm_name` | The name of the metric alarm, unique within the AWS account and region. Useful for CloudWatch API calls, CLI operations, and dashboard widgets. |

## What Is Deliberately Omitted (v1)

- **Composite alarms** — Combining the states of multiple alarms with boolean logic is its own first-class kind, `AwsCloudwatchCompositeAlarm` (independent AWS identity and lifecycle; compose by this alarm's exported `alarm_name`).
- **Dashboard integration** — CloudWatch Dashboards are a distinct resource type. Alarm widgets on dashboards reference the alarm ARN from `status.outputs.alarm_arn`.
- **Extended statistic regex validation** — The `extendedStatistic` field accepts freeform strings (`p95`, `TM(10%:90%)`, etc.). The full pattern grammar (trimmed means, winsorized ranges) is validated by the AWS API at deploy time.
- **Tags** — Tags are automatically derived from `metadata` fields (org, env, resource kind, resource ID); custom user tags are a platform-wide concern, not per-kind scope.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
