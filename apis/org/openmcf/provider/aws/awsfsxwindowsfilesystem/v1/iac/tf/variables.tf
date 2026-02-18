variable "provider_config" {
  description = "AWS provider configuration"
  type = object({
    region            = string
    access_key_id     = optional(string)
    secret_access_key = optional(string)
    session_token     = optional(string)
  })
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    org  = string
    env  = string
    name = string
    id   = string
  })
}

variable "spec" {
  description = "AwsFsxWindowsFileSystem spec"
  type = object({
    # The AWS region where the resource will be created.
    region = string
    deployment_type    = optional(string, "SINGLE_AZ_2")
    storage_capacity_gib = number
    storage_type       = optional(string, "SSD")
    throughput_capacity = number
    subnet_ids         = list(string)
    preferred_subnet_id = optional(string)
    security_group_ids = optional(list(string), [])
    kms_key_id         = optional(string)
    active_directory_id = optional(string)
    self_managed_active_directory = optional(object({
      domain_name                            = string
      dns_ips                                = list(string)
      username                               = optional(string)
      password                               = optional(string)
      domain_join_service_account_secret_arn  = optional(string)
      file_system_administrators_group       = optional(string, "Domain Admins")
      organizational_unit_distinguished_name = optional(string)
    }))
    aliases = optional(list(string), [])
    audit_log_configuration = optional(object({
      file_access_audit_log_level       = optional(string, "DISABLED")
      file_share_access_audit_log_level = optional(string, "DISABLED")
      audit_log_destination             = optional(string)
    }))
    disk_iops_configuration = optional(object({
      mode = optional(string, "AUTOMATIC")
      iops = optional(number)
    }))
    automatic_backup_retention_days    = optional(number, 7)
    daily_automatic_backup_start_time  = optional(string)
    copy_tags_to_backups               = optional(bool, false)
    skip_final_backup                  = optional(bool, true)
    weekly_maintenance_start_time      = optional(string)
  })
}
