# Typed mirror of KubernetesAltinityOperatorSpec (spec.proto). The spec
# arrives from the proto->tfvars converter in snake_case with every
# StringValueOrRef foreign key -- `namespace` (KubernetesNamespace) and
# `operator_credentials.password` -- resolved to a literal string before
# Terraform runs.
#
# Presence-tracked proto optionals (chart_version, metrics.enabled,
# crd_hook.enabled, operator_credentials.username) carry no optional()
# default here — their proto defaults are resolved in locals.tf, so the
# module renders the same resource whether or not the platform's
# defaulting middleware ran.

variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesAltinityOperator specification"
  type = object({
    namespace             = string
    create_namespace      = optional(bool, false)
    chart_version         = optional(string)
    watch_namespaces      = optional(list(string), [])
    namespace_scoped_rbac = optional(bool, false)

    operator_credentials = optional(object({
      username = optional(string)
      password = string
    }))

    metrics = optional(object({
      enabled = optional(bool)
      resources = optional(object({
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
    }))

    crd_hook = optional(object({
      enabled = optional(bool)
      image = optional(object({
        repo             = optional(string, "")
        tag              = optional(string, "")
        pull_secret_name = optional(string, "")
      }))
    }))

    resources = optional(object({
      limits = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
      requests = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
    }))

    service_monitor_enabled = optional(bool, false)
    node_selector           = optional(map(string), {})

    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])

    image_pull_secrets = optional(list(string), [])

    image = optional(object({
      repo             = optional(string, "")
      tag              = optional(string, "")
      pull_secret_name = optional(string, "")
    }))

    helm_values = optional(string, "")
  })
}
