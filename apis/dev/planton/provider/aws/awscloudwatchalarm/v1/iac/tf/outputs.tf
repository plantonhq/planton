output "alarm_arn" {
  description = "The ARN of the CloudWatch Metric Alarm."
  value       = aws_cloudwatch_metric_alarm.this.arn
}

output "alarm_name" {
  description = "The name of the CloudWatch Metric Alarm."
  value       = aws_cloudwatch_metric_alarm.this.alarm_name
}
