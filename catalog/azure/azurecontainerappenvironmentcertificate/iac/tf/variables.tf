variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AzureContainerAppEnvironmentCertificate specification"
  type = object({
    certificate_name             = string
    container_app_environment_id = string
    certificate_blob_base64      = optional(string, "")
    certificate_password         = optional(string, "")
    certificate_key_vault = optional(object({
      key_vault_secret_id = string
      identity            = optional(string, "")
    }))
    tags = optional(map(string), {})
  })
}