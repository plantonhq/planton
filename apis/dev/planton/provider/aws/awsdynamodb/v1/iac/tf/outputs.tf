output "table_name" {
  description = "The DynamoDB table name."
  value       = aws_dynamodb_table.this.name
}

output "table_arn" {
  description = "The DynamoDB table ARN."
  value       = aws_dynamodb_table.this.arn
}

output "table_id" {
  description = "The provider-assigned table ID."
  value       = aws_dynamodb_table.this.id
}

output "stream_arn" {
  description = "The stream ARN if streams are enabled."
  value       = aws_dynamodb_table.this.stream_arn
}

output "stream_label" {
  description = "The stream label if streams are enabled."
  value       = aws_dynamodb_table.this.stream_label
}


