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
  description = "OciKmsKey specification"
  type = object({
    compartment_id = object({
      value = string
    })

    display_name = optional(string, "")

    management_endpoint = object({
      value = string
    })

    key_shape = object({
      algorithm = string
      length    = number
      curve_id  = optional(string, "")
    })

    protection_mode = optional(string, "")

    is_auto_rotation_enabled = optional(bool, false)

    auto_key_rotation_details = optional(object({
      rotation_interval_in_days = optional(number, 0)
      time_of_schedule_start    = optional(string, "")
    }), null)

    external_key_reference = optional(object({
      external_key_id = string
    }), null)
  })
}
