# Enable the Vertex AI API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one online store must
# never disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The online store: exactly one of a managed Bigtable instance or Optimized
# serving (the spec's CEL). The id, location, and storage kind are
# immutable; Bigtable scaling, the dedicated endpoint, the key, and labels
# update in place.
resource "google_vertex_ai_feature_online_store" "this" {
  project = local.project_id
  # The provider names the axis `region`; the spec keeps the Vertex
  # family's single word, `location`.
  region        = var.spec.location
  name          = var.spec.feature_online_store_id
  labels        = local.final_labels
  force_destroy = var.spec.force_destroy

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON,
  # fanned to every feature view below. Sent only when set so the provider
  # default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  dynamic "bigtable" {
    for_each = var.spec.bigtable != null ? [var.spec.bigtable] : []
    content {
      enable_direct_bigtable_access = bigtable.value.enable_direct_bigtable_access ? true : null
      # Optional+Computed: Google picks a zone in the region when unset.
      zone = bigtable.value.zone != "" ? bigtable.value.zone : null

      auto_scaling {
        min_node_count = bigtable.value.auto_scaling.min_node_count
        max_node_count = bigtable.value.auto_scaling.max_node_count
        # Optional+Computed: Google defaults to 50 percent when unset.
        cpu_utilization_target = bigtable.value.auto_scaling.cpu_utilization_target
      }
    }
  }

  # The spec's bool is the honest form of Google's empty optimized block.
  dynamic "optimized" {
    for_each = var.spec.optimized ? [true] : []
    content {}
  }

  # Optional+Computed: Optimized stores get a public dedicated endpoint by
  # default; the block is sent only when the spec shapes it.
  dynamic "dedicated_serving_endpoint" {
    for_each = var.spec.dedicated_serving_endpoint != null ? [var.spec.dedicated_serving_endpoint] : []
    content {
      dynamic "private_service_connect_config" {
        for_each = dedicated_serving_endpoint.value.private_service_connect_config != null ? [dedicated_serving_endpoint.value.private_service_connect_config] : []
        content {
          enable_private_service_connect = private_service_connect_config.value.enable_private_service_connect
          project_allowlist              = length(private_service_connect_config.value.project_allowlist) > 0 ? private_service_connect_config.value.project_allowlist : null
        }
      }
    }
  }

  dynamic "encryption_spec" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      kms_key_name = encryption_spec.value
    }
  }

  depends_on = [google_project_service.aiplatform_api]
}

# The folded feature views, one resource per spec.feature_views[] entry.
# Each has exactly one source (the spec's CEL). A GcpBigQueryTable
# reference resolves to project.dataset.table; Google stores the bq://
# form, so the prefix is added when missing (the Pulumi module's rule).
resource "google_vertex_ai_feature_online_store_featureview" "this" {
  for_each = local.feature_views

  project              = local.project_id
  region               = var.spec.location
  feature_online_store = google_vertex_ai_feature_online_store.this.name
  name                 = each.key
  labels               = merge(each.value.labels, local.final_labels)
  deletion_policy      = local.deletion_policy

  dynamic "big_query_source" {
    for_each = each.value.big_query_source != null ? [each.value.big_query_source] : []
    content {
      uri               = startswith(big_query_source.value.uri, "bq://") ? big_query_source.value.uri : "bq://${big_query_source.value.uri}"
      entity_id_columns = big_query_source.value.entity_id_columns
    }
  }

  dynamic "feature_registry_source" {
    for_each = each.value.feature_registry_source != null ? [each.value.feature_registry_source] : []
    content {
      project_number = feature_registry_source.value.project_number != "" ? feature_registry_source.value.project_number : null

      dynamic "feature_groups" {
        for_each = feature_registry_source.value.feature_groups
        content {
          feature_group_id = feature_groups.value.feature_group_id
          feature_ids      = feature_groups.value.feature_ids
        }
      }
    }
  }

  # cron is Optional+Computed and sent only when set; continuous is sent as
  # declared inside a declared block.
  dynamic "sync_config" {
    for_each = each.value.sync_config != null ? [each.value.sync_config] : []
    content {
      cron       = sync_config.value.cron != "" ? sync_config.value.cron : null
      continuous = sync_config.value.continuous ? true : null
    }
  }
}
