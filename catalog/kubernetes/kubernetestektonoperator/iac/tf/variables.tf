# Typed mirror of KubernetesTektonOperatorSpec (spec.proto). The spec
# arrives from the proto->tfvars converter in snake_case. The kind has no
# foreign keys and no namespace field — the release manifest installs into
# its fixed `tekton-operator` namespace (see the spec).

variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "KubernetesTektonOperator specification"
  type = object({
    operator_image = optional(object({
      repo             = optional(string, "")
      tag              = optional(string, "")
      pull_secret_name = optional(string, "")
    }))
    webhook_image = optional(object({
      repo             = optional(string, "")
      tag              = optional(string, "")
      pull_secret_name = optional(string, "")
    }))
    operator_resources = optional(object({
      requests = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
      limits = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
    }))
    webhook_resources = optional(object({
      requests = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
      limits = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
    }))
    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    image_pull_secrets = optional(list(string), [])
    image_registry     = optional(string, "")
  })
}
