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
  description = "AzureAiFoundry specification"
  type = object({
    region             = string
    resource_group     = string
    name               = string
    key_vault_id       = string
    storage_account_id = string
    identity = object({
      type         = string
      identity_ids = optional(list(string), [])
    })
    application_insights_id        = optional(string, "")
    container_registry_id          = optional(string, "")
    primary_user_assigned_identity = optional(string, "")
    public_network_access_enabled  = optional(bool)
    encryption = optional(object({
      key_vault_id              = string
      key_id                    = string
      user_assigned_identity_id = optional(string, "")
    }))
    managed_network = optional(object({
      isolation_mode = optional(string, "")
    }))
    high_business_impact_enabled = optional(bool, false)
    description                  = optional(string, "")
    friendly_name                = optional(string, "")
    tags                         = optional(map(string), {})
  })
}