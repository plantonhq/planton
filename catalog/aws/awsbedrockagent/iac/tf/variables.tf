variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsBedrockAgent specification"
  type = object({
    region = string
    description = optional(string, "")
    foundation_model = string
    agent_resource_role_arn = string
    instruction = string
    idle_session_ttl_seconds = optional(number, 0)
    customer_encryption_key_arn = optional(string, "")
    agent_collaboration = optional(string, "")
    guardrail = optional(object({
      guardrail_id = string
      version = string
    }))
    memory = optional(object({
      enabled = optional(bool)
      storage_days = optional(number, 0)
      max_recent_sessions = optional(number, 0)
    }))
    prompt_override = optional(object({
      override_lambda = optional(string, "")
      prompt_configurations = list(object({
        prompt_type = optional(string, "")
        base_prompt_template = string
        parser_mode = optional(string, "")
        prompt_state = optional(string, "")
        inference_configuration = optional(object({
          max_length = optional(number)
          stop_sequences = optional(list(string), [])
          temperature = optional(number)
          top_k = optional(number)
          top_p = optional(number)
        }))
      }))
    }))
    action_groups = optional(list(object({
      name = string
      description = optional(string, "")
      state = optional(string, "")
      parent_action_group_signature = optional(string, "")
      executor = optional(object({
        lambda = optional(string, "")
        return_control = optional(bool, false)
      }))
      api_schema = optional(object({
        payload = optional(string, "")
        s3 = optional(object({
          bucket_name = string
          object_key = string
        }))
      }))
      function_schema = optional(object({
        functions = list(object({
          name = string
          description = optional(string, "")
          parameters = optional(list(object({
            name = string
            type = optional(string, "")
            description = optional(string, "")
            required = optional(bool, false)
          })), [])
        }))
      }))
    })), [])
    aliases = optional(list(object({
      name = string
      description = optional(string, "")
      routing = optional(object({
        agent_version = optional(string, "")
        provisioned_throughput = optional(string, "")
      }))
    })), [])
    collaborators = optional(list(object({
      name = string
      collaboration_instruction = string
      collaborator_alias_arn = string
      relay_conversation_history = optional(string, "")
    })), [])
    knowledge_base_associations = optional(list(object({
      name = string
      knowledge_base_id = string
      description = string
      state = optional(string, "")
    })), [])
  })
}
