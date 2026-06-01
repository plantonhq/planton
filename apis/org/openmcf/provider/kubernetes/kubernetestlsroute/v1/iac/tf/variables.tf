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
  description = "Specification for KubernetesTlsRoute"
  type = object({
    # Namespace the TLSRoute is created in (resolved foreign key).
    namespace = string

    # Parent resources (usually Gateways) this route attaches to.
    parent_refs = optional(list(object({
      group        = optional(string)
      kind         = optional(string)
      namespace    = optional(string)
      name         = string
      section_name = optional(string)
      port         = optional(number)
    })))

    # SNI hostnames matched against the TLS ClientHello. At least one required.
    hostnames = list(string)

    # Routing rules. Upstream permits exactly one rule for a TLSRoute.
    rules = list(object({
      name = optional(string)
      backend_refs = list(object({
        group     = optional(string)
        kind      = optional(string)
        name      = string
        namespace = optional(string)
        port      = optional(number)
        weight    = optional(number)
      }))
    }))
  })
}
