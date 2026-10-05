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
  description = "AzureVirtualHub specification"
  type = object({
    region                                 = string
    resource_group                         = string
    name                                   = string
    virtual_wan_id                         = string
    address_prefix                         = string
    sku                                    = optional(string)
    hub_routing_preference                 = optional(string)
    branch_to_branch_traffic_enabled       = optional(bool, false)
    virtual_router_auto_scale_min_capacity = optional(number)
    routes = optional(list(object({
      address_prefixes    = list(string)
      next_hop_ip_address = string
    })), [])
    route_tables = optional(list(object({
      name   = string
      labels = optional(list(string), [])
      routes = optional(list(object({
        name              = string
        destinations_type = optional(string, "")
        destinations      = list(string)
        next_hop          = string
      })), [])
    })), [])
    route_maps = optional(list(object({
      name = string
      rules = optional(list(object({
        name = string
        match_criteria = optional(list(object({
          match_condition = optional(string, "")
          as_path         = optional(list(string), [])
          community       = optional(list(string), [])
          route_prefix    = optional(list(string), [])
        })), [])
        actions = optional(list(object({
          type = optional(string, "")
          parameters = optional(list(object({
            as_path      = optional(list(string), [])
            community    = optional(list(string), [])
            route_prefix = optional(list(string), [])
          })), [])
        })), [])
        next_step_if_matched = optional(string)
      })), [])
    })), [])
    bgp_connections = optional(list(object({
      name                          = string
      peer_asn                      = optional(number, 0)
      peer_ip                       = string
      virtual_network_connection_id = optional(string, "")
    })), [])
    routing_intent = optional(object({
      name = string
      routing_policies = list(object({
        name         = string
        destinations = list(string)
        next_hop     = string
      }))
    }))
    tags = optional(map(string), {})
  })
}