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
  description = "KubernetesKafkaConnector specification"
  type = object({
    namespace = string
    connect_cluster = string
    connector_class = string
    tasks_max = optional(number)
    version = optional(string, "")
    config = optional(map(string), {})
    state = optional(string, "")
    auto_restart = optional(object({
      enabled = optional(bool, false)
      max_restarts = optional(number)
    }))
    list_offsets = optional(object({
      to_config_map = string
    }))
    alter_offsets = optional(object({
      from_config_map = string
    }))
  })
}
