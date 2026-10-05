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
  description = "AwsBedrockGuardrail specification"
  type = object({
    region                    = string
    description               = optional(string, "")
    blocked_input_messaging   = string
    blocked_outputs_messaging = string
    kms_key_arn               = optional(string, "")
    content_policy = optional(object({
      tier = optional(string, "")
      filters = list(object({
        type              = optional(string, "")
        input_strength    = optional(string, "")
        output_strength   = optional(string, "")
        input_action      = optional(string, "")
        output_action     = optional(string, "")
        input_enabled     = optional(bool)
        output_enabled    = optional(bool)
        input_modalities  = optional(list(string), [])
        output_modalities = optional(list(string), [])
      }))
    }))
    topic_policy = optional(object({
      tier = optional(string, "")
      topics = list(object({
        name       = string
        definition = string
        examples   = optional(list(string), [])
      }))
    }))
    word_policy = optional(object({
      profanity_filter = optional(object({
        input_action   = optional(string, "")
        output_action  = optional(string, "")
        input_enabled  = optional(bool)
        output_enabled = optional(bool)
      }))
      custom_words = optional(list(object({
        text           = string
        input_action   = optional(string, "")
        output_action  = optional(string, "")
        input_enabled  = optional(bool)
        output_enabled = optional(bool)
      })), [])
    }))
    sensitive_information_policy = optional(object({
      pii_entities = optional(list(object({
        type           = optional(string, "")
        action         = optional(string, "")
        input_action   = optional(string, "")
        output_action  = optional(string, "")
        input_enabled  = optional(bool)
        output_enabled = optional(bool)
      })), [])
      regexes = optional(list(object({
        name           = string
        pattern        = string
        description    = optional(string, "")
        action         = optional(string, "")
        input_action   = optional(string, "")
        output_action  = optional(string, "")
        input_enabled  = optional(bool)
        output_enabled = optional(bool)
      })), [])
    }))
    contextual_grounding_policy = optional(object({
      filters = list(object({
        type      = optional(string, "")
        threshold = optional(number, 0)
      }))
    }))
    cross_region_profile = optional(string, "")
    versions = optional(list(object({
      name           = string
      description    = optional(string, "")
      keep_on_delete = optional(bool, false)
    })), [])
  })
}