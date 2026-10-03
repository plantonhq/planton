# Typed mirror of KubernetesKeycloakOperatorSpec (spec.proto). The spec
# arrives from the proto->tfvars converter in snake_case; the namespace
# value-or-ref foreign key resolves to a literal string before Terraform
# runs. There is NO version field BY DESIGN — the module pins the
# keycloak-k8s-resources release (see locals.tf).

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
  description = "KubernetesKeycloakOperator specification"
  type = object({
    namespace              = string
    create_namespace       = optional(bool, false)
    cluster_wide           = optional(bool, false)
    operator_image         = optional(string, "")
    default_keycloak_image = optional(string, "")
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
    scheduling = optional(object({
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
    }))
  })
}
