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
  description = "Specification for KubernetesEnvoyFilter"
  type = object({
    # Namespace the EnvoyFilter is created in (resolved foreign key).
    namespace = string

    # Selects in-mesh workloads by label. Matched at runtime by istiod; not an OpenMCF
    # foreign key. Mutually exclusive with target_refs.
    workload_selector = optional(object({
      labels = optional(map(string))
    }))

    # Ordered list of patches with match conditions, applied in list order within a context.
    config_patches = optional(list(object({
      apply_to = optional(string)

      match = optional(object({
        context = optional(string)

        proxy = optional(object({
          proxy_version = optional(string)
          metadata      = optional(map(string))
        }))

        listener = optional(object({
          port_number = optional(number)
          filter_chain = optional(object({
            name                  = optional(string)
            sni                   = optional(string)
            transport_protocol    = optional(string)
            application_protocols = optional(string)
            filter = optional(object({
              name = optional(string)
              sub_filter = optional(object({
                name = optional(string)
              }))
            }))
            destination_port = optional(number)
          }))
          listener_filter = optional(string)
          name            = optional(string)
        }))

        route_configuration = optional(object({
          port_number = optional(number)
          port_name   = optional(string)
          gateway     = optional(string)
          vhost = optional(object({
            name        = optional(string)
            domain_name = optional(string)
            route = optional(object({
              name   = optional(string)
              action = optional(string)
            }))
          }))
          name = optional(string)
        }))

        cluster = optional(object({
          port_number = optional(number)
          service     = optional(string)
          subset      = optional(string)
          name        = optional(string)
        }))
      }))

      patch = optional(object({
        operation = optional(string)
        # Free-form xDS JSON config (preserveUnknownFields upstream); passed through unmodified.
        value        = optional(any)
        filter_class = optional(string)
      }))
    })))

    # Patch-set ordering within a context. Default 0; negatives before, positives after.
    priority = optional(number)

    # Attach to specific resources instead of selecting workloads by label. Each entry is a
    # plain cross-resource reference, not an OpenMCF foreign key. Mutually exclusive with
    # workload_selector. At most 16.
    target_refs = optional(list(object({
      group     = optional(string)
      kind      = optional(string)
      name      = optional(string)
      namespace = optional(string)
    })))
  })
}
