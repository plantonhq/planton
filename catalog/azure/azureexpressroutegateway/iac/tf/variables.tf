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
  description = "AzureExpressRouteGateway specification"
  type = object({
    region                        = string
    resource_group                = string
    name                          = string
    virtual_hub_id                = string
    scale_units                   = number
    allow_non_virtual_wan_traffic = optional(bool, false)
    connections = optional(list(object({
      name                                 = string
      express_route_circuit_peering_id     = string
      authorization_key                    = optional(string, "")
      internet_security_enabled            = optional(bool, false)
      express_route_gateway_bypass_enabled = optional(bool, false)
      routing = optional(object({
        associated_route_table_id = optional(string, "")
        inbound_route_map_id      = optional(string, "")
        outbound_route_map_id     = optional(string, "")
        propagated_route_table = optional(object({
          labels          = optional(list(string), [])
          route_table_ids = optional(list(string), [])
        }))
      }))
      routing_weight = optional(number, 0)
    })), [])
    tags = optional(map(string), {})
  })
}