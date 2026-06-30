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
  description = "GcpSpannerInstance specification"
  type = object({
    project_id = object({
      value = string
    })
    instance_name = string
    config        = string
    display_name  = string
    num_nodes     = optional(number, 0)
    processing_units = optional(number, 0)
    autoscaling_config = optional(object({
      autoscaling_limits = object({
        min_nodes            = optional(number, 0)
        max_nodes            = optional(number, 0)
        min_processing_units = optional(number, 0)
        max_processing_units = optional(number, 0)
      })
      autoscaling_targets = optional(object({
        high_priority_cpu_utilization_percent = optional(number, 0)
        storage_utilization_percent           = optional(number, 0)
      }), null)
    }), null)
    instance_type                 = optional(string, "")
    edition                       = optional(string, "")
    default_backup_schedule_type  = optional(string, "")
    force_destroy                 = optional(bool, false)
  })
}
