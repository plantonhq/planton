# Enable the Certificate Manager API so a fresh project can host issuance
# configs. disable_on_destroy is false: tearing down one config must never
# disable the API for everything else in the project.
resource "google_project_service" "certificatemanager_api" {
  project = local.project_id
  service = "certificatemanager.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# One certificate issuance config: how Google-managed certificates that name
# it are issued from a Certificate Authority Service pool. Every argument
# except labels and deletion_policy forces replacement.
resource "google_certificate_manager_certificate_issuance_config" "this" {
  name                       = local.issuance_config_name
  project                    = local.project_id
  location                   = local.location
  description                = var.spec.description != "" ? var.spec.description : null
  labels                     = local.final_labels
  key_algorithm              = var.spec.key_algorithm
  lifetime                   = var.spec.lifetime
  rotation_window_percentage = var.spec.rotation_window_percentage

  # Empty defers to the provider default (DELETE).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The spec lifts the provider's two single-field wrappers into ca_pool.
  certificate_authority_config {
    certificate_authority_service_config {
      ca_pool = var.spec.ca_pool
    }
  }

  depends_on = [google_project_service.certificatemanager_api]
}
