output "bus_name" {
  description = "The name of the EventBridge custom event bus."
  value       = aws_cloudwatch_event_bus.this.name
}

output "bus_arn" {
  description = "The ARN of the EventBridge custom event bus."
  value       = aws_cloudwatch_event_bus.this.arn
}
