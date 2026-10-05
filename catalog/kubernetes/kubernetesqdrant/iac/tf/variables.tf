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
  description = "KubernetesQdrant specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    replicas         = optional(number)
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
    storage = optional(object({
      size          = optional(string)
      storage_class = optional(string, "")
    }))
    snapshots = optional(object({
      size          = optional(string)
      storage_class = optional(string, "")
    }))
    api_key = optional(object({
      generate = optional(bool)
      existing_secret = optional(object({
        name = string
        key  = string
      }))
    }))
    read_only_api_key = optional(object({
      generate = optional(bool)
      existing_secret = optional(object({
        name = string
        key  = string
      }))
    }))
    tls = optional(object({
      secret = string
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
      pod_anti_affinity   = optional(bool, false)
      priority_class_name = optional(string, "")
    }))
    service_monitor_enabled = optional(bool, false)
    image = optional(object({
      repository       = optional(string, "")
      tag              = optional(string, "")
      use_unprivileged = optional(bool, false)
    }))
    helm_values = optional(string, "")
  })
}
