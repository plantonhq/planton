# main.tf

# Lookup the Cloudflare zone to get the account id (Access policies and
# account-scoped applications require it).
data "cloudflare_zone" "main" {
  zone_id = var.spec.zone_id
}

# Create the Access Policy. In v5 a policy is a standalone, account-scoped
# object; the application references it via its policies list.
resource "cloudflare_zero_trust_access_policy" "main" {
  account_id = data.cloudflare_zone.main.account.id
  name       = "default-policy"
  decision   = var.spec.policy_type == "BLOCK" ? "deny" : "allow"

  # Include rules. Each element carries the full attribute set (unused keys are
  # null) so all elements share one object type.
  include = concat(
    [for e in var.spec.allowed_emails : { email = { email = e }, group = null }],
    [for g in var.spec.allowed_google_groups : { email = null, group = { id = g } }],
  )

  # Require MFA when requested.
  require = var.spec.require_mfa ? [{ auth_method = { auth_method = "mfa" } }] : []
}

# Create the Cloudflare Zero Trust Access Application and attach the policy.
resource "cloudflare_zero_trust_access_application" "main" {
  account_id = data.cloudflare_zone.main.account.id
  name       = var.spec.application_name
  domain     = var.spec.hostname
  type       = "self_hosted"

  # Session duration as a duration string (defaults to 24h).
  session_duration = var.spec.session_duration_minutes > 0 ? "${var.spec.session_duration_minutes}m" : "24h"

  # Reference the policy; precedence is the evaluation order.
  policies = [{
    id         = cloudflare_zero_trust_access_policy.main.id
    precedence = 1
  }]
}
