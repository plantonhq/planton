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
  description = "KubernetesPerconaMongoOperator specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    chart_version = optional(string)
    replicas = optional(number)
    watch = optional(object({
      cluster_wide = optional(bool, false)
      namespaces = optional(list(string), [])
    }))
    max_concurrent_reconciles = optional(number)
    log = optional(object({
      structured = optional(bool, false)
      level = optional(string)
    }))
    disable_telemetry = optional(bool, false)
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
    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key = optional(string, "")
      operator = optional(string, "")
      value = optional(string, "")
      effect = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    image_pull_secrets = optional(list(string), [])
    image = optional(object({
      repository = optional(string, "")
      tag = optional(string, "")
    }))
    helm_values = optional(string, "")
  })
}
