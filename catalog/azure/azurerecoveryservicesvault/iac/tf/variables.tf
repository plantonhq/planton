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
  description = "AzureRecoveryServicesVault specification"
  type = object({
    region                        = string
    resource_group                = string
    name                          = string
    sku                           = optional(string)
    storage_mode_type             = optional(string)
    cross_region_restore_enabled  = optional(bool, false)
    public_network_access_enabled = optional(bool)
    immutability                  = optional(string, "")
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    encryption = optional(object({
      key_id                            = string
      infrastructure_encryption_enabled = optional(bool, false)
      use_system_assigned_identity      = optional(bool)
      user_assigned_identity_id         = optional(string, "")
    }))
    monitoring = optional(object({
      alerts_for_all_job_failures_enabled            = optional(bool)
      alerts_for_all_failover_issues_enabled         = optional(bool)
      alerts_for_all_replication_issues_enabled      = optional(bool)
      alerts_for_critical_operation_failures_enabled = optional(bool)
      email_notifications_for_site_recovery_enabled  = optional(bool)
    }))
    resource_guard_id                  = optional(string, "")
    classic_vmware_replication_enabled = optional(bool, false)
    tags                               = optional(map(string), {})
  })
}