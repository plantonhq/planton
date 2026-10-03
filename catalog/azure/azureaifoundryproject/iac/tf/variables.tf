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
  description = "AzureAiFoundryProject specification"
  type = object({
    region             = string
    name               = string
    ai_services_hub_id = string
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    primary_user_assigned_identity = optional(string, "")
    high_business_impact_enabled   = optional(bool, false)
    description                    = optional(string, "")
    friendly_name                  = optional(string, "")
    tags                           = optional(map(string), {})
  })
}