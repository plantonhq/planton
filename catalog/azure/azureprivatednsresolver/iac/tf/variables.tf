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
  description = "AzurePrivateDnsResolver specification"
  type = object({
    region             = string
    resource_group     = string
    name               = string
    virtual_network_id = string
    inbound_endpoints = optional(list(object({
      name                         = string
      subnet_id                    = string
      private_ip_allocation_method = optional(string, "")
      private_ip_address           = optional(string, "")
    })), [])
    outbound_endpoints = optional(list(object({
      name      = string
      subnet_id = string
    })), [])
    tags = optional(map(string), {})
  })
}
