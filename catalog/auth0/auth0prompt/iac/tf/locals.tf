# Local values for the Auth0Prompt module.
#
# A field left unset is NOT MANAGED: it renders as null, so the provider never
# sends it and the tenant keeps whatever value it already carries. The empty
# string is the proto's zero value for an unset experience, and an unset
# optional bool arrives as null.
locals {
  universal_login_experience     = var.spec.universal_login_experience != "" ? var.spec.universal_login_experience : null
  identifier_first               = try(var.spec.identifier_first, null)
  webauthn_platform_first_factor = try(var.spec.webauthn_platform_first_factor, null)
}
