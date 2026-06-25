output "application_id" {
  description = "The unique ID of the Access application"
  value       = cloudflare_zero_trust_access_application.main.id
}

output "aud" {
  description = "The application's audience (AUD) tag, used to validate Access JWTs"
  value       = cloudflare_zero_trust_access_application.main.aud
}

output "domain" {
  description = "The primary domain protected by this application"
  value       = cloudflare_zero_trust_access_application.main.domain
}

output "saas_client_id" {
  description = "The issued OAuth client ID (SaaS/OIDC applications)"
  value       = try(cloudflare_zero_trust_access_application.main.saas_app.client_id, "")
}

output "saas_client_secret" {
  description = "The issued OAuth client secret (SaaS/OIDC applications)"
  value       = try(cloudflare_zero_trust_access_application.main.saas_app.client_secret, "")
  sensitive   = true
}

output "saas_public_key" {
  description = "The IdP-facing public key (SaaS/SAML applications)"
  value       = try(cloudflare_zero_trust_access_application.main.saas_app.public_key, "")
}

output "saas_sso_endpoint" {
  description = "The single sign-on endpoint URL (SaaS/SAML applications)"
  value       = try(cloudflare_zero_trust_access_application.main.saas_app.sso_endpoint, "")
}

output "saas_idp_entity_id" {
  description = "The IdP entity ID (SaaS/SAML applications)"
  value       = try(cloudflare_zero_trust_access_application.main.saas_app.idp_entity_id, "")
}
