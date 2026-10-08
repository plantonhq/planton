locals {
  # Scope selection, the same for the project, folder, and organization
  # resources: exactly one is created. The project arm (or no scope, the
  # provider's default project, read from the provider's own configuration
  # with google_client_config -- no API call) is a project resource; the
  # folder and organization arms take Google's bare numeric IDs. Identical
  # to the Pulumi module.
  scope_project_id = var.spec.scope != null ? var.spec.scope.project_id : ""
  scope_folder_id  = var.spec.scope != null ? var.spec.scope.folder_id : ""
  scope_org_id     = var.spec.scope != null ? var.spec.scope.organization_id : ""

  # The project the spec names (bare ID), or null -- handed to the provider
  # so an import records it. The google_client_config fallback stays out of
  # it: a provider cannot depend on its own data source.
  project_id = local.scope_project_id != "" ? trimprefix(local.scope_project_id, "projects/") : null

  is_folder  = local.scope_folder_id != ""
  is_org     = local.scope_org_id != ""
  is_project = !local.is_folder && !local.is_org

  needs_client_project = local.is_project && local.scope_project_id == ""
  project = (
    !local.is_project ? null :
    local.scope_project_id != "" ? trimprefix(local.scope_project_id, "projects/") :
    data.google_client_config.current[0].project
  )
  folder_id = trimprefix(local.scope_folder_id, "folders/")

  # Security Command Center stores configurations at "global" unless data
  # residency was set up at activation.
  location        = var.spec.location != "" ? var.spec.location : "global"
  description     = var.spec.description != "" ? var.spec.description : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  pubsub_topic = var.spec.pubsub_topic != "" ? var.spec.pubsub_topic : null

  # The publisher of whichever config was created (exactly one is), for the
  # service_account and service_account_member outputs.
  service_account = one(concat(
    google_scc_v2_project_notification_config.this[*].service_account,
    google_scc_v2_folder_notification_config.this[*].service_account,
    google_scc_v2_organization_notification_config.this[*].service_account,
  ))
}
