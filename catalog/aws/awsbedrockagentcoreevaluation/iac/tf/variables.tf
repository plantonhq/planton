variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsBedrockAgentCoreEvaluation specification"
  type = object({
    region     = string
    evaluators = optional(any, [])
    harnesses = optional(list(object({
      name               = string
      execution_role_arn = string
      model = object({
        bedrock = optional(object({
          model_id    = string
          max_tokens  = optional(number, 0)
          temperature = optional(number)
          top_p       = optional(number)
        }))
        gemini = optional(object({
          api_key_arn = string
          model_id    = string
          max_tokens  = optional(number, 0)
          temperature = optional(number)
          top_p       = optional(number)
          top_k       = optional(number)
        }))
        openai = optional(object({
          api_key_arn = string
          model_id    = string
          max_tokens  = optional(number, 0)
          temperature = optional(number)
          top_p       = optional(number)
        }))
      })
      system_prompts = optional(list(object({
        text = string
      })), [])
      tools = optional(list(object({
        name = optional(string, "")
        type = optional(string, "")
        remote_mcp = optional(object({
          url     = string
          headers = optional(map(string), {})
        }))
        agentcore_browser = optional(object({
          browser_arn = optional(string, "")
        }))
        agentcore_gateway = optional(object({
          gateway_arn = string
          outbound_auth = optional(object({
            aws_iam = optional(bool, false)
            no_auth = optional(bool, false)
            oauth = optional(object({
              provider_arn       = string
              scopes             = list(string)
              custom_parameters  = optional(map(string), {})
              default_return_url = optional(string, "")
              grant_type         = optional(string, "")
            }))
          }))
        }))
        inline_function = optional(object({
          description  = string
          input_schema = string
        }))
        agentcore_code_interpreter = optional(object({
          code_interpreter_arn = optional(string, "")
        }))
      })), [])
      skill_paths = optional(list(string), [])
      memory = optional(object({
        memory_arn     = string
        actor_id       = optional(string, "")
        messages_count = optional(number, 0)
        retrieval = optional(object({
          namespace       = string
          relevance_score = optional(number)
          strategy_id     = optional(string, "")
          top_k           = optional(number, 0)
        }))
      }))
      environment_variables = optional(map(string), {})
      runtime_environment = optional(object({
        agent_runtime_arn = optional(string, "")
        filesystems = optional(list(object({
          mount_path           = string
          efs_access_point_arn = optional(string, "")
          s3_access_point_arn  = optional(string, "")
          session_storage      = optional(bool, false)
        })), [])
        lifecycle = optional(object({
          idle_runtime_session_timeout_seconds = optional(number, 0)
          max_lifetime_seconds                 = optional(number, 0)
        }))
        network = optional(object({
          mode = optional(string, "")
          vpc_config = optional(object({
            subnets                     = list(string)
            security_groups             = list(string)
            require_service_s3_endpoint = optional(bool, false)
          }))
        }))
      }))
      container_image_uri = optional(string, "")
      custom_jwt_authorizer = optional(object({
        discovery_url    = string
        allowed_audience = optional(list(string), [])
        allowed_clients  = optional(list(string), [])
        allowed_scopes   = optional(list(string), [])
        allowed_workloads = optional(object({
          workload_identities      = list(string)
          hosting_environment_arns = list(string)
        }))
        custom_claims = optional(list(object({
          claim_name     = string
          value_type     = optional(string, "")
          match_operator = optional(string, "")
          match_value    = optional(string, "")
          match_values   = optional(list(string), [])
        })), [])
        private_endpoint = optional(object({
          managed_vpc = optional(object({
            vpc_id                   = string
            subnet_ids               = list(string)
            security_group_ids       = optional(list(string), [])
            endpoint_ip_address_type = optional(string, "")
            routing_domain           = optional(string, "")
            tags                     = optional(map(string), {})
          }))
          self_managed_lattice = optional(object({
            resource_configuration_id = string
          }))
        }))
        private_endpoint_overrides = optional(list(object({
          domain = string
          private_endpoint = object({
            managed_vpc = optional(object({
              vpc_id                   = string
              subnet_ids               = list(string)
              security_group_ids       = optional(list(string), [])
              endpoint_ip_address_type = optional(string, "")
              routing_domain           = optional(string, "")
              tags                     = optional(map(string), {})
            }))
            self_managed_lattice = optional(object({
              resource_configuration_id = string
            }))
          })
        })), [])
      }))
      allowed_tools   = optional(list(string), [])
      max_iterations  = optional(number, 0)
      max_tokens      = optional(number, 0)
      timeout_seconds = optional(number, 0)
      truncation = optional(object({
        strategy = optional(string, "")
        sliding_window = optional(object({
          messages_count = optional(number, 0)
        }))
        summarization = optional(object({
          summary_ratio               = optional(number)
          preserve_recent_messages    = optional(number, 0)
          summarization_system_prompt = optional(string, "")
        }))
      }))
    })), [])
    online_evaluation_configs = optional(list(object({
      name               = string
      description        = optional(string, "")
      execution_role_arn = string
      enabled            = optional(bool)
      data_source = object({
        log_group_names = list(string)
        service_names   = list(string)
      })
      evaluator_ids = list(string)
      rule = object({
        sampling_percentage = optional(number, 0)
        filters = optional(list(object({
          key           = string
          operator      = optional(string, "")
          string_value  = optional(string, "")
          boolean_value = optional(bool)
          double_value  = optional(number)
        })), [])
        session_timeout_minutes = optional(number, 0)
      })
    })), [])
  })
}