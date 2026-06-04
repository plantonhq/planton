# Auth0Role Variables
# Maps to the Auth0RoleSpec protobuf message

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
  description = "Auth0Role specification"
  type = object({
    # name is the human-readable name of the role.
    # Optional — defaults to metadata.name when omitted.
    name = optional(string)

    # description is a human-readable explanation of what the role grants.
    description = optional(string)

    # permissions is the authoritative set of API permissions (scopes) granted
    # to the role. Each entry requires the scope name and the identifier
    # (audience) of the resource server that owns it.
    permissions = optional(list(object({
      name                       = string
      resource_server_identifier = string
    })), [])
  })
}
