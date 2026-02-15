output "api_id" {
  description = "The API Gateway API identifier. Used for constructing resource ARNs and referencing the API in other AWS services."
  value       = aws_apigatewayv2_api.this.id
}

output "api_endpoint" {
  description = "The default endpoint URL of the API. Format: https://{api-id}.execute-api.{region}.amazonaws.com"
  value       = aws_apigatewayv2_api.this.api_endpoint
}

output "api_arn" {
  description = "The Amazon Resource Name (ARN) of the API. Used for IAM policies and resource-based permissions."
  value       = aws_apigatewayv2_api.this.arn
}

output "execution_arn" {
  description = "The execution ARN prefix for the API. Used in Lambda resource-based policies to grant API Gateway permission to invoke Lambda functions. Format: arn:aws:execute-api:{region}:{account-id}:{api-id}"
  value       = aws_apigatewayv2_api.this.execution_arn
}

output "stage_invoke_url" {
  description = "The invoke URL for the deployed stage. For the \"$default\" stage this is the same as api_endpoint. For named stages the URL includes the stage name."
  value       = aws_apigatewayv2_stage.this.invoke_url
}

output "stage_name" {
  description = "The name of the deployed stage (e.g., \"$default\", \"prod\")."
  value       = aws_apigatewayv2_stage.this.name
}
