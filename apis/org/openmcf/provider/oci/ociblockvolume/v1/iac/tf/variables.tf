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
  description = "OciBlockVolume specification"
  type = object({
    compartment_id = object({
      value = string
    })

    availability_domain = string

    display_name = optional(string, "")

    size_in_gbs = number

    vpus_per_gb = optional(number, null)

    kms_key_id = optional(object({
      value = string
    }), null)

    is_reservations_enabled = optional(bool, false)

    autotune_policies = optional(list(object({
      autotune_type  = string
      max_vpus_per_gb = optional(number, 0)
    })), [])

    block_volume_replicas = optional(list(object({
      availability_domain = string
      display_name        = optional(string, "")
      xrr_kms_key_id = optional(object({
        value = string
      }), null)
    })), [])

    backup_policy_id = optional(object({
      value = string
    }), null)

    xrc_kms_key_id = optional(object({
      value = string
    }), null)
  })
}
