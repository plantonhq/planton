# Enable the Dialogflow API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one agent must never
# disable the API for everything else in the project.
resource "google_project_service" "dialogflow_api" {
  project = local.project_id
  service = "dialogflow.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The agent. project, location, and default_language_code are immutable;
# everything else updates in place. Google creates the default start flow
# and the default playbook with it.
resource "google_dialogflow_cx_agent" "this" {
  project                  = local.project_id
  location                 = var.spec.location
  display_name             = local.display_name
  default_language_code    = var.spec.default_language_code
  time_zone                = var.spec.time_zone
  description              = local.description
  avatar_uri               = local.avatar_uri
  supported_language_codes = local.supported_language_codes
  security_settings        = local.security_settings
  start_playbook           = local.start_playbook

  # Optional booleans are sent only when true (Google's default is false);
  # turning one off plans the change back to false.
  enable_multi_language_training = var.spec.enable_multi_language_training ? true : null
  enable_spell_correction        = var.spec.enable_spell_correction ? true : null
  locked                         = var.spec.locked ? true : null

  # Client-side: also delete the linked Vertex AI Search engine on destroy.
  delete_chat_engine_on_destroy = var.spec.delete_chat_engine_on_destroy ? true : null

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON,
  # fanned to every webhook, tool, tool version, version, and environment
  # below. Sent only when set so the provider default stays in charge.
  deletion_policy = local.deletion_policy

  # Optional and computed: an unset block leaves Google's values alone, and
  # each nested block is sent only when declared.
  dynamic "advanced_settings" {
    for_each = var.spec.advanced_settings != null ? [var.spec.advanced_settings] : []
    content {
      dynamic "audio_export_gcs_destination" {
        for_each = advanced_settings.value.audio_export_gcs_destination != null ? [advanced_settings.value.audio_export_gcs_destination] : []
        content {
          uri = audio_export_gcs_destination.value.uri != "" ? audio_export_gcs_destination.value.uri : null
        }
      }
      dynamic "dtmf_settings" {
        for_each = advanced_settings.value.dtmf_settings != null ? [advanced_settings.value.dtmf_settings] : []
        content {
          enabled      = dtmf_settings.value.enabled ? true : null
          finish_digit = dtmf_settings.value.finish_digit != "" ? dtmf_settings.value.finish_digit : null
          max_digits   = dtmf_settings.value.max_digits > 0 ? dtmf_settings.value.max_digits : null
        }
      }
      dynamic "logging_settings" {
        for_each = advanced_settings.value.logging_settings != null ? [advanced_settings.value.logging_settings] : []
        content {
          enable_consent_based_redaction = logging_settings.value.enable_consent_based_redaction ? true : null
          enable_interaction_logging     = logging_settings.value.enable_interaction_logging ? true : null
          enable_stackdriver_logging     = logging_settings.value.enable_stackdriver_logging ? true : null
        }
      }
      dynamic "speech_settings" {
        for_each = advanced_settings.value.speech_settings != null ? [advanced_settings.value.speech_settings] : []
        content {
          endpointer_sensitivity        = speech_settings.value.endpointer_sensitivity > 0 ? speech_settings.value.endpointer_sensitivity : null
          models                        = length(speech_settings.value.models) > 0 ? speech_settings.value.models : null
          no_speech_timeout             = speech_settings.value.no_speech_timeout != "" ? speech_settings.value.no_speech_timeout : null
          use_timeout_based_endpointing = speech_settings.value.use_timeout_based_endpointing ? true : null
        }
      }
    }
  }

  # The spec lifts the block's one flag.
  dynamic "answer_feedback_settings" {
    for_each = var.spec.enable_answer_feedback ? [true] : []
    content {
      enable_answer_feedback = true
    }
  }

  dynamic "client_certificate_settings" {
    for_each = var.spec.client_certificate_settings != null ? [var.spec.client_certificate_settings] : []
    content {
      ssl_certificate = client_certificate_settings.value.ssl_certificate
      private_key     = client_certificate_settings.value.private_key
      passphrase      = client_certificate_settings.value.passphrase != "" ? client_certificate_settings.value.passphrase : null
    }
  }

  dynamic "gen_app_builder_settings" {
    for_each = var.spec.gen_app_builder_settings != null ? [var.spec.gen_app_builder_settings] : []
    content {
      engine = gen_app_builder_settings.value.engine
    }
  }

  dynamic "git_integration_settings" {
    for_each = var.spec.git_integration_settings != null ? [var.spec.git_integration_settings] : []
    content {
      dynamic "github_settings" {
        for_each = git_integration_settings.value.github_settings != null ? [git_integration_settings.value.github_settings] : []
        content {
          display_name    = github_settings.value.display_name != "" ? github_settings.value.display_name : null
          repository_uri  = github_settings.value.repository_uri != "" ? github_settings.value.repository_uri : null
          tracking_branch = github_settings.value.tracking_branch != "" ? github_settings.value.tracking_branch : null
          branches        = length(github_settings.value.branches) > 0 ? github_settings.value.branches : null
          access_token    = github_settings.value.access_token != "" ? github_settings.value.access_token : null
        }
      }
    }
  }

  # The spec lifts these three blocks' single leaves.
  dynamic "personalization_settings" {
    for_each = local.default_end_user_metadata != null ? [local.default_end_user_metadata] : []
    content {
      default_end_user_metadata = personalization_settings.value
    }
  }

  dynamic "speech_to_text_settings" {
    for_each = var.spec.enable_speech_adaptation ? [true] : []
    content {
      enable_speech_adaptation = true
    }
  }

  dynamic "text_to_speech_settings" {
    for_each = local.synthesize_speech_configs != null ? [local.synthesize_speech_configs] : []
    content {
      synthesize_speech_configs = text_to_speech_settings.value
    }
  }

  depends_on = [google_project_service.dialogflow_api]
}

# The folded webhooks, one per spec.webhooks[] entry, keyed by display name.
# A webhook is a generic web service or a Service Directory service; both
# carry the same endpoint block, written out twice below because Terraform
# cannot share a block template.
resource "google_dialogflow_cx_webhook" "this" {
  for_each = local.webhooks

  parent          = google_dialogflow_cx_agent.this.id
  display_name    = each.key
  disabled        = each.value.disabled ? true : null
  timeout         = each.value.timeout != "" ? each.value.timeout : null
  deletion_policy = local.deletion_policy

  dynamic "generic_web_service" {
    for_each = each.value.generic_web_service != null ? [each.value.generic_web_service] : []
    content {
      uri                                  = generic_web_service.value.uri
      webhook_type                         = generic_web_service.value.webhook_type != "" ? generic_web_service.value.webhook_type : null
      http_method                          = generic_web_service.value.http_method != "" ? generic_web_service.value.http_method : null
      request_body                         = generic_web_service.value.request_body != "" ? generic_web_service.value.request_body : null
      parameter_mapping                    = length(generic_web_service.value.parameter_mapping) > 0 ? generic_web_service.value.parameter_mapping : null
      request_headers                      = length(generic_web_service.value.request_headers) > 0 ? generic_web_service.value.request_headers : null
      secret_version_for_username_password = generic_web_service.value.secret_version_for_username_password != "" ? generic_web_service.value.secret_version_for_username_password : null
      service_agent_auth                   = generic_web_service.value.service_agent_auth != "" ? generic_web_service.value.service_agent_auth : null
      allowed_ca_certs                     = length(generic_web_service.value.allowed_ca_certs) > 0 ? generic_web_service.value.allowed_ca_certs : null

      dynamic "secret_versions_for_request_headers" {
        for_each = generic_web_service.value.secret_versions_for_request_headers
        content {
          key            = secret_versions_for_request_headers.value.key
          secret_version = secret_versions_for_request_headers.value.secret_version
        }
      }

      dynamic "oauth_config" {
        for_each = generic_web_service.value.oauth_config != null ? [generic_web_service.value.oauth_config] : []
        content {
          client_id                        = oauth_config.value.client_id
          token_endpoint                   = oauth_config.value.token_endpoint
          client_secret                    = oauth_config.value.client_secret != "" ? oauth_config.value.client_secret : null
          scopes                           = length(oauth_config.value.scopes) > 0 ? oauth_config.value.scopes : null
          secret_version_for_client_secret = oauth_config.value.secret_version_for_client_secret != "" ? oauth_config.value.secret_version_for_client_secret : null
        }
      }

      # The spec lifts the block's one field.
      dynamic "service_account_auth_config" {
        for_each = generic_web_service.value.service_account != "" ? [generic_web_service.value.service_account] : []
        content {
          service_account = service_account_auth_config.value
        }
      }
    }
  }

  dynamic "service_directory" {
    for_each = each.value.service_directory != null ? [each.value.service_directory] : []
    content {
      service = service_directory.value.service

      dynamic "generic_web_service" {
        for_each = service_directory.value.generic_web_service != null ? [service_directory.value.generic_web_service] : []
        content {
          uri                                  = generic_web_service.value.uri
          webhook_type                         = generic_web_service.value.webhook_type != "" ? generic_web_service.value.webhook_type : null
          http_method                          = generic_web_service.value.http_method != "" ? generic_web_service.value.http_method : null
          request_body                         = generic_web_service.value.request_body != "" ? generic_web_service.value.request_body : null
          parameter_mapping                    = length(generic_web_service.value.parameter_mapping) > 0 ? generic_web_service.value.parameter_mapping : null
          request_headers                      = length(generic_web_service.value.request_headers) > 0 ? generic_web_service.value.request_headers : null
          secret_version_for_username_password = generic_web_service.value.secret_version_for_username_password != "" ? generic_web_service.value.secret_version_for_username_password : null
          service_agent_auth                   = generic_web_service.value.service_agent_auth != "" ? generic_web_service.value.service_agent_auth : null
          allowed_ca_certs                     = length(generic_web_service.value.allowed_ca_certs) > 0 ? generic_web_service.value.allowed_ca_certs : null

          dynamic "secret_versions_for_request_headers" {
            for_each = generic_web_service.value.secret_versions_for_request_headers
            content {
              key            = secret_versions_for_request_headers.value.key
              secret_version = secret_versions_for_request_headers.value.secret_version
            }
          }

          dynamic "oauth_config" {
            for_each = generic_web_service.value.oauth_config != null ? [generic_web_service.value.oauth_config] : []
            content {
              client_id                        = oauth_config.value.client_id
              token_endpoint                   = oauth_config.value.token_endpoint
              client_secret                    = oauth_config.value.client_secret != "" ? oauth_config.value.client_secret : null
              scopes                           = length(oauth_config.value.scopes) > 0 ? oauth_config.value.scopes : null
              secret_version_for_client_secret = oauth_config.value.secret_version_for_client_secret != "" ? oauth_config.value.secret_version_for_client_secret : null
            }
          }

          dynamic "service_account_auth_config" {
            for_each = generic_web_service.value.service_account != "" ? [generic_web_service.value.service_account] : []
            content {
              service_account = service_account_auth_config.value
            }
          }
        }
      }
    }
  }
}

# The folded tools, one per spec.tools[] entry, keyed by display name. The
# tool definition is written out again under tool versions below, because
# Terraform cannot share a block template.
resource "google_dialogflow_cx_tool" "this" {
  for_each = local.tools

  parent          = google_dialogflow_cx_agent.this.id
  display_name    = each.key
  description     = each.value.description
  deletion_policy = local.deletion_policy

  dynamic "open_api_spec" {
    for_each = each.value.open_api_spec != null ? [each.value.open_api_spec] : []
    content {
      text_schema = open_api_spec.value.text_schema

      dynamic "authentication" {
        for_each = open_api_spec.value.authentication != null ? [open_api_spec.value.authentication] : []
        content {
          dynamic "api_key_config" {
            for_each = authentication.value.api_key_config != null ? [authentication.value.api_key_config] : []
            content {
              key_name                   = api_key_config.value.key_name
              request_location           = api_key_config.value.request_location
              api_key                    = api_key_config.value.api_key != "" ? api_key_config.value.api_key : null
              secret_version_for_api_key = api_key_config.value.secret_version_for_api_key != "" ? api_key_config.value.secret_version_for_api_key : null
            }
          }
          dynamic "bearer_token_config" {
            for_each = authentication.value.bearer_token_config != null ? [authentication.value.bearer_token_config] : []
            content {
              token                    = bearer_token_config.value.token != "" ? bearer_token_config.value.token : null
              secret_version_for_token = bearer_token_config.value.secret_version_for_token != "" ? bearer_token_config.value.secret_version_for_token : null
            }
          }
          dynamic "oauth_config" {
            for_each = authentication.value.oauth_config != null ? [authentication.value.oauth_config] : []
            content {
              client_id                        = oauth_config.value.client_id
              oauth_grant_type                 = oauth_config.value.oauth_grant_type
              token_endpoint                   = oauth_config.value.token_endpoint
              client_secret                    = oauth_config.value.client_secret != "" ? oauth_config.value.client_secret : null
              scopes                           = length(oauth_config.value.scopes) > 0 ? oauth_config.value.scopes : null
              secret_version_for_client_secret = oauth_config.value.secret_version_for_client_secret != "" ? oauth_config.value.secret_version_for_client_secret : null
            }
          }
          dynamic "service_agent_auth_config" {
            for_each = authentication.value.service_agent_auth_config != null ? [authentication.value.service_agent_auth_config] : []
            content {
              service_agent_auth = service_agent_auth_config.value.service_agent_auth != "" ? service_agent_auth_config.value.service_agent_auth : null
            }
          }
        }
      }

      dynamic "service_directory_config" {
        for_each = open_api_spec.value.service_directory_config != null ? [open_api_spec.value.service_directory_config] : []
        content {
          service = service_directory_config.value.service
        }
      }

      dynamic "tls_config" {
        for_each = open_api_spec.value.tls_config != null ? [open_api_spec.value.tls_config] : []
        content {
          dynamic "ca_certs" {
            for_each = tls_config.value.ca_certs
            content {
              display_name = ca_certs.value.display_name
              cert         = ca_certs.value.cert
            }
          }
        }
      }
    }
  }

  dynamic "data_store_spec" {
    for_each = each.value.data_store_spec != null ? [each.value.data_store_spec] : []
    content {
      dynamic "data_store_connections" {
        for_each = data_store_spec.value.data_store_connections
        content {
          data_store               = data_store_connections.value.data_store != "" ? data_store_connections.value.data_store : null
          data_store_type          = data_store_connections.value.data_store_type != "" ? data_store_connections.value.data_store_type : null
          document_processing_mode = data_store_connections.value.document_processing_mode != "" ? data_store_connections.value.document_processing_mode : null
        }
      }

      # Required by Google and carries no settings.
      fallback_prompt {}
    }
  }

  dynamic "function_spec" {
    for_each = each.value.function_spec != null ? [each.value.function_spec] : []
    content {
      input_schema  = function_spec.value.input_schema != "" ? function_spec.value.input_schema : null
      output_schema = function_spec.value.output_schema != "" ? function_spec.value.output_schema : null
    }
  }
}

# The folded tool versions, keyed "{tool}/{version}". Every argument is
# immutable: any change replaces the version.
resource "google_dialogflow_cx_tool_version" "this" {
  for_each = local.tool_versions

  parent          = google_dialogflow_cx_tool.this[each.value.tool_key].id
  display_name    = each.value.version.display_name
  deletion_policy = local.deletion_policy

  tool {
    display_name = each.value.version.tool.display_name
    description  = each.value.version.tool.description

    dynamic "open_api_spec" {
      for_each = each.value.version.tool.open_api_spec != null ? [each.value.version.tool.open_api_spec] : []
      content {
        text_schema = open_api_spec.value.text_schema

        dynamic "authentication" {
          for_each = open_api_spec.value.authentication != null ? [open_api_spec.value.authentication] : []
          content {
            dynamic "api_key_config" {
              for_each = authentication.value.api_key_config != null ? [authentication.value.api_key_config] : []
              content {
                key_name                   = api_key_config.value.key_name
                request_location           = api_key_config.value.request_location
                api_key                    = api_key_config.value.api_key != "" ? api_key_config.value.api_key : null
                secret_version_for_api_key = api_key_config.value.secret_version_for_api_key != "" ? api_key_config.value.secret_version_for_api_key : null
              }
            }
            dynamic "bearer_token_config" {
              for_each = authentication.value.bearer_token_config != null ? [authentication.value.bearer_token_config] : []
              content {
                token                    = bearer_token_config.value.token != "" ? bearer_token_config.value.token : null
                secret_version_for_token = bearer_token_config.value.secret_version_for_token != "" ? bearer_token_config.value.secret_version_for_token : null
              }
            }
            dynamic "oauth_config" {
              for_each = authentication.value.oauth_config != null ? [authentication.value.oauth_config] : []
              content {
                client_id                        = oauth_config.value.client_id
                oauth_grant_type                 = oauth_config.value.oauth_grant_type
                token_endpoint                   = oauth_config.value.token_endpoint
                client_secret                    = oauth_config.value.client_secret != "" ? oauth_config.value.client_secret : null
                scopes                           = length(oauth_config.value.scopes) > 0 ? oauth_config.value.scopes : null
                secret_version_for_client_secret = oauth_config.value.secret_version_for_client_secret != "" ? oauth_config.value.secret_version_for_client_secret : null
              }
            }
            dynamic "service_agent_auth_config" {
              for_each = authentication.value.service_agent_auth_config != null ? [authentication.value.service_agent_auth_config] : []
              content {
                service_agent_auth = service_agent_auth_config.value.service_agent_auth != "" ? service_agent_auth_config.value.service_agent_auth : null
              }
            }
          }
        }

        dynamic "service_directory_config" {
          for_each = open_api_spec.value.service_directory_config != null ? [open_api_spec.value.service_directory_config] : []
          content {
            service = service_directory_config.value.service
          }
        }

        dynamic "tls_config" {
          for_each = open_api_spec.value.tls_config != null ? [open_api_spec.value.tls_config] : []
          content {
            dynamic "ca_certs" {
              for_each = tls_config.value.ca_certs
              content {
                display_name = ca_certs.value.display_name
                cert         = ca_certs.value.cert
              }
            }
          }
        }
      }
    }

    dynamic "data_store_spec" {
      for_each = each.value.version.tool.data_store_spec != null ? [each.value.version.tool.data_store_spec] : []
      content {
        dynamic "data_store_connections" {
          for_each = data_store_spec.value.data_store_connections
          content {
            data_store               = data_store_connections.value.data_store != "" ? data_store_connections.value.data_store : null
            data_store_type          = data_store_connections.value.data_store_type != "" ? data_store_connections.value.data_store_type : null
            document_processing_mode = data_store_connections.value.document_processing_mode != "" ? data_store_connections.value.document_processing_mode : null
          }
        }

        fallback_prompt {}
      }
    }

    dynamic "function_spec" {
      for_each = each.value.version.tool.function_spec != null ? [each.value.version.tool.function_spec] : []
      content {
        input_schema  = function_spec.value.input_schema != "" ? function_spec.value.input_schema : null
        output_schema = function_spec.value.output_schema != "" ? function_spec.value.output_schema : null
      }
    }
  }
}

# The folded flow versions, keyed "{flow_id}/{display name}". A version
# snapshots its flow at creation and waits for the flow's model to train.
resource "google_dialogflow_cx_version" "this" {
  for_each = local.versions

  parent          = "${google_dialogflow_cx_agent.this.id}/flows/${each.value.flow_id}"
  display_name    = each.value.display_name
  description     = each.value.description != "" ? each.value.description : null
  deletion_policy = local.deletion_policy
}

# The folded environments, keyed by display name. Each version config
# resolves to a declared version or composes an outside version's path
# under this agent.
resource "google_dialogflow_cx_environment" "this" {
  for_each = local.environments

  parent          = google_dialogflow_cx_agent.this.id
  display_name    = each.key
  description     = each.value.description != "" ? each.value.description : null
  deletion_policy = local.deletion_policy

  dynamic "version_configs" {
    for_each = each.value.version_configs
    content {
      version = (
        version_configs.value.version != ""
        ? google_dialogflow_cx_version.this["${version_configs.value.flow_id != "" ? version_configs.value.flow_id : local.start_flow_id}/${version_configs.value.version}"].id
        : "${google_dialogflow_cx_agent.this.id}/flows/${version_configs.value.flow_id != "" ? version_configs.value.flow_id : local.start_flow_id}/versions/${version_configs.value.version_id}"
      )
    }
  }
}

# The folded generative settings, one per language. Created by a PATCH on
# the agent's per-language settings; destroy only stops managing them.
resource "google_dialogflow_cx_generative_settings" "this" {
  for_each = local.generative_settings

  parent        = google_dialogflow_cx_agent.this.id
  language_code = each.key

  dynamic "fallback_settings" {
    for_each = each.value.fallback_settings != null ? [each.value.fallback_settings] : []
    content {
      selected_prompt = fallback_settings.value.selected_prompt != "" ? fallback_settings.value.selected_prompt : null

      dynamic "prompt_templates" {
        for_each = fallback_settings.value.prompt_templates
        content {
          display_name = prompt_templates.value.display_name != "" ? prompt_templates.value.display_name : null
          frozen       = prompt_templates.value.frozen ? true : null
          prompt_text  = prompt_templates.value.prompt_text != "" ? prompt_templates.value.prompt_text : null
        }
      }
    }
  }

  dynamic "generative_safety_settings" {
    for_each = each.value.generative_safety_settings != null ? [each.value.generative_safety_settings] : []
    content {
      default_banned_phrase_match_strategy = generative_safety_settings.value.default_banned_phrase_match_strategy != "" ? generative_safety_settings.value.default_banned_phrase_match_strategy : null

      dynamic "banned_phrases" {
        for_each = generative_safety_settings.value.banned_phrases
        content {
          language_code = banned_phrases.value.language_code
          text          = banned_phrases.value.text
        }
      }
    }
  }

  dynamic "knowledge_connector_settings" {
    for_each = each.value.knowledge_connector_settings != null ? [each.value.knowledge_connector_settings] : []
    content {
      agent                       = knowledge_connector_settings.value.agent != "" ? knowledge_connector_settings.value.agent : null
      agent_identity              = knowledge_connector_settings.value.agent_identity != "" ? knowledge_connector_settings.value.agent_identity : null
      agent_scope                 = knowledge_connector_settings.value.agent_scope != "" ? knowledge_connector_settings.value.agent_scope : null
      business                    = knowledge_connector_settings.value.business != "" ? knowledge_connector_settings.value.business : null
      business_description        = knowledge_connector_settings.value.business_description != "" ? knowledge_connector_settings.value.business_description : null
      disable_data_store_fallback = knowledge_connector_settings.value.disable_data_store_fallback ? true : null
    }
  }

  dynamic "llm_model_settings" {
    for_each = each.value.llm_model_settings != null ? [each.value.llm_model_settings] : []
    content {
      model       = llm_model_settings.value.model != "" ? llm_model_settings.value.model : null
      prompt_text = llm_model_settings.value.prompt_text != "" ? llm_model_settings.value.prompt_text : null
    }
  }
}
