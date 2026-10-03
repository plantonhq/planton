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
  description = "KubernetesFlinkOperator specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    chart_version = optional(string)
    watch_namespaces = optional(list(string), [])
    webhook_enabled = optional(bool)
    replicas = optional(number)
    operator_config = optional(map(string), {})
    job_service_account = optional(string)
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
    image_pull_secrets = optional(list(string), [])
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
    helm_values = optional(string, "")
  })
}