# The provider's default project, read only when the manifest names none.
data "google_client_config" "current" {
  count = local.needs_client_project ? 1 : 0
}

# The Fleet API (GKE Hub). disable_on_destroy is false: tearing down the
# fleet's settings must never disable the API for the clusters and
# features that still use it.
resource "google_project_service" "gkehub_api" {
  project = local.project
  service = "gkehub.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The project's one fleet ("default" in "global"). Everything but the
# project updates in place. Fleet labels and the compliance posture
# default are not sent: the pinned Pulumi SDK lacks both, and the two
# engines send the same arguments.
resource "google_gke_hub_fleet" "this" {
  project         = local.project
  display_name    = local.display_name
  deletion_policy = local.deletion_policy

  dynamic "default_cluster_config" {
    for_each = (local.binary_authorization_config != null || local.security_posture_config != null) ? [1] : []
    content {
      dynamic "binary_authorization_config" {
        for_each = local.binary_authorization_config != null ? [local.binary_authorization_config] : []
        content {
          evaluation_mode = binary_authorization_config.value.evaluation_mode != "" ? binary_authorization_config.value.evaluation_mode : null

          dynamic "policy_bindings" {
            for_each = binary_authorization_config.value.policy_bindings
            content {
              name = policy_bindings.value
            }
          }
        }
      }

      dynamic "security_posture_config" {
        for_each = local.security_posture_config != null ? [local.security_posture_config] : []
        content {
          mode               = security_posture_config.value.mode != "" ? security_posture_config.value.mode : null
          vulnerability_mode = security_posture_config.value.vulnerability_mode != "" ? security_posture_config.value.vulnerability_mode : null
        }
      }
    }
  }

  depends_on = [google_project_service.gkehub_api]
}
