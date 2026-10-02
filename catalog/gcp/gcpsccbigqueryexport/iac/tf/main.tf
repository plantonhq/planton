# The provider's default project, read only for a project config whose
# manifest names no project.
data "google_client_config" "current" {
  count = local.needs_client_project ? 1 : 0
}

# A project config calls the Security Command Center API on its project; a
# folder or organization config has no project of its own to enable it on.
# disable_on_destroy is false: activation, findings, and other configs in
# the project depend on the API.
resource "google_project_service" "securitycenter_api" {
  count   = local.is_project ? 1 : 0
  project = local.project
  service = "securitycenter.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# One export, at whichever scope the spec selects.
resource "google_scc_v2_project_scc_big_query_export" "this" {
  count = local.is_project ? 1 : 0

  project             = local.project
  big_query_export_id = var.spec.big_query_export_id
  location            = local.location
  dataset             = local.dataset
  filter              = local.filter
  description         = local.description
  deletion_policy     = local.deletion_policy

  depends_on = [google_project_service.securitycenter_api]
}

resource "google_scc_v2_folder_scc_big_query_export" "this" {
  count = local.is_folder ? 1 : 0

  folder              = local.folder_id
  big_query_export_id = var.spec.big_query_export_id
  location            = local.location
  dataset             = local.dataset
  filter              = local.filter
  description         = local.description
  deletion_policy     = local.deletion_policy
}

resource "google_scc_v2_organization_scc_big_query_export" "this" {
  count = local.is_org ? 1 : 0

  name                = local.organization_export_name
  organization        = local.scope_org_id
  big_query_export_id = var.spec.big_query_export_id
  location            = local.location
  dataset             = local.dataset
  filter              = local.filter
  description         = local.description
  deletion_policy     = local.deletion_policy
}
