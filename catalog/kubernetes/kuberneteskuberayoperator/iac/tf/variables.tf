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
  description = "KubernetesKubeRayOperator specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    chart_version = optional(string)
    watch_namespaces = optional(list(string), [])
    leader_election_enabled = optional(bool)
    batch_scheduler = optional(string, "")
    feature_gates = optional(list(object({
      name = string
      enabled = optional(bool, false)
    })), [])
    metrics_enabled = optional(bool)
    service_monitor_enabled = optional(bool, false)
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
    image_registry = optional(string, "")
    helm_values = optional(string, "")
  })
}