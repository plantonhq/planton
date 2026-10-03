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
  description = "KubernetesExternalSecretsOperator specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    crds = optional(object({
      install           = optional(bool)
      keep_on_uninstall = optional(bool)
    }))
    replicas         = optional(number)
    leader_elect     = optional(bool, false)
    concurrent       = optional(number)
    controller_class = optional(string, "")
    scoped_namespace = optional(string, "")
    scoped_rbac      = optional(bool, false)
    log_level        = optional(string)
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
    workload_identity = optional(object({
      gke = optional(object({
        service_account_email = string
      }))
      eks = optional(object({
        role_arn = string
      }))
      aks = optional(object({
        client_id = string
        tenant_id = optional(string)
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
    priority_class_name   = optional(string, "")
    pod_disruption_budget = optional(bool, false)
    prometheus = optional(object({
      service_monitor          = optional(bool, false)
      service_monitor_interval = optional(string)
      service_monitor_labels   = optional(map(string), {})
    }))
    webhook = optional(object({
      enabled  = optional(bool)
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
    cert_controller = optional(object({
      enabled  = optional(bool)
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
    image_repository = optional(string, "")
    helm_values      = optional(string, "")
  })
}
