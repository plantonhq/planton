# Enable the Vertex AI API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one agent must never
# disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Vertex AI Agent Engine instance (Google's ReasoningEngine): the
# agent's code source, identity, deployment shape, and optional Memory
# Bank. Location and the encryption key are immutable; everything else
# updates in place, and a new source archive redeploys the agent's code.
# Optional strings, booleans, and Optional+Computed numbers are sent only
# when set so Google's defaults stay in charge -- the Pulumi module's
# posture. build_spec.service_account is not sent: the pinned Pulumi SDK
# lacks it, and an argument one engine cannot send is never a one-engine
# field.
resource "google_vertex_ai_reasoning_engine" "this" {
  project      = local.project_id
  display_name = local.display_name
  description  = local.description
  labels       = local.final_labels
  # The provider names the axis `region`; the spec keeps the Vertex
  # family's single word, `location`.
  region = var.spec.location

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  deletion_policy = local.deletion_policy

  # CMEK: the agent's data encrypted under this key.
  dynamic "encryption_spec" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      kms_key_name = encryption_spec.value
    }
  }

  dynamic "spec" {
    for_each = local.agent != null ? [local.agent] : []
    content {
      agent_framework = spec.value.agent_framework != "" ? spec.value.agent_framework : null
      class_methods   = spec.value.class_methods != "" ? spec.value.class_methods : null
      identity_type   = spec.value.identity_type != "" ? spec.value.identity_type : null
      service_account = spec.value.service_account != "" ? spec.value.service_account : null

      dynamic "container_spec" {
        for_each = spec.value.container_spec != null ? [spec.value.container_spec] : []
        content {
          image_uri = container_spec.value.image_uri
          port      = container_spec.value.port
        }
      }

      dynamic "source_code_spec" {
        for_each = spec.value.source_code_spec != null ? [spec.value.source_code_spec] : []
        content {
          dynamic "inline_source" {
            for_each = source_code_spec.value.inline_source != null ? [source_code_spec.value.inline_source] : []
            content {
              source_archive = inline_source.value.source_archive
            }
          }
          dynamic "developer_connect_source" {
            for_each = source_code_spec.value.developer_connect_source != null ? [source_code_spec.value.developer_connect_source] : []
            content {
              config {
                git_repository_link = developer_connect_source.value.config.git_repository_link
                dir                 = developer_connect_source.value.config.dir
                revision            = developer_connect_source.value.config.revision
              }
            }
          }
          dynamic "agent_config_source" {
            for_each = source_code_spec.value.agent_config_source != null ? [source_code_spec.value.agent_config_source] : []
            content {
              adk_config {
                json_config = agent_config_source.value.adk_config.json_config
              }
              dynamic "inline_source" {
                for_each = agent_config_source.value.inline_source != null ? [agent_config_source.value.inline_source] : []
                content {
                  source_archive = inline_source.value.source_archive
                }
              }
            }
          }
          dynamic "python_spec" {
            for_each = source_code_spec.value.python_spec != null ? [source_code_spec.value.python_spec] : []
            content {
              version           = python_spec.value.version != "" ? python_spec.value.version : null
              entrypoint_module = python_spec.value.entrypoint_module != "" ? python_spec.value.entrypoint_module : null
              entrypoint_object = python_spec.value.entrypoint_object != "" ? python_spec.value.entrypoint_object : null
              requirements_file = python_spec.value.requirements_file != "" ? python_spec.value.requirements_file : null
            }
          }
          dynamic "image_spec" {
            for_each = source_code_spec.value.image_spec != null ? [source_code_spec.value.image_spec] : []
            content {
              build_args = length(image_spec.value.build_args) > 0 ? image_spec.value.build_args : null
            }
          }
        }
      }

      dynamic "package_spec" {
        for_each = spec.value.package_spec != null ? [spec.value.package_spec] : []
        content {
          pickle_object_gcs_uri    = package_spec.value.pickle_object_gcs_uri != "" ? package_spec.value.pickle_object_gcs_uri : null
          dependency_files_gcs_uri = package_spec.value.dependency_files_gcs_uri != "" ? package_spec.value.dependency_files_gcs_uri : null
          requirements_gcs_uri     = package_spec.value.requirements_gcs_uri != "" ? package_spec.value.requirements_gcs_uri : null
          python_version           = package_spec.value.python_version != "" ? package_spec.value.python_version : null
        }
      }

      dynamic "build_spec" {
        for_each = spec.value.build_spec != null && spec.value.build_spec.worker_pool != "" ? [spec.value.build_spec] : []
        content {
          worker_pool = build_spec.value.worker_pool
        }
      }

      dynamic "deployment_spec" {
        for_each = spec.value.deployment_spec != null ? [spec.value.deployment_spec] : []
        content {
          # Optional+Computed: sent only when set.
          min_instances         = deployment_spec.value.min_instances
          max_instances         = deployment_spec.value.max_instances
          container_concurrency = deployment_spec.value.container_concurrency
          resource_limits       = length(deployment_spec.value.resource_limits) > 0 ? deployment_spec.value.resource_limits : null

          dynamic "env" {
            for_each = deployment_spec.value.env
            content {
              name  = env.value.name
              value = env.value.value
            }
          }

          dynamic "secret_env" {
            for_each = deployment_spec.value.secret_env
            content {
              name = secret_env.value.name
              secret_ref {
                secret  = secret_env.value.secret_ref.secret
                version = secret_env.value.secret_ref.version != "" ? secret_env.value.secret_ref.version : null
              }
            }
          }

          dynamic "psc_interface_config" {
            for_each = deployment_spec.value.psc_interface_config != null ? [deployment_spec.value.psc_interface_config] : []
            content {
              network_attachment = psc_interface_config.value.network_attachment != "" ? psc_interface_config.value.network_attachment : null

              dynamic "dns_peering_configs" {
                for_each = psc_interface_config.value.dns_peering_configs
                content {
                  domain         = dns_peering_configs.value.domain
                  target_project = dns_peering_configs.value.target_project
                  target_network = dns_peering_configs.value.target_network
                }
              }
            }
          }

          dynamic "agent_gateway_config" {
            for_each = deployment_spec.value.agent_gateway_config != null ? [deployment_spec.value.agent_gateway_config] : []
            content {
              dynamic "client_to_agent_config" {
                for_each = agent_gateway_config.value.client_to_agent_config != null ? [agent_gateway_config.value.client_to_agent_config] : []
                content {
                  agent_gateway = client_to_agent_config.value.agent_gateway
                }
              }
              dynamic "agent_to_anywhere_config" {
                for_each = agent_gateway_config.value.agent_to_anywhere_config != null ? [agent_gateway_config.value.agent_to_anywhere_config] : []
                content {
                  agent_gateway = agent_to_anywhere_config.value.agent_gateway
                }
              }
            }
          }
        }
      }
    }
  }

  # The Memory Bank. Example conversations carry Gemini's Content.Part
  # union; the audio_transcription payload is not offered because the
  # pinned Pulumi SDK lacks it.
  dynamic "context_spec" {
    for_each = local.memory_bank != null ? [local.memory_bank] : []
    content {
      memory_bank_config {
        disable_memory_revisions = context_spec.value.disable_memory_revisions ? true : null

        dynamic "generation_config" {
          for_each = context_spec.value.generation_config != null ? [context_spec.value.generation_config] : []
          content {
            model = generation_config.value.model

            dynamic "generation_trigger_config" {
              for_each = generation_config.value.generation_trigger_config != null ? [generation_config.value.generation_trigger_config] : []
              content {
                dynamic "generation_rule" {
                  for_each = generation_trigger_config.value.generation_rule != null ? [generation_trigger_config.value.generation_rule] : []
                  content {
                    event_count         = generation_rule.value.event_count
                    fixed_interval      = generation_rule.value.fixed_interval != "" ? generation_rule.value.fixed_interval : null
                    idle_duration       = generation_rule.value.idle_duration != "" ? generation_rule.value.idle_duration : null
                    overlap_event_count = generation_rule.value.overlap_event_count
                  }
                }
              }
            }
          }
        }

        dynamic "similarity_search_config" {
          for_each = context_spec.value.similarity_search_config != null ? [context_spec.value.similarity_search_config] : []
          content {
            embedding_model = similarity_search_config.value.embedding_model
          }
        }

        dynamic "ttl_config" {
          for_each = context_spec.value.ttl_config != null ? [context_spec.value.ttl_config] : []
          content {
            default_ttl                 = ttl_config.value.default_ttl != "" ? ttl_config.value.default_ttl : null
            memory_revision_default_ttl = ttl_config.value.memory_revision_default_ttl != "" ? ttl_config.value.memory_revision_default_ttl : null

            dynamic "granular_ttl_config" {
              for_each = ttl_config.value.granular_ttl_config != null ? [ttl_config.value.granular_ttl_config] : []
              content {
                create_ttl           = granular_ttl_config.value.create_ttl != "" ? granular_ttl_config.value.create_ttl : null
                generate_created_ttl = granular_ttl_config.value.generate_created_ttl != "" ? granular_ttl_config.value.generate_created_ttl : null
                generate_updated_ttl = granular_ttl_config.value.generate_updated_ttl != "" ? granular_ttl_config.value.generate_updated_ttl : null
              }
            }
          }
        }

        dynamic "structured_memory_configs" {
          for_each = context_spec.value.structured_memory_configs
          content {
            scope_keys = length(structured_memory_configs.value.scope_keys) > 0 ? structured_memory_configs.value.scope_keys : null

            dynamic "schema_configs" {
              for_each = structured_memory_configs.value.schema_configs
              content {
                id            = schema_configs.value.id
                memory_schema = schema_configs.value.memory_schema != "" ? schema_configs.value.memory_schema : null
              }
            }
          }
        }

        dynamic "customization_configs" {
          for_each = context_spec.value.customization_configs
          content {
            scope_keys                        = length(customization_configs.value.scope_keys) > 0 ? customization_configs.value.scope_keys : null
            disable_natural_language_memories = customization_configs.value.disable_natural_language_memories ? true : null
            enable_third_person_memories      = customization_configs.value.enable_third_person_memories ? true : null

            dynamic "memory_topics" {
              for_each = customization_configs.value.memory_topics
              content {
                dynamic "custom_memory_topic" {
                  for_each = memory_topics.value.custom_memory_topic != null ? [memory_topics.value.custom_memory_topic] : []
                  content {
                    label       = custom_memory_topic.value.label
                    description = custom_memory_topic.value.description != "" ? custom_memory_topic.value.description : null
                  }
                }
                dynamic "managed_memory_topic" {
                  for_each = memory_topics.value.managed_memory_topic != null ? [memory_topics.value.managed_memory_topic] : []
                  content {
                    managed_topic_enum = managed_memory_topic.value.managed_topic_enum
                  }
                }
              }
            }

            dynamic "generate_memories_examples" {
              for_each = customization_configs.value.generate_memories_examples
              content {
                dynamic "conversation_source" {
                  for_each = generate_memories_examples.value.conversation_source != null ? [generate_memories_examples.value.conversation_source] : []
                  content {
                    dynamic "events" {
                      for_each = conversation_source.value.events
                      content {
                        content {
                          role = events.value.content.role != "" ? events.value.content.role : null

                          dynamic "parts" {
                            for_each = events.value.content.parts
                            content {
                              text    = parts.value.text != "" ? parts.value.text : null
                              thought = parts.value.thought ? true : null

                              dynamic "inline_data" {
                                for_each = parts.value.inline_data != null ? [parts.value.inline_data] : []
                                content {
                                  mime_type = inline_data.value.mime_type
                                  data      = inline_data.value.data
                                }
                              }
                              dynamic "file_data" {
                                for_each = parts.value.file_data != null ? [parts.value.file_data] : []
                                content {
                                  mime_type = file_data.value.mime_type
                                  file_uri  = file_data.value.file_uri
                                }
                              }
                              dynamic "function_call" {
                                for_each = parts.value.function_call != null ? [parts.value.function_call] : []
                                content {
                                  id   = function_call.value.id != "" ? function_call.value.id : null
                                  name = function_call.value.name != "" ? function_call.value.name : null
                                  args = function_call.value.args != "" ? function_call.value.args : null
                                }
                              }
                              dynamic "function_response" {
                                for_each = parts.value.function_response != null ? [parts.value.function_response] : []
                                content {
                                  id       = function_response.value.id != "" ? function_response.value.id : null
                                  name     = function_response.value.name
                                  response = function_response.value.response != "" ? function_response.value.response : null
                                }
                              }
                              dynamic "executable_code" {
                                for_each = parts.value.executable_code != null ? [parts.value.executable_code] : []
                                content {
                                  id       = executable_code.value.id != "" ? executable_code.value.id : null
                                  language = executable_code.value.language
                                  code     = executable_code.value.code
                                }
                              }
                              dynamic "code_execution_result" {
                                for_each = parts.value.code_execution_result != null ? [parts.value.code_execution_result] : []
                                content {
                                  id      = code_execution_result.value.id != "" ? code_execution_result.value.id : null
                                  outcome = code_execution_result.value.outcome
                                  output  = code_execution_result.value.output != "" ? code_execution_result.value.output : null
                                }
                              }
                              dynamic "video_metadata" {
                                for_each = parts.value.video_metadata != null ? [parts.value.video_metadata] : []
                                content {
                                  start_offset = video_metadata.value.start_offset != "" ? video_metadata.value.start_offset : null
                                  end_offset   = video_metadata.value.end_offset != "" ? video_metadata.value.end_offset : null
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }

                dynamic "generated_memories" {
                  for_each = generate_memories_examples.value.generated_memories
                  content {
                    fact = generated_memories.value.fact

                    dynamic "topics" {
                      for_each = generated_memories.value.topics
                      content {
                        custom_memory_topic_label = topics.value.custom_memory_topic_label != "" ? topics.value.custom_memory_topic_label : null
                        managed_memory_topic      = topics.value.managed_memory_topic != "" ? topics.value.managed_memory_topic : null
                      }
                    }
                  }
                }
              }
            }

            dynamic "consolidation_config" {
              for_each = customization_configs.value.consolidation_config != null ? [customization_configs.value.consolidation_config] : []
              content {
                revisions_per_candidate_count = consolidation_config.value.revisions_per_candidate_count
              }
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_service.aiplatform_api]
}
