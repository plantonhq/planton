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
  description = "AzureVirtualHubConnection specification"
  type = object({
    name                      = string
    virtual_hub_id            = string
    remote_virtual_network_id = string
    internet_security_enabled = optional(bool, false)
    routing = optional(object({
      associated_route_table_id = optional(string, "")
      inbound_route_map_id      = optional(string, "")
      outbound_route_map_id     = optional(string, "")
      propagated_route_table = optional(object({
        labels          = optional(list(string), [])
        route_table_ids = optional(list(string), [])
      }))
      static_vnet_routes = optional(list(object({
        name                = string
        address_prefixes    = list(string)
        next_hop_ip_address = string
      })), [])
      static_vnet_local_route_override_criteria   = optional(string)
      static_vnet_propagate_static_routes_enabled = optional(bool)
    }))
  })
}