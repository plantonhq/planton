variable "spec" {
  description = "GcpKmsKeyRing spec"
  type = object({
    project_id    = object({ value = string })
    key_ring_name = string
    location      = string
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key_base64 = optional(string, "")
  })
  default = { service_account_key_base64 = "" }
}
