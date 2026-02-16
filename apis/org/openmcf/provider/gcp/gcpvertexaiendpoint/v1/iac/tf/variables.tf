variable "metadata" {
  description = "OpenMCF resource metadata"
  type = object({
    name    = string
    id      = optional(string, "")
    org     = optional(string, "")
    env     = optional(string, "")
    labels  = optional(map(string), {})
    tags    = optional(list(string), [])
    version = optional(string, "")
  })
}

variable "spec" {
  description = "GcpVertexAiEndpoint spec"
  type = object({
    project_id   = object({ value = string })
    location     = string
    display_name = string

    description   = optional(string, "")
    endpoint_name = optional(string, "")

    network      = optional(object({ value = string }), null)
    kms_key_name = optional(object({ value = string }), null)

    dedicated_endpoint_enabled = optional(bool, false)

    private_service_connect_config = optional(object({
      project_allowlist = optional(list(string), [])
    }), null)
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key_base64 = optional(string, "")
  })
  default = { service_account_key_base64 = "" }
}
