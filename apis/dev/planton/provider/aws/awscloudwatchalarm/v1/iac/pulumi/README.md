# AwsCloudwatchAlarm Pulumi Module

Provisions an AWS CloudWatch metric alarm with support for simple metric alarms and metric math expressions, configurable M-of-N evaluation, and multi-action notifications.

## Resources Created

- `aws:cloudwatch:MetricAlarm` — The CloudWatch metric alarm

## Inputs

Accepts `AwsCloudwatchAlarmStackInput` which includes:
- `target` — The AwsCloudwatchAlarm KRM resource (metadata + spec)
- `provider_config` — AWS provider credentials and region

## Outputs

| Key | Description |
|-----|-------------|
| `alarm_arn` | The ARN of the metric alarm |
| `alarm_name` | The name of the metric alarm |

## Local Development

```bash
./debug.sh preview  # dry-run
./debug.sh up       # deploy
./debug.sh destroy  # tear down
```
