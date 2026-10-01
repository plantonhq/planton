# Auth0Prompt Main Resources
#
# auth0_prompt manages the login-flow settings of the tenant the provider's
# credential belongs to; a tenant has exactly one set. Only the settings the
# spec declares are sent, each null when the spec leaves it unset, so the
# tenant keeps the others as they are. Destroy drops the resource from state
# and leaves the last-applied values in place: Auth0 has no delete for the
# prompt settings.
resource "auth0_prompt" "this" {
  universal_login_experience     = local.universal_login_experience
  identifier_first               = local.identifier_first
  webauthn_platform_first_factor = local.webauthn_platform_first_factor
}
