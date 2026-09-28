# Auth0TenantSettings Main Resources
#
# auth0_tenant manages the settings of the EXISTING tenant the provider's
# credential belongs to; the Management API cannot create or delete a tenant.
# Only the four presentation settings are set here, each null when the spec
# leaves it unset, so the tenant's other settings (session lifetimes, flags,
# error pages) are never touched; a spec that manages only the default domain
# sends none of them. Destroy drops the resource from state and
# leaves the last-applied values in place: Auth0 has no delete for tenant
# settings.
resource "auth0_tenant" "this" {
  friendly_name = local.friendly_name
  picture_url   = local.picture_url
  support_email = local.support_email
  support_url   = local.support_url
}

# The tenant's default domain -- the one its email links and Management API
# notifications use -- declared only when the spec manages it. Auth0 has no way
# to unset a default, so the resource's delete only forgets it: destroy leaves
# the last-applied default in place.
resource "auth0_custom_domain_default" "this" {
  count  = local.default_custom_domain != null ? 1 : 0
  domain = local.default_custom_domain
}
