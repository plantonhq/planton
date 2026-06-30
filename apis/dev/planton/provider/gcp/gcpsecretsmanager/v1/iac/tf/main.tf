resource "google_secret_manager_secret" "secrets" {
  for_each = { for secret in local.secrets : secret.name => secret }

  # Support both literal value and reference (reference resolution handled externally)
  project   = var.spec.project_id.value
  secret_id = each.value.secret_id
  labels    = local.final_labels

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret_versions" {
  for_each = google_secret_manager_secret.secrets

  secret     = each.value.id
  secret_data = "placeholder"

  enabled = true

  lifecycle {
    ignore_changes = [ secret_data ]
  }
}
