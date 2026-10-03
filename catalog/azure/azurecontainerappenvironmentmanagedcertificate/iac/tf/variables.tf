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
  description = "AzureContainerAppEnvironmentManagedCertificate specification"
  type = object({
    certificate_name             = string
    container_app_environment_id = string
    subject_name                 = string
    domain_control_validation    = optional(string, "")
    tags                         = optional(map(string), {})
  })
}