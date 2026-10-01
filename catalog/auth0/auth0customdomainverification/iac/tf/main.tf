# Auth0CustomDomainVerification Main Resources
#
# Verification asks Auth0 to check the custom domain's DNS record and waits
# until the domain is ready -- the provider polls until Auth0 reports "ready"
# and fails naming the last status otherwise. It is a one-time action with no
# update; its delete is the provider's no-op, so destroy leaves the domain
# verified.
resource "auth0_custom_domain_verification" "this" {
  custom_domain_id = local.custom_domain_id
}

# The domain's name, read back by id once verification has finished, so the
# name exported is one Auth0 has verified.
data "auth0_custom_domain" "verified" {
  custom_domain_id = auth0_custom_domain_verification.this.custom_domain_id
}
