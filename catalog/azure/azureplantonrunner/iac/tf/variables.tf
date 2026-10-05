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
  description = "AzurePlantonRunner specification"
  type = object({
    resource_group               = string
    container_app_environment_id = string
    token                        = string
    control_plane_endpoint       = optional(string, "")
    runner_version               = optional(string)
    image_repository             = optional(string)
    cpu                          = optional(number)
    memory                       = optional(string)
  })
}