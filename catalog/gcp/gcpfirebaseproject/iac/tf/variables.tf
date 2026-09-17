variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "GcpFirebaseProject specification"
  type = object({
    project_id               = optional(string, "")
    default_storage_location = optional(string, "")
    app_check = optional(object({
      service_configs = optional(list(object({
        service_id       = string
        enforcement_mode = optional(string, "")
      })), [])
      resource_policies = optional(list(object({
        service_id       = string
        target_resource  = string
        enforcement_mode = optional(string, "")
      })), [])
    }))
    deletion_policy = optional(string, "")
  })
}
