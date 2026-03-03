variable "spec" {
  description = "GcpKmsKey spec"
  type = object({
    key_ring_id = object({ value = string })
    key_name    = string
    purpose     = optional(string, "")
    rotation_period = optional(string, "")
    destroy_scheduled_duration = optional(string, "")
    version_template = optional(object({
      algorithm        = string
      protection_level = optional(string, "")
    }), null)
    skip_initial_version_creation = optional(bool, false)
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key = optional(string, "")
  })
  default = { service_account_key = "" }
}
