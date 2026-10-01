# The provider's default project, read only when the manifest names no
# scope; a scoped configuration never performs this read.
data "google_client_config" "current" {
  count = local.needs_client_project ? 1 : 0
}

# Autokey creates keys through the Cloud KMS API. disable_on_destroy is
# false: the keys Autokey created outlive this block and keep encrypting
# through the API.
resource "google_project_service" "cloudkms_api" {
  for_each = local.api_projects
  project  = each.value
  service  = "cloudkms.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# Exactly one configuration exists per folder and per project: create and
# update are the same PATCH (applying takes over an existing
# configuration), and destroy under DELETE clears it.
resource "google_kms_autokey_config" "this" {
  count = local.is_folder ? 1 : 0

  folder                      = local.folder_id
  key_project                 = local.key_project != null ? "projects/${local.key_project}" : null
  key_project_resolution_mode = local.key_project_resolution_mode
  deletion_policy             = local.deletion_policy

  depends_on = [google_project_service.cloudkms_api]
}

resource "google_kms_project_autokey_config" "this" {
  count = local.is_folder ? 0 : 1

  project                     = local.config_project
  key_project_resolution_mode = local.key_project_resolution_mode
  deletion_policy             = local.deletion_policy

  depends_on = [google_project_service.cloudkms_api]
}
