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
  description = "KubernetesKafkaUser specification"
  type = object({
    namespace = string
    kafka_cluster = string
    authentication = optional(object({
      type = string
    }))
    authorization = optional(object({
      type = optional(string)
      acls = list(object({
        resource = object({
          type = string
          name = optional(string, "")
          pattern_type = optional(string)
        })
        operations = list(string)
        host = optional(string, "")
      }))
    }))
    quotas = optional(object({
      producer_byte_rate = optional(number)
      consumer_byte_rate = optional(number)
      request_percentage = optional(number)
      controller_mutation_rate = optional(number)
    }))
  })
}
