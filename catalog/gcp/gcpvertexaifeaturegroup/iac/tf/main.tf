# Enable the Vertex AI API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one feature group must
# never disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The feature group: the registry entry pointing Feature Store at a
# BigQuery table or view. The id, location, and source table are immutable;
# the description, labels, and entity ID columns update in place.
resource "google_vertex_ai_feature_group" "this" {
  project = local.project_id
  # The provider names the axis `region`; the spec keeps the Vertex
  # family's single word, `location`.
  region      = var.spec.location
  name        = var.spec.feature_group_id
  description = local.description
  labels      = local.final_labels

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON,
  # fanned to every feature below. Sent only when set so the provider
  # default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  # The spec lifts the one-leaf big_query_source wrapper; the module
  # restores Google's nesting.
  dynamic "big_query" {
    for_each = var.spec.big_query != null ? [var.spec.big_query] : []
    content {
      entity_id_columns = length(big_query.value.entity_id_columns) > 0 ? big_query.value.entity_id_columns : null

      big_query_source {
        input_uri = local.input_uri
      }
    }
  }

  depends_on = [google_project_service.aiplatform_api]
}

# The folded features, one resource per spec.features[] entry. The feature
# id is immutable; version_column_name is Optional+Computed and sent only
# when set so Google's default (the column named like the feature) stays
# in charge (the Pulumi module's posture).
resource "google_vertex_ai_feature_group_feature" "this" {
  for_each = local.features

  project             = local.project_id
  region              = var.spec.location
  feature_group       = google_vertex_ai_feature_group.this.name
  name                = each.key
  description         = each.value.description != "" ? each.value.description : null
  labels              = merge(each.value.labels, local.final_labels)
  version_column_name = each.value.version_column_name != "" ? each.value.version_column_name : null
  deletion_policy     = local.deletion_policy
}
