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
  description = "GcpFirestoreDatabase specification"
  type = object({
    project_id = object({
      value = string
    })
    location_id   = string
    database_name = string
    type          = string

    concurrency_mode                  = optional(string, "")
    point_in_time_recovery_enablement = optional(string, "")
    delete_protection_state           = optional(string, "DELETE_PROTECTION_DISABLED")
    database_edition                  = optional(string, "")
    kms_key_name = optional(object({
      value = string
    }), null)
  })
}
