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
  description = "AzureVpnGateway specification"
  type = object({
    region                                = string
    resource_group                        = string
    name                                  = string
    virtual_hub_id                        = string
    routing_preference                    = optional(string)
    scale_unit                            = optional(number)
    bgp_route_translation_for_nat_enabled = optional(bool, false)
    bgp_settings = optional(object({
      asn         = optional(number, 0)
      peer_weight = optional(number, 0)
      instance_0_bgp_peering_address = optional(object({
        custom_ips = list(string)
      }))
      instance_1_bgp_peering_address = optional(object({
        custom_ips = list(string)
      }))
    }))
    nat_rules = optional(list(object({
      name = string
      mode = optional(string, "")
      type = optional(string, "")
      external_mappings = list(object({
        address_space = string
        port_range    = optional(string, "")
      }))
      internal_mappings = list(object({
        address_space = string
        port_range    = optional(string, "")
      }))
      ip_configuration = optional(string, "")
    })), [])
    tags = optional(map(string), {})
  })
}
