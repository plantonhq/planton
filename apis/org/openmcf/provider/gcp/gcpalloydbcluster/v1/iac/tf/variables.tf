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
    org  = optional(string, "")
    env  = optional(string, "")
    id   = optional(string, "")
  })
}

variable "spec" {
  description = "GcpAlloydbCluster specification"
  type = object({
    project_id         = string
    cluster_name       = string
    location           = string
    network            = string
    allocated_ip_range = optional(string, "")
    database_version   = optional(string, "")
    display_name       = optional(string, "")
    initial_user = optional(object({
      password = string
      user     = optional(string, "")
    }), null)
    automated_backup_policy = optional(object({
      enabled                        = optional(bool, true)
      backup_window                  = optional(string, "")
      location                       = optional(string, "")
      quantity_based_retention_count  = optional(number, 0)
      time_based_retention_period    = optional(string, "")
      weekly_schedule = optional(object({
        days_of_week = optional(list(string), [])
        start_hour   = optional(number, 0)
      }), null)
      encryption_kms_key_name = optional(string, "")
    }), null)
    continuous_backup_config = optional(object({
      enabled              = optional(bool, true)
      recovery_window_days = optional(number, 0)
      encryption_kms_key_name = optional(string, "")
    }), null)
    kms_key_name = optional(string, "")
    maintenance_window = optional(object({
      day        = string
      start_hour = number
    }), null)
    deletion_protection = optional(bool, true)
    primary_instance = object({
      instance_id   = string
      cpu_count     = optional(number, 0)
      machine_type  = optional(string, "")
      availability_type = optional(string, "")
      database_flags    = optional(map(string), {})
      display_name      = optional(string, "")
      query_insights_config = optional(object({
        query_plans_per_minute  = optional(number, 5)
        query_string_length     = optional(number, 1024)
        record_application_tags = optional(bool, true)
        record_client_address   = optional(bool, true)
      }), null)
      require_connectors = optional(bool, false)
      ssl_mode           = optional(string, "")
    })
  })
}
