variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesValkey specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    chart_version = optional(string)
    image = optional(object({
      registry = optional(string, "")
      repository = optional(string, "")
      tag = optional(string, "")
    }))
    replication = optional(object({
      replicas = optional(number)
      persistence = object({
        size = string
        storage_class = optional(string, "")
        keep_on_uninstall = optional(bool, false)
      })
      replication_user = optional(string)
      diskless_sync = optional(bool, false)
      min_replicas_to_write = optional(number, 0)
      min_replicas_max_lag = optional(number)
      read_service = optional(object({
        enabled = optional(bool)
        type = optional(string)
        annotations = optional(map(string), {})
      }))
    }))
    persistence = optional(object({
      size = string
      storage_class = optional(string, "")
      keep_on_uninstall = optional(bool, false)
    }))
    config = optional(object({
      append_only = optional(bool, false)
      rdb_save_points = optional(list(string), [])
      snapshots_disabled = optional(bool, false)
      max_memory = optional(string, "")
      max_memory_policy = optional(string, "")
      extra_directives = optional(string, "")
    }))
    auth = optional(object({
      users = list(object({
        name = string
        password = optional(string, "")
        permissions = optional(string)
      }))
    }))
    tls = optional(object({
      enabled = optional(bool, false)
      certificate_secret = optional(string, "")
      require_client_certificate = optional(bool, false)
    }))
    service = optional(object({
      type = optional(string)
      port = optional(number)
      annotations = optional(map(string), {})
    }))
    resources = optional(object({
      limits = optional(object({
        cpu = optional(string, "")
        memory = optional(string, "")
      }))
      requests = optional(object({
        cpu = optional(string, "")
        memory = optional(string, "")
      }))
    }))
    metrics = optional(object({
      enabled = optional(bool, false)
      service_monitor_enabled = optional(bool, false)
    }))
    scheduling = optional(object({
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        key = optional(string, "")
        operator = optional(string, "")
        value = optional(string, "")
        effect = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
      priority_class_name = optional(string, "")
    }))
    pod_disruption_budget = optional(object({
      enabled = optional(bool, false)
      max_unavailable = optional(number, 0)
      min_available = optional(number, 0)
    }))
    log_level = optional(string)
    image_pull_secrets = optional(list(string), [])
    helm_values = optional(string, "")
  })
}
