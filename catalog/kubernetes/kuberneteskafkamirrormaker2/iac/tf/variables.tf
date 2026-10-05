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
  description = "KubernetesKafkaMirrorMaker2 specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    version = optional(string, "")
    replicas = optional(number)
    target = object({
      alias = optional(string)
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
    })
    mirrors = list(object({
      source = object({
        alias = string
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
        config = optional(map(string), {})
      })
      topics_pattern = optional(string)
      topics_exclude_pattern = optional(string, "")
      groups_pattern = optional(string)
      groups_exclude_pattern = optional(string, "")
      source_connector = optional(object({
        tasks_max = optional(number)
        config = optional(map(string), {})
        auto_restart = optional(object({
          enabled = optional(bool, false)
          max_restarts = optional(number)
        }))
      }))
      checkpoint_connector = optional(object({
        tasks_max = optional(number)
        config = optional(map(string), {})
        auto_restart = optional(object({
          enabled = optional(bool, false)
          max_restarts = optional(number)
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
