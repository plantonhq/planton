# AwsCloudwatchAlarm Terraform Module

Provisions an AWS CloudWatch metric alarm with support for simple metric alarms and metric math expressions, configurable M-of-N evaluation, and multi-action notifications.

## Resources Created

- `aws_cloudwatch_metric_alarm.this` — The CloudWatch metric alarm

## Inputs

| Variable | Description |
|----------|-------------|
| `metadata` | Resource metadata (name, org, env, id, labels) |
| `spec` | AwsCloudwatchAlarmSpec — desired configuration |

## Outputs

| Output | Description |
|--------|-------------|
| `alarm_arn` | The ARN of the metric alarm |
| `alarm_name` | The name of the metric alarm |

## Provider

Requires `hashicorp/aws` provider version `5.82.0`.
