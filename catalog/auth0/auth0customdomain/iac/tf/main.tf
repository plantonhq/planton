# Auth0CustomDomain Main Resources
#
# The custom domain is created in the tenant the provider's credential belongs
# to. Auth0 answers with the DNS record that proves control of the name
# (exported as dns_record_*); the domain serves nothing until an
# Auth0CustomDomainVerification has confirmed that record. Changing the domain
# or its type replaces it. Destroy deletes the domain.
resource "auth0_custom_domain" "this" {
  domain                   = var.spec.domain
  type                     = var.spec.type
  custom_client_ip_header  = local.custom_client_ip_header
  tls_policy               = local.tls_policy
  relying_party_identifier = local.relying_party_identifier
  domain_metadata          = local.domain_metadata
}
