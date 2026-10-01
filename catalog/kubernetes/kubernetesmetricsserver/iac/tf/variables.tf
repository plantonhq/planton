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
  description = "KubernetesMetricsServer specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    replicas         = optional(number)

    kubelet_insecure_tls            = optional(bool, false)
    kubelet_preferred_address_types = optional(list(string), [])
    metric_resolution               = optional(string)
    host_network                    = optional(bool, false)

    api_service = optional(object({
      create                   = optional(bool)
      insecure_skip_tls_verify = optional(bool)
      ca_bundle                = optional(string, "")
    }))

    tls = optional(object({
      type = optional(string)
      cert_manager_issuer = optional(object({
        issuer = optional(object({
          name = string
        }))
        cluster_issuer = optional(object({
          name = string
        }))
      }))
      existing_secret_name = optional(string, "")
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

    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    priority_class_name   = optional(string)
    pod_disruption_budget = optional(bool, false)

    prometheus = optional(object({
      enabled                  = optional(bool, false)
      service_monitor          = optional(bool, false)
      service_monitor_interval = optional(string)
      service_monitor_labels   = optional(map(string), {})
    }))

    image = optional(object({
      repository = optional(string, "")
      tag        = optional(string, "")
    }))

    helm_values = optional(string, "")
  })
}
