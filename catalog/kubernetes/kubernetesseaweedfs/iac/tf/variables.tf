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
  description = "KubernetesSeaweedFs specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    master = optional(object({
      replicas = optional(number)
      data_volume = optional(object({
        size          = optional(string, "")
        storage_class = optional(string, "")
      }))
      volume_size_limit_mb = optional(number)
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
    volume = optional(object({
      replicas = optional(number)
      data_volume = optional(object({
        size          = optional(string, "")
        storage_class = optional(string, "")
      }))
      max_volumes            = optional(number, 0)
      index_mode             = optional(string, "")
      min_free_space_percent = optional(number)
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
    filer = optional(object({
      replicas = optional(number)
      data_volume = optional(object({
        size          = optional(string, "")
        storage_class = optional(string, "")
      }))
      encrypt_volume_data    = optional(bool, false)
      extra_environment_vars = optional(map(string), {})
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
    s3 = optional(object({
      enabled                = optional(bool)
      enable_auth            = optional(bool)
      existing_config_secret = optional(string, "")
      buckets = optional(list(object({
        name           = string
        anonymous_read = optional(bool, false)
        ttl            = optional(string, "")
        object_lock    = optional(bool, false)
        versioning     = optional(bool, false)
      })), [])
      domain_name = optional(string, "")
      dedicated = optional(object({
        replicas = optional(number)
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
    }))
    replication = optional(string, "")
    admin = optional(object({
      enabled              = optional(bool, false)
      existing_auth_secret = optional(string, "")
      data_volume = optional(object({
        size          = optional(string, "")
        storage_class = optional(string, "")
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
    service_monitor_enabled = optional(bool, false)
    image = optional(object({
      registry   = optional(string, "")
      repository = optional(string, "")
      tag        = optional(string, "")
    }))
    helm_values = optional(string, "")
  })
}
