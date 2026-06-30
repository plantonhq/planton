variable "metadata" {
  description = "Resource metadata."
  type = object({
    name   = string
    id     = optional(string, "")
    org    = optional(string, "")
    env    = optional(string, "")
    labels = optional(map(string), {})
  })
}

variable "spec" {
  description = "OciDynamicRoutingGateway specification."
  type = object({
    compartment_id = object({
      value = string
    })
    display_name = optional(string, "")

    attachments = optional(list(object({
      display_name = string
      network_details = object({
        type = string
        id = object({
          value = string
        })
        route_table_id = optional(string, "")
        vcn_route_type = optional(string, "")
      })
      drg_route_table_name                = optional(string, "")
      export_drg_route_distribution_name = optional(string, "")
    })), [])

    route_tables = optional(list(object({
      display_name                        = string
      import_drg_route_distribution_name = optional(string, "")
      is_ecmp_enabled                     = optional(bool, false)
      static_route_rules = optional(list(object({
        destination              = string
        next_hop_attachment_name = string
      })), [])
    })), [])

    route_distributions = optional(list(object({
      display_name      = string
      distribution_type = string
      statements = optional(list(object({
        priority = number
        match_criteria = object({
          match_type          = string
          attachment_type     = optional(string, "")
          drg_attachment_name = optional(string, "")
        })
      })), [])
    })), [])
  })
}
