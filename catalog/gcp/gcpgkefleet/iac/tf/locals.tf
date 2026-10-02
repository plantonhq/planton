locals {
  # Empty project means the provider's default project, read from the
  # provider's own configuration (google_client_config, no API call) --
  # identical to the Pulumi module's GetClientConfig fallback. The fleet
  # exports its project as the value every fleet child references, so it
  # must always be concrete.
  needs_client_project = var.spec.project_id == ""
  project = (
    var.spec.project_id != "" ? trimprefix(var.spec.project_id, "projects/") :
    data.google_client_config.current[0].project
  )

  # Empty optional strings become null so the provider omits them instead
  # of sending values it would reject or diff on.
  display_name    = var.spec.display_name != "" ? var.spec.display_name : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  binary_authorization_config = try(var.spec.default_cluster_config.binary_authorization_config, null)
  security_posture_config     = try(var.spec.default_cluster_config.security_posture_config, null)
}
