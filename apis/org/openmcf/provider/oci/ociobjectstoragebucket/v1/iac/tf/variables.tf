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
  description = "OciObjectStorageBucket specification"
  type = object({
    compartment_id = object({
      value = string
    })

    namespace = string
    name      = string

    access_type  = optional(string, "")
    storage_tier = optional(string, "")
    versioning   = optional(string, "")
    auto_tiering = optional(string, "")

    object_events_enabled = optional(bool, false)

    kms_key_id = optional(object({
      value = string
    }), null)

    metadata = optional(map(string), {})

    retention_rules = optional(list(object({
      display_name = string
      duration = optional(object({
        time_amount = number
        time_unit   = string
      }), null)
      time_rule_locked = optional(string, "")
    })), [])

    lifecycle_rules = optional(list(object({
      name        = string
      action      = string
      is_enabled  = bool
      time_amount = number
      time_unit   = string
      target      = optional(string, "objects")
      object_name_filter = optional(object({
        inclusion_patterns = optional(list(string), [])
        inclusion_prefixes = optional(list(string), [])
        exclusion_patterns = optional(list(string), [])
      }), null)
    })), [])

    replication_policies = optional(list(object({
      name                    = string
      destination_bucket_name = string
      destination_region_name = string
    })), [])
  })
}
