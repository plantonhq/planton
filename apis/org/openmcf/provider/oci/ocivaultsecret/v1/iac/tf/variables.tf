variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "OciVaultSecret specification"
  type = object({
    compartment_id = object({
      value = string
    })

    secret_name = string

    vault_id = object({
      value = string
    })

    key_id = object({
      value = string
    })

    description = optional(string, "")

    secret_content = optional(object({
      content = optional(string, "")
      name    = optional(string, "")
      stage   = optional(string, "")
    }), null)

    enable_auto_generation = optional(bool, false)

    secret_generation_context = optional(object({
      generation_type     = string
      generation_template = string
      passphrase_length   = optional(number, 0)
      secret_template     = optional(string, "")
    }), null)

    secret_rules = optional(list(object({
      rule_type                                        = string
      is_secret_content_retrieval_blocked_on_expiry     = optional(bool, false)
      secret_version_expiry_interval                    = optional(string, "")
      time_of_absolute_expiry                           = optional(string, "")
      is_enforced_on_deleted_secret_versions            = optional(bool, false)
    })), [])

    rotation_config = optional(object({
      is_scheduled_rotation_enabled = optional(bool, false)
      rotation_interval             = optional(string, "")
      target_system_details = object({
        target_system_type = string
        adb_id = optional(object({
          value = string
        }), null)
        function_id = optional(object({
          value = string
        }), null)
      })
    }), null)

    secret_metadata = optional(map(string), {})
  })
}
