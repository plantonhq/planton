# AwsCloudwatchAlarm

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsCloudwatchAlarmSpec defines the desired configuration for an AWS
CloudWatch metric alarm.

A CloudWatch alarm watches a metric (or a computed signal) and performs one
or more actions when the value crosses a threshold for a sustained number of
evaluation periods.

Three metric definition modes are supported (mutually exclusive — exactly
one must be used):

1. **Simple metric mode** — Set `metric_name`, `namespace`, `period`, and
   one of `statistic` or `extended_statistic`. Suitable for alarms on a
   single metric like CPUUtilization or RequestCount.

2. **Metric query mode** — Use `metric_queries` for metric math expressions,
   anomaly detection, or multi-metric alarms. Up to 20 metric queries can
   be combined. Exactly one query must set `return_data = true` to serve as
   the alarm's evaluation signal.

3. **PromQL mode** — Use `evaluation_criteria` to alarm on a PromQL query
   against an Amazon Managed Service for Prometheus workspace. In this mode
   the alerting contract lives in the query itself, so the threshold-based
   fields (`comparison_operator`, `evaluation_periods`, `threshold`, the
   simple-metric fields) must NOT be set.

Actions are specified as repeated StringValueOrRef fields for alarm, OK, and
insufficient-data state transitions. The most common action target is an SNS
topic, so `default_kind = AwsSnsTopic` is set for convenience.

Credentials, region, and deployment workflow live outside this spec in IaC
inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchAlarm
metadata:
  name: test-cpu-alarm
  org: test-org
  env: dev
  id: test-cpu-alarm-dev
spec:
  region: us-west-2
  comparisonOperator: GreaterThanThreshold
  evaluationPeriods: 3
  datapointsToAlarm: 2
  threshold: 80.0
  metricName: CPUUtilization
  namespace: AWS/EC2
  period: 300
  statistic: Average
  dimensions:
    InstanceId: i-1234567890abcdef0
  treatMissingData: breaching
  alarmDescription: "CPU utilization exceeds 80% for 2 of 3 evaluation periods"
  alarmActions:
    - value: arn:aws:sns:us-east-1:123456789012:ops-alerts
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.comparisonOperator` | `string` |  |  |  |
| `spec.evaluationPeriods` | `int32` |  |  |  |
| `spec.datapointsToAlarm` | `int32` |  |  |  |
| `spec.threshold` | `double` |  |  |  |
| `spec.thresholdMetricId` | `string` |  |  |  |
| `spec.treatMissingData` | `string` |  |  |  |
| `spec.actionsEnabled` | `bool` |  |  |  |
| `spec.metricName` | `string` |  |  |  |
| `spec.namespace` | `string` |  |  |  |
| `spec.period` | `int32` |  |  |  |
| `spec.statistic` | `string` |  |  |  |
| `spec.extendedStatistic` | `string` |  |  |  |
| `spec.dimensions` | `map<string, string>` |  |  |  |
| `spec.unit` | `string` |  |  |  |
| `spec.metricQueries` | `[]AwsCloudwatchAlarmMetricQuery` |  |  |  |
| `spec.metricQueries[].id` | `string` | yes |  |  |
| `spec.metricQueries[].expression` | `string` |  |  |  |
| `spec.metricQueries[].metric` | `AwsCloudwatchAlarmMetricQueryMetric` |  |  |  |
| `spec.metricQueries[].metric.metricName` | `string` | yes |  |  |
| `spec.metricQueries[].metric.namespace` | `string` |  |  |  |
| `spec.metricQueries[].metric.period` | `int32` |  |  |  |
| `spec.metricQueries[].metric.stat` | `string` | yes |  |  |
| `spec.metricQueries[].metric.dimensions` | `map<string, string>` |  |  |  |
| `spec.metricQueries[].metric.unit` | `string` |  |  |  |
| `spec.metricQueries[].label` | `string` |  |  |  |
| `spec.metricQueries[].period` | `int32` |  |  |  |
| `spec.metricQueries[].returnData` | `bool` |  |  |  |
| `spec.metricQueries[].accountId` | `string` |  |  |  |
| `spec.alarmActions` | `[]string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.okActions` | `[]string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.insufficientDataActions` | `[]string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.alarmDescription` | `string` |  |  |  |
| `spec.evaluateLowSampleCountPercentiles` | `string` |  |  |  |
| `spec.evaluationCriteria` | `AwsCloudwatchAlarmEvaluationCriteria` |  |  |  |
| `spec.evaluationCriteria.promqlCriteria` | `AwsCloudwatchAlarmPromqlCriteria` | yes |  |  |
| `spec.evaluationCriteria.promqlCriteria.query` | `string` | yes |  |  |
| `spec.evaluationCriteria.promqlCriteria.pendingPeriod` | `int32` |  |  |  |
| `spec.evaluationCriteria.promqlCriteria.recoveryPeriod` | `int32` |  |  |  |
| `spec.evaluationInterval` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.comparisonOperator

`string`

Arithmetic operation to compare the statistic against the threshold.

Standard comparison operators:
- "GreaterThanOrEqualToThreshold"
- "GreaterThanThreshold"
- "LessThanThreshold"
- "LessThanOrEqualToThreshold"

Anomaly detection operators (use with metric_queries and threshold_metric_id):
- "LessThanLowerOrGreaterThanUpperThreshold"
- "LessThanLowerThreshold"
- "GreaterThanUpperThreshold"

Required for simple-metric and metric-query alarms. Must NOT be set for
PromQL alarms (`evaluation_criteria`) — the query expresses the condition.

### spec.evaluationPeriods

`int32`

Number of consecutive periods over which the metric is compared to the
threshold.

Combined with `datapoints_to_alarm` this creates an M-of-N evaluation
window. For example, evaluation_periods=5 with datapoints_to_alarm=3
means "3 out of the last 5 periods must breach to trigger the alarm."

Required (must be at least 1) for simple-metric and metric-query alarms.
Must NOT be set for PromQL alarms — pending/recovery periods on the
PromQL criteria play that role instead.

### spec.datapointsToAlarm

`int32`

Number of data points within the evaluation window that must breach the
threshold to trigger the alarm. Must be less than or equal to
`evaluation_periods`. When omitted, defaults to `evaluation_periods`
(every period must breach).

Use a value lower than `evaluation_periods` for M-of-N evaluation to
reduce false positives caused by transient spikes.

### spec.threshold

`double`

Static threshold value. The statistic (or metric math result) is compared
against this value using `comparison_operator`.

Required for static threshold alarms. Must not be set for anomaly
detection alarms (use `threshold_metric_id` instead) or PromQL alarms.

### spec.thresholdMetricId

`string`

For anomaly detection alarms, set this to the ID of the
ANOMALY_DETECTION_BAND function defined in `metric_queries`. The anomaly
band serves as a dynamic threshold instead of a static `threshold` value.

Mutually exclusive with `threshold`.

- rule: {"string":{"maxLen":"255"}}

### spec.treatMissingData

`string`

How the alarm treats missing data points during evaluation.

Valid values:
- "missing"      — (Default) Missing data is treated as missing; the alarm
                    state does not change. Can cause delayed transitions.
- "notBreaching" — Missing data is within the threshold. Best for alarms
                    on intermittent metrics (e.g., error count on low-traffic
                    services) to avoid false alarms during idle periods.
- "breaching"    — Missing data is treated as breaching. Use for metrics
                    that must always report (e.g., heartbeat checks).
- "ignore"       — The current alarm state is maintained regardless of
                    missing data.

### spec.actionsEnabled

`bool` · optional (explicit presence)

Whether actions execute during alarm state transitions. When unset, AWS
defaults to true (actions enabled). Set explicitly to false to suppress
actions during maintenance windows or alarm tuning — the alarm still
evaluates and changes state, it just does not act.

### spec.metricName

`string`

Name of the CloudWatch metric to alarm on (e.g., "CPUUtilization",
"4xxErrorRate", "ApproximateNumberOfMessagesVisible").

When set, `namespace`, `period`, and one of `statistic` or
`extended_statistic` must also be set.

Mutually exclusive with `metric_queries` and `evaluation_criteria`.

- rule: {"string":{"maxLen":"255"}}

### spec.namespace

`string`

Namespace of the metric (e.g., "AWS/EC2", "AWS/ECS", "AWS/SQS",
"AWS/ApplicationELB"). Custom namespaces are also supported.

Required when `metric_name` is set. Must not start with a colon.

- rule: {"string":{"maxLen":"255"}}

### spec.period

`int32`

Period in seconds over which the statistic is applied. Defines the
granularity of the alarm evaluation.

Valid values: 10, 20, 30, or any multiple of 60.
High-resolution metrics support 10/20/30-second periods.
Standard-resolution metrics require multiples of 60.

Required when `metric_name` is set.

### spec.statistic

`string`

Standard statistic to apply to the metric.

Valid values: "SampleCount", "Average", "Sum", "Minimum", "Maximum".

Mutually exclusive with `extended_statistic`. Exactly one must be set
when using simple metric mode.

### spec.extendedStatistic

`string`

Percentile or extended statistic to apply to the metric. Used for
percentile-based alarms (e.g., p95 latency, p99 error rate).

Examples: "p95", "p99", "p99.9", "IQM", "TM(10%:90%)"

Mutually exclusive with `statistic`. Exactly one must be set when
using simple metric mode.

### spec.dimensions

`map<string, string>`

Dimensions that identify the specific metric stream to alarm on.
Dimensions narrow the metric to a specific resource or subset.

Example for EC2 CPU: {"InstanceId": "i-1234567890abcdef0"}
Example for ECS service: {"ClusterName": "prod", "ServiceName": "api"}

Mutually exclusive with `metric_queries`.

### spec.unit

`string`

Unit for the metric. When specified, only data points with a matching unit
are used for evaluation. Most alarms omit this and use the metric's
published unit. Must be a valid CloudWatch StandardUnit when set.

### spec.metricQueries

`[]AwsCloudwatchAlarmMetricQuery`

Metric math expressions or multi-metric queries. Use this mode for:
- Metric math (e.g., error rate = errors / total * 100)
- Anomaly detection (ANOMALY_DETECTION_BAND function)
- Cross-account metric monitoring
- Multi-metric composite evaluations

Up to 20 queries are supported. Exactly one query must set
`return_data = true` to serve as the alarm's evaluation signal.

Mutually exclusive with simple metric fields (`metric_name`,
`namespace`, `period`, `statistic`, `extended_statistic`, `dimensions`)
and with `evaluation_criteria`.

- rule: each metric query must set exactly one of expression or metric — an expression computes from other queries, a metric retrieves data from CloudWatch
- rule: period must be 1, 5, 10, 20, 30, or a multiple of 60 when set

### spec.metricQueries[].id

`string` · required

Unique identifier for this query. Used as a variable name in metric math
expressions. Must start with a lowercase letter; valid characters are
lowercase letters, digits, and underscores.

- rule: {"required":true,"string":{"maxLen":"255"}}

### spec.metricQueries[].expression

`string`

Metric math expression or Metrics Insights query. References other
queries by their `id` field.

Examples:
  "m1/m2*100"                           — error rate percentage
  "ANOMALY_DETECTION_BAND(m1, 2)"       — anomaly detection with 2 std devs
  "METRICS('AWS/EC2')"                   — Metrics Insights query

Mutually exclusive with `metric`.

- rule: {"string":{"maxLen":"1024"}}

### spec.metricQueries[].metric

`AwsCloudwatchAlarmMetricQueryMetric`

Raw metric definition. Use when this query retrieves a metric from
CloudWatch rather than computing a value from other queries.

Mutually exclusive with `expression`.

- rule: period must be 1, 5, 10, 20, 30, or a multiple of 60
- rule: stat must be a standard statistic (SampleCount, Average, Sum, Minimum, Maximum) or a percentile/trimmed statistic, e.g. 'p95', 'IQM', 'TM(10%:90%)'
- rule: namespace must not start with a colon

### spec.metricQueries[].metric.metricName

`string` · required

Name of the CloudWatch metric (e.g., "CPUUtilization", "5XXError").

- rule: {"required":true,"string":{"maxLen":"255"}}

### spec.metricQueries[].metric.namespace

`string`

Namespace of the metric (e.g., "AWS/EC2", "AWS/ApplicationELB").
Optional in the CloudWatch API, but virtually always needed to address
a metric unambiguously. Must not start with a colon.

- rule: {"string":{"maxLen":"255"}}

### spec.metricQueries[].metric.period

`int32`

Period in seconds for the metric data points. Determines the granularity
of the data used in expressions.

Valid values: 1, 5, 10, 20, 30, or any multiple of 60.
High-resolution metrics support sub-minute periods (1, 5, 10, 20, 30).

- rule: {"int32":{"gte":1}}

### spec.metricQueries[].metric.stat

`string` · required

Statistic to apply — either a standard statistic (SampleCount, Average,
Sum, Minimum, Maximum) or a percentile/extended statistic (p95, p99.9,
IQM, TM(10%:90%), etc.).

- rule: {"required":true}

### spec.metricQueries[].metric.dimensions

`map<string, string>`

Dimensions that identify the specific metric stream.

### spec.metricQueries[].metric.unit

`string`

Unit for the metric. Optional; filters data points to those matching
the specified unit.

### spec.metricQueries[].label

`string`

Human-readable label for this query. Displayed in the CloudWatch console
and alarm history. Especially useful for expressions to describe what the
computed value represents.

### spec.metricQueries[].period

`int32`

Override period (in seconds) for this query. When omitted, uses the period
from the `metric` definition (if present) or inherits the alarm's period.

Valid values: 1, 5, 10, 20, 30, or any multiple of 60.

### spec.metricQueries[].returnData

`bool`

Whether this query's result is used as the alarm's evaluation signal.
Exactly one query in the `metric_queries` list must set this to true.
Queries with `return_data = false` are intermediate values used by
expressions but not directly evaluated by the alarm.

### spec.metricQueries[].accountId

`string`

AWS account ID where the metric is located. Use for cross-account alarms
that monitor metrics from a different account.

- rule: {"string":{"maxLen":"255"}}

### spec.alarmActions

`[]string | valueFrom`

Actions to execute when the alarm transitions to ALARM state. Each action
is an ARN — typically an SNS topic ARN, but can also be an Auto Scaling
policy, EC2 automation action, Lambda function, or SSM OpsItem.

Maximum 5 actions — AWS's quota is 5 actions per alarm state. (The
provider caps only ok_actions and insufficient_data_actions; the missing
cap on alarm_actions is provider looseness — the spec holds AWS's
contract.)

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.okActions

`[]string | valueFrom`

Actions to execute when the alarm transitions to OK state.
Maximum 5 actions.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.insufficientDataActions

`[]string | valueFrom`

Actions to execute when the alarm transitions to INSUFFICIENT_DATA state.
Maximum 5 actions.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.alarmDescription

`string`

Human-readable description of the alarm's purpose. Include details about
what the alarm monitors, expected thresholds, and remediation steps.
Maximum 1024 characters.

- rule: {"string":{"maxLen":"1024"}}

### spec.evaluateLowSampleCountPercentiles

`string`

Controls alarm behavior during periods with too few data points for
statistically significant percentile calculations.

Valid values:
- "evaluate" — Always evaluate the alarm regardless of sample count.
- "ignore"   — Do not change alarm state during low-sample-count periods.

Only meaningful when using percentile statistics (extended_statistic or
a percentile stat in metric_queries).

### spec.evaluationCriteria

`AwsCloudwatchAlarmEvaluationCriteria`

PromQL evaluation criteria for alarming on an Amazon Managed Service for
Prometheus workspace. In this mode the alarm evaluates a PromQL query on
a schedule (`evaluation_interval`) and transitions based on the query's
firing state, so the threshold-based fields (`comparison_operator`,
`evaluation_periods`, `threshold`, `datapoints_to_alarm`, and all
simple-metric fields) must NOT be set.

### spec.evaluationCriteria.promqlCriteria

`AwsCloudwatchAlarmPromqlCriteria` · required

The PromQL criteria to evaluate. Required.

- rule: {"required":true}
- rule: pending_period must be between 0 and 86400 seconds (24 hours)
- rule: recovery_period must be between 0 and 86400 seconds (24 hours)

### spec.evaluationCriteria.promqlCriteria.query

`string` · required

The PromQL query to evaluate, addressing an Amazon Managed Service for
Prometheus workspace. The query must produce a boolean-style firing
signal, e.g.:
  'avg(rate(node_cpu_seconds_total{mode!="idle"}[5m])) > 0.8'
Maximum 10000 characters.

- rule: {"required":true,"string":{"maxLen":"10000"}}

### spec.evaluationCriteria.promqlCriteria.pendingPeriod

`int32` · optional (explicit presence)

How long (in seconds) the query must continuously return a firing result
before the alarm transitions to ALARM — the PromQL analog of "for:" in
Prometheus alerting rules. 0 transitions immediately. Maximum 86400
(24 hours). When unset, AWS applies its default pending behavior.

### spec.evaluationCriteria.promqlCriteria.recoveryPeriod

`int32` · optional (explicit presence)

How long (in seconds) the query must continuously return a non-firing
result before the alarm transitions back to OK — dampens flapping.
0 recovers immediately. Maximum 86400 (24 hours). When unset, AWS
applies its default recovery behavior.

### spec.evaluationInterval

`int32`

How often (in seconds) the PromQL query is evaluated. Only valid together
with `evaluation_criteria`. Valid values: 10, 20, 30, or any multiple
of 60. When unset, AWS uses its default evaluation interval.

## Validation Rules

- `comparison_operator_valid`: comparison_operator must be one of: GreaterThanOrEqualToThreshold, GreaterThanThreshold, LessThanThreshold, LessThanOrEqualToThreshold, LessThanLowerOrGreaterThanUpperThreshold, LessThanLowerThreshold, GreaterThanUpperThreshold
- `statistic_valid`: statistic must be one of: SampleCount, Average, Sum, Minimum, Maximum when set
- `treat_missing_data_valid`: treat_missing_data must be one of: missing, ignore, breaching, notBreaching when set
- `evaluate_low_sample_count_percentiles_valid`: evaluate_low_sample_count_percentiles must be 'evaluate' or 'ignore' when set
- `unit_valid_values`: unit must be a valid CloudWatch unit (e.g. Seconds, Milliseconds, Bytes, Percent, Count, Bytes/Second, Count/Second, None) when set
- `statistic_extended_statistic_exclusive`: only one of statistic or extended_statistic may be set
- `metric_source_required`: one of metric_name, metric_queries, or evaluation_criteria must be set — the alarm needs a signal to evaluate
- `simple_metric_or_metric_queries`: metric_name and metric_queries are mutually exclusive — use one mode or the other
- `promql_exclusive_with_other_modes`: evaluation_criteria (PromQL mode) cannot be combined with metric_name or metric_queries — use exactly one metric definition mode
- `promql_forbids_threshold_fields`: PromQL alarms (evaluation_criteria) must not set comparison_operator, evaluation_periods, datapoints_to_alarm, threshold, threshold_metric_id, namespace, period, statistic, extended_statistic, dimensions, or unit — the PromQL query expresses the alarm condition
- `comparison_operator_required_for_threshold_modes`: comparison_operator is required for simple-metric and metric-query alarms
- `evaluation_periods_required_for_threshold_modes`: evaluation_periods must be at least 1 for simple-metric and metric-query alarms
- `evaluation_interval_requires_promql`: evaluation_interval is only valid together with evaluation_criteria (PromQL mode)
- `evaluation_interval_valid_values`: evaluation_interval must be 10, 20, 30, or a multiple of 60 when set
- `namespace_required_with_metric_name`: namespace is required when metric_name is set
- `period_required_with_metric_name`: period is required when metric_name is set
- `statistic_required_with_metric_name`: one of statistic or extended_statistic is required when metric_name is set
- `period_valid_values`: period must be 10, 20, 30, or a multiple of 60 when set
- `datapoints_to_alarm_lte_evaluation_periods`: datapoints_to_alarm must be less than or equal to evaluation_periods
- `metric_queries_max_20`: maximum 20 metric_queries allowed
- `metric_queries_exactly_one_return_data`: exactly one metric query must set return_data = true — that query is the alarm's evaluation signal
- `alarm_actions_max_5`: maximum 5 alarm_actions allowed
- `ok_actions_max_5`: maximum 5 ok_actions allowed
- `insufficient_data_actions_max_5`: maximum 5 insufficient_data_actions allowed
- `threshold_conflicts_with_threshold_metric_id`: threshold and threshold_metric_id are mutually exclusive — use a static threshold or an anomaly detection band, not both
- `metric_queries_forbid_simple_metric_fields`: namespace, period, statistic, extended_statistic, dimensions, and unit must not be set together with metric_queries — define them inside each query's metric instead
- `extended_statistic_format`: extended_statistic must be a percentile or trimmed statistic, e.g. 'p95', 'tm99', 'IQM', 'TM(10%:90%)', 'PR(:300)'
- `namespace_no_leading_colon`: namespace must not start with a colon

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudwatchAlarm, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.alarm_arn` | `string` | The Amazon Resource Name (ARN) of the metric alarm. This is the primary identifier used to reference the alarm in composite alarms, dashboards, and operational tooling. |
| `status.outputs.alarm_name` | `string` | The name of the metric alarm. The alarm name is unique within the AWS account and region. Useful for CloudWatch API calls, CLI operations, and dashboard widgets. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.alarmActions` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.okActions` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.insufficientDataActions` | AwsSnsTopic | `status.outputs.topic_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAutoScalingGroup | `spec.instanceRefresh.preferences.alarms` | `status.outputs.alarm_name` |
| AwsCloudwatchCompositeAlarm | `spec.actionsSuppressor.alarm` | `status.outputs.alarm_name` |
| AwsEcsService | `spec.alarms.alarmNames` | `status.outputs.alarm_name` |

## See Also

- [Overview](../README.md)
