# Enable the Model Armor API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one template must never
# disable the API for everything else in the project.
resource "google_project_service" "modelarmor_api" {
  project = local.project_id
  service = "modelarmor.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Model Armor template. location and template_id are immutable; the
# filters, metadata, and labels update in place. Every optional string is
# sent only when set, so Google's defaults stay in charge otherwise.
resource "google_model_armor_template" "this" {
  project     = local.project_id
  location    = var.spec.location
  template_id = local.template_id
  labels      = local.final_labels

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  deletion_policy = local.deletion_policy

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

  dynamic "template_metadata" {
    for_each = local.template_metadata != null ? [local.template_metadata] : []
    content {
      log_template_operations            = template_metadata.value.log_template_operations
      log_sanitize_operations            = template_metadata.value.log_sanitize_operations
      ignore_partial_invocation_failures = template_metadata.value.ignore_partial_invocation_failures

      custom_prompt_safety_error_code          = template_metadata.value.custom_prompt_safety_error_code != 0 ? template_metadata.value.custom_prompt_safety_error_code : null
      custom_prompt_safety_error_message       = template_metadata.value.custom_prompt_safety_error_message != "" ? template_metadata.value.custom_prompt_safety_error_message : null
      custom_llm_response_safety_error_code    = template_metadata.value.custom_llm_response_safety_error_code != 0 ? template_metadata.value.custom_llm_response_safety_error_code : null
      custom_llm_response_safety_error_message = template_metadata.value.custom_llm_response_safety_error_message != "" ? template_metadata.value.custom_llm_response_safety_error_message : null
      enforcement_type                         = template_metadata.value.enforcement_type != "" ? template_metadata.value.enforcement_type : null

      # The spec lifts Google's one-leaf multi_language_detection wrapper to
      # a bool; the block is sent only when it is true.
      dynamic "multi_language_detection" {
        for_each = template_metadata.value.enable_multi_language_detection ? [true] : []
        content {
          enable_multi_language_detection = true
        }
      }

      dynamic "filter_version_selector" {
        for_each = template_metadata.value.filter_version_selector != null ? [template_metadata.value.filter_version_selector] : []
        content {
          alias   = filter_version_selector.value.alias != "" ? filter_version_selector.value.alias : null
          version = filter_version_selector.value.version != "" ? filter_version_selector.value.version : null
        }
      }
    }
  }

  depends_on = [google_project_service.modelarmor_api]
}
