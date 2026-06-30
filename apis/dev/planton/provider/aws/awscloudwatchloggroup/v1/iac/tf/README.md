# AwsCloudwatchLogGroup Terraform Module

Provisions an AWS CloudWatch Logs log group with configurable retention, KMS encryption, and log group class.

## Resources Created

- `aws_cloudwatch_log_group.this` — The CloudWatch log group

## Inputs

| Variable | Description |
|----------|-------------|
| `metadata` | Resource metadata (name, org, env, id, labels) |
| `spec` | AwsCloudwatchLogGroupSpec — desired configuration |

## Outputs

| Output | Description |
|--------|-------------|
| `log_group_arn` | The ARN of the log group |
| `log_group_name` | The name of the log group |

## Provider

Requires `hashicorp/aws` provider version `5.82.0`.
