# Enable the Certificate Manager API so a fresh project can host trust
# configs. disable_on_destroy is false: tearing down one trust config must
# never disable the API for everything else in the project.
resource "google_project_service" "certificatemanager_api" {
  project = local.project_id
  service = "certificatemanager.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# One Certificate Manager trust config: the CAs (and individually allowlisted
# certificates) a load balancer validates client or backend certificates
# against. Updates are in place -- rotating a CA never replaces the config.
resource "google_certificate_manager_trust_config" "this" {
  name        = local.trust_config_name
  project     = local.project_id
  location    = local.location
  description = var.spec.description != "" ? var.spec.description : null
  labels      = local.final_labels

  # Empty defers to the provider default (DELETE).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Each PEM is its own one-field block in the provider; the spec lifts those
  # wrappers into plain string lists.
  dynamic "trust_stores" {
    for_each = var.spec.trust_stores
    content {
      dynamic "trust_anchors" {
        for_each = trust_stores.value.trust_anchors
        content {
          pem_certificate = trust_anchors.value
        }
      }
      dynamic "intermediate_cas" {
        for_each = trust_stores.value.intermediate_cas
        content {
          pem_certificate = intermediate_cas.value
        }
      }
    }
  }

  dynamic "allowlisted_certificates" {
    for_each = var.spec.allowlisted_certificates
    content {
      pem_certificate = allowlisted_certificates.value
    }
  }

  depends_on = [google_project_service.certificatemanager_api]
}
