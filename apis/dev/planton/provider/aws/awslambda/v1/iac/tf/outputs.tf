output "function_arn" {
  description = "The ARN of the Lambda function."
  value       = try(aws_lambda_function.this[0].arn, null)
}

output "function_name" {
  description = "The name of the Lambda function."
  value       = try(aws_lambda_function.this[0].function_name, null)
}

output "log_group_name" {
  description = "CloudWatch Logs log group name for the function."
  value       = try(aws_cloudwatch_log_group.lambda[0].name, null)
}

output "role_arn" {
  description = "Execution role ARN used by the Lambda function."
  value       = aws_iam_role.lambda.arn
}

output "layer_arns" {
  description = "List of layer ARNs attached to the Lambda function."
  value       = try(aws_lambda_function.this[0].layers, [])
}


