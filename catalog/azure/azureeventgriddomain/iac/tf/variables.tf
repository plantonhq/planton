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
  description = "AzureEventgridDomain specification"
  type = object({
    resource_group = string
    name           = string
    region         = string

    # The schema incoming events must arrive in; the platform default
    # (EventGridSchema) mirrors Azure's own.
    input_schema = optional(string, "EventGridSchema")

    # Custom-schema envelope mapping -- only meaningful with
    # input_schema = "CustomEventSchema".
    input_mapping_fields = optional(object({
      id           = optional(string, "")
      topic        = optional(string, "")
      event_time   = optional(string, "")
      event_type   = optional(string, "")
      subject      = optional(string, "")
      data_version = optional(string, "")
    }))

    input_mapping_default_values = optional(object({
      event_type   = optional(string, "")
      subject      = optional(string, "")
      data_version = optional(string, "")
    }))

    # Domain-topic lifecycle: auto-managed (both true, Azure's default)
    # or pinned (both false -- topics exist only as declared
    # AzureEventgridDomainTopic resources)
    auto_create_topic_with_first_subscription = optional(bool, true)
    auto_delete_topic_with_last_subscription  = optional(bool, true)

    # Whether the endpoint accepts publishes from the public internet
    public_network_access_enabled = optional(bool, true)

    # Whether access-key (SAS) authentication works alongside Entra ID
    local_auth_enabled = optional(bool, true)

    # IPv4 CIDR ranges allowed to publish (Azure's rule action is
    # Allow-only at v5; the module sends it explicitly)
    inbound_ip_rules = optional(list(string), [])

    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))

    tags = optional(map(string), {})
  })
}
