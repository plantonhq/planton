# Local values for the Auth0CustomDomain module.
#
# An optional setting left unset renders as null, so the provider never sends
# it and Auth0 applies its own default. Empty strings are the proto's zero value
# for "unset", so each maps "" to null here.
locals {
  custom_client_ip_header  = var.spec.custom_client_ip_header != "" ? var.spec.custom_client_ip_header : null
  tls_policy               = var.spec.tls_policy != "" ? var.spec.tls_policy : null
  relying_party_identifier = var.spec.relying_party_identifier != "" ? var.spec.relying_party_identifier : null
  domain_metadata          = length(var.spec.domain_metadata) > 0 ? var.spec.domain_metadata : null

  # The DNS record that proves control of the domain, picked from the
  # verification methods Auth0 returns. The CNAME is preferred when offered --
  # an Auth0-managed domain is served through it, so it is the one record that
  # both proves control and carries traffic -- and otherwise the first method
  # (the TXT record of a self-managed domain). A method's "domain" names a TXT
  # record; a CNAME's name is the custom domain itself. The Pulumi module's
  # verificationRecord applies the same rule -- keep them in lockstep.
  verification_methods = flatten([for v in auth0_custom_domain.this.verification : v.methods])
  cname_methods        = [for m in local.verification_methods : m if lower(lookup(m, "name", "")) == "cname"]
  chosen_method = (
    length(local.cname_methods) > 0 ? local.cname_methods[0] :
    length(local.verification_methods) > 0 ? local.verification_methods[0] : {}
  )
  dns_record_name  = lookup(local.chosen_method, "domain", "") != "" ? lookup(local.chosen_method, "domain", "") : auth0_custom_domain.this.domain
  dns_record_type  = upper(lookup(local.chosen_method, "name", ""))
  dns_record_value = lookup(local.chosen_method, "record", "")
}
