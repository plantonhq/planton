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
  description = "AzureMongoCluster specification"
  type = object({
    resource_group             = string
    name                       = string
    region                     = string
    create_mode                = optional(string)
    administrator_username     = optional(string, "")
    administrator_password     = optional(string, "")
    version                    = optional(string)
    compute_tier               = optional(string)
    storage_size_in_gb         = optional(number)
    storage_type               = optional(string)
    shard_count                = optional(number)
    high_availability_mode     = optional(string)
    authentication_methods     = optional(list(string), [])
    user_assigned_identity_ids = optional(list(string), [])
    customer_managed_key = optional(object({
      key_vault_key_id          = string
      user_assigned_identity_id = string
    }))
    preview_features = optional(list(string), [])
    source_server_id = optional(string, "")
    source_location  = optional(string, "")
    restore = optional(object({
      point_in_time_utc = string
      source_id         = string
    }))
    data_api_mode_enabled         = optional(bool)
    public_network_access_enabled = optional(bool)
    firewall_rules = optional(list(object({
      name             = string
      start_ip_address = string
      end_ip_address   = string
    })), [])
    tags = optional(map(string), {})
  })
}