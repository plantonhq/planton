# Enable the Discovery Engine API first so a fresh project works on the
# first deploy. disable_on_destroy is false: tearing down one engine must
# never disable the API for everything else in the project.
resource "google_project_service" "discoveryengine_api" {
  project = local.project_id
  service = "discoveryengine.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A chat engine creates (or links) a Dialogflow CX agent, which needs the
# Dialogflow API on -- enabled by the module on the chat arm so a fresh
# project never has to remember it.
resource "google_project_service" "dialogflow_api" {
  count = local.is_chat ? 1 : 0

  project = local.project_id
  service = "dialogflow.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# ---------------------------------------------------------------------------
# The engine: exactly one of three resources, chosen by engine_type
# ---------------------------------------------------------------------------

# SEARCH: search_engine_config is ALWAYS sent because Google requires the
# block (empty when the spec leaves it out, so the Standard tier applies);
# app_type, features, kms_key_name, and knowledge_graph_config are sent
# only when set -- features and the knowledge graph are Optional+Computed
# on Google's side and must not be defaulted by the module (the Pulumi
# module's posture).
resource "google_discovery_engine_search_engine" "this" {
  count = local.is_search ? 1 : 0

  project           = local.project_id
  location          = var.spec.location
  collection_id     = local.collection_id
  engine_id         = local.engine_id
  display_name      = local.display_name
  data_store_ids    = var.spec.data_store_ids
  industry_vertical = local.industry_vertical
  app_type          = local.app_type
  disable_analytics = var.spec.disable_analytics
  features          = local.features
  kms_key_name      = local.kms_key_name

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON,
  # fanned to the controls and assistants. Sent only when set so the
  # provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  search_engine_config {
    search_tier                = var.spec.search_engine_config != null && var.spec.search_engine_config.search_tier != "" ? var.spec.search_engine_config.search_tier : null
    search_add_ons             = var.spec.search_engine_config != null && length(var.spec.search_engine_config.search_add_ons) > 0 ? var.spec.search_engine_config.search_add_ons : null
    required_subscription_tier = var.spec.search_engine_config != null && var.spec.search_engine_config.required_subscription_tier != "" ? var.spec.search_engine_config.required_subscription_tier : null
  }

  dynamic "common_config" {
    for_each = local.company_name != null ? [local.company_name] : []
    content {
      company_name = common_config.value
    }
  }

  dynamic "knowledge_graph_config" {
    for_each = var.spec.knowledge_graph_config != null ? [var.spec.knowledge_graph_config] : []
    content {
      enable_cloud_knowledge_graph   = knowledge_graph_config.value.enable_cloud_knowledge_graph
      enable_private_knowledge_graph = knowledge_graph_config.value.enable_private_knowledge_graph
      cloud_knowledge_graph_types    = length(knowledge_graph_config.value.cloud_knowledge_graph_types) > 0 ? knowledge_graph_config.value.cloud_knowledge_graph_types : null

      dynamic "feature_config" {
        for_each = knowledge_graph_config.value.feature_config != null ? [knowledge_graph_config.value.feature_config] : []
        content {
          disable_private_kg_auto_complete       = feature_config.value.disable_private_kg_auto_complete
          disable_private_kg_enrichment          = feature_config.value.disable_private_kg_enrichment
          disable_private_kg_query_ui_chips      = feature_config.value.disable_private_kg_query_ui_chips
          disable_private_kg_query_understanding = feature_config.value.disable_private_kg_query_understanding
        }
      }
    }
  }

  depends_on = [google_project_service.discoveryengine_api]
}

# CHAT: a chat app that creates or links a Dialogflow CX agent (exactly
# one, a spec rule). The whole chat_engine_config is immutable.
resource "google_discovery_engine_chat_engine" "this" {
  count = local.is_chat ? 1 : 0

  project           = local.project_id
  location          = var.spec.location
  collection_id     = local.collection_id
  engine_id         = local.engine_id
  display_name      = local.display_name
  data_store_ids    = var.spec.data_store_ids
  industry_vertical = local.industry_vertical

  deletion_policy = local.deletion_policy

  chat_engine_config {
    allow_cross_region       = var.spec.chat_engine_config.allow_cross_region
    dialogflow_agent_to_link = var.spec.chat_engine_config.dialogflow_agent_to_link != "" ? var.spec.chat_engine_config.dialogflow_agent_to_link : null

    dynamic "agent_creation_config" {
      for_each = var.spec.chat_engine_config.agent_creation_config != null ? [var.spec.chat_engine_config.agent_creation_config] : []
      content {
        business              = agent_creation_config.value.business != "" ? agent_creation_config.value.business : null
        default_language_code = agent_creation_config.value.default_language_code
        time_zone             = agent_creation_config.value.time_zone
        location              = agent_creation_config.value.location != "" ? agent_creation_config.value.location : null
      }
    }
  }

  dynamic "common_config" {
    for_each = local.company_name != null ? [local.company_name] : []
    content {
      company_name = common_config.value
    }
  }

  depends_on = [google_project_service.discoveryengine_api, google_project_service.dialogflow_api]
}

# RECOMMENDATION: the resource has no collection_id (always
# default_collection) and no kms_key_name; the spec walls both. The media
# config's levers are sent only when set so Google's per-type defaults
# apply.
resource "google_discovery_engine_recommendation_engine" "this" {
  count = local.is_recommendation ? 1 : 0

  project           = local.project_id
  location          = var.spec.location
  engine_id         = local.engine_id
  display_name      = local.display_name
  data_store_ids    = var.spec.data_store_ids
  industry_vertical = local.industry_vertical

  deletion_policy = local.deletion_policy

  dynamic "common_config" {
    for_each = local.company_name != null ? [local.company_name] : []
    content {
      company_name = common_config.value
    }
  }

  dynamic "media_recommendation_engine_config" {
    for_each = var.spec.media_recommendation_engine_config != null ? [var.spec.media_recommendation_engine_config] : []
    content {
      type                   = media_recommendation_engine_config.value.type != "" ? media_recommendation_engine_config.value.type : null
      optimization_objective = media_recommendation_engine_config.value.optimization_objective != "" ? media_recommendation_engine_config.value.optimization_objective : null
      training_state         = media_recommendation_engine_config.value.training_state != "" ? media_recommendation_engine_config.value.training_state : null

      dynamic "optimization_objective_config" {
        for_each = media_recommendation_engine_config.value.optimization_objective_config != null ? [media_recommendation_engine_config.value.optimization_objective_config] : []
        content {
          target_field             = optimization_objective_config.value.target_field
          target_field_value_float = optimization_objective_config.value.target_field_value_float
        }
      }

      dynamic "engine_features_config" {
        for_each = media_recommendation_engine_config.value.engine_features_config != null ? [media_recommendation_engine_config.value.engine_features_config] : []
        content {
          dynamic "most_popular_config" {
            for_each = engine_features_config.value.most_popular_config != null ? [engine_features_config.value.most_popular_config] : []
            content {
              time_window_days = most_popular_config.value.time_window_days
            }
          }
          dynamic "recommended_for_you_config" {
            for_each = engine_features_config.value.recommended_for_you_config != null ? [engine_features_config.value.recommended_for_you_config] : []
            content {
              context_event_type = recommended_for_you_config.value.context_event_type != "" ? recommended_for_you_config.value.context_event_type : null
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_service.discoveryengine_api]
}

# ---------------------------------------------------------------------------
# Folded companions: controls, the serving config, the widget config,
# assistants -- each addressed by the created engine's id, whichever arm
# built it
# ---------------------------------------------------------------------------

# One control per spec.controls[] entry, keyed by control_id; exactly one
# action each (a spec rule). solution_type is derived from the arm.
resource "google_discovery_engine_control" "this" {
  for_each = local.controls

  project       = local.project_id
  location      = var.spec.location
  collection_id = local.collection_id
  engine_id     = local.created_engine_id
  control_id    = each.value.control_id
  display_name  = each.value.display_name
  solution_type = local.solution_type
  use_cases     = length(each.value.use_cases) > 0 ? each.value.use_cases : null

  deletion_policy = local.deletion_policy

  dynamic "conditions" {
    for_each = each.value.conditions
    content {
      query_regex = conditions.value.query_regex != "" ? conditions.value.query_regex : null

      dynamic "query_terms" {
        for_each = conditions.value.query_terms
        content {
          value      = query_terms.value.value
          full_match = query_terms.value.full_match
        }
      }

      dynamic "active_time_range" {
        for_each = conditions.value.active_time_ranges
        content {
          start_time = active_time_range.value.start_time != "" ? active_time_range.value.start_time : null
          end_time   = active_time_range.value.end_time != "" ? active_time_range.value.end_time : null
        }
      }
    }
  }

  dynamic "boost_action" {
    for_each = each.value.boost_action != null ? [each.value.boost_action] : []
    content {
      data_store  = boost_action.value.data_store
      filter      = boost_action.value.filter
      fixed_boost = boost_action.value.fixed_boost

      dynamic "interpolation_boost_spec" {
        for_each = boost_action.value.interpolation_boost_spec != null ? [boost_action.value.interpolation_boost_spec] : []
        content {
          field_name         = interpolation_boost_spec.value.field_name != "" ? interpolation_boost_spec.value.field_name : null
          attribute_type     = interpolation_boost_spec.value.attribute_type != "" ? interpolation_boost_spec.value.attribute_type : null
          interpolation_type = interpolation_boost_spec.value.interpolation_type != "" ? interpolation_boost_spec.value.interpolation_type : null

          dynamic "control_point" {
            for_each = interpolation_boost_spec.value.control_point != null ? [interpolation_boost_spec.value.control_point] : []
            content {
              attribute_value = control_point.value.attribute_value != "" ? control_point.value.attribute_value : null
              boost_amount    = control_point.value.boost_amount
            }
          }
        }
      }
    }
  }

  dynamic "filter_action" {
    for_each = each.value.filter_action != null ? [each.value.filter_action] : []
    content {
      data_store = filter_action.value.data_store
      filter     = filter_action.value.filter
    }
  }

  dynamic "promote_action" {
    for_each = each.value.promote_action != null ? [each.value.promote_action] : []
    content {
      data_store = promote_action.value.data_store

      search_link_promotion {
        title       = promote_action.value.search_link_promotion.title
        uri         = promote_action.value.search_link_promotion.uri != "" ? promote_action.value.search_link_promotion.uri : null
        document    = promote_action.value.search_link_promotion.document != "" ? promote_action.value.search_link_promotion.document : null
        description = promote_action.value.search_link_promotion.description != "" ? promote_action.value.search_link_promotion.description : null
        image_uri   = promote_action.value.search_link_promotion.image_uri != "" ? promote_action.value.search_link_promotion.image_uri : null
        enabled     = promote_action.value.search_link_promotion.enabled
      }
    }
  }

  dynamic "redirect_action" {
    for_each = each.value.redirect_action != null ? [each.value.redirect_action] : []
    content {
      redirect_uri = redirect_action.value.redirect_uri
    }
  }

  dynamic "synonyms_action" {
    for_each = each.value.synonyms_action != null ? [each.value.synonyms_action] : []
    content {
      synonyms = synonyms_action.value.synonyms
    }
  }
}

# The engine's default serving config ("default_search", which Google
# creates with the engine): the provider's create is a PATCH and its
# delete a no-op, so this block configures rather than owns. It depends on
# the controls so every id it lists exists when the PATCH lands.
resource "google_discovery_engine_serving_config" "this" {
  count = var.spec.serving_config != null ? 1 : 0

  project       = local.project_id
  location      = var.spec.location
  collection_id = local.collection_id
  engine_id     = local.created_engine_id

  boost_control_ids    = length(var.spec.serving_config.boost_control_ids) > 0 ? var.spec.serving_config.boost_control_ids : null
  filter_control_ids   = length(var.spec.serving_config.filter_control_ids) > 0 ? var.spec.serving_config.filter_control_ids : null
  promote_control_ids  = length(var.spec.serving_config.promote_control_ids) > 0 ? var.spec.serving_config.promote_control_ids : null
  redirect_control_ids = length(var.spec.serving_config.redirect_control_ids) > 0 ? var.spec.serving_config.redirect_control_ids : null
  synonyms_control_ids = length(var.spec.serving_config.synonyms_control_ids) > 0 ? var.spec.serving_config.synonyms_control_ids : null

  depends_on = [google_discovery_engine_control.this]
}

# The engine's search widget ("default_search_widget_config", which Google
# creates with a search engine): a PATCH on create; a widget config cannot
# be deleted, so destroy leaves it as configured. Optional levers are sent
# only when set.
resource "google_discovery_engine_widget_config" "this" {
  count = var.spec.widget_config != null ? 1 : 0

  project       = local.project_id
  location      = var.spec.location
  collection_id = local.collection_id
  engine_id     = local.created_engine_id

  dynamic "access_settings" {
    for_each = var.spec.widget_config.access_settings != null ? [var.spec.widget_config.access_settings] : []
    content {
      allow_public_access              = access_settings.value.allow_public_access
      allowlisted_domains              = length(access_settings.value.allowlisted_domains) > 0 ? access_settings.value.allowlisted_domains : null
      enable_web_app                   = access_settings.value.enable_web_app
      language_code                    = access_settings.value.language_code != "" ? access_settings.value.language_code : null
      workforce_identity_pool_provider = access_settings.value.workforce_identity_pool_provider != "" ? access_settings.value.workforce_identity_pool_provider : null
    }
  }

  dynamic "homepage_setting" {
    for_each = var.spec.widget_config.homepage_setting != null ? [var.spec.widget_config.homepage_setting] : []
    content {
      dynamic "shortcuts" {
        for_each = homepage_setting.value.shortcuts
        content {
          title           = shortcuts.value.title != "" ? shortcuts.value.title : null
          destination_uri = shortcuts.value.destination_uri != "" ? shortcuts.value.destination_uri : null

          dynamic "icon" {
            for_each = shortcuts.value.icon_url != "" ? [shortcuts.value.icon_url] : []
            content {
              url = icon.value
            }
          }
        }
      }
    }
  }

  dynamic "ui_branding" {
    for_each = var.spec.widget_config.ui_branding != null ? [var.spec.widget_config.ui_branding] : []
    content {
      dynamic "logo" {
        for_each = ui_branding.value.logo_url != "" ? [ui_branding.value.logo_url] : []
        content {
          url = logo.value
        }
      }
    }
  }

  dynamic "ui_settings" {
    for_each = var.spec.widget_config.ui_settings != null ? [var.spec.widget_config.ui_settings] : []
    content {
      interaction_type                = ui_settings.value.interaction_type != "" ? ui_settings.value.interaction_type : null
      result_description_type         = ui_settings.value.result_description_type != "" ? ui_settings.value.result_description_type : null
      default_search_request_order_by = ui_settings.value.default_search_request_order_by != "" ? ui_settings.value.default_search_request_order_by : null
      disable_user_events_collection  = ui_settings.value.disable_user_events_collection
      enable_autocomplete             = ui_settings.value.enable_autocomplete
      enable_create_agent_button      = ui_settings.value.enable_create_agent_button
      enable_people_search            = ui_settings.value.enable_people_search
      enable_quality_feedback         = ui_settings.value.enable_quality_feedback
      enable_safe_search              = ui_settings.value.enable_safe_search
      enable_search_as_you_type       = ui_settings.value.enable_search_as_you_type
      enable_visual_content_summary   = ui_settings.value.enable_visual_content_summary

      dynamic "data_store_ui_configs" {
        for_each = ui_settings.value.data_store_ui_configs
        content {
          name = data_store_ui_configs.value.name

          dynamic "facet_field" {
            for_each = data_store_ui_configs.value.facet_fields
            content {
              field        = facet_field.value.field
              display_name = facet_field.value.display_name != "" ? facet_field.value.display_name : null
            }
          }

          dynamic "fields_ui_components_map" {
            for_each = data_store_ui_configs.value.fields_ui_components_map
            content {
              ui_component      = fields_ui_components_map.value.ui_component
              field             = fields_ui_components_map.value.field
              device_visibility = length(fields_ui_components_map.value.device_visibility) > 0 ? fields_ui_components_map.value.device_visibility : null
              display_template  = fields_ui_components_map.value.display_template != "" ? fields_ui_components_map.value.display_template : null
            }
          }
        }
      }

      dynamic "generative_answer_config" {
        for_each = ui_settings.value.generative_answer_config != null ? [ui_settings.value.generative_answer_config] : []
        content {
          disable_related_questions       = generative_answer_config.value.disable_related_questions
          ignore_adversarial_query        = generative_answer_config.value.ignore_adversarial_query
          ignore_low_relevant_content     = generative_answer_config.value.ignore_low_relevant_content
          ignore_non_answer_seeking_query = generative_answer_config.value.ignore_non_answer_seeking_query
          image_source                    = generative_answer_config.value.image_source != "" ? generative_answer_config.value.image_source : null
          language_code                   = generative_answer_config.value.language_code != "" ? generative_answer_config.value.language_code : null
          max_rephrase_steps              = generative_answer_config.value.max_rephrase_steps
          model_prompt_preamble           = generative_answer_config.value.model_prompt_preamble != "" ? generative_answer_config.value.model_prompt_preamble : null
          model_version                   = generative_answer_config.value.model_version != "" ? generative_answer_config.value.model_version : null
          result_count                    = generative_answer_config.value.result_count
        }
      }
    }
  }
}

# One Gemini Enterprise assistant per spec.assistants[] entry, keyed by
# assistant_id. Optional levers are sent only when set.
resource "google_discovery_engine_assistant" "this" {
  for_each = local.assistants

  project       = local.project_id
  location      = var.spec.location
  collection_id = local.collection_id
  engine_id     = local.created_engine_id
  assistant_id  = each.value.assistant_id
  display_name  = each.value.display_name
  description   = each.value.description != "" ? each.value.description : null

  web_grounding_type = each.value.web_grounding_type != "" ? each.value.web_grounding_type : null

  deletion_policy = local.deletion_policy

  dynamic "customer_policy" {
    for_each = each.value.customer_policy != null ? [each.value.customer_policy] : []
    content {
      dynamic "banned_phrases" {
        for_each = customer_policy.value.banned_phrases
        content {
          phrase            = banned_phrases.value.phrase
          match_type        = banned_phrases.value.match_type != "" ? banned_phrases.value.match_type : null
          ignore_diacritics = banned_phrases.value.ignore_diacritics
        }
      }

      dynamic "model_armor_config" {
        for_each = customer_policy.value.model_armor_config != null ? [customer_policy.value.model_armor_config] : []
        content {
          user_prompt_template = model_armor_config.value.user_prompt_template
          response_template    = model_armor_config.value.response_template
          failure_mode         = model_armor_config.value.failure_mode != "" ? model_armor_config.value.failure_mode : null
        }
      }
    }
  }

  dynamic "generation_config" {
    for_each = each.value.generation_config != null ? [each.value.generation_config] : []
    content {
      default_language = generation_config.value.default_language != "" ? generation_config.value.default_language : null

      dynamic "system_instruction" {
        for_each = generation_config.value.additional_system_instruction != "" ? [generation_config.value.additional_system_instruction] : []
        content {
          additional_system_instruction = system_instruction.value
        }
      }
    }
  }
}
