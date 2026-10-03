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
  description = "AzureCognitiveDeployment specification"
  type = object({
    cognitive_account_id = string
    name                 = string
    model = object({
      format  = string
      name    = string
      version = optional(string, "")
    })
    sku = object({
      name     = string
      tier     = optional(string, "")
      size     = optional(string, "")
      family   = optional(string, "")
      capacity = optional(number)
    })
    rai_policy_name            = optional(string, "")
    version_upgrade_option     = optional(string, "")
    dynamic_throttling_enabled = optional(bool, false)
  })
}