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
  description = "StripeEventDestination specification"
  type = object({
    name                 = string
    description          = optional(string, "")
    event_payload        = optional(string, "")
    enabled_events       = list(string)
    events_from          = optional(list(string), [])
    snapshot_api_version = optional(string, "")
    metadata             = optional(map(string), {})
    webhook_endpoint = optional(object({
      url = string
    }))
    amazon_eventbridge = optional(object({
      aws_account_id = optional(string, "")
      aws_region     = optional(string, "")
    }))
    azure_event_grid = optional(object({
      azure_subscription_id     = optional(string, "")
      azure_resource_group_name = string
      azure_region              = string
    }))
  })
}
