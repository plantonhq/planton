variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "OpenStackRoleAssignmentSpec defines the configuration for a role assignment"
  type = object({
    # (Required) UUID of the role to assign.
    role_id = string

    # (Optional) Project UUID to scope the assignment. XOR with domain_id.
    # Middleware resolves StringValueOrRef to a literal value before TF runs.
    project_id = optional(string, "")

    # (Optional) Domain UUID to scope the assignment. XOR with project_id.
    domain_id = optional(string, "")

    # (Optional) User UUID to assign the role to. XOR with group_id.
    user_id = optional(string, "")

    # (Optional) Group UUID to assign the role to. XOR with user_id.
    group_id = optional(string, "")

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
