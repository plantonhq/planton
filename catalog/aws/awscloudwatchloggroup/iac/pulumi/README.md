# AwsCloudwatchLogGroup Pulumi Module

Provisions an AWS CloudWatch Logs log group with configurable retention, KMS
encryption, log group class, deletion protection, and the group-scoped
satellites: metric filters, subscription filters, a data protection policy,
and a field index policy.

## Resources Created

- `aws:cloudwatch:LogGroup` — The CloudWatch log group
- `aws:cloudwatch:LogMetricFilter` — One per named metric filter
- `aws:cloudwatch:LogSubscriptionFilter` — One per named subscription filter (max 2)
- `aws:cloudwatch:LogDataProtectionPolicy` — When a data protection policy is configured
- `aws:cloudwatch:LogIndexPolicy` — When a field index policy is configured

## Inputs

Accepts `AwsCloudwatchLogGroupIacInput` which includes:
- `target` — The AwsCloudwatchLogGroup KRM resource (metadata + spec)
- `provider_config` — AWS provider credentials and region

## Outputs

| Key | Description |
|-----|-------------|
| `log_group_arn` | The ARN of the log group (without the `:*` suffix) |
| `log_group_name` | The name of the log group |

## Local Development

```bash
./debug.sh preview  # dry-run
./debug.sh up       # deploy
./debug.sh destroy  # tear down
```
