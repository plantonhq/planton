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
  description = "AzureSearchService specification"
  type = object({
    region                                   = string
    resource_group                           = string
    name                                     = string
    sku                                      = string
    replica_count                            = optional(number)
    partition_count                          = optional(number)
    hosting_mode                             = optional(string, "")
    local_authentication_enabled             = optional(bool)
    authentication_failure_mode              = optional(string, "")
    customer_managed_key_enforcement_enabled = optional(bool, false)
    public_network_access_enabled            = optional(bool)
    semantic_search_sku                      = optional(string, "")
    allowed_ips                              = optional(list(string), [])
    network_rule_bypass_option               = optional(string, "")
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    tags = optional(map(string), {})
    shared_private_link_services = optional(list(object({
      name               = string
      subresource_name   = string
      target_resource_id = string
      request_message    = optional(string, "")
    })), [])
  })
}