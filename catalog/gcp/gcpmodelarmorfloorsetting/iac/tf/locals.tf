locals {
  # Scope selection. Exactly one arm renders Google's parent; an empty scope
  # means the provider's default project, read from the provider's own
  # configuration (google_client_config, no API call) -- the same fallback
  # the Pulumi module takes through GetClientConfig.
  scope_project_id = var.spec.scope != null ? var.spec.scope.project_id : ""
  scope_folder_id  = var.spec.scope != null ? var.spec.scope.folder_id : ""
  scope_org_id     = var.spec.scope != null ? var.spec.scope.organization_id : ""

  needs_client_project = local.scope_project_id == "" && local.scope_folder_id == "" && local.scope_org_id == ""

  # The project a project-scoped floor governs (bare id), or null for a
  # folder or organization floor.
  floor_project = (
    local.scope_project_id != "" ? trimprefix(local.scope_project_id, "projects/") :
    local.needs_client_project ? data.google_client_config.current[0].project :
    null
  )

  parent = (
    local.floor_project != null ? "projects/${local.floor_project}" :
    local.scope_folder_id != "" ? (startswith(local.scope_folder_id, "folders/") ? local.scope_folder_id : "folders/${local.scope_folder_id}") :
    "organizations/${local.scope_org_id}"
  )

  # Google manages floor settings at "global"; the spec defaults to it --
  # identical to the Pulumi module.
  location = var.spec.location != "" ? var.spec.location : "global"

  filter_config = var.spec.filter_config
}
