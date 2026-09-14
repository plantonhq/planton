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
  description = "AzureVpnSite specification"
  type = object({
    region         = string
    resource_group = string
    name           = string
    virtual_wan_id = string
    address_cidrs  = optional(list(string), [])
    device_vendor  = optional(string, "")
    device_model   = optional(string, "")
    links = optional(list(object({
      name          = string
      provider_name = optional(string, "")
      speed_in_mbps = optional(number, 0)
      ip_address    = optional(string, "")
      fqdn          = optional(string, "")
      bgp = optional(object({
        asn             = optional(number, 0)
        peering_address = string
      }))
    })), [])
    o365_policy = optional(object({
      traffic_category = optional(object({
        allow_endpoint_enabled    = optional(bool)
        default_endpoint_enabled  = optional(bool)
        optimize_endpoint_enabled = optional(bool)
      }))
    }))
    tags = optional(map(string), {})
  })
}
