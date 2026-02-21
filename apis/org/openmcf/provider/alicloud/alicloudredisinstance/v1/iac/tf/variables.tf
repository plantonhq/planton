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
  description = "Alibaba Cloud Redis (KVStore) Instance specification"
  type = object({
    region                      = string
    vswitch_id                  = string
    instance_class              = string
    password                    = string
    engine_version              = optional(string, "7.0")
    instance_type               = optional(string, "Redis")
    db_instance_name            = optional(string, "")
    zone_id                     = optional(string, "")
    secondary_zone_id           = optional(string, "")
    payment_type                = optional(string, "PostPaid")
    security_ips                = optional(list(string), [])
    security_group_id           = optional(string, "")
    resource_group_id           = optional(string, "")
    tags                        = optional(map(string), {})
    shard_count                 = optional(number, null)
    read_only_count             = optional(number, null)
    ssl_enable                  = optional(string, "")
    tde_status                  = optional(string, "")
    encryption_key              = optional(string, "")
    vpc_auth_mode               = optional(string, "Open")
    config                      = optional(map(string), {})
    instance_release_protection = optional(bool, false)
    maintain_start_time         = optional(string, "")
    maintain_end_time           = optional(string, "")
    backup_period               = optional(list(string), [])
    backup_time                 = optional(string, "")
    private_connection_prefix   = optional(string, "")
    auto_renew                  = optional(bool, false)
    auto_renew_period           = optional(number, null)
    period                      = optional(string, "")
  })

  validation {
    condition     = contains(["Redis", "Memcache"], var.spec.instance_type)
    error_message = "instance_type must be one of: Redis, Memcache."
  }

  validation {
    condition     = contains(["2.8", "4.0", "5.0", "6.0", "7.0"], var.spec.engine_version)
    error_message = "engine_version must be one of: 2.8, 4.0, 5.0, 6.0, 7.0."
  }

  validation {
    condition     = contains(["PostPaid", "PrePaid"], var.spec.payment_type)
    error_message = "payment_type must be one of: PostPaid, PrePaid."
  }

  validation {
    condition = (
      var.spec.vpc_auth_mode == "" ||
      contains(["Open", "Close"], var.spec.vpc_auth_mode)
    )
    error_message = "vpc_auth_mode must be one of: Open, Close."
  }

  validation {
    condition = (
      var.spec.ssl_enable == "" ||
      contains(["Enable", "Disable", "Update"], var.spec.ssl_enable)
    )
    error_message = "ssl_enable must be one of: Enable, Disable, Update."
  }

  validation {
    condition = (
      var.spec.tde_status == "" ||
      var.spec.tde_status == "Enabled"
    )
    error_message = "tde_status must be Enabled."
  }
}
