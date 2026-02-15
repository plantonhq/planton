# AwsCloudwatchLogGroup Pulumi Module

Provisions an AWS CloudWatch Logs log group with configurable retention, KMS encryption, and log group class.

## Resources Created

- `aws:cloudwatch:LogGroup` — The CloudWatch log group

## Inputs

Accepts `AwsCloudwatchLogGroupStackInput` which includes:
- `target` — The AwsCloudwatchLogGroup KRM resource (metadata + spec)
- `provider_config` — AWS provider credentials and region

## Outputs

| Key | Description |
|-----|-------------|
| `log_group_arn` | The ARN of the log group |
| `log_group_name` | The name of the log group |

## Local Development

```bash
./debug.sh preview  # dry-run
./debug.sh up       # deploy
./debug.sh destroy  # tear down
```
