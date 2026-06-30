output "topic_arn" {
  description = "The ARN of the SNS topic."
  value       = aws_sns_topic.this.arn
}

output "topic_name" {
  description = "The name of the SNS topic."
  value       = aws_sns_topic.this.name
}

output "subscription_arns" {
  description = "Map of subscription name to subscription ARN."
  value = {
    for key, sub in aws_sns_topic_subscription.this : key => sub.arn
  }
}
