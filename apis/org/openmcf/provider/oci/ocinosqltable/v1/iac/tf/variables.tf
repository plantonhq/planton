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
  description = "OciNosqlTable specification"
  type = object({
    compartment_id = object({
      value = string
    })

    name          = string
    ddl_statement = string

    table_limits = object({
      capacity_mode    = optional(string, "")
      max_read_units   = optional(number, 0)
      max_write_units  = optional(number, 0)
      max_storage_in_gbs = number
    })

    is_auto_reclaimable = optional(bool, false)

    indexes = optional(list(object({
      name = string
      keys = list(object({
        column_name     = string
        json_field_type = optional(string, "")
        json_path       = optional(string, "")
      }))
    })), [])
  })
}
