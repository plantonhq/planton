# Enable the Vertex AI API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one TensorBoard must
# never disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The TensorBoard instance. Google assigns its numeric id at creation;
# location and the encryption key are immutable, the display name,
# description, and labels update in place.
resource "google_vertex_ai_tensorboard" "this" {
  project = local.project_id
  # The provider names the axis `region`; the spec keeps the Vertex
  # family's single word, `location`.
  region       = var.spec.location
  display_name = local.display_name
  description  = local.description
  labels       = local.final_labels

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON,
  # fanned to every experiment and run below. Sent only when set so the
  # provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  # CMEK: the TensorBoard and every logged series encrypted under this key.
  dynamic "encryption_spec" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      kms_key_name = encryption_spec.value
    }
  }

  depends_on = [google_project_service.aiplatform_api]
}

# The folded experiments, one resource per spec.experiments[] entry.
# experiment_id and source are immutable.
resource "google_vertex_ai_tensorboard_experiment" "this" {
  for_each = local.experiments

  project                   = local.project_id
  location                  = var.spec.location
  tensorboard               = local.tensorboard_id
  tensorboard_experiment_id = each.key
  display_name              = each.value.display_name != "" ? each.value.display_name : null
  description               = each.value.description != "" ? each.value.description : null
  source                    = each.value.source != "" ? each.value.source : null
  labels                    = merge(each.value.labels, local.final_labels)
  deletion_policy           = local.deletion_policy
}

# The folded runs, one resource per experiments[].runs[] entry. Google
# requires a display name unique within the experiment; it defaults to the
# run id (the Pulumi module's rule).
resource "google_vertex_ai_tensorboard_run" "this" {
  for_each = local.runs

  project            = local.project_id
  location           = var.spec.location
  tensorboard        = local.tensorboard_id
  experiment         = google_vertex_ai_tensorboard_experiment.this[each.value.experiment_id].tensorboard_experiment_id
  tensorboard_run_id = each.value.run_id
  display_name       = each.value.display_name != "" ? each.value.display_name : each.value.run_id
  description        = each.value.description != "" ? each.value.description : null
  labels             = merge(each.value.labels, local.final_labels)
  deletion_policy    = local.deletion_policy
}
