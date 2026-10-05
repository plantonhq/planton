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
  description = "AzureTrafficManagerEndpoint specification"
  type = object({
    profile_id = string
    name       = string
    azure = optional(object({
      target_resource_id   = string
      always_serve_enabled = optional(bool)
    }))
    external = optional(object({
      target               = string
      endpoint_location    = optional(string, "")
      always_serve_enabled = optional(bool)
    }))
    nested = optional(object({
      target_profile_id                     = string
      minimum_child_endpoints               = number
      minimum_required_child_endpoints_ipv4 = optional(number)
      minimum_required_child_endpoints_ipv6 = optional(number)
      endpoint_location                     = optional(string, "")
    }))
    weight       = optional(number)
    priority     = optional(number)
    enabled      = optional(bool)
    geo_mappings = optional(list(string), [])
    subnets = optional(list(object({
      first = string
      last  = optional(string, "")
      scope = optional(number)
    })), [])
    custom_headers = optional(list(object({
      name  = string
      value = string
    })), [])
  })
}
