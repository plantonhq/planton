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
  description = "OciFunctionsApplication specification"
  type = object({
    compartment_id = object({
      value = string
    })

    subnet_ids = list(object({
      value = string
    }))

    display_name = optional(string, "")

    shape = optional(string, "")

    config = optional(map(string), {})

    network_security_group_ids = optional(list(object({
      value = string
    })), [])

    syslog_url = optional(string, "")

    image_policy_config = optional(object({
      is_policy_enabled = bool
      key_details = optional(list(object({
        kms_key_id = object({
          value = string
        })
      })), [])
    }), null)

    trace_config = optional(object({
      is_enabled = optional(bool, null)
      domain_id  = optional(string, "")
    }), null)
  })
}
