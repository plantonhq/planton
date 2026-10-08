locals {
  # Scope selection. A folder arm configures the folder
  # (google_kms_autokey_config); otherwise the project arm, or the
  # provider's default project read from the provider's own configuration
  # (google_client_config, no API call), configures a project
  # (google_kms_project_autokey_config) -- identical to the Pulumi module.
  scope_project_id = var.spec.scope != null ? var.spec.scope.project_id : ""
  scope_folder_id  = var.spec.scope != null ? var.spec.scope.folder_id : ""

  # The project the spec names (bare ID), or null -- handed to the provider
  # so an import records it. The google_client_config fallback stays out of
  # it: a provider cannot depend on its own data source.
  project_id = local.scope_project_id != "" ? trimprefix(local.scope_project_id, "projects/") : null

  is_folder            = local.scope_folder_id != ""
  needs_client_project = !local.is_folder && local.scope_project_id == ""

  # Google's folder argument is the bare numeric ID.
  folder_id = trimprefix(local.scope_folder_id, "folders/")

  # The project a project-scoped configuration governs (bare ID), or null
  # for a folder configuration.
  config_project = (
    local.is_folder ? null :
    local.scope_project_id != "" ? trimprefix(local.scope_project_id, "projects/") :
    data.google_client_config.current[0].project
  )

  # The dedicated key project, bare; Google takes it as projects/{id}.
  key_project = var.spec.key_project != "" ? trimprefix(var.spec.key_project, "projects/") : null

  # Autokey creates keys through the Cloud KMS API, so it must be on in the
  # project a project configuration governs and in a folder's key project.
  api_projects = toset(compact([local.config_project, local.key_project]))

  key_project_resolution_mode = var.spec.key_project_resolution_mode != "" ? var.spec.key_project_resolution_mode : null
  deletion_policy             = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
