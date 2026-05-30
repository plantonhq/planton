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
  description = "Specification for KubernetesHttpRoute"
  type = object({
    # Namespace the HTTPRoute is created in (resolved foreign key).
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

    # Hostnames matched against the HTTP Host header.
    hostnames = optional(list(string))

    # Routing rules. At least one is required.
    rules = list(object({
      name = optional(string)

      matches = optional(list(object({
        path = optional(object({
          type  = optional(string)
          value = optional(string)
        }))
        headers = optional(list(object({
          type  = optional(string)
          name  = string
          value = string
        })))
        query_params = optional(list(object({
          type  = optional(string)
          name  = string
          value = string
        })))
        method = optional(string)
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
        request_redirect = optional(object({
          scheme   = optional(string)
          hostname = optional(string)
          path = optional(object({
            type                 = string
            replace_full_path    = optional(string)
            replace_prefix_match = optional(string)
          }))
          port        = optional(number)
          status_code = optional(number)
        }))
        url_rewrite = optional(object({
          hostname = optional(string)
          path = optional(object({
            type                 = string
            replace_full_path    = optional(string)
            replace_prefix_match = optional(string)
          }))
        }))
        cors = optional(object({
          allow_origins     = optional(list(string))
          allow_credentials = optional(bool)
          allow_methods     = optional(list(string))
          allow_headers     = optional(list(string))
          expose_headers    = optional(list(string))
          max_age           = optional(number)
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
          request_redirect = optional(object({
            scheme   = optional(string)
            hostname = optional(string)
            path = optional(object({
              type                 = string
              replace_full_path    = optional(string)
              replace_prefix_match = optional(string)
            }))
            port        = optional(number)
            status_code = optional(number)
          }))
          url_rewrite = optional(object({
            hostname = optional(string)
            path = optional(object({
              type                 = string
              replace_full_path    = optional(string)
              replace_prefix_match = optional(string)
            }))
          }))
          cors = optional(object({
            allow_origins     = optional(list(string))
            allow_credentials = optional(bool)
            allow_methods     = optional(list(string))
            allow_headers     = optional(list(string))
            expose_headers    = optional(list(string))
            max_age           = optional(number)
          }))
          extension_ref = optional(object({
            group = string
            kind  = string
            name  = string
          }))
        })))
      })))

      timeouts = optional(object({
        request         = optional(string)
        backend_request = optional(string)
      }))
    }))
  })
}
