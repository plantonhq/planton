# The provider's default project, read only when the manifest names none.
data "google_client_config" "current" {
  count = local.needs_client_project ? 1 : 0
}

# Same-project key storage creates the key in the handle's own project
# through the Cloud KMS API, so the API must be on there even when Autokey
# was switched on at the folder. disable_on_destroy is false: the key
# outlives the handle and keeps encrypting through the API.
resource "google_project_service" "cloudkms_api" {
  project = local.project
  service = "cloudkms.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The key handle. Autokey (on for the project or a folder above it)
# creates or reuses a key for the resource type and location and returns it
# as kms_key. Every argument is immutable; destroy only removes the handle
# from state -- Google keeps it, and the key keeps protecting its resources.
resource "google_kms_key_handle" "this" {
  project                = local.project
  location               = var.spec.location
  name                   = local.key_handle_name
  resource_type_selector = var.spec.resource_type_selector

  depends_on = [google_project_service.cloudkms_api]
}
