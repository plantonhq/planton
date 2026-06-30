variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key = optional(string, "")
  })
  default = {}
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    name = string
    id   = optional(string, "")
    org  = optional(string, "")
    env  = optional(string, "")
  })
}

variable "spec" {
  description = "GcpRedisInstance specification"
  type = object({
    project_id = object({
      value = string
    })
    instance_name           = string
    region                  = string
    tier                    = string
    memory_size_gb          = number
    redis_version           = optional(string, "")
    display_name            = optional(string, "")
    location_id             = optional(string, "")
    authorized_network      = optional(object({ value = string }), null)
    connect_mode            = optional(string, "")
    reserved_ip_range       = optional(string, "")
    auth_enabled            = optional(bool, false)
    transit_encryption_mode = optional(string, "")
    redis_configs           = optional(map(string), {})
    maintenance_window = optional(object({
      day  = string
      hour = number
    }), null)
    read_replicas_mode = optional(string, "")
    replica_count      = optional(number, 0)
    persistence_config = optional(object({
      persistence_mode    = string
      rdb_snapshot_period = optional(string, "")
    }), null)
    customer_managed_key = optional(object({ value = string }), null)
    deletion_protection  = optional(bool, false)
  })
}
