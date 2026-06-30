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
  description = "OpenStackProjectSpec defines the configuration for a Keystone project"
  type = object({
    # (Optional) Human-readable description of the project.
    description = optional(string, "")

    # (Optional) Keystone domain UUID. ForceNew.
    domain_id = optional(string, "")

    # (Optional) Whether the project is active. Default: true.
    enabled = optional(bool, true)

    # (Optional) Parent project UUID for nested hierarchies. ForceNew.
    parent_id = optional(string, "")

    # (Optional) Tags for filtering and organization.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
