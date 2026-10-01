# Auth0CustomDomain Outputs
# Maps to the Auth0CustomDomainStackOutputs protobuf message: the domain as
# Auth0 created it, and the DNS record that proves control of it.

output "id" {
  description = "The custom domain's identifier in Auth0 (cd_...)"
  value       = auth0_custom_domain.this.id
}

output "domain" {
  description = "The custom domain's name, as created"
  value       = auth0_custom_domain.this.domain
}

output "status" {
  description = "Where the domain is in its life (pending_verification, ready, ...)"
  value       = auth0_custom_domain.this.status
}

output "origin_domain_name" {
  description = "The tenant host the domain serves from"
  value       = auth0_custom_domain.this.origin_domain_name
}

output "dns_record_name" {
  description = "The fully qualified name of the DNS record that proves control of the domain"
  value       = local.dns_record_name
}

output "dns_record_type" {
  description = "The type of the DNS record that proves control of the domain (CNAME or TXT)"
  value       = local.dns_record_type
}

output "dns_record_value" {
  description = "The value of the DNS record that proves control of the domain"
  value       = local.dns_record_value
}
