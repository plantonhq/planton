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
  description = "AzureDataProtectionBackupVault specification"
  type = object({
    region                       = string
    resource_group               = string
    name                         = string
    datastore_type               = string
    redundancy                   = string
    cross_region_restore_enabled = optional(bool, false)
    retention_duration_in_days   = optional(number)
    soft_delete                  = optional(string)
    immutability                 = optional(string)
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    encryption = optional(object({
      key_id = string
    }))
    tags = optional(map(string), {})
  })
}
