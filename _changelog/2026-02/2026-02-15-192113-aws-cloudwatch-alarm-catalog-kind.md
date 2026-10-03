# AWS CloudWatch Metric Alarm Catalog Kind

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Added AwsCloudwatchAlarm (R15) — the eighteenth new AWS resource kind and the final Phase 1 kind in the AWS resource expansion project. This kind deploys CloudWatch metric alarms with support for both simple metric mode and metric math expressions, M-of-N evaluation, anomaly detection, and multi-target actions.

## Problem Statement / Motivation

CloudWatch metric alarms are the primary mechanism for automated monitoring and alerting on AWS. Without this kind, Planton users could not define metric-based alarms declaratively or compose them into infra charts alongside the resources they monitor (ECS services, SQS queues, ALBs, etc.).

### Pain Points

- No declarative alarm definition in Planton — users had to manage alarms outside the IaC workflow
- Metric math alarms (error rates, latency percentiles) are common in production (~35% of alarms) but were not addressable
- Alarm actions (SNS notifications) could not participate in the infra chart DAG via `valueFrom` references
- Phase 1 of the AWS expansion had one remaining kind blocking completion

## Solution / What's New

### Two Metric Definition Modes

The kind supports two mutually exclusive modes, covering ~100% of CloudWatch alarm use cases:

1. **Simple metric mode** (~65% of production alarms) — direct metric name, namespace, period, and statistic
2. **Metric query mode** (~35% of production alarms) — up to 20 metric math expressions, anomaly detection bands, or cross-account metrics

### Key Capabilities

- **M-of-N evaluation** — `datapointsToAlarm` < `evaluationPeriods` reduces false positives from transient spikes
- **Missing data handling** — 4 treatment modes (missing, ignore, breaching, notBreaching) for different operational contexts
- **Anomaly detection** — `thresholdMetricId` references an ANOMALY_DETECTION_BAND function for ML-based dynamic thresholds
- **Multi-target actions** — alarm, OK, and insufficient-data actions with `StringValueOrRef` referencing AwsSnsTopic
- **14 CEL validations** — cross-field mutual exclusivity, range constraints, and mode consistency checks

## Implementation Details

### Proto API

- `spec.proto` — 20 top-level fields, 2 nested messages (~30 total fields), 14 CEL validations
- `AwsCloudwatchAlarmMetricQuery` — id, expression/metric (mutually exclusive), label, period, return_data, account_id
- `AwsCloudwatchAlarmMetricQueryMetric` — metric_name, namespace, period, stat, dimensions, unit
- Enum: `AwsCloudwatchAlarm = 311` in `catalog_kind.proto`

### Pulumi Module (4 files)

- `main.go` — entry point with provider setup
- `locals.go` — tag initialization with AwsCloudwatchAlarm kind
- `alarm.go` — single `cloudwatch.NewMetricAlarm` with conditional branching for simple vs metric query mode
- `outputs.go` — alarm_arn, alarm_name constants

### Terraform Module (5 files)

- `main.tf` — single `aws_cloudwatch_metric_alarm` with dynamic `metric_query` blocks
- Feature parity with Pulumi module

### Validation Tests

- 39 spec tests (15 happy path + 24 failure scenarios), all passing
- Tests cover all 14 CEL validations including mode mutual exclusivity, action limits, period validation, and API envelope checks

### Presets

- `01-cpu-utilization-alarm` — EC2 CPU with 2-of-3 M-of-N evaluation
- `02-error-rate-metric-math` — ALB 5xx error rate via 3-query metric math pattern
- `03-production-multi-action` — SQS depth with alarm/OK/insufficient-data actions

## Benefits

- **Completes Phase 1** — all 15 core AWS resource kinds (+ 3 splits) are now forged
- **Metric math support** — enables error rate, latency ratio, and computed metric alarms that ~35% of production users need
- **Infra chart composability** — alarm actions use `StringValueOrRef` with `default_kind = AwsSnsTopic`, enabling DAG wiring in infra charts
- **M-of-N evaluation** — reduces false positive alerts, a common operational pain point

## Impact

- **Users**: Can define CloudWatch metric alarms declaratively alongside the resources they monitor
- **Infra charts**: Monitoring alarms can be composed into environment charts (e.g., ECS + ALB + SQS + alarms)
- **Project milestone**: Phase 1 of the AWS expansion is complete (18 new resource kinds forged)

## Related Work

- R14: AwsCloudwatchLogGroup — sibling in the CloudWatch family, commonly referenced by log-based metric filters that alarms monitor
- R01-R14: All previous Phase 1 AWS kinds that produce metrics CloudWatch alarms can monitor
- Phase 2 (R16-R25): 10 important services queued next (Kinesis, Athena, Glue, Redshift, MSK, SageMaker, App Runner, MWAA, Transit Gateway)

---

**Status**: Production Ready
**Timeline**: Single session
