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
  description = "StripeWebhookEndpoint specification"
  type = object({
    url            = string
    enabled_events = list(string)
    description    = optional(string, "")
    metadata       = optional(map(string), {})
    api_version    = optional(string, "")
    connect        = optional(bool, false)
  })
}
