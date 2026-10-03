# Typed mirror of KubernetesGhaRunnerScaleSetControllerSpec (spec.proto).
# The spec arrives from the proto->tfvars converter in snake_case;
# StringValueOrRef fields arrive resolved to their literal string values.

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
  description = "KubernetesGhaRunnerScaleSetController specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string, "0.14.2")
    replicas         = optional(number, 1)
    image = optional(object({
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
    flags = optional(object({
      log_level                          = optional(string, "")
      log_format                         = optional(string, "")
      watch_single_namespace             = optional(string, "")
      runner_max_concurrent_reconciles   = optional(number)
      update_strategy                    = optional(string, "")
      exclude_label_propagation_prefixes = optional(list(string), [])
      k8s_client_rate_limiter_qps        = optional(number)
      k8s_client_rate_limiter_burst      = optional(number)
      rate_limiter                       = optional(string, "")
      health_probe_bind_address          = optional(string, "")
      priority_class_name                = optional(string, "")
    }))
    metrics = optional(object({
      controller_manager_addr = string
      listener_addr           = string
      listener_endpoint       = string
    }))
    image_pull_secrets = optional(list(string), [])
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
    helm_values = optional(string, "")
  })
}
