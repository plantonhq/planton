# Enable the Discovery Engine API first so a fresh project works on the
# first deploy. disable_on_destroy is false: tearing down one store must
# never disable the API for everything else in the project.
resource "google_project_service" "discoveryengine_api" {
  project = local.project_id
  service = "discoveryengine.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The data store: structured records, unstructured documents, or a public
# website, enrolled in one or more solutions. Everything but display_name
# and kms_key_name is immutable, so a change replaces the store (documents
# included). Optional strings and blocks are sent only when set so Google's
# defaults stay in charge (the Pulumi module's posture); the two virtual
# inputs are sent as declared and never read back.
resource "google_discovery_engine_data_store" "this" {
  project           = local.project_id
  location          = var.spec.location
  data_store_id     = local.data_store_id
  display_name      = local.display_name
  industry_vertical = var.spec.industry_vertical
  content_config    = local.content_config
  solution_types    = local.solution_types
  acl_enabled       = var.spec.acl_enabled
  kms_key_name      = local.kms_key_name

  create_advanced_site_search  = var.spec.create_advanced_site_search
  skip_default_schema_creation = var.spec.skip_default_schema_creation

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON,
  # fanned to the schema, target sites, and sitemaps below. Sent only when
  # set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  dynamic "advanced_site_search_config" {
    for_each = var.spec.advanced_site_search_config != null ? [var.spec.advanced_site_search_config] : []
    content {
      disable_initial_index     = advanced_site_search_config.value.disable_initial_index
      disable_automatic_refresh = advanced_site_search_config.value.disable_automatic_refresh
    }
  }

  # How documents are parsed and chunked. The spec's digital_parsing bool
  # becomes Google's empty digital_parsing_config block (the marker-block
  # idiom); the layout and OCR arms carry their settings. The default
  # parser and each per-file-type override share one shape.
  dynamic "document_processing_config" {
    for_each = var.spec.document_processing_config != null ? [var.spec.document_processing_config] : []
    content {
      dynamic "chunking_config" {
        for_each = document_processing_config.value.chunking_config != null ? [document_processing_config.value.chunking_config] : []
        content {
          layout_based_chunking_config {
            chunk_size                = chunking_config.value.chunk_size
            include_ancestor_headings = chunking_config.value.include_ancestor_headings
          }
        }
      }

      dynamic "default_parsing_config" {
        for_each = document_processing_config.value.default_parsing_config != null ? [document_processing_config.value.default_parsing_config] : []
        content {
          dynamic "digital_parsing_config" {
            for_each = default_parsing_config.value.digital_parsing ? [true] : []
            content {}
          }
          dynamic "layout_parsing_config" {
            for_each = default_parsing_config.value.layout_parsing_config != null ? [default_parsing_config.value.layout_parsing_config] : []
            content {
              enable_get_processed_document = layout_parsing_config.value.enable_get_processed_document
              enable_image_annotation       = layout_parsing_config.value.enable_image_annotation
              enable_llm_layout_parsing     = layout_parsing_config.value.enable_llm_layout_parsing
              enable_table_annotation       = layout_parsing_config.value.enable_table_annotation
              exclude_html_classes          = length(layout_parsing_config.value.exclude_html_classes) > 0 ? layout_parsing_config.value.exclude_html_classes : null
              exclude_html_elements         = length(layout_parsing_config.value.exclude_html_elements) > 0 ? layout_parsing_config.value.exclude_html_elements : null
              exclude_html_ids              = length(layout_parsing_config.value.exclude_html_ids) > 0 ? layout_parsing_config.value.exclude_html_ids : null
              structured_content_types      = length(layout_parsing_config.value.structured_content_types) > 0 ? layout_parsing_config.value.structured_content_types : null
            }
          }
          dynamic "ocr_parsing_config" {
            for_each = default_parsing_config.value.ocr_parsing_config != null ? [default_parsing_config.value.ocr_parsing_config] : []
            content {
              use_native_text = ocr_parsing_config.value.use_native_text
            }
          }
        }
      }

      dynamic "parsing_config_overrides" {
        for_each = document_processing_config.value.parsing_config_overrides
        content {
          file_type = parsing_config_overrides.value.file_type

          dynamic "digital_parsing_config" {
            for_each = parsing_config_overrides.value.parsing_config.digital_parsing ? [true] : []
            content {}
          }
          dynamic "layout_parsing_config" {
            for_each = parsing_config_overrides.value.parsing_config.layout_parsing_config != null ? [parsing_config_overrides.value.parsing_config.layout_parsing_config] : []
            content {
              enable_get_processed_document = layout_parsing_config.value.enable_get_processed_document
              enable_image_annotation       = layout_parsing_config.value.enable_image_annotation
              enable_llm_layout_parsing     = layout_parsing_config.value.enable_llm_layout_parsing
              enable_table_annotation       = layout_parsing_config.value.enable_table_annotation
              exclude_html_classes          = length(layout_parsing_config.value.exclude_html_classes) > 0 ? layout_parsing_config.value.exclude_html_classes : null
              exclude_html_elements         = length(layout_parsing_config.value.exclude_html_elements) > 0 ? layout_parsing_config.value.exclude_html_elements : null
              exclude_html_ids              = length(layout_parsing_config.value.exclude_html_ids) > 0 ? layout_parsing_config.value.exclude_html_ids : null
              structured_content_types      = length(layout_parsing_config.value.structured_content_types) > 0 ? layout_parsing_config.value.structured_content_types : null
            }
          }
          dynamic "ocr_parsing_config" {
            for_each = parsing_config_overrides.value.parsing_config.ocr_parsing_config != null ? [parsing_config_overrides.value.parsing_config.ocr_parsing_config] : []
            content {
              use_native_text = ocr_parsing_config.value.use_native_text
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_service.discoveryengine_api]
}

# The store's one custom schema, when declared (the spec requires the
# default schema skipped first: Google keeps exactly one per store). Fully
# immutable; a change replaces it.
resource "google_discovery_engine_schema" "this" {
  count = var.spec.schema != null ? 1 : 0

  project       = local.project_id
  location      = var.spec.location
  data_store_id = google_discovery_engine_data_store.this.data_store_id
  schema_id     = var.spec.schema.schema_id
  json_schema   = var.spec.schema.json_schema

  deletion_policy = local.deletion_policy
}

# The URL patterns a PUBLIC_WEBSITE store crawls or excludes, one resource
# each, keyed by pattern. Every field is immutable; a change replaces that
# target site. Empty type lets Google default to INCLUDE.
resource "google_discovery_engine_target_site" "this" {
  for_each = local.target_sites

  project              = local.project_id
  location             = var.spec.location
  data_store_id        = google_discovery_engine_data_store.this.data_store_id
  provided_uri_pattern = each.value.provided_uri_pattern
  type                 = each.value.type != "" ? each.value.type : null
  exact_match          = each.value.exact_match

  deletion_policy = local.deletion_policy
}

# The sitemaps an advanced site search store reads, one resource each,
# keyed by URI. Fully immutable.
resource "google_discovery_engine_sitemap" "this" {
  for_each = local.sitemaps

  project       = local.project_id
  location      = var.spec.location
  data_store_id = google_discovery_engine_data_store.this.data_store_id
  uri           = each.value

  deletion_policy = local.deletion_policy
}
