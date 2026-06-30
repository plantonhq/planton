variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud MongoDB Instance specification"
  type = object({
    region                         = string
    vswitch_id                     = string
    engine_version                 = string
    db_instance_class              = string
    db_instance_storage            = number
    account_password               = string
    db_instance_name               = optional(string, "")
    zone_id                        = optional(string, "")
    secondary_zone_id              = optional(string, "")
    hidden_zone_id                 = optional(string, "")
    replication_factor             = optional(number, 3)
    readonly_replicas              = optional(number, null)
    storage_engine                 = optional(string, "WiredTiger")
    storage_type                   = optional(string, "")
    provisioned_iops               = optional(number, null)
    instance_charge_type           = optional(string, "PostPaid")
    security_ip_list               = optional(list(string), [])
    security_group_id              = optional(string, "")
    resource_group_id              = optional(string, "")
    tags                           = optional(map(string), {})
    ssl_action                     = optional(string, "")
    tde_status                     = optional(string, "")
    encryption_key                 = optional(string, "")
    encrypted                      = optional(bool, false)
    cloud_disk_encryption_key      = optional(string, "")
    maintain_start_time            = optional(string, "")
    maintain_end_time              = optional(string, "")
    backup_time                    = optional(string, "")
    backup_period                  = optional(list(string), [])
    parameters                     = optional(map(string), {})
    db_instance_release_protection = optional(bool, false)
    period                         = optional(number, null)
    auto_renew                     = optional(bool, false)
    auto_renew_duration            = optional(number, null)
  })

  validation {
    condition     = contains(["4.0", "4.2", "4.4", "5.0", "6.0", "7.0"], var.spec.engine_version)
    error_message = "engine_version must be one of: 4.0, 4.2, 4.4, 5.0, 6.0, 7.0."
  }

  validation {
    condition     = contains(["WiredTiger", "RocksDB"], var.spec.storage_engine)
    error_message = "storage_engine must be one of: WiredTiger, RocksDB."
  }

  validation {
    condition     = contains(["PostPaid", "PrePaid"], var.spec.instance_charge_type)
    error_message = "instance_charge_type must be one of: PostPaid, PrePaid."
  }

  validation {
    condition     = contains([1, 3, 5, 7], var.spec.replication_factor)
    error_message = "replication_factor must be one of: 1, 3, 5, 7."
  }

  validation {
    condition = (
      var.spec.storage_type == "" ||
      contains(["cloud_essd1", "cloud_essd2", "cloud_essd3", "cloud_auto", "local_ssd"], var.spec.storage_type)
    )
    error_message = "storage_type must be one of: cloud_essd1, cloud_essd2, cloud_essd3, cloud_auto, local_ssd."
  }

  validation {
    condition = (
      var.spec.ssl_action == "" ||
      contains(["Open", "Close", "Update"], var.spec.ssl_action)
    )
    error_message = "ssl_action must be one of: Open, Close, Update."
  }

  validation {
    condition = (
      var.spec.tde_status == "" ||
      var.spec.tde_status == "enabled"
    )
    error_message = "tde_status must be enabled."
  }
}
