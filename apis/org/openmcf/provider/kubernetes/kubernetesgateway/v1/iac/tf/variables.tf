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
  description = "Specification for KubernetesGateway"
  type = object({
    # Namespace the Gateway is created in (resolved foreign key).
    namespace = string

    # Name of the GatewayClass this Gateway belongs to (resolved foreign key).
    gateway_class_name = string

    # Logical endpoints bound on this Gateway. At least one is required.
    listeners = list(object({
      name     = string
      hostname = optional(string)
      port     = number
      protocol = string
      tls = optional(object({
        mode = optional(string)
        certificate_refs = optional(list(object({
          group     = optional(string)
          kind      = optional(string)
          name      = string
          namespace = optional(string)
        })))
        options = optional(map(string))
      }))
      allowed_routes = optional(object({
        namespaces = optional(object({
          from = optional(string)
          selector = optional(object({
            match_labels = optional(map(string))
            match_expressions = optional(list(object({
              key      = string
              operator = string
              values   = optional(list(string))
            })))
          }))
        }))
        kinds = optional(list(object({
          group = optional(string)
          kind  = string
        })))
      }))
    }))

    # Requested Gateway addresses (optional).
    addresses = optional(list(object({
      type  = optional(string)
      value = optional(string)
    })))

    # Infrastructure-level attributes for created resources (optional).
    infrastructure = optional(object({
      labels      = optional(map(string))
      annotations = optional(map(string))
      parameters_ref = optional(object({
        group = string
        kind  = string
        name  = string
      }))
    }))

    # Which ListenerSets may attach to this Gateway (optional).
    allowed_listeners = optional(object({
      namespaces = optional(object({
        from = optional(string)
        selector = optional(object({
          match_labels = optional(map(string))
          match_expressions = optional(list(object({
            key      = string
            operator = string
            values   = optional(list(string))
          })))
        }))
      }))
    }))

    # Gateway-wide frontend/backend TLS configuration (optional).
    tls = optional(object({
      backend = optional(object({
        client_certificate_ref = optional(object({
          group     = optional(string)
          kind      = optional(string)
          name      = string
          namespace = optional(string)
        }))
      }))
      frontend = optional(object({
        default = object({
          validation = optional(object({
            ca_certificate_refs = list(object({
              group     = string
              kind      = string
              name      = string
              namespace = optional(string)
            }))
            mode = optional(string)
          }))
        })
        per_port = optional(list(object({
          port = number
          tls = object({
            validation = optional(object({
              ca_certificate_refs = list(object({
                group     = string
                kind      = string
                name      = string
                namespace = optional(string)
              }))
              mode = optional(string)
            }))
          })
        })))
      }))
    }))
  })
}
