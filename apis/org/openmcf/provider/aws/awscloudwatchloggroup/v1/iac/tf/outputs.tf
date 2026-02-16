output "log_group_arn" {
  description = "The ARN of the CloudWatch Log Group."
  value       = aws_cloudwatch_log_group.this.arn
}

output "log_group_name" {
  description = "The name of the CloudWatch Log Group."
  value       = aws_cloudwatch_log_group.this.name
}
