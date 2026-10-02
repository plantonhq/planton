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

# One rule, at whichever scope the spec selects.
resource "google_scc_v2_project_mute_config" "this" {
  count = local.is_project ? 1 : 0

  project         = local.project
  mute_config_id  = var.spec.mute_config_id
  location        = local.location
  filter          = var.spec.filter
  type            = var.spec.type
  description     = local.description
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.securitycenter_api]
}

resource "google_scc_v2_folder_mute_config" "this" {
  count = local.is_folder ? 1 : 0

  folder          = local.folder_id
  mute_config_id  = var.spec.mute_config_id
  location        = local.location
  filter          = var.spec.filter
  type            = var.spec.type
  description     = local.description
  deletion_policy = local.deletion_policy
}

resource "google_scc_v2_organization_mute_config" "this" {
  count = local.is_org ? 1 : 0

  organization    = local.scope_org_id
  mute_config_id  = var.spec.mute_config_id
  location        = local.location
  filter          = var.spec.filter
  type            = var.spec.type
  description     = local.description
  deletion_policy = local.deletion_policy
}
