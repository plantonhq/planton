# The provider's default project, read only when the manifest names none.
data "google_client_config" "current" {
  count = local.needs_client_project ? 1 : 0
}

# disable_on_destroy is false: destroying the policy writes Google's
# default policy back, and GKE clusters keep evaluating it through the API.
resource "google_project_service" "binaryauthorization_api" {
  project = local.project
  service = "binaryauthorization.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The project's one policy. Every create and update is a full PUT (this
# block replaces whatever policy the project had); destroy under DELETE
# writes Google's default policy back -- allow every image.
resource "google_binary_authorization_policy" "this" {
  project                       = local.project
  description                   = local.description
  global_policy_evaluation_mode = local.global_policy_evaluation_mode
  deletion_policy               = local.deletion_policy

  # The spec lifts Google's one-field pattern blocks to strings.
  dynamic "admission_whitelist_patterns" {
    for_each = var.spec.admission_whitelist_patterns
    content {
      name_pattern = admission_whitelist_patterns.value
    }
  }

  default_admission_rule {
    evaluation_mode         = var.spec.default_admission_rule.evaluation_mode
    enforcement_mode        = var.spec.default_admission_rule.enforcement_mode
    require_attestations_by = length(var.spec.default_admission_rule.require_attestations_by) > 0 ? var.spec.default_admission_rule.require_attestations_by : null
  }

  dynamic "cluster_admission_rules" {
    for_each = var.spec.cluster_admission_rules
    content {
      cluster                 = cluster_admission_rules.value.cluster
      evaluation_mode         = cluster_admission_rules.value.evaluation_mode
      enforcement_mode        = cluster_admission_rules.value.enforcement_mode
      require_attestations_by = length(cluster_admission_rules.value.require_attestations_by) > 0 ? cluster_admission_rules.value.require_attestations_by : null
    }
  }

  depends_on = [google_project_service.binaryauthorization_api]
}
