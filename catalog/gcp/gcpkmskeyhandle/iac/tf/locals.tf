locals {
  # Empty project means the provider's default project, read from the
  # provider's own configuration (google_client_config, no API call) --
  # identical to the Pulumi module's GetClientConfig fallback.
  needs_client_project = var.spec.project_id == ""
  project = (
    var.spec.project_id != "" ? trimprefix(var.spec.project_id, "projects/") :
    data.google_client_config.current[0].project
  )

  # The handle's ID defaults to metadata.name -- identical to the Pulumi
  # module.
  key_handle_name = var.spec.key_handle_name != "" ? var.spec.key_handle_name : var.metadata.name
}
