variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key_base64 = optional(string, "")
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
  description = "GcpMemorystoreInstance specification"
  type = object({
    project_id = object({
      value = string
    })
    instance_name = string
    location      = string
    shard_count   = number
    mode          = optional(string, "")
    node_type     = optional(string, "")
    engine_version = optional(string, "")
    engine_configs = optional(map(string), {})
    replica_count  = optional(number, 0)
    psc_auto_connections = optional(list(object({
      network    = object({ value = string })
      project_id = object({ value = string })
    })), [])
    authorization_mode       = optional(string, "")
    transit_encryption_mode  = optional(string, "")
    kms_key                  = optional(object({ value = string }), null)
    persistence_config = optional(object({
      mode = string
      rdb_config = optional(object({
        rdb_snapshot_period     = string
        rdb_snapshot_start_time = optional(string, "")
      }), null)
      aof_config = optional(object({
        append_fsync = string
      }), null)
    }), null)
    zone_distribution_config = optional(object({
      mode = string
      zone = optional(string, "")
    }), null)
    maintenance_policy = optional(object({
      weekly_maintenance_window = object({
        day  = string
        hour = number
      })
    }), null)
    automated_backup_config = optional(object({
      start_hour = number
      retention  = string
    }), null)
    deletion_protection_enabled = optional(bool, false)
  })
}
