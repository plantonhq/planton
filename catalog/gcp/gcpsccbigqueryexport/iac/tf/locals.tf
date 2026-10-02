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

  # Google takes projects/{project}/datasets/{dataset}; a GcpBigQueryDataset
  # reference resolves to the dataset's self link, so the API prefix is
  # trimmed.
  dataset = trimprefix(var.spec.dataset, "https://bigquery.googleapis.com/bigquery/v2/")
  filter  = var.spec.filter != "" ? var.spec.filter : null

  # The organization resource carries name as an argument (Optional, not
  # Computed); the module always composes Google's own value so a plan
  # never shows it drifting.
  organization_export_name = "organizations/${local.scope_org_id}/locations/${local.location}/bigQueryExports/${var.spec.big_query_export_id}"
}
