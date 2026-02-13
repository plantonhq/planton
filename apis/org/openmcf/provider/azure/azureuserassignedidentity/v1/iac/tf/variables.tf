variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Azure User-Assigned Managed Identity specification"
  type = object({
    region         = string
    resource_group = string
    name           = string
    role_assignments = optional(list(object({
      scope                = string
      role_definition_name = string
    })), [])
  })
}
