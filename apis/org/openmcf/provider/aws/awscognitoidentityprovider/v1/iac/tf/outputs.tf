output "provider_name" {
  description = "The name of the identity provider as registered in the User Pool."
  value       = aws_cognito_identity_provider.this.provider_name
}

output "provider_type" {
  description = "The type of the identity provider."
  value       = aws_cognito_identity_provider.this.provider_type
}
