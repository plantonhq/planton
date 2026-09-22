# Enable the Vector Search API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one collection must
# never disable the API for everything else in the project.
resource "google_project_service" "vectorsearch_api" {
  project = local.project_id
  service = "vectorsearch.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Vector Search collection: a schema'd store of data objects with one
# or more vector fields. collection_id, location, and the encryption key
# are immutable; the schema, description, display name, and labels update
# in place.
resource "google_vector_search_collection" "this" {
  project       = local.project_id
  location      = var.spec.location
  collection_id = local.collection_id
  display_name  = local.display_name
  description   = local.description
  data_schema   = local.data_schema
  labels        = local.final_labels

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON,
  # fanned to every index below. Sent only when set so the provider default
  # stays in charge otherwise.
  deletion_policy = local.deletion_policy

  # CMEK: the collection and its indexes encrypted under this key.
  dynamic "encryption_spec" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      crypto_key_name = encryption_spec.value
    }
  }

  # The searchable vector fields. A dense field carries its dimensions and
  # optional Vertex embedding config; the sparse flag becomes Google's
  # empty sparse_vector block (the spec's bool is the honest form of a
  # block with no settings).
  dynamic "vector_schema" {
    for_each = var.spec.vector_schemas
    content {
      field_name = vector_schema.value.field_name

      dynamic "dense_vector" {
        for_each = vector_schema.value.dense_vector != null ? [vector_schema.value.dense_vector] : []
        content {
          dimensions = dense_vector.value.dimensions

          dynamic "vertex_embedding_config" {
            for_each = dense_vector.value.vertex_embedding_config != null ? [dense_vector.value.vertex_embedding_config] : []
            content {
              model_id      = vertex_embedding_config.value.model_id
              task_type     = vertex_embedding_config.value.task_type
              text_template = vertex_embedding_config.value.text_template
            }
          }
        }
      }

      dynamic "sparse_vector" {
        for_each = vector_schema.value.sparse_vector ? [true] : []
        content {}
      }
    }
  }

  depends_on = [google_project_service.vectorsearch_api]
}

# The folded indexes, one resource per spec.indexes[] entry. Every setting
# but labels is immutable on Google's side, so a change here replaces the
# index (rebuilt from the collection's data). Optional+Computed levers --
# distance_metric, dense_scann, dedicated infrastructure and its replica
# bounds -- are sent only when set so Google's defaults stay in charge
# (the Pulumi module's posture).
resource "google_vector_search_index" "this" {
  for_each = local.indexes

  project       = local.project_id
  location      = var.spec.location
  collection_id = google_vector_search_collection.this.collection_id
  index_id      = each.value.index_id
  index_field   = each.value.index_field
  display_name  = each.value.display_name != "" ? each.value.display_name : null
  description   = each.value.description != "" ? each.value.description : null
  labels        = merge(each.value.labels, local.final_labels)

  distance_metric = each.value.distance_metric != "" ? each.value.distance_metric : null
  filter_fields   = length(each.value.filter_fields) > 0 ? each.value.filter_fields : null
  store_fields    = length(each.value.store_fields) > 0 ? each.value.store_fields : null

  deletion_policy = local.deletion_policy

  dynamic "dense_scann" {
    for_each = each.value.feature_norm_type != "" ? [each.value.feature_norm_type] : []
    content {
      feature_norm_type = dense_scann.value
    }
  }

  dynamic "dedicated_infrastructure" {
    for_each = each.value.dedicated_infrastructure != null ? [each.value.dedicated_infrastructure] : []
    content {
      mode = dedicated_infrastructure.value.mode != "" ? dedicated_infrastructure.value.mode : null

      dynamic "autoscaling_spec" {
        for_each = dedicated_infrastructure.value.autoscaling_spec != null ? [dedicated_infrastructure.value.autoscaling_spec] : []
        content {
          min_replica_count = autoscaling_spec.value.min_replica_count
          max_replica_count = autoscaling_spec.value.max_replica_count
        }
      }
    }
  }
}
