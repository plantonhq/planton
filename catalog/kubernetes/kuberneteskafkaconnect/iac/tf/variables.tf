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
  description = "KubernetesKafkaConnect specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    version = optional(string, "")
    replicas = optional(number, 1)
    bootstrap_servers = string
    tls = optional(object({
      trusted_certificates = list(object({
        secret_name = string
        certificate = optional(string, "")
        pattern = optional(string, "")
      }))
    }))
    authentication = optional(object({
      type = string
      certificate_and_key = optional(object({
        secret_name = string
        certificate = optional(string)
        key = optional(string)
      }))
      username = optional(string, "")
      password_secret = optional(object({
        secret_name = string
        password = optional(string)
      }))
      sasl = optional(bool, false)
      config = optional(map(string), {})
    }))
    group_id = optional(string, "")
    config_storage_topic = optional(string, "")
    status_storage_topic = optional(string, "")
    offset_storage_topic = optional(string, "")
    config = optional(map(string), {})
    image = optional(string, "")
    plugins = optional(list(object({
      name = string
      artifacts = list(object({
        reference = string
        pull_policy = optional(string)
      }))
    })), [])
    build = optional(object({
      output = object({
        type = optional(string)
        image = string
        push_secret = optional(string, "")
        additional_build_options = optional(list(string), [])
        additional_push_options = optional(list(string), [])
      })
      plugins = list(object({
        name = string
        artifacts = list(object({
          type = string
          url = optional(string, "")
          sha512sum = optional(string, "")
          insecure = optional(bool, false)
          file_name = optional(string, "")
          repository = optional(string, "")
          group = optional(string, "")
          artifact = optional(string, "")
          version = optional(string, "")
        }))
      }))
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
    jvm = optional(object({
      xms = optional(string, "")
      xmx = optional(string, "")
    }))
    rack = optional(object({
      topology_key = string
    }))
    metrics = optional(object({
      enabled = optional(bool, false)
    }))
    node_selector = optional(map(string), {})
    tolerations = optional(list(object({
      key = optional(string, "")
      operator = optional(string, "")
      value = optional(string, "")
      effect = optional(string, "")
      toleration_seconds = optional(number)
    })), [])
  })
}
