output "user_pool_id" {
  description = "The Cognito user pool identifier"
  value       = aws_cognito_user_pool.this.id
}

output "user_pool_arn" {
  description = "The ARN of the Cognito user pool"
  value       = aws_cognito_user_pool.this.arn
}

output "user_pool_endpoint" {
  description = "The OIDC issuer endpoint URL"
  value       = aws_cognito_user_pool.this.endpoint
}

output "user_pool_domain" {
  description = "The full domain URL for the hosted UI"
  value = local.has_domain ? (
    local.is_custom_domain
    ? "https://${local.spec.domain.domain}"
    : "https://${local.spec.domain.domain}.auth.${data.aws_region.current.name}.amazoncognito.com"
  ) : ""
}

output "cloudfront_distribution_arn" {
  description = "The CloudFront distribution ARN for custom domains"
  value       = local.has_domain && local.is_custom_domain ? try(aws_cognito_user_pool_domain.this[0].cloudfront_distribution_arn, "") : ""
}

output "client_ids" {
  description = "Map of client name to client ID"
  value       = { for k, v in aws_cognito_user_pool_client.this : k => v.id }
}

output "client_secrets" {
  description = "Map of client name to client secret (sensitive)"
  value       = { for k, v in aws_cognito_user_pool_client.this : k => v.client_secret if try(local.client_map[k].generate_secret, false) }
  sensitive   = true
}

data "aws_region" "current" {}
