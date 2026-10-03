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
  description = "AzurePointToSiteVpnGateway specification"
  type = object({
    region                      = string
    resource_group              = string
    name                        = string
    virtual_hub_id              = string
    vpn_server_configuration_id = string
    connection_configurations = list(object({
      name             = string
      address_prefixes = list(string)
      route = optional(object({
        associated_route_table_id = string
        inbound_route_map_id      = optional(string, "")
        outbound_route_map_id     = optional(string, "")
        propagated_route_table = optional(object({
          route_table_ids = list(string)
          labels          = optional(list(string), [])
        }))
      }))
      internet_security_enabled = optional(bool, false)
    }))
    scale_unit                          = optional(number)
    routing_preference_internet_enabled = optional(bool, false)
    dns_servers                         = optional(list(string), [])
    tags                                = optional(map(string), {})
  })
}
