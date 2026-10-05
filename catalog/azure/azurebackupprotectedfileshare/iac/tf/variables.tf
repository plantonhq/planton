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
  description = "AzureBackupProtectedFileShare specification"
  type = object({
    resource_group            = string
    recovery_vault_name       = string
    source_storage_account_id = string
    source_file_share_name    = string
    backup_policy_id          = string
  })
}