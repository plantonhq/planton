output "delivery_stream_arn" {
  description = "The ARN of the Kinesis Firehose delivery stream."
  value       = aws_kinesis_firehose_delivery_stream.this.arn
}

output "delivery_stream_name" {
  description = "The name of the Kinesis Firehose delivery stream."
  value       = aws_kinesis_firehose_delivery_stream.this.name
}
