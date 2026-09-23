# Enable the Vertex AI API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one dataset must never
# disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Vertex AI managed dataset. Google assigns its numeric id at creation.
# location, metadata_schema_uri, and the encryption key are immutable; the
# display name and labels update in place.
resource "google_vertex_ai_dataset" "this" {
  project = local.project_id
  # The provider names the axis `region`; the spec keeps the Vertex
  # family's single word, `location`.
  region              = var.spec.location
  display_name        = local.display_name
  metadata_schema_uri = var.spec.metadata_schema_uri
  labels              = local.final_labels

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  # CMEK: the dataset and everything imported into it encrypted under this
  # key.
  dynamic "encryption_spec" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      kms_key_name = encryption_spec.value
    }
  }

  depends_on = [google_project_service.aiplatform_api]
}
