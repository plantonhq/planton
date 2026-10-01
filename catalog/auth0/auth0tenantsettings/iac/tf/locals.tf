# Local values for the Auth0TenantSettings module. The Pulumi module's
# locals.go and tenantArgs (iac/pulumi/module/tenant.go) are its twins -- keep
# them in lockstep.
#
# A setting left unset is NOT MANAGED: it renders as null, so the provider never
# sends it and the tenant keeps whatever value it already carries.
# - The four presentation strings and the two references arrive as "" when
#   unset (the proto's zero value), so each maps "" to null here.
# - Lists arrive as [] when unset, so each maps an empty list to null (an empty
#   list would be sent, clearing the tenant's).
# - Every other scalar is optional in the proto and arrives as null when unset;
#   main.tf passes those through as they are.
locals {
  friendly_name = var.spec.friendly_name != "" ? var.spec.friendly_name : null
  picture_url   = var.spec.picture_url != "" ? var.spec.picture_url : null
  support_email = var.spec.support_email != "" ? var.spec.support_email : null
  support_url   = var.spec.support_url != "" ? var.spec.support_url : null

  # References arrive resolved to their values.
  default_audience  = var.spec.default_audience != "" ? var.spec.default_audience : null
  default_directory = var.spec.default_directory != "" ? var.spec.default_directory : null

  enabled_locales      = length(var.spec.enabled_locales) > 0 ? var.spec.enabled_locales : null
  acr_values_supported = length(var.spec.acr_values_supported) > 0 ? var.spec.acr_values_supported : null
  allowed_logout_urls  = length(var.spec.allowed_logout_urls) > 0 ? var.spec.allowed_logout_urls : null

  # The provider takes this one as the string "true" or "false" (its third
  # value, "null", is what leaving it unset sends).
  skip_non_verifiable_callback_uri_confirmation_prompt = (
    var.spec.skip_non_verifiable_callback_uri_confirmation_prompt == null ? null :
    tostring(var.spec.skip_non_verifiable_callback_uri_confirmation_prompt)
  )

  # The tenant's default domain, or null when the spec leaves it unmanaged (the
  # default-domain resource is then not declared). A reference arrives resolved
  # to its value.
  default_custom_domain = var.spec.default_custom_domain != "" ? var.spec.default_custom_domain : null

  # Whether the spec declares any tenant setting beyond the default domain,
  # which has a resource of its own. Auth0 refuses a tenant update that
  # carries no setting ("Too few properties defined (0)"), so a spec that
  # declares none does not declare auth0_tenant at all, and its settings are
  # read through the provider's data source instead. A declared setting is a
  # non-null value that is not an empty string or an empty list (every
  # optional scalar and block arrives as null when unset; a false toggle is
  # declared). The Pulumi module's managesTenantSettings is its twin.
  manages_tenant_settings = length([
    for name, value in var.spec : name
    if name != "default_custom_domain" && value != null && try(length(value) > 0, true)
  ]) > 0

  # The tenant as it carries its settings after the apply, from whichever of
  # the resource and the data source this spec declares.
  tenant = local.manages_tenant_settings ? auth0_tenant.this[0] : data.auth0_tenant.current[0]
}
