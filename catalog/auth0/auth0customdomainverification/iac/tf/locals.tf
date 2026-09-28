# Local values for the Auth0CustomDomainVerification module.
#
# custom_domain_id arrives as a plain string: a reference to an
# Auth0CustomDomain is resolved to its status.outputs.id before the module runs.
locals {
  custom_domain_id = var.spec.custom_domain_id
}
