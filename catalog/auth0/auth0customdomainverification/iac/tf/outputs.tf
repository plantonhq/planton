# Auth0CustomDomainVerification Outputs
# Maps to the Auth0CustomDomainVerificationOutputs protobuf message: the
# custom domain once Auth0 has verified it.

output "custom_domain_id" {
  description = "The verified custom domain's identifier in Auth0 (cd_...)"
  value       = auth0_custom_domain_verification.this.custom_domain_id
}

output "domain" {
  description = "The verified domain's name, ready to serve sign-in"
  value       = data.auth0_custom_domain.verified.domain
}

output "origin_domain_name" {
  description = "The tenant host the domain serves from"
  value       = auth0_custom_domain_verification.this.origin_domain_name
}

output "cname_api_key" {
  description = "The key a self-managed domain's proxy sends in the cname-api-key header (empty for an Auth0-managed domain)"
  value       = auth0_custom_domain_verification.this.cname_api_key
  sensitive   = true
}
