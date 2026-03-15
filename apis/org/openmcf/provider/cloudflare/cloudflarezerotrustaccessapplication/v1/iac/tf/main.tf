# main.tf

# Lookup the Cloudflare zone to get account_id
data "cloudflare_zone" "main" {
  zone_id = var.spec.zone_id.value
}

# Create the Cloudflare Zero Trust Access Application
resource "cloudflare_access_application" "main" {
  account_id = data.cloudflare_zone.main.account_id
  name       = var.spec.application_name
  domain     = var.spec.hostname
  type       = "self_hosted"

  # Set session duration if specified (convert minutes to duration string)
  session_duration = var.spec.session_duration_minutes > 0 ? "${var.spec.session_duration_minutes}m" : "24h"
}

# Create the Access Policy for the application
resource "cloudflare_access_policy" "main" {
  account_id     = data.cloudflare_zone.main.account_id
  application_id = cloudflare_access_application.main.id
  name           = "default-policy"
  precedence     = 1
  decision       = var.spec.policy_type == "BLOCK" ? "deny" : "allow"

  # Include rules for allowed emails
  dynamic "include" {
    for_each = var.spec.allowed_emails
    content {
      email = [include.value]
    }
  }

  # Include rules for allowed Google groups
  dynamic "include" {
    for_each = var.spec.allowed_google_groups
    content {
      group = [include.value]
    }
  }

  # Require MFA if specified
  dynamic "require" {
    for_each = var.spec.require_mfa ? [1] : []
    content {
      auth_method = ["mfa"]
    }
  }
}

