variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesSuperset specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    image = optional(object({
      repository = optional(string)
      tag        = optional(string)
    }))
    image_pull_secrets = optional(list(string), [])
    metadata_database = object({
      host          = string
      port          = optional(number)
      database_name = optional(string)
      username      = optional(string)
      password_secret = object({
        secret_name = string
        secret_key  = optional(string)
      })
      ssl = optional(object({
        enabled = optional(bool, false)
        mode    = optional(string)
      }))
    })
    cache = optional(object({
      host     = string
      port     = optional(number)
      username = optional(string, "")
      password_secret = optional(object({
        secret_name = string
        secret_key  = optional(string)
      }))
      cache_db  = optional(number)
      celery_db = optional(number)
    }))
    secret_key_secret = optional(object({
      secret_name = string
      secret_key  = string
    }))
    web = optional(object({
      replicas = optional(number)
      hpa = optional(object({
        min_replicas                   = optional(number)
        max_replicas                   = optional(number, 0)
        target_cpu_utilization_percent = optional(number)
      }))
      resources = optional(object({
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
    }))
    worker = optional(object({
      enabled  = optional(bool)
      replicas = optional(number)
      hpa = optional(object({
        min_replicas                   = optional(number)
        max_replicas                   = optional(number, 0)
        target_cpu_utilization_percent = optional(number)
      }))
      resources = optional(object({
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
    }))
    beat = optional(object({
      enabled = optional(bool, false)
    }))
    flower = optional(object({
      enabled = optional(bool, false)
    }))
    websockets = optional(object({
      enabled = optional(bool, false)
      image = optional(object({
        repository = optional(string)
        tag        = optional(string)
      }))
      replicas = optional(number)
    }))
    mcp = optional(object({
      enabled = optional(bool, false)
    }))
    init = optional(object({
      admin = optional(object({
        username = optional(string)
        email    = optional(string)
        password_secret = optional(object({
          secret_name = string
          secret_key  = string
        }))
      }))
      load_examples = optional(bool, false)
    }))
    feature_flags    = optional(map(bool), {})
    config_overrides = optional(map(string), {})
    extra_env        = optional(map(string), {})
    extra_env_from_secret = optional(map(object({
      secret_name = string
      secret_key  = string
    })), {})
    bootstrap_script = optional(string, "")
    service = optional(object({
      type        = optional(string)
      annotations = optional(map(string), {})
    }))
    scheduling = optional(object({
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
    }))
    helm_values = optional(string, "")
  })
}