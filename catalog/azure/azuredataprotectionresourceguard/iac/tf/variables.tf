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
  description = "AzureDataProtectionResourceGuard specification"
  type = object({
    region                                  = string
    resource_group                          = string
    name                                    = string
    vault_critical_operation_exclusion_list = optional(list(string), [])
    tags                                    = optional(map(string), {})
  })
}
