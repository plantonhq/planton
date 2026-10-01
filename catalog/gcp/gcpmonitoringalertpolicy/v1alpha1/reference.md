# GcpMonitoringAlertPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpMonitoringAlertPolicySpec defines a Cloud Monitoring alerting policy —
the rule that watches metrics or logs and opens an incident (and
notifies the configured channels) when its conditions are met.

A policy is built from three parts:
  1. conditions      -- WHAT to watch (a metric threshold, metric absence,
                        a matched log entry, an MQL/PromQL query, or a
                        SQL condition), combined by `combiner`;
  2. notification_channels -- WHO to tell (references to
                        GcpMonitoringNotificationChannel resources — the
                        composition edge charts wire);
  3. alert_strategy  -- HOW to notify (auto-close, rate limiting,
                        re-notification cadence).

The GCP API allows one to six conditions per policy. Each condition
carries exactly one condition type — the API models the choice as a
oneof, and this spec enforces it even though the Terraform provider
leaves it unchecked client-side (the API rejects violations at apply).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMonitoringAlertPolicy
metadata:
  name: my-sample-alert-policy
spec:
  # GCP project that owns the alert policy.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Shown in the console, incidents, and notifications; omit to default
  # to metadata.name.
  displayName: CPU saturation

  # How condition results combine (required by GCP even with one
  # condition): AND, OR, or AND_WITH_MATCHING_RESOURCE.
  combiner: OR

  # Severity reported on opened incidents: CRITICAL, ERROR, or WARNING.
  severity: WARNING

  # Whether the policy evaluates (default true). Disable to silence a
  # noisy rule without deleting it.
  enabled: true

  # 1-6 conditions, each with exactly one condition-type arm.
  conditions:
    - displayName: cpu above 80%
      conditionThreshold:
        # Which time series to evaluate (must name a metric.type).
        filter: metric.type="compute.googleapis.com/instance/cpu/utilization" AND resource.type="gce_instance"
        comparison: COMPARISON_GT
        thresholdValue: 0.8
        # Sustained-violation window (whole minutes; 0s pages on a
        # single point — the classic noisy-policy mistake).
        duration: 300s
        aggregations:
          - alignmentPeriod: 60s
            perSeriesAligner: ALIGN_MEAN

  # The channels to page — reference GcpMonitoringNotificationChannel
  # resources (their channel_name output) or supply literal names.
  notificationChannels:
    - valueFrom:
        kind: GcpMonitoringNotificationChannel
        # The channel kind's PUBLISHED prerequisite fixture name.
        name: planton-oss-e2e-gcpnotifchan-prereq
        fieldPath: status.outputs.channel_name

  # Notification behavior for open incidents.
  alertStrategy:
    # Close incidents automatically after the condition has stopped
    # violating for this long.
    autoClose: 1800s

  # The runbook the on-call engineer sees in every notification.
  documentation:
    content: |
      1. Check the instance dashboard for the saturated VM.
      2. If load is legitimate, scale the instance group; otherwise
         investigate the offending process.
    mimeType: text/markdown
    subject: CPU saturation on ${resource.label.instance_id}

  # User metadata labels, merged with Planton's platform labels.
  labels:
    team: platform

  # What a destroy does: DELETE (default), PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.displayName` | `string` |  |  |  |
| `spec.combiner` | `string` | yes |  |  |
| `spec.severity` | `string` |  |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.conditions` | `[]GcpMonitoringAlertPolicyCondition` | yes |  |  |
| `spec.conditions[].displayName` | `string` | yes |  |  |
| `spec.conditions[].conditionThreshold` | `GcpMonitoringAlertPolicyConditionThreshold` |  |  |  |
| `spec.conditions[].conditionThreshold.filter` | `string` | yes |  |  |
| `spec.conditions[].conditionThreshold.comparison` | `string` | yes |  |  |
| `spec.conditions[].conditionThreshold.thresholdValue` | `double` |  |  |  |
| `spec.conditions[].conditionThreshold.duration` | `string` | yes |  |  |
| `spec.conditions[].conditionThreshold.aggregations` | `[]GcpMonitoringAlertPolicyAggregation` |  |  |  |
| `spec.conditions[].conditionThreshold.aggregations[].alignmentPeriod` | `string` |  |  |  |
| `spec.conditions[].conditionThreshold.aggregations[].perSeriesAligner` | `string` |  |  |  |
| `spec.conditions[].conditionThreshold.aggregations[].crossSeriesReducer` | `string` |  |  |  |
| `spec.conditions[].conditionThreshold.aggregations[].groupByFields` | `[]string` |  |  |  |
| `spec.conditions[].conditionThreshold.denominatorFilter` | `string` |  |  |  |
| `spec.conditions[].conditionThreshold.denominatorAggregations` | `[]GcpMonitoringAlertPolicyAggregation` |  |  |  |
| `spec.conditions[].conditionThreshold.denominatorAggregations[].alignmentPeriod` | `string` |  |  |  |
| `spec.conditions[].conditionThreshold.denominatorAggregations[].perSeriesAligner` | `string` |  |  |  |
| `spec.conditions[].conditionThreshold.denominatorAggregations[].crossSeriesReducer` | `string` |  |  |  |
| `spec.conditions[].conditionThreshold.denominatorAggregations[].groupByFields` | `[]string` |  |  |  |
| `spec.conditions[].conditionThreshold.forecastOptions` | `GcpMonitoringAlertPolicyForecastOptions` |  |  |  |
| `spec.conditions[].conditionThreshold.forecastOptions.forecastHorizon` | `string` | yes |  |  |
| `spec.conditions[].conditionThreshold.trigger` | `GcpMonitoringAlertPolicyTrigger` |  |  |  |
| `spec.conditions[].conditionThreshold.trigger.count` | `int32` |  |  |  |
| `spec.conditions[].conditionThreshold.trigger.percent` | `double` |  |  |  |
| `spec.conditions[].conditionThreshold.evaluationMissingData` | `string` |  |  |  |
| `spec.conditions[].conditionAbsent` | `GcpMonitoringAlertPolicyConditionAbsent` |  |  |  |
| `spec.conditions[].conditionAbsent.filter` | `string` | yes |  |  |
| `spec.conditions[].conditionAbsent.duration` | `string` | yes |  |  |
| `spec.conditions[].conditionAbsent.aggregations` | `[]GcpMonitoringAlertPolicyAggregation` |  |  |  |
| `spec.conditions[].conditionAbsent.aggregations[].alignmentPeriod` | `string` |  |  |  |
| `spec.conditions[].conditionAbsent.aggregations[].perSeriesAligner` | `string` |  |  |  |
| `spec.conditions[].conditionAbsent.aggregations[].crossSeriesReducer` | `string` |  |  |  |
| `spec.conditions[].conditionAbsent.aggregations[].groupByFields` | `[]string` |  |  |  |
| `spec.conditions[].conditionAbsent.trigger` | `GcpMonitoringAlertPolicyTrigger` |  |  |  |
| `spec.conditions[].conditionAbsent.trigger.count` | `int32` |  |  |  |
| `spec.conditions[].conditionAbsent.trigger.percent` | `double` |  |  |  |
| `spec.conditions[].conditionMatchedLog` | `GcpMonitoringAlertPolicyConditionMatchedLog` |  |  |  |
| `spec.conditions[].conditionMatchedLog.filter` | `string` | yes |  |  |
| `spec.conditions[].conditionMatchedLog.labelExtractors` | `map<string, string>` |  |  |  |
| `spec.conditions[].conditionMonitoringQueryLanguage` | `GcpMonitoringAlertPolicyConditionMql` |  |  |  |
| `spec.conditions[].conditionMonitoringQueryLanguage.query` | `string` | yes |  |  |
| `spec.conditions[].conditionMonitoringQueryLanguage.duration` | `string` | yes |  |  |
| `spec.conditions[].conditionMonitoringQueryLanguage.trigger` | `GcpMonitoringAlertPolicyTrigger` |  |  |  |
| `spec.conditions[].conditionMonitoringQueryLanguage.trigger.count` | `int32` |  |  |  |
| `spec.conditions[].conditionMonitoringQueryLanguage.trigger.percent` | `double` |  |  |  |
| `spec.conditions[].conditionMonitoringQueryLanguage.evaluationMissingData` | `string` |  |  |  |
| `spec.conditions[].conditionPrometheusQueryLanguage` | `GcpMonitoringAlertPolicyConditionPromql` |  |  |  |
| `spec.conditions[].conditionPrometheusQueryLanguage.query` | `string` | yes |  |  |
| `spec.conditions[].conditionPrometheusQueryLanguage.duration` | `string` |  |  |  |
| `spec.conditions[].conditionPrometheusQueryLanguage.evaluationInterval` | `string` |  |  |  |
| `spec.conditions[].conditionPrometheusQueryLanguage.labels` | `map<string, string>` |  |  |  |
| `spec.conditions[].conditionPrometheusQueryLanguage.ruleGroup` | `string` |  |  |  |
| `spec.conditions[].conditionPrometheusQueryLanguage.alertRule` | `string` |  |  |  |
| `spec.conditions[].conditionPrometheusQueryLanguage.disableMetricValidation` | `bool` |  |  |  |
| `spec.conditions[].conditionSql` | `GcpMonitoringAlertPolicyConditionSql` |  |  |  |
| `spec.conditions[].conditionSql.query` | `string` | yes |  |  |
| `spec.conditions[].conditionSql.minutes` | `GcpMonitoringAlertPolicySqlMinutes` |  |  |  |
| `spec.conditions[].conditionSql.minutes.periodicity` | `int32` |  |  |  |
| `spec.conditions[].conditionSql.hourly` | `GcpMonitoringAlertPolicySqlHourly` |  |  |  |
| `spec.conditions[].conditionSql.hourly.periodicity` | `int32` |  |  |  |
| `spec.conditions[].conditionSql.hourly.minuteOffset` | `int32` |  |  |  |
| `spec.conditions[].conditionSql.daily` | `GcpMonitoringAlertPolicySqlDaily` |  |  |  |
| `spec.conditions[].conditionSql.daily.periodicity` | `int32` |  |  |  |
| `spec.conditions[].conditionSql.daily.executionTime` | `GcpMonitoringAlertPolicyTimeOfDay` |  |  |  |
| `spec.conditions[].conditionSql.daily.executionTime.hours` | `int32` |  |  |  |
| `spec.conditions[].conditionSql.daily.executionTime.minutes` | `int32` |  |  |  |
| `spec.conditions[].conditionSql.daily.executionTime.seconds` | `int32` |  |  |  |
| `spec.conditions[].conditionSql.daily.executionTime.nanos` | `int32` |  |  |  |
| `spec.conditions[].conditionSql.rowCountTest` | `GcpMonitoringAlertPolicySqlRowCountTest` |  |  |  |
| `spec.conditions[].conditionSql.rowCountTest.comparison` | `string` | yes |  |  |
| `spec.conditions[].conditionSql.rowCountTest.threshold` | `int64` |  |  |  |
| `spec.conditions[].conditionSql.booleanTest` | `GcpMonitoringAlertPolicySqlBooleanTest` |  |  |  |
| `spec.conditions[].conditionSql.booleanTest.column` | `string` | yes |  |  |
| `spec.notificationChannels` | `[]string \| valueFrom` |  |  | GcpMonitoringNotificationChannel (`status.outputs.channel_name`) |
| `spec.alertStrategy` | `GcpMonitoringAlertPolicyAlertStrategy` |  |  |  |
| `spec.alertStrategy.autoClose` | `string` |  |  |  |
| `spec.alertStrategy.notificationRateLimit` | `GcpMonitoringAlertPolicyNotificationRateLimit` |  |  |  |
| `spec.alertStrategy.notificationRateLimit.period` | `string` |  |  |  |
| `spec.alertStrategy.notificationChannelStrategy` | `[]GcpMonitoringAlertPolicyNotificationChannelStrategy` |  |  |  |
| `spec.alertStrategy.notificationChannelStrategy[].notificationChannelNames` | `[]string \| valueFrom` |  |  | GcpMonitoringNotificationChannel (`status.outputs.channel_name`) |
| `spec.alertStrategy.notificationChannelStrategy[].renotifyInterval` | `string` |  |  |  |
| `spec.alertStrategy.notificationPrompts` | `[]string` |  |  |  |
| `spec.documentation` | `GcpMonitoringAlertPolicyDocumentation` |  |  |  |
| `spec.documentation.content` | `string` |  |  |  |
| `spec.documentation.mimeType` | `string` |  |  |  |
| `spec.documentation.subject` | `string` |  |  |  |
| `spec.documentation.links` | `[]GcpMonitoringAlertPolicyDocumentationLink` |  |  |  |
| `spec.documentation.links[].displayName` | `string` |  |  |  |
| `spec.documentation.links[].url` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the alert policy. Can be a literal project ID
or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.displayName

`string`

Human-friendly name shown in the console, in incidents, and in
notifications. Defaults to metadata.name when left empty (the GCP API
requires a display name).

### spec.combiner

`string` · required

How the results of multiple conditions combine into one incident
decision:
  AND -- an incident opens only while ALL conditions are met
  OR  -- an incident opens while ANY condition is met
  AND_WITH_MATCHING_RESOURCE -- AND, but only when the conditions
         trigger on the SAME monitored resource (the strictest form)
Required by the GCP API even for single-condition policies (use OR or
AND — they are equivalent with one condition).

- rule: combiner must be one of: AND, OR, AND_WITH_MATCHING_RESOURCE
- rule: {"required":true}

### spec.severity

`string`

Severity reported on incidents this policy opens: CRITICAL, ERROR, or
WARNING. Empty leaves the severity unset in GCP (incidents show no
severity level).

- rule: severity must be one of: CRITICAL, ERROR, WARNING

### spec.enabled

`bool` · optional (explicit presence)

Whether the policy is evaluated (default true). A disabled policy keeps
its configuration but opens no incidents — the safe way to silence a
noisy rule while tuning it. Both IaC engines send the value explicitly
so behavior is identical regardless of engine.

- default: `true`

### spec.conditions

`[]GcpMonitoringAlertPolicyCondition` · required

The conditions this policy evaluates (1 to 6 — the GCP API's own
bounds), combined by `combiner`.

- rule: {"repeated":{"minItems":"1","maxItems":"6"}}
- rule: each condition carries exactly one of: condition_threshold, condition_absent, condition_matched_log, condition_monitoring_query_language, condition_prometheus_query_language, condition_sql

### spec.conditions[].displayName

`string` · required

Name shown for this condition in the console and in incident detail.

- rule: {"required":true}

### spec.conditions[].conditionThreshold

`GcpMonitoringAlertPolicyConditionThreshold`

Alert when a metric crosses a threshold — the workhorse condition type
(CPU above 80%, error rate above 1%, uptime-check failures).

### spec.conditions[].conditionThreshold.filter

`string` · required

A monitoring filter (https://cloud.google.com/monitoring/api/v3/filters)
selecting the time series to evaluate, e.g.
  metric.type="compute.googleapis.com/instance/cpu/utilization" AND
  resource.type="gce_instance"
The API requires a filter that names a metric.type.

- rule: {"required":true}

### spec.conditions[].conditionThreshold.comparison

`string` · required

How the aggregated value compares to the threshold to trigger:
COMPARISON_GT / _GE / _LT / _LE / _EQ / _NE. GT ("above threshold") is
by far the most common.

- rule: comparison must be one of: COMPARISON_GT, COMPARISON_GE, COMPARISON_LT, COMPARISON_LE, COMPARISON_EQ, COMPARISON_NE
- rule: {"required":true}

### spec.conditions[].conditionThreshold.thresholdValue

`double`

The threshold value compared against the aggregated metric.

### spec.conditions[].conditionThreshold.duration

`string` · required

How long the comparison must hold before the condition triggers, as a
duration in seconds. Must be a multiple of 60s ("0s" triggers on a
single violating point; "300s" requires five sustained minutes — the
usual flap guard). The GCP API rejects non-minute-aligned values.

- rule: duration must be a seconds duration such as 0s, 60s, or 300s (the GCP API additionally requires whole minutes — a multiple of 60)
- rule: {"required":true}

### spec.conditions[].conditionThreshold.aggregations

`[]GcpMonitoringAlertPolicyAggregation`

How raw time series are aligned and combined before comparison. Most
metric-type conditions need at least one aggregation (e.g. align
ALIGN_MEAN over the alignment_period) — without one, GCP compares every
raw point of every series.

### spec.conditions[].conditionThreshold.aggregations[].alignmentPeriod

`string`

The window each series is aligned over, as a duration in seconds
(minute-aligned, e.g. "60s", "300s").

### spec.conditions[].conditionThreshold.aggregations[].perSeriesAligner

`string`

How points within the window combine per series: ALIGN_NONE,
ALIGN_DELTA, ALIGN_RATE, ALIGN_INTERPOLATE, ALIGN_NEXT_OLDER,
ALIGN_MIN, ALIGN_MAX, ALIGN_MEAN, ALIGN_COUNT, ALIGN_SUM, ALIGN_STDDEV,
ALIGN_COUNT_TRUE, ALIGN_COUNT_FALSE, ALIGN_FRACTION_TRUE,
ALIGN_PERCENTILE_99, ALIGN_PERCENTILE_95, ALIGN_PERCENTILE_50,
ALIGN_PERCENTILE_05, ALIGN_PERCENT_CHANGE.

- rule: per_series_aligner must be one of the ALIGN_* values documented on the field

### spec.conditions[].conditionThreshold.aggregations[].crossSeriesReducer

`string`

How aligned series combine ACROSS series: REDUCE_NONE, REDUCE_MEAN,
REDUCE_MIN, REDUCE_MAX, REDUCE_SUM, REDUCE_STDDEV, REDUCE_COUNT,
REDUCE_COUNT_TRUE, REDUCE_COUNT_FALSE, REDUCE_FRACTION_TRUE,
REDUCE_PERCENTILE_99, REDUCE_PERCENTILE_95, REDUCE_PERCENTILE_50,
REDUCE_PERCENTILE_05.

- rule: cross_series_reducer must be one of the REDUCE_* values documented on the field

### spec.conditions[].conditionThreshold.aggregations[].groupByFields

`[]string`

Resource/metric label keys that survive cross-series reduction — series
sharing these label values reduce together (e.g. group error rates by
resource.label.zone).

### spec.conditions[].conditionThreshold.denominatorFilter

`string`

For ratio conditions: the filter selecting the DENOMINATOR time series
(the condition then evaluates numerator/denominator against the
threshold, e.g. error requests / total requests).

### spec.conditions[].conditionThreshold.denominatorAggregations

`[]GcpMonitoringAlertPolicyAggregation`

Aggregations applied to the denominator series. The API requires the
denominator to be aligned identically to the numerator for the ratio to
be meaningful.

### spec.conditions[].conditionThreshold.denominatorAggregations[].alignmentPeriod

`string`

The window each series is aligned over, as a duration in seconds
(minute-aligned, e.g. "60s", "300s").

### spec.conditions[].conditionThreshold.denominatorAggregations[].perSeriesAligner

`string`

How points within the window combine per series: ALIGN_NONE,
ALIGN_DELTA, ALIGN_RATE, ALIGN_INTERPOLATE, ALIGN_NEXT_OLDER,
ALIGN_MIN, ALIGN_MAX, ALIGN_MEAN, ALIGN_COUNT, ALIGN_SUM, ALIGN_STDDEV,
ALIGN_COUNT_TRUE, ALIGN_COUNT_FALSE, ALIGN_FRACTION_TRUE,
ALIGN_PERCENTILE_99, ALIGN_PERCENTILE_95, ALIGN_PERCENTILE_50,
ALIGN_PERCENTILE_05, ALIGN_PERCENT_CHANGE.

- rule: per_series_aligner must be one of the ALIGN_* values documented on the field

### spec.conditions[].conditionThreshold.denominatorAggregations[].crossSeriesReducer

`string`

How aligned series combine ACROSS series: REDUCE_NONE, REDUCE_MEAN,
REDUCE_MIN, REDUCE_MAX, REDUCE_SUM, REDUCE_STDDEV, REDUCE_COUNT,
REDUCE_COUNT_TRUE, REDUCE_COUNT_FALSE, REDUCE_FRACTION_TRUE,
REDUCE_PERCENTILE_99, REDUCE_PERCENTILE_95, REDUCE_PERCENTILE_50,
REDUCE_PERCENTILE_05.

- rule: cross_series_reducer must be one of the REDUCE_* values documented on the field

### spec.conditions[].conditionThreshold.denominatorAggregations[].groupByFields

`[]string`

Resource/metric label keys that survive cross-series reduction — series
sharing these label values reduce together (e.g. group error rates by
resource.label.zone).

### spec.conditions[].conditionThreshold.forecastOptions

`GcpMonitoringAlertPolicyForecastOptions`

Alert on the PREDICTED value instead of the current one: the condition
triggers when GCP forecasts the threshold will be crossed within this
horizon (a duration, minimum "3600s"). The capacity-planning form of a
threshold alert.

### spec.conditions[].conditionThreshold.forecastOptions.forecastHorizon

`string` · required

How far ahead GCP forecasts, as a duration in seconds (minimum "3600s"
— one hour — per the API).

- rule: {"required":true}

### spec.conditions[].conditionThreshold.trigger

`GcpMonitoringAlertPolicyTrigger`

How many time series (count) or what fraction of them (percent) must
violate before the condition triggers. Default: any single series.

- rule: set either count or percent, not both

### spec.conditions[].conditionThreshold.trigger.count

`int32`

Absolute number of violating series.

- rule: {"int32":{"gte":0}}

### spec.conditions[].conditionThreshold.trigger.percent

`double`

Percentage of violating series (0-100).

- rule: {"double":{"lte":100,"gte":0}}

### spec.conditions[].conditionThreshold.evaluationMissingData

`string`

How the condition evaluates when data stops arriving:
EVALUATION_MISSING_DATA_INACTIVE (missing data closes/keeps the
incident closed), _ACTIVE (missing data violates — the paranoid
setting), or _NO_OP (missing data changes nothing, the API default).

- rule: evaluation_missing_data must be one of: EVALUATION_MISSING_DATA_INACTIVE, EVALUATION_MISSING_DATA_ACTIVE, EVALUATION_MISSING_DATA_NO_OP

### spec.conditions[].conditionAbsent

`GcpMonitoringAlertPolicyConditionAbsent`

Alert when a metric stops reporting for a duration — the "silence is
failure" condition (a heartbeat metric going quiet).

### spec.conditions[].conditionAbsent.filter

`string` · required

A monitoring filter selecting the time series whose ABSENCE triggers
the condition.

- rule: {"required":true}

### spec.conditions[].conditionAbsent.duration

`string` · required

How long the data must be absent before triggering, as a duration in
seconds (minute-aligned, maximum 24 hours per the API — e.g. "300s").

- rule: {"required":true}

### spec.conditions[].conditionAbsent.aggregations

`[]GcpMonitoringAlertPolicyAggregation`

Alignment applied to the series before absence is judged.

### spec.conditions[].conditionAbsent.aggregations[].alignmentPeriod

`string`

The window each series is aligned over, as a duration in seconds
(minute-aligned, e.g. "60s", "300s").

### spec.conditions[].conditionAbsent.aggregations[].perSeriesAligner

`string`

How points within the window combine per series: ALIGN_NONE,
ALIGN_DELTA, ALIGN_RATE, ALIGN_INTERPOLATE, ALIGN_NEXT_OLDER,
ALIGN_MIN, ALIGN_MAX, ALIGN_MEAN, ALIGN_COUNT, ALIGN_SUM, ALIGN_STDDEV,
ALIGN_COUNT_TRUE, ALIGN_COUNT_FALSE, ALIGN_FRACTION_TRUE,
ALIGN_PERCENTILE_99, ALIGN_PERCENTILE_95, ALIGN_PERCENTILE_50,
ALIGN_PERCENTILE_05, ALIGN_PERCENT_CHANGE.

- rule: per_series_aligner must be one of the ALIGN_* values documented on the field

### spec.conditions[].conditionAbsent.aggregations[].crossSeriesReducer

`string`

How aligned series combine ACROSS series: REDUCE_NONE, REDUCE_MEAN,
REDUCE_MIN, REDUCE_MAX, REDUCE_SUM, REDUCE_STDDEV, REDUCE_COUNT,
REDUCE_COUNT_TRUE, REDUCE_COUNT_FALSE, REDUCE_FRACTION_TRUE,
REDUCE_PERCENTILE_99, REDUCE_PERCENTILE_95, REDUCE_PERCENTILE_50,
REDUCE_PERCENTILE_05.

- rule: cross_series_reducer must be one of the REDUCE_* values documented on the field

### spec.conditions[].conditionAbsent.aggregations[].groupByFields

`[]string`

Resource/metric label keys that survive cross-series reduction — series
sharing these label values reduce together (e.g. group error rates by
resource.label.zone).

### spec.conditions[].conditionAbsent.trigger

`GcpMonitoringAlertPolicyTrigger`

How many series (count) or what fraction (percent) must be absent
before triggering.

- rule: set either count or percent, not both

### spec.conditions[].conditionAbsent.trigger.count

`int32`

Absolute number of violating series.

- rule: {"int32":{"gte":0}}

### spec.conditions[].conditionAbsent.trigger.percent

`double`

Percentage of violating series (0-100).

- rule: {"double":{"lte":100,"gte":0}}

### spec.conditions[].conditionMatchedLog

`GcpMonitoringAlertPolicyConditionMatchedLog`

Alert on every log entry matching a filter. Requires
alert_strategy.notification_rate_limit (the API's own pairing for
log-based policies).

### spec.conditions[].conditionMatchedLog.filter

`string` · required

A logging filter (https://cloud.google.com/logging/docs/view/logging-query-language)
selecting the entries that trigger notifications, e.g.
  resource.type="gce_instance" AND severity>=ERROR

- rule: {"required":true}

### spec.conditions[].conditionMatchedLog.labelExtractors

`map<string, string>`

Extract values from matched entries into incident labels: map label
name -> extractor expression, e.g.
  { "vm": "EXTRACT(resource.labels.instance_id)" }
Extracted labels appear in notifications and documentation variables.

### spec.conditions[].conditionMonitoringQueryLanguage

`GcpMonitoringAlertPolicyConditionMql`

Alert on a Monitoring Query Language (MQL) query — for conditions the
structured threshold/absent forms cannot express. MQL is deprecated by
Google in favor of PromQL; prefer condition_prometheus_query_language
for new policies.

### spec.conditions[].conditionMonitoringQueryLanguage.query

`string` · required

The Monitoring Query Language query whose boolean output drives the
condition.

- rule: {"required":true}

### spec.conditions[].conditionMonitoringQueryLanguage.duration

`string` · required

How long the query output must be true before triggering (duration in
seconds, minute-aligned).

- rule: {"required":true}

### spec.conditions[].conditionMonitoringQueryLanguage.trigger

`GcpMonitoringAlertPolicyTrigger`

How many series (count) or what fraction (percent) must violate before
triggering.

- rule: set either count or percent, not both

### spec.conditions[].conditionMonitoringQueryLanguage.trigger.count

`int32`

Absolute number of violating series.

- rule: {"int32":{"gte":0}}

### spec.conditions[].conditionMonitoringQueryLanguage.trigger.percent

`double`

Percentage of violating series (0-100).

- rule: {"double":{"lte":100,"gte":0}}

### spec.conditions[].conditionMonitoringQueryLanguage.evaluationMissingData

`string`

How the condition evaluates when data stops arriving (same values as
the threshold condition's field).

- rule: evaluation_missing_data must be one of: EVALUATION_MISSING_DATA_INACTIVE, EVALUATION_MISSING_DATA_ACTIVE, EVALUATION_MISSING_DATA_NO_OP

### spec.conditions[].conditionPrometheusQueryLanguage

`GcpMonitoringAlertPolicyConditionPromql`

Alert on a PromQL query — the expressive form for rate/ratio/quantile
conditions and for teams porting Prometheus alert rules.

### spec.conditions[].conditionPrometheusQueryLanguage.query

`string` · required

The PromQL expression. The condition fires while the expression
produces any series (Prometheus alert-rule semantics), e.g.
  rate(http_requests_total{code=~"5.."}[5m]) > 0.1

- rule: {"required":true}

### spec.conditions[].conditionPrometheusQueryLanguage.duration

`string`

How long the expression must produce output before triggering
(Prometheus "for" semantics). A duration in seconds, e.g. "300s";
empty means fire immediately.

### spec.conditions[].conditionPrometheusQueryLanguage.evaluationInterval

`string`

How often the query is evaluated (default "30s"; must be a multiple of
30 seconds per the API).

### spec.conditions[].conditionPrometheusQueryLanguage.labels

`map<string, string>`

Labels added to every incident this condition opens (Prometheus
alert-rule labels; values support PromQL template syntax).

### spec.conditions[].conditionPrometheusQueryLanguage.ruleGroup

`string`

The rule group name this condition belongs to when imported from a
Prometheus rule file — carried for round-trip fidelity, no behavioral
effect in GCP.

### spec.conditions[].conditionPrometheusQueryLanguage.alertRule

`string`

The alert rule name when imported from a Prometheus rule file. Becomes
the incident's alertname label.

### spec.conditions[].conditionPrometheusQueryLanguage.disableMetricValidation

`bool`

Skip GCP's validation that the query only references known metrics
(default false). Enable only when alerting on metrics that do not exist
yet — a typo in the metric name then goes undetected until it never
fires.

### spec.conditions[].conditionSql

`GcpMonitoringAlertPolicyConditionSql`

Alert on a SQL query against log analytics data, evaluated on a
schedule (minutes/hourly/daily) with a row-count or boolean test on the
results.

- rule: set exactly one schedule: minutes, hourly, or daily
- rule: set exactly one result test: row_count_test or boolean_test

### spec.conditions[].conditionSql.query

`string` · required

The SQL query to run against log analytics (GoogleSQL). The query's
results feed row_count_test or boolean_test.

- rule: {"required":true}

### spec.conditions[].conditionSql.minutes

`GcpMonitoringAlertPolicySqlMinutes`

Run the query every N minutes.

### spec.conditions[].conditionSql.minutes.periodicity

`int32`

Number of minutes between runs (5 to 1440 per the GCP API).

- rule: {"int32":{"lte":1440,"gte":5}}

### spec.conditions[].conditionSql.hourly

`GcpMonitoringAlertPolicySqlHourly`

Run the query every N hours.

### spec.conditions[].conditionSql.hourly.periodicity

`int32`

Number of hours between runs (1 to 48 per the GCP API).

- rule: {"int32":{"lte":48,"gte":1}}

### spec.conditions[].conditionSql.hourly.minuteOffset

`int32` · optional (explicit presence)

Minute of the hour the run starts (0-59). Optional; GCP picks one when
unset.

- rule: {"int32":{"lte":59,"gte":0}}

### spec.conditions[].conditionSql.daily

`GcpMonitoringAlertPolicySqlDaily`

Run the query every N days at a fixed time.

### spec.conditions[].conditionSql.daily.periodicity

`int32`

Number of days between runs (1 to 31 per the GCP API).

- rule: {"int32":{"lte":31,"gte":1}}

### spec.conditions[].conditionSql.daily.executionTime

`GcpMonitoringAlertPolicyTimeOfDay`

Time of day the run starts (UTC). Omit for GCP's default.

### spec.conditions[].conditionSql.daily.executionTime.hours

`int32`

Hour (0-23).

- rule: {"int32":{"lte":23,"gte":0}}

### spec.conditions[].conditionSql.daily.executionTime.minutes

`int32`

Minute (0-59).

- rule: {"int32":{"lte":59,"gte":0}}

### spec.conditions[].conditionSql.daily.executionTime.seconds

`int32`

Second (0-59).

- rule: {"int32":{"lte":59,"gte":0}}

### spec.conditions[].conditionSql.daily.executionTime.nanos

`int32`

Nanoseconds (0-999999999). Kept for API-shape fidelity; alert schedules
realistically use whole seconds.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.conditions[].conditionSql.rowCountTest

`GcpMonitoringAlertPolicySqlRowCountTest`

Trigger when the query's ROW COUNT compares against a threshold.

### spec.conditions[].conditionSql.rowCountTest.comparison

`string` · required

How the row count compares to the threshold (same comparison values as
the threshold condition).

- rule: comparison must be one of: COMPARISON_GT, COMPARISON_GE, COMPARISON_LT, COMPARISON_LE, COMPARISON_EQ, COMPARISON_NE
- rule: {"required":true}

### spec.conditions[].conditionSql.rowCountTest.threshold

`int64`

The row-count threshold.

### spec.conditions[].conditionSql.booleanTest

`GcpMonitoringAlertPolicySqlBooleanTest`

Trigger when a BOOLEAN COLUMN of the first result row is true.

### spec.conditions[].conditionSql.booleanTest.column

`string` · required

The BOOL column of the first result row that decides the condition.

- rule: {"required":true}

### spec.notificationChannels

`[]string | valueFrom`

The notification channels to notify when incidents open, close, or gain
new violations. Each entry is a channel resource name
(projects/{project}/notificationChannels/{id}) — a literal or a
reference to a GcpMonitoringNotificationChannel resource (its
channel_name output is exactly this value).

- references: GcpMonitoringNotificationChannel (`status.outputs.channel_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpMonitoringNotificationChannel, name: <that resource's name>, fieldPath: status.outputs.channel_name}} -- a bare string does not parse

### spec.alertStrategy

`GcpMonitoringAlertPolicyAlertStrategy`

How notifications behave once an incident is open: auto-close timing,
rate limiting (log-based policies), and per-channel re-notification.

### spec.alertStrategy.autoClose

`string`

Close an incident automatically after its condition has stopped
violating for this long (duration in seconds, minimum "1800s" per the
API; GCP's default is 7 days).

### spec.alertStrategy.notificationRateLimit

`GcpMonitoringAlertPolicyNotificationRateLimit`

Rate limit for LOG-BASED policies (condition_matched_log): at most one
notification per period. The API requires this block for log-match
conditions and rejects it for metric conditions.

### spec.alertStrategy.notificationRateLimit.period

`string`

Not more than one notification per period (duration in seconds, e.g.
"300s").

### spec.alertStrategy.notificationChannelStrategy

`[]GcpMonitoringAlertPolicyNotificationChannelStrategy`

Per-channel-subset re-notification cadence for open incidents.

### spec.alertStrategy.notificationChannelStrategy[].notificationChannelNames

`[]string | valueFrom`

The channels this cadence applies to — full channel resource names,
each a literal or a reference to a GcpMonitoringNotificationChannel.
Every entry must also appear in the policy's notification_channels.

- references: GcpMonitoringNotificationChannel (`status.outputs.channel_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpMonitoringNotificationChannel, name: <that resource's name>, fieldPath: status.outputs.channel_name}} -- a bare string does not parse

### spec.alertStrategy.notificationChannelStrategy[].renotifyInterval

`string`

Send reminder notifications for still-open incidents at this interval
(duration in seconds).

### spec.alertStrategy.notificationPrompts

`[]string`

When to prompt the configured channels beyond incident open/close:
OPENED and/or CLOSED.

- rule: {"repeated":{"items":{"cel":[{"id":"valid_notification_prompt","message":"each notification prompt must be OPENED or CLOSED","expression":"this in ['OPENED', 'CLOSED']"}]}}}

### spec.documentation

`GcpMonitoringAlertPolicyDocumentation`

Troubleshooting content attached to every notification — the runbook
the on-call engineer sees. Supports Markdown and ${variable}
substitution (e.g. ${resource.label.instance_id}).

### spec.documentation.content

`string`

The runbook body (at most 8192 bytes). Supports the mime_type's syntax
and ${variable} substitution — write the steps the on-call engineer
should take.

- rule: {"string":{"maxLen":"8192"}}

### spec.documentation.mimeType

`string`

Content format; "text/markdown" (the default) is the only value the
API currently accepts.

- rule: mime_type must be text/markdown (the only format the GCP API accepts)

### spec.documentation.subject

`string`

Custom subject line for notifications (at most 255 bytes; not all
channel types render it).

- rule: {"string":{"maxLen":"255"}}

### spec.documentation.links

`[]GcpMonitoringAlertPolicyDocumentationLink`

Reference links (runbooks, dashboards, playbooks) attached to
notifications — at most 3 per the API.

- rule: {"repeated":{"maxItems":"3"}}

### spec.documentation.links[].displayName

`string`

Short display text for the link.

### spec.documentation.links[].url

`string`

The URL (http/https). Supports ${variable} substitution.

### spec.labels

`map<string, string>`

User labels attached to the policy for organizing and identifying it
(maps to the provider's user_labels), merged with Planton's platform
labels (which win on key conflicts). Keys and values may contain only
lowercase letters, numerals, underscores, and dashes; keys must begin
with a letter.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the policy is deleted; its open incidents close and no
               further notifications are sent
  "PREVENT" -- destroy FAILS; protects production alerting from
               accidental teardown
  "ABANDON" -- the policy is removed from management but keeps
               evaluating (and paging) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpMonitoringAlertPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_name` | `string` | The server-assigned resource name of the policy. Format: projects/{project}/alertPolicies/{policy_id} The handle for cross-referencing the policy in dashboards, snooze configurations, and the Monitoring API. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.notificationChannels` | GcpMonitoringNotificationChannel | `status.outputs.channel_name` |
| `spec.alertStrategy.notificationChannelStrategy[].notificationChannelNames` | GcpMonitoringNotificationChannel | `status.outputs.channel_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDeliveryPipeline | `spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud.alertPolicyChecks[].alertPolicies` | `status.outputs.policy_name` |
| GcpDeliveryPipeline | `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud.alertPolicyChecks[].alertPolicies` | `status.outputs.policy_name` |
| GcpDeliveryPipeline | `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud.alertPolicyChecks[].alertPolicies` | `status.outputs.policy_name` |

## See Also

- [Overview](../README.md)
