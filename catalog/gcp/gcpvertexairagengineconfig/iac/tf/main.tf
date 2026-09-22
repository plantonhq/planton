# Enable the Vertex AI API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down this block must never
# disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The tier of RAG Engine's managed vector database for one project and
# location. Google owns the singleton
# (projects/{p}/locations/{l}/ragEngineConfig); the provider creates and
# deletes it by PATCH, so an apply over an already-configured location
# changes the tier in place and a destroy sets the location to
# UNPROVISIONED -- which deletes the managed database's data. ABANDON is
# the teardown for anyone who keeps corpora.
resource "google_vertex_ai_rag_engine_config" "this" {
  project = local.project_id
  # The provider names the axis `region`; the spec keeps the Vertex
  # family's single word, `location`.
  region = var.spec.location

  # Client-side destroy behavior: DELETE (default; unprovisions the
  # location), PREVENT, or ABANDON. Sent only when set so the provider
  # default stays in charge otherwise.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  rag_managed_db_config {
    dynamic "basic" {
      for_each = local.tier_basic
      content {}
    }
    dynamic "scaled" {
      for_each = local.tier_scaled
      content {}
    }
    dynamic "unprovisioned" {
      for_each = local.tier_unprovisioned
      content {}
    }
  }

  depends_on = [google_project_service.aiplatform_api]
}
