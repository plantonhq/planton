output "consumer_arn" {
  description = "The ARN of the registered Kinesis stream consumer."
  value       = aws_kinesis_stream_consumer.this.arn
}

output "consumer_name" {
  description = "The name of the registered stream consumer."
  value       = aws_kinesis_stream_consumer.this.name
}

output "stream_arn" {
  description = "The ARN of the parent Kinesis Data Stream."
  value       = aws_kinesis_stream_consumer.this.stream_arn
}

output "creation_timestamp" {
  description = "RFC3339 timestamp of when the consumer was registered."
  value       = aws_kinesis_stream_consumer.this.creation_timestamp
}
