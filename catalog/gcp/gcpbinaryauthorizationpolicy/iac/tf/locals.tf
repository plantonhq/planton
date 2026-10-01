locals {
  # Empty project means the provider's default project, read from the
  # provider's own configuration (google_client_config, no API call) --
  # identical to the Pulumi module's GetClientConfig fallback.
  needs_client_project = var.spec.project_id == ""
  project = (
    var.spec.project_id != "" ? trimprefix(var.spec.project_id, "projects/") :
    data.google_client_config.current[0].project
  )

  description                   = var.spec.description != "" ? var.spec.description : null
  global_policy_evaluation_mode = var.spec.global_policy_evaluation_mode != "" ? var.spec.global_policy_evaluation_mode : null
  deletion_policy               = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
