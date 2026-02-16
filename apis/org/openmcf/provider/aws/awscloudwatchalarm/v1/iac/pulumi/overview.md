# AwsCloudwatchAlarm — Pulumi Module Architecture

## Overview

This module provisions a single CloudWatch metric alarm. It supports two modes: simple metric alarms (single metric with namespace/metricName/statistic) and metric math alarms (using `metricQueries` with expressions). The module conditionally maps fields based on which mode is active.

## Resource Graph

```
AwsCloudwatchAlarmStackInput
    └── cloudwatch.MetricAlarm (the alarm)
```

## File Structure

| File | Purpose |
|------|---------|
| `module/main.go` | Entry point — creates AWS provider, invokes `metricAlarm()`, exports outputs |
| `module/locals.go` | Initializes `Locals` struct with tags and resource reference |
| `module/outputs.go` | Output key constants (`alarm_arn`, `alarm_name`) |
| `module/metric_alarm.go` | Creates the `cloudwatch.NewMetricAlarm` resource |

## Field Mapping

### Simple Metric Mode

| Spec Field | Pulumi Arg | Notes |
|------------|------------|-------|
| `metricName` | `MetricName` | Required for simple mode |
| `namespace` | `Namespace` | Required for simple mode |
| `statistic` | `Statistic` | Average, Sum, Maximum, etc. |
| `period` | `Period` | Evaluation period in seconds |

### Metric Math Mode

| Spec Field | Pulumi Arg | Notes |
|------------|------------|-------|
| `metricQueries` | `MetricQueries` | Array of metric query objects |
| `metricQueries[].expression` | `Expression` | Math expression referencing other query IDs |
| `metricQueries[].metric` | `MetricStat` | Raw metric definition with dimensions |
| `metricQueries[].returnData` | `ReturnData` | Only one query should return data |

### Common Fields

| Spec Field | Pulumi Arg | Notes |
|------------|------------|-------|
| `comparisonOperator` | `ComparisonOperator` | GreaterThanThreshold, LessThanThreshold, etc. |
| `evaluationPeriods` | `EvaluationPeriods` | Number of periods in the evaluation window |
| `datapointsToAlarm` | `DatapointsToAlarm` | M-of-N — how many must breach |
| `threshold` | `Threshold` | Numeric threshold value |
| `treatMissingData` | `TreatMissingData` | breaching, notBreaching, missing, ignore |
| `alarmDescription` | `AlarmDescription` | Human-readable description |
| `alarmActions` | `AlarmActions` | SNS topic ARNs for ALARM state |
| `okActions` | `OkActions` | SNS topic ARNs for OK state |
| `insufficientDataActions` | `InsufficientDataActions` | SNS topic ARNs for INSUFFICIENT_DATA state |

## Naming

The alarm name is derived from `metadata.name` via the Pulumi resource name argument.
