# The API key. Its identity is the spec's key_id (the provider's `name`);
# the project and the optional service-account binding are immutable with
# it, so a change to any of the three destroys and recreates the key --
# which rotates the key string every client holds. Restrictions are
# updated in place.
#
# The spec's deletion_policy maps straight onto the provider's: DELETE
# soft-deletes the key (recoverable for 30 days, key_id reserved for the
# window), PREVENT fails the destroy, ABANDON drops it from state and
# leaves the key live. Sent only when set so the provider's default
# (DELETE) stays the provider's.
resource "google_apikeys_key" "this" {
  project = local.project_id
  name    = var.spec.key_id

  display_name          = var.spec.display_name != "" ? var.spec.display_name : null
  service_account_email = var.spec.service_account_email != "" ? var.spec.service_account_email : null
  deletion_policy       = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Restrictions: at most one client arm (the spec's CEL enforces Google's
  # rule) plus any number of API targets. Every arm is a dynamic block
  # emitted exactly when the spec declares it.
  dynamic "restrictions" {
    for_each = local.restrictions
    content {
      dynamic "android_key_restrictions" {
        for_each = restrictions.value.android_key_restrictions != null ? [restrictions.value.android_key_restrictions] : []
        content {
          dynamic "allowed_applications" {
            for_each = android_key_restrictions.value.allowed_applications
            content {
              package_name     = allowed_applications.value.package_name
              sha1_fingerprint = allowed_applications.value.sha1_fingerprint
            }
          }
        }
      }

      dynamic "ios_key_restrictions" {
        for_each = restrictions.value.ios_key_restrictions != null ? [restrictions.value.ios_key_restrictions] : []
        content {
          allowed_bundle_ids = ios_key_restrictions.value.allowed_bundle_ids
        }
      }

      dynamic "browser_key_restrictions" {
        for_each = restrictions.value.browser_key_restrictions != null ? [restrictions.value.browser_key_restrictions] : []
        content {
          allowed_referrers = browser_key_restrictions.value.allowed_referrers
        }
      }

      dynamic "server_key_restrictions" {
        for_each = restrictions.value.server_key_restrictions != null ? [restrictions.value.server_key_restrictions] : []
        content {
          allowed_ips = server_key_restrictions.value.allowed_ips
        }
      }

      dynamic "api_targets" {
        for_each = restrictions.value.api_targets
        content {
          service = api_targets.value.service
          methods = length(api_targets.value.methods) > 0 ? api_targets.value.methods : null
        }
      }
    }
  }
}
