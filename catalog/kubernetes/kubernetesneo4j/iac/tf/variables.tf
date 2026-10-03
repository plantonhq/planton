variable "metadata" {
  description = "Catalog object metadata"
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
  description = "KubernetesNeo4j specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    chart_version = optional(string)
    edition = optional(string)
    accept_license_agreement = optional(bool, false)
    auth = optional(object({
      password = optional(string)
      existing_secret = optional(string)
    }))
    cluster_name = optional(string, "")
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
    data_volume = optional(object({
      size = optional(string)
      storage_class = optional(string, "")
    }))
    memory = optional(object({
      heap_initial = optional(string, "")
      heap_max = optional(string, "")
      page_cache = optional(string, "")
    }))
    config = optional(map(string), {})
    apoc_config = optional(map(string), {})
    additional_jvm_arguments = optional(list(string), [])
    use_default_jvm_arguments = optional(bool)
    service = optional(object({
      type = optional(string)
      annotations = optional(map(string), {})
    }))
    ssl = optional(object({
      bolt = optional(object({
        secret = optional(string, "")
      }))
      https = optional(object({
        secret = optional(string, "")
      }))
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
      pod_anti_affinity = optional(bool)
      priority_class_name = optional(string, "")
    }))
    service_monitor_enabled = optional(bool, false)
    image = optional(object({
      registry = optional(string, "")
      repository = optional(string, "")
      tag = optional(string, "")
    }))
    helm_values = optional(string, "")
  })
}
