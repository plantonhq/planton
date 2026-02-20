variable "metadata" {
  type = object({
    name   = string
    id     = optional(string, "")
    org    = optional(string, "")
    env    = optional(string, "")
    labels = optional(map(string), {})
  })
}

variable "spec" {
  type = object({
    compartment_id = object({
      value = string
    })
    zone_type = string
    scope     = optional(string, "scope_unspecified")
    view_id = optional(object({
      value = string
    }))
    is_dnssec_enabled = optional(bool)

    external_masters = optional(list(object({
      address     = string
      port        = optional(number)
      tsig_key_id = optional(string, "")
    })), [])

    external_downstreams = optional(list(object({
      address     = string
      port        = optional(number)
      tsig_key_id = optional(string, "")
    })), [])
  })
}
