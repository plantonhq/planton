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
  description = "AzurePrivateLinkService specification"
  type = object({
    region         = string
    resource_group = string
    name           = string
    nat_ip_configurations = list(object({
      name                       = string
      subnet_id                  = string
      private_ip_address         = optional(string, "")
      private_ip_address_version = optional(string)
      primary                    = optional(bool, false)
    }))
    load_balancer_frontend_ip_configuration_ids = optional(list(string), [])
    destination_ip_address                      = optional(string, "")
    proxy_protocol_enabled                      = optional(bool, false)
    auto_approval_subscription_ids              = optional(list(string), [])
    visibility_subscription_ids                 = optional(list(string), [])
    fqdns                                       = optional(list(string), [])
    tags                                        = optional(map(string), {})
  })
}