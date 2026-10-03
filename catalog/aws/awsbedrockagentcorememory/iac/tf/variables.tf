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
  description = "AwsBedrockAgentCoreMemory specification"
  type = object({
    region             = string
    memory_name        = string
    description        = optional(string, "")
    event_expiry_days  = optional(number, 0)
    encryption_key_arn = optional(string, "")
    execution_role_arn = optional(string, "")
    indexed_keys = optional(list(object({
      key  = string
      type = optional(string, "")
    })), [])
    kinesis_delivery = optional(object({
      data_stream_arn = string
      content_level   = optional(string, "")
    }))
    strategies = optional(list(object({
      name                = string
      type                = optional(string, "")
      description         = optional(string, "")
      namespace_templates = list(string)
      custom = optional(object({
        type = optional(string, "")
        extraction = optional(object({
          append_to_prompt = string
          model_id         = string
        }))
        consolidation = optional(object({
          append_to_prompt = string
          model_id         = string
        }))
        reflection = optional(object({
          append_to_prompt    = string
          model_id            = string
          namespace_templates = list(string)
        }))
      }))
      reflection_namespace_templates = optional(list(string), [])
    })), [])
  })
}
