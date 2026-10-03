# Pulumi Module to Deploy AwsCloudwatchCompositeAlarm

This module provisions a CloudWatch composite alarm — the boolean combination
of other alarms' states used to page once for shared-cause outages, express
alerting dependencies, and gate actions behind maintenance suppression.

## Features

- **Boolean rule expression** over other alarms (`ALARM`/`OK`/
  `INSUFFICIENT_DATA` with `AND`/`OR`/`NOT`), addressed by alarm name
- **Three action lists** (alarm / ok / insufficient-data), pre-resolved from
  references to SNS topics
- **Actions suppressor** with wait/extension windows — the maintenance-window
  mechanism that silences paging without pausing evaluation
- **Presence-aware `actions_enabled`** (ForceNew on this resource — a change
  replaces the alarm)

## Usage

The module is executed by the Planton runtime with an IaC input carrying the
target resource and provider config. For a local run:

```shell
planton pulumi up --manifest e2e/manifest.yaml
planton pulumi destroy --manifest e2e/manifest.yaml
```

## Outputs

| Output | Description |
| --- | --- |
| `alarm_arn` | ARN of the composite alarm |
| `alarm_name` | Name of the composite alarm (the join key for parent composites) |
