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
  description = "KubernetesKeda specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)

    crds = optional(object({
      install           = optional(bool)
      keep_on_uninstall = optional(bool)
    }))

    watch_namespace = optional(string, "")

    operator = optional(object({
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
    }))

    metrics_server = optional(object({
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
    }))

    webhooks = optional(object({
      enabled        = optional(bool)
      failure_policy = optional(string)
      replicas       = optional(number)
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

    pod_identity = optional(object({
      aws_irsa = optional(object({
        enabled  = optional(bool, false)
        role_arn = optional(string, "")
      }))
      azure_workload_identity = optional(object({
        enabled   = optional(bool, false)
        client_id = optional(string, "")
        tenant_id = optional(string, "")
      }))
      gcp_workload_identity = optional(object({
        enabled               = optional(bool, false)
        service_account_email = optional(string, "")
      }))
    }))

    certificates = optional(object({
      type = optional(string)
      cert_manager_issuer = optional(object({
        issuer = optional(object({
          name = string
        }))
        cluster_issuer = optional(object({
          name = string
        }))
      }))
    }))

    http_timeout_ms     = optional(number)
    priority_class_name = optional(string, "")
    node_selector       = optional(map(string), {})
    tolerations = optional(list(object({
      key                = optional(string, "")
      operator           = optional(string, "")
      value              = optional(string, "")
      effect             = optional(string, "")
      toleration_seconds = optional(number)
    })), [])

    prometheus = optional(object({
      enabled         = optional(bool, false)
      service_monitor = optional(bool, false)
    }))

    helm_values = optional(string, "")
  })
}
