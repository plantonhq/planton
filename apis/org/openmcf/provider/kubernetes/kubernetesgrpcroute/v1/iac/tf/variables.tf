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
  description = "Specification for KubernetesGrpcRoute"
  type = object({
    # Namespace the GRPCRoute is created in (resolved foreign key).
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

    # Hostnames matched against the authority (Host) pseudo-header.
    hostnames = optional(list(string))

    # Routing rules. At least one is required.
    rules = list(object({
      name = optional(string)

      matches = optional(list(object({
        method = optional(object({
          type    = optional(string)
          service = optional(string)
          method  = optional(string)
        }))
        headers = optional(list(object({
          type  = optional(string)
          name  = string
          value = string
        })))
      })))

      filters = optional(list(object({
        type = string
        request_header_modifier = optional(object({
          set    = optional(list(object({ name = string, value = string })))
          add    = optional(list(object({ name = string, value = string })))
          remove = optional(list(string))
        }))
        response_header_modifier = optional(object({
          set    = optional(list(object({ name = string, value = string })))
          add    = optional(list(object({ name = string, value = string })))
          remove = optional(list(string))
        }))
        request_mirror = optional(object({
          backend_ref = object({
            group     = optional(string)
            kind      = optional(string)
            name      = string
            namespace = optional(string)
            port      = optional(number)
          })
          percent = optional(number)
          fraction = optional(object({
            numerator   = number
            denominator = optional(number)
          }))
        }))
        extension_ref = optional(object({
          group = string
          kind  = string
          name  = string
        }))
      })))

      backend_refs = optional(list(object({
        group     = optional(string)
        kind      = optional(string)
        name      = string
        namespace = optional(string)
        port      = optional(number)
        weight    = optional(number)
        filters = optional(list(object({
          type = string
          request_header_modifier = optional(object({
            set    = optional(list(object({ name = string, value = string })))
            add    = optional(list(object({ name = string, value = string })))
            remove = optional(list(string))
          }))
          response_header_modifier = optional(object({
            set    = optional(list(object({ name = string, value = string })))
            add    = optional(list(object({ name = string, value = string })))
            remove = optional(list(string))
          }))
          request_mirror = optional(object({
            backend_ref = object({
              group     = optional(string)
              kind      = optional(string)
              name      = string
              namespace = optional(string)
              port      = optional(number)
            })
            percent = optional(number)
            fraction = optional(object({
              numerator   = number
              denominator = optional(number)
            }))
          }))
          extension_ref = optional(object({
            group = string
            kind  = string
            name  = string
          }))
        })))
      })))
    }))
  })
}
