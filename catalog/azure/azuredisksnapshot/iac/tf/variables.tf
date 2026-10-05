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
  description = "AzureDiskSnapshot specification"
  type = object({
    resource_group        = string
    name                  = string
    region                = string
    create_option         = string
    source_resource_id    = optional(string, "")
    source_uri            = optional(string, "")
    storage_account_id    = optional(string, "")
    incremental_enabled   = optional(bool, false)
    disk_size_gb          = optional(number)
    network_access_policy = optional(string, "")
    disk_access_id        = optional(string, "")
    public_network_access_enabled = optional(bool)
    encryption_settings = optional(object({
      disk_encryption_key = object({
        secret_url      = string
        source_vault_id = string
      })
      key_encryption_key = optional(object({
        key_url         = string
        source_vault_id = string
      }))
    }))
    tags = optional(map(string), {})
  })
}
