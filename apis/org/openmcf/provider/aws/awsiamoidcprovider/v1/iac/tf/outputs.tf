output "provider_arn" {
  description = "The ARN of the IAM OIDC provider (referenced as a Federated principal in IAM role trust policies)."
  value       = aws_iam_openid_connect_provider.this.arn
}

output "provider_url" {
  description = "The issuer URL AWS stored for this provider, with the https:// scheme stripped."
  value       = aws_iam_openid_connect_provider.this.url
}
