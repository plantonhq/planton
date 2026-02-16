# Auto-generate a numeric endpoint name when not explicitly provided.
# Vertex AI endpoints require numeric-only names (max 10 digits).
# The generated value is stable as long as the keepers don't change.
resource "random_integer" "endpoint_name" {
  count = var.spec.endpoint_name == "" ? 1 : 0
  min   = 1000000000
  max   = 9999999999
  keepers = {
    display_name = var.spec.display_name
    location     = var.spec.location
  }
}

resource "google_vertex_ai_endpoint" "this" {
  name         = tostring(local.endpoint_name)
  display_name = local.display_name
  location     = local.location
  project      = local.project_id
  labels       = local.gcp_labels

  description                = var.spec.description != "" ? var.spec.description : null
  dedicated_endpoint_enabled = var.spec.dedicated_endpoint_enabled ? true : null

  # VPC-peered private networking.
  network = var.spec.network != null ? var.spec.network.value : null

  # CMEK encryption.
  dynamic "encryption_spec" {
    for_each = var.spec.kms_key_name != null ? [var.spec.kms_key_name] : []
    content {
      kms_key_name = encryption_spec.value.value
    }
  }

  # Private Service Connect configuration.
  dynamic "private_service_connect_config" {
    for_each = var.spec.private_service_connect_config != null ? [var.spec.private_service_connect_config] : []
    content {
      enable_private_service_connect = true
      project_allowlist              = length(private_service_connect_config.value.project_allowlist) > 0 ? private_service_connect_config.value.project_allowlist : null
    }
  }
}
