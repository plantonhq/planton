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
  description = "AzureNetworkWatcherFlowLog specification"
  type = object({
    region             = string
    name               = string
    target_resource_id = string
    storage_account_id = string
    enabled            = optional(bool)
    version            = optional(number)
    retention_policy = object({
      enabled = optional(bool, false)
      days    = optional(number, 0)
    })
    traffic_analytics = optional(object({
      enabled               = optional(bool)
      workspace_id          = string
      workspace_region      = string
      workspace_resource_id = string
      interval_in_minutes   = optional(number)
    }))
    network_watcher_name           = optional(string, "")
    network_watcher_resource_group = optional(string, "")
    tags                           = optional(map(string), {})
  })
}
