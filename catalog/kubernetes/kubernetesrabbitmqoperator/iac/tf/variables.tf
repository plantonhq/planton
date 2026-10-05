# Typed mirror of KubernetesRabbitMqOperatorSpec (spec.proto). The spec
# arrives from the proto->tfvars converter in snake_case. The kind has no
# foreign keys and no namespace field — the release manifest installs into
# its fixed `rabbitmq-system` namespace (see the spec).

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
  description = "KubernetesRabbitMqOperator specification"
  type = object({
    watch_namespaces           = optional(list(string), [])
    default_rabbitmq_image     = optional(string, "")
    default_user_updater_image = optional(string, "")
    operator_image = optional(object({
      repo             = optional(string, "")
      tag              = optional(string, "")
      pull_secret_name = optional(string, "")
    }))
    resources = optional(object({
      requests = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
      limits = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
    }))
    node_selector      = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    image_pull_secrets = optional(list(string), [])
  })
}
