variable "spec" {
  description = "GcpGlobalAddress spec"
  type = object({
    project_id    = object({ value = string })
    address_name  = string
    address       = optional(string, "")
    address_type  = optional(string, "EXTERNAL")
    description   = optional(string, "")
    ip_version    = optional(string, "IPV4")
    network       = optional(object({ value = string }), null)
    prefix_length = optional(number, null)
    purpose       = optional(string, "")
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key = optional(string, "")
  })
  default = { service_account_key = "" }
}

variable "labels" {
  description = "Labels to apply to the global address"
  type        = map(string)
  default     = {}
}
