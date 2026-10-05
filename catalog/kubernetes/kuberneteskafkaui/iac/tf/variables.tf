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
  description = "KubernetesKafkaUi specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    chart_version = optional(string)
    clusters = list(object({
      name = string
      bootstrap_servers = string
      read_only = optional(bool, false)
      tls = optional(object({
        ca_secret_name = string
        ca_certificate = optional(string)
      }))
      sasl = optional(object({
        mechanism = string
        username = string
        password_secret = object({
          secret_name = string
          key = optional(string)
        })
      }))
      schema_registry = optional(object({
        url = string
        username = optional(string, "")
        password_secret = optional(object({
          secret_name = string
          key = optional(string)
        }))
      }))
      kafka_connect = optional(list(object({
        name = string
        address = string
        username = optional(string, "")
        password_secret = optional(object({
          secret_name = string
          key = optional(string)
        }))
      })), [])
      properties = optional(map(string), {})
    }))
    auth = optional(object({
      type = string
      user = object({
        username = string
        password = string
      })
    }))
    replicas = optional(number)
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
    service_type = optional(string)
    service_port = optional(number)
    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key = optional(string, "")
      operator = optional(string, "")
      value = optional(string, "")
      effect = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
    image_registry = optional(string, "")
    helm_values = optional(string, "")
  })
}
