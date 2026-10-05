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
  description = "AzureEventgridNamespace specification"
  type = object({
    resource_group                = string
    name                          = string
    region                        = string
    capacity                      = optional(number)
    public_network_access_enabled = optional(bool)
    inbound_ip_rules              = optional(list(string), [])
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    topic_spaces_configuration = optional(object({
      alternative_authentication_name_sources         = optional(list(string), [])
      maximum_client_sessions_per_authentication_name = optional(number)
      maximum_session_expiry_in_hours                 = optional(number)
      route_topic_id                                  = optional(string, "")
      dynamic_routing_enrichments = optional(list(object({
        key   = string
        value = string
      })), [])
      static_routing_enrichments = optional(list(object({
        key   = string
        value = string
      })), [])
    }))
    tags = optional(map(string), {})
  })
}