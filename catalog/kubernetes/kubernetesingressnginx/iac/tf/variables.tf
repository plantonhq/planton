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
  description = "KubernetesIngressNginx specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)

    ingress_class = optional(object({
      name                        = optional(string)
      is_default_class           = optional(bool, false)
      controller_value           = optional(string, "")
      watch_ingress_without_class = optional(bool, false)
    }))

    replicas = optional(number)

    autoscaling = optional(object({
      enabled                           = optional(bool, false)
      min_replicas                      = optional(number)
      max_replicas                      = optional(number)
      target_cpu_utilization_percent    = optional(number)
      target_memory_utilization_percent = optional(number)
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

    service = optional(object({
      type                         = optional(string)
      annotations                  = optional(map(string), {})
      external_traffic_policy      = optional(string)
      load_balancer_source_ranges  = optional(list(string), [])
      load_balancer_class          = optional(string, "")
      enable_http                  = optional(bool)
      enable_https                 = optional(bool)
      http_node_port               = optional(number)
      https_node_port              = optional(number)
      internal = optional(object({
        enabled     = optional(bool, false)
        annotations = optional(map(string), {})
      }))
    }))

    controller_kind = optional(string)
    host_network    = optional(bool, false)
    host_ports      = optional(bool, false)

    nginx_config              = optional(map(string), {})
    allow_snippet_annotations = optional(bool, false)

    default_tls_certificate = optional(object({
      secret_name = string
      namespace   = optional(string, "")
    }))

    default_backend = optional(object({
      enabled  = optional(bool, false)
      replicas = optional(number)
      image    = optional(string, "")
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

    admission_webhooks = optional(object({
      enabled         = optional(bool)
      failure_policy  = optional(string)
      timeout_seconds = optional(number)
    }))

    metrics = optional(object({
      enabled                  = optional(bool, false)
      service_monitor          = optional(bool, false)
      service_monitor_interval = optional(string)
      service_monitor_labels   = optional(map(string), {})
    }))

    tcp_services = optional(map(string), {})
    udp_services = optional(map(string), {})

    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    priority_class_name = optional(string, "")

    image_registry = optional(string, "")
    helm_values    = optional(string, "")
  })
}
