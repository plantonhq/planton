# The provider's default project, read only when the manifest names no
# scope; an explicitly scoped floor never performs this read.
data "google_client_config" "current" {
  count = local.needs_client_project ? 1 : 0
}

# A project floor needs the Model Armor API on that project; a folder or
# organization floor has no project of its own to enable it on.
# disable_on_destroy is false: the floor outlives this block in Google, and
# templates in the project depend on the API.
resource "google_project_service" "modelarmor_api" {
  count   = local.floor_project != null ? 1 : 0
  project = local.floor_project
  service = "modelarmor.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The floor setting. Google keeps exactly one per parent: create and update
# are the same PATCH, and destroy only stops managing it -- the last applied
# floor stays in force in Google Cloud.
resource "google_model_armor_floorsetting" "this" {
  parent   = local.parent
  location = local.location

  enable_floor_setting_enforcement = var.spec.enable_floor_setting_enforcement
  integrated_services              = length(var.spec.integrated_services) > 0 ? var.spec.integrated_services : null

  filter_config {
    dynamic "malicious_uri_filter_settings" {
      for_each = local.filter_config.malicious_uri_filter_settings != null ? [local.filter_config.malicious_uri_filter_settings] : []
      content {
        filter_enforcement = malicious_uri_filter_settings.value.filter_enforcement != "" ? malicious_uri_filter_settings.value.filter_enforcement : null
      }
    }

    dynamic "pi_and_jailbreak_filter_settings" {
      for_each = local.filter_config.pi_and_jailbreak_filter_settings != null ? [local.filter_config.pi_and_jailbreak_filter_settings] : []
      content {
        filter_enforcement = pi_and_jailbreak_filter_settings.value.filter_enforcement != "" ? pi_and_jailbreak_filter_settings.value.filter_enforcement : null
        confidence_level   = pi_and_jailbreak_filter_settings.value.confidence_level != "" ? pi_and_jailbreak_filter_settings.value.confidence_level : null
      }
    }

    dynamic "rai_settings" {
      for_each = local.filter_config.rai_settings != null ? [local.filter_config.rai_settings] : []
      content {
        dynamic "rai_filters" {
          for_each = rai_settings.value.rai_filters
          content {
            filter_type      = rai_filters.value.filter_type
            confidence_level = rai_filters.value.confidence_level != "" ? rai_filters.value.confidence_level : null
          }
        }
      }
    }

    dynamic "sdp_settings" {
      for_each = local.filter_config.sdp_settings != null ? [local.filter_config.sdp_settings] : []
      content {
        dynamic "basic_config" {
          for_each = sdp_settings.value.basic_config != null ? [sdp_settings.value.basic_config] : []
          content {
            filter_enforcement = basic_config.value.filter_enforcement != "" ? basic_config.value.filter_enforcement : null
          }
        }

        dynamic "advanced_config" {
          for_each = sdp_settings.value.advanced_config != null ? [sdp_settings.value.advanced_config] : []
          content {
            inspect_template    = advanced_config.value.inspect_template != "" ? advanced_config.value.inspect_template : null
            deidentify_template = advanced_config.value.deidentify_template != "" ? advanced_config.value.deidentify_template : null
          }
        }
      }
    }
  }

  # Google wants exactly one of inspect_only / inspect_and_block; the spec
  # carries the choice as enforcement_type and only the chosen flag is sent.
  dynamic "ai_platform_floor_setting" {
    for_each = var.spec.ai_platform_floor_setting != null ? [var.spec.ai_platform_floor_setting] : []
    content {
      inspect_only         = ai_platform_floor_setting.value.enforcement_type == "INSPECT_ONLY" ? true : null
      inspect_and_block    = ai_platform_floor_setting.value.enforcement_type == "INSPECT_AND_BLOCK" ? true : null
      enable_cloud_logging = ai_platform_floor_setting.value.enable_cloud_logging
    }
  }

  dynamic "google_mcp_server_floor_setting" {
    for_each = var.spec.google_mcp_server_floor_setting != null ? [var.spec.google_mcp_server_floor_setting] : []
    content {
      inspect_only         = google_mcp_server_floor_setting.value.enforcement_type == "INSPECT_ONLY" ? true : null
      inspect_and_block    = google_mcp_server_floor_setting.value.enforcement_type == "INSPECT_AND_BLOCK" ? true : null
      enable_cloud_logging = google_mcp_server_floor_setting.value.enable_cloud_logging
    }
  }

  # The spec lifts Google's two one-leaf wrappers to a bool; the blocks are
  # sent only when it is true.
  dynamic "floor_setting_metadata" {
    for_each = var.spec.enable_multi_language_detection ? [true] : []
    content {
      multi_language_detection {
        enable_multi_language_detection = true
      }
    }
  }

  depends_on = [google_project_service.modelarmor_api]
}
