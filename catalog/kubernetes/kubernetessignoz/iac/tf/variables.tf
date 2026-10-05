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
  description = "KubernetesSignoz specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    # The ClickHouse SigNoz stores telemetry in — required, composed,
    # never bundled (nothing ClickHouse-related installs).
    clickhouse = object({
      host         = string
      cluster_name = optional(string, "")
      tcp_port     = optional(number)
      http_port    = optional(number)
      username     = string
      password_secret = object({
        secret_name = string
        secret_key  = string
      })
      secure = optional(bool, false)
      verify = optional(bool, false)
    })
    server = optional(object({
      disk_size     = optional(string)
      storage_class = optional(string, "")
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
      external_url = optional(string, "")
      smtp = optional(object({
        address  = string
        from     = string
        username = optional(string, "")
        password_secret = optional(object({
          name = string
          key  = string
        }))
        tls_enabled = optional(bool, false)
      }))
      env = optional(map(string), {})
    }))
    otel_collector = optional(object({
      replicas = optional(number)
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
      autoscaling = optional(object({
        enabled                           = optional(bool, false)
        min_replicas                      = optional(number)
        max_replicas                      = optional(number)
        target_cpu_utilization_percent    = optional(number)
        target_memory_utilization_percent = optional(number)
      }))
      jaeger_receiver_enabled            = optional(bool)
      zipkin_receiver_enabled            = optional(bool, false)
      http_logs_receivers_enabled        = optional(bool)
      low_cardinality_exception_grouping = optional(bool, false)
    }))
    cluster_name       = optional(string, "")
    image_registry     = optional(string, "")
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
      priority_class_name = optional(string, "")
    }))
    helm_values = optional(string, "")
  })
}
