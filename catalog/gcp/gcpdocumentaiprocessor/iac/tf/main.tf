# Enable the Document AI API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one processor must never
# disable the API for everything else in the project.
resource "google_project_service" "documentai_api" {
  project = local.project_id
  service = "documentai.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The processor. Every argument is immutable: a change replaces the
# processor, and its id and endpoint change with it.
resource "google_document_ai_processor" "this" {
  project      = local.project_id
  location     = var.spec.location
  type         = var.spec.type
  display_name = local.display_name
  kms_key_name = local.kms_key_name

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.documentai_api]
}

# The processor version that serves requests which name none. The spec
# carries the short version id and the full path is composed under this
# processor -- the Pulumi module composes it the same way. A change
# re-creates this binding, which sets the new default; destroy is a no-op in
# Google (there is no "unset"), so the last default stays.
resource "google_document_ai_processor_default_version" "this" {
  count     = var.spec.default_version != "" ? 1 : 0
  processor = google_document_ai_processor.this.id
  version   = "${google_document_ai_processor.this.id}/processorVersions/${var.spec.default_version}"
}
