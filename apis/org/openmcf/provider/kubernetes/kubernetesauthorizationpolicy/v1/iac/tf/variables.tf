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
  description = "Specification for KubernetesAuthorizationPolicy"
  type = object({
    # Namespace the AuthorizationPolicy is created in (resolved foreign key).
    namespace = string

    # Workload selector. When omitted (and target_refs is also omitted), the
    # policy applies to all workloads in the namespace. Matched by label at
    # runtime by istiod; not an OpenMCF foreign key. Mutually exclusive with
    # target_refs.
    selector = optional(object({
      match_labels = optional(map(string))
    }))

    # Attaches the policy to specific resources (Gateway, Service, ServiceEntry)
    # instead of selecting workloads by label. Mutually exclusive with selector.
    # Plain cross-resource references, not OpenMCF foreign keys.
    target_refs = optional(list(object({
      group     = optional(string)
      kind      = string
      name      = string
      namespace = optional(string)
    })))

    # Rules evaluated against each request. A request matches when at least one
    # rule matches. An empty rule list with action ALLOW denies all requests.
    rules = optional(list(object({
      from = optional(list(object({
        source = optional(object({
          principals             = optional(list(string))
          not_principals         = optional(list(string))
          request_principals     = optional(list(string))
          not_request_principals = optional(list(string))
          namespaces             = optional(list(string))
          not_namespaces         = optional(list(string))
          service_accounts       = optional(list(string))
          not_service_accounts   = optional(list(string))
          ip_blocks              = optional(list(string))
          not_ip_blocks          = optional(list(string))
          remote_ip_blocks       = optional(list(string))
          not_remote_ip_blocks   = optional(list(string))
        }))
      })))
      to = optional(list(object({
        operation = optional(object({
          hosts       = optional(list(string))
          not_hosts   = optional(list(string))
          ports       = optional(list(string))
          not_ports   = optional(list(string))
          methods     = optional(list(string))
          not_methods = optional(list(string))
          paths       = optional(list(string))
          not_paths   = optional(list(string))
        }))
      })))
      when = optional(list(object({
        key        = string
        values     = optional(list(string))
        not_values = optional(list(string))
      })))
    })))

    # The action to take on a matched request: ALLOW (default), DENY, AUDIT, or
    # CUSTOM. Left unset to inherit the upstream default (ALLOW).
    action = optional(string)

    # The external authorizer for the CUSTOM action (names a MeshConfig provider).
    provider = optional(object({
      name = string
    }))
  })
}
