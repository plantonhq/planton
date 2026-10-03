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
  description = "AzureContainerAppCustomDomain specification"
  type = object({
    domain_name                              = string
    container_app_id                         = string
    container_app_environment_certificate_id = optional(string, "")
    certificate_binding_type                 = optional(string, "")
  })
}