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
  description = "AzureDataProtectionBackupInstance specification"
  type = object({
    vault_id         = string
    name             = string
    region           = string
    backup_policy_id = string
    blob_storage = optional(object({
      storage_account_id              = string
      storage_account_container_names = optional(list(string), [])
    }))
    disk = optional(object({
      disk_id                      = string
      snapshot_resource_group_name = string
      snapshot_subscription_id     = optional(string)
    }))
    kubernetes_cluster = optional(object({
      kubernetes_cluster_id        = string
      snapshot_resource_group_name = string
      backup_datasource_parameters = optional(object({
        included_namespaces              = optional(list(string), [])
        excluded_namespaces              = optional(list(string), [])
        included_resource_types          = optional(list(string), [])
        excluded_resource_types          = optional(list(string), [])
        label_selectors                  = optional(list(string), [])
        cluster_scoped_resources_enabled = optional(bool)
        volume_snapshot_enabled          = optional(bool)
      }))
    }))
    mysql_flexible_server = optional(object({
      server_id = string
    }))
    postgresql_flexible_server = optional(object({
      server_id = string
    }))
    data_lake_storage = optional(object({
      storage_account_id      = string
      storage_container_names = list(string)
    }))
  })
}