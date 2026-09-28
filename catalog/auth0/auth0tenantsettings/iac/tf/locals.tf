# Local values for the Auth0TenantSettings module.
#
# A field left unset is NOT MANAGED: it renders as null, so the provider never
# sends it and the tenant keeps whatever value it already carries. Empty strings
# are the proto's zero value for "unset", so every field maps "" to null here.
locals {
  friendly_name = var.spec.friendly_name != "" ? var.spec.friendly_name : null
  picture_url   = var.spec.picture_url != "" ? var.spec.picture_url : null
  support_email = var.spec.support_email != "" ? var.spec.support_email : null
  support_url   = var.spec.support_url != "" ? var.spec.support_url : null

  # The tenant's default domain, or null when the spec leaves it unmanaged (the
  # default-domain resource is then not declared). A reference arrives resolved
  # to its value.
  default_custom_domain = var.spec.default_custom_domain != "" ? var.spec.default_custom_domain : null
}
