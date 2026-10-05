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
  description = "AzureBackupProtectedVm specification"
  type = object({
    resource_group      = string
    recovery_vault_name = string
    source_vm_id        = string
    backup_policy_id    = string
    exclude_disk_luns   = optional(list(number), [])
    include_disk_luns   = optional(list(number), [])
    protection_state    = optional(string, "")
  })
}