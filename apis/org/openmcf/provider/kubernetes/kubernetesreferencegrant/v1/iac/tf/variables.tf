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
  description = "Specification for KubernetesReferenceGrant"
  type = object({
    # Namespace the ReferenceGrant is created in (resolved foreign key). This is
    # the "to" namespace -- the one whose resources the grant authorizes inbound
    # references to.
    namespace = string

    # Trusted sources: the namespaces and kinds permitted to reference into this
    # grant's namespace. At least one required (max 16). group may be "" (core).
    from = list(object({
      group     = optional(string)
      kind      = string
      namespace = string
    }))

    # Referenceable targets: the kinds (and optionally a specific name) in this
    # grant's namespace that may be referenced. At least one required (max 16).
    to = list(object({
      group = optional(string)
      kind  = string
      name  = optional(string)
    }))
  })
}
