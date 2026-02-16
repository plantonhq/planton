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
  description = "GcpSpannerDatabase specification"
  type = object({
    project_id = object({
      value = string
    })
    instance = object({
      value = string
    })
    database_name            = string
    database_dialect         = optional(string, "")
    version_retention_period = optional(string, "")
    ddl                      = optional(list(string), [])
    enable_drop_protection   = optional(bool, false)
    kms_key_name = optional(object({
      value = string
    }), null)
    default_time_zone = optional(string, "")
  })
}
