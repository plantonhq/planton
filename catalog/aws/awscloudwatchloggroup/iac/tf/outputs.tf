# Outputs — must stay in lockstep with
# AwsCloudwatchLogGroupOutputs. The provider trims the API's ":*" suffix
# from the ARN on read, so downstream consumers that need the wildcard form
# (e.g. Step Functions logging) append it themselves.
output "log_group_arn" {
  description = "ARN of the CloudWatch log group (without the :* suffix)."
  value       = aws_cloudwatch_log_group.this.arn
}

output "log_group_name" {
  description = "Name of the CloudWatch log group — the join key for services that address log groups by name."
  value       = aws_cloudwatch_log_group.this.name
}
