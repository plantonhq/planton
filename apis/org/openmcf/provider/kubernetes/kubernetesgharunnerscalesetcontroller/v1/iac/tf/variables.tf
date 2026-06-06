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
  description = "Specification for KubernetesGhaRunnerScaleSetController deployment"
  type = object({
    # Kubernetes namespace where the controller will be installed
    namespace = optional(string, "arc-system")

    # Whether to create the namespace
    create_namespace = optional(bool, false)

    # Version of the Helm chart to deploy
    helm_chart_version = optional(string, "0.13.1")

    # Number of controller replicas
    replica_count = optional(number, 1)

    # Container configuration
    container = optional(object({
      resources = optional(object({
        requests = optional(object({
          cpu    = optional(string, "100m")
          memory = optional(string, "128Mi")
        }), {})
        limits = optional(object({
          cpu    = optional(string, "500m")
          memory = optional(string, "512Mi")
        }), {})
      }), {})
      image = optional(object({
        repository  = optional(string, "")
        tag         = optional(string, "")
        pull_policy = optional(string, "IfNotPresent")
      }), {})
    }), {})

    # Controller behavior flags
    flags = optional(object({
      log_level                          = optional(string, "debug")
      log_format                         = optional(string, "text")
      watch_single_namespace             = optional(string, "")
      runner_max_concurrent_reconciles   = optional(number, 2)
      update_strategy                    = optional(string, "immediate")
      exclude_label_propagation_prefixes = optional(list(string), [])
      k8s_client_rate_limiter_qps        = optional(number, 0)
      k8s_client_rate_limiter_burst      = optional(number, 0)
    }), {})

    # Metrics configuration for monitoring
    metrics = optional(object({
      controller_manager_addr = optional(string, "")
      listener_addr           = optional(string, "")
      listener_endpoint       = optional(string, "")
    }))

    # List of image pull secret names
    image_pull_secrets = optional(list(string), [])

    # Priority class name for the controller pods
    priority_class_name = optional(string, "")
  })
}
