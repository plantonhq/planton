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
  description = "KubernetesKafkaTopic specification"
  type = object({
    namespace = string
    kafka_cluster = string
    topic_name = optional(string, "")
    partitions = optional(number)
    replicas = optional(number)
    config = optional(map(string), {})
  })
}
