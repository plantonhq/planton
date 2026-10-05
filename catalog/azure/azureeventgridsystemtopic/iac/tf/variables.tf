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
  description = "AzureEventgridSystemTopic specification"
  type = object({
    resource_group = string
    name           = string

    # Must match the source resource's region; "Global" for global
    # sources (subscriptions, resource groups).
    region = string

    # The ARM ID of the resource whose events this topic surfaces.
    source_resource_id = string

    # Which Azure service's event stream the source emits (validated
    # by Azure against its live topic-type catalog).
    topic_type = string

    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))

    tags = optional(map(string), {})
  })
}
