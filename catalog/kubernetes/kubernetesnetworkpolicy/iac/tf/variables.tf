# Input variables for KubernetesNetworkPolicy Terraform module.
# These mirror the KubernetesNetworkPolicySpec protobuf schema; the namespace
# StringValueOrRef arrives flattened to a plain string, and enum fields arrive
# as the proto enum value names (e.g. "ingress", "egress", "TCP").

variable "metadata" {
  description = "Metadata for the network policy resource"
  type = object({
    name = string
    id   = optional(string)
    org  = optional(string)
    env  = optional(string)
  })
}

variable "spec" {
  description = "Specification for the Kubernetes NetworkPolicy"
  type = object({
    namespace   = optional(string, "default")
    name        = string
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})

    # Selects the governed pods; absent/empty selects ALL pods in the
    # namespace (the default-deny building block).
    pod_selector = optional(object({
      match_labels = optional(map(string), {})
      match_expressions = optional(list(object({
        key      = string
        operator = string
        values   = optional(list(string), [])
      })), [])
      # Selects everything; renders as the empty selector. Alternative to labels.
      match_all = optional(bool)
    }))

    # Governed directions: "ingress" / "egress". Empty defers to the API
    # server's inference (ingress always; egress only with egress rules).
    policy_types = optional(list(string), [])

    ingress_rules = optional(list(object({
      from = optional(list(object({
        pod_selector = optional(object({
          match_labels = optional(map(string), {})
          match_expressions = optional(list(object({
            key      = string
            operator = string
            values   = optional(list(string), [])
          })), [])
          # Selects everything; renders as the empty selector. Alternative to labels.
          match_all = optional(bool)
        }))
        namespace_selector = optional(object({
          match_labels = optional(map(string), {})
          match_expressions = optional(list(object({
            key      = string
            operator = string
            values   = optional(list(string), [])
          })), [])
          # Selects everything; renders as the empty selector. Alternative to labels.
          match_all = optional(bool)
        }))
        ip_block = optional(object({
          cidr   = string
          except = optional(list(string), [])
        }))
      })), [])
      ports = optional(list(object({
        protocol = optional(string, "TCP")
        # Number ("5432") or named container port; "" matches all ports.
        port     = optional(string, "")
        end_port = optional(number, 0)
      })), [])
    })), [])

    egress_rules = optional(list(object({
      to = optional(list(object({
        pod_selector = optional(object({
          match_labels = optional(map(string), {})
          match_expressions = optional(list(object({
            key      = string
            operator = string
            values   = optional(list(string), [])
          })), [])
          # Selects everything; renders as the empty selector. Alternative to labels.
          match_all = optional(bool)
        }))
        namespace_selector = optional(object({
          match_labels = optional(map(string), {})
          match_expressions = optional(list(object({
            key      = string
            operator = string
            values   = optional(list(string), [])
          })), [])
          # Selects everything; renders as the empty selector. Alternative to labels.
          match_all = optional(bool)
        }))
        ip_block = optional(object({
          cidr   = string
          except = optional(list(string), [])
        }))
      })), [])
      ports = optional(list(object({
        protocol = optional(string, "TCP")
        port     = optional(string, "")
        end_port = optional(number, 0)
      })), [])
    })), [])
  })
}
