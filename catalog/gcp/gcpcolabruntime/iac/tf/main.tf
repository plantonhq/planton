# Enable the Vertex AI API (Colab Enterprise's API) first so a fresh project
# works on the first deploy. disable_on_destroy is false: tearing down one
# runtime must never disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The runtime, assigned to runtime_user from the template. desired_state
# (RUNNING / STOPPED) and auto_upgrade are client-side controls the provider
# enforces with start, stop, and upgrade calls on every apply; everything
# else is fixed at creation.
resource "google_colab_runtime" "this" {
  project      = local.project_id
  location     = var.spec.location
  name         = local.runtime_id
  display_name = local.display_name
  description  = local.description
  runtime_user = var.spec.runtime_user

  notebook_runtime_template_ref {
    notebook_runtime_template = var.spec.runtime_template
  }

  desired_state = local.desired_state
  auto_upgrade  = var.spec.auto_upgrade

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.aiplatform_api]
}
