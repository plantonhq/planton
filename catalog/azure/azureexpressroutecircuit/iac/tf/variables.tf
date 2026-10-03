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
  description = "AzureExpressRouteCircuit specification"
  type = object({
    region                   = string
    resource_group           = string
    name                     = string
    sku_tier                 = optional(string, "")
    sku_family               = optional(string, "")
    service_provider_name    = optional(string, "")
    peering_location         = optional(string, "")
    bandwidth_in_mbps        = optional(number, 0)
    express_route_port_id    = optional(string, "")
    bandwidth_in_gbps        = optional(number, 0)
    rate_limiting_enabled    = optional(bool, false)
    allow_classic_operations = optional(bool, false)
    authorization_key        = optional(string, "")
    authorizations = optional(list(object({
      name = string
    })), [])
    tags = optional(map(string), {})
  })
}