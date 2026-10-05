# Outputs — must stay in lockstep with
# AwsCloudwatchCompositeAlarmOutputs.
output "alarm_arn" {
  description = "ARN of the CloudWatch composite alarm."
  value       = aws_cloudwatch_composite_alarm.this.arn
}

output "alarm_name" {
  description = "Name of the composite alarm — the join key other composite alarms use in their rule expressions."
  value       = aws_cloudwatch_composite_alarm.this.alarm_name
}
