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
  description = "KubernetesTrino specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    image = optional(object({
      registry   = optional(string, "")
      repository = optional(string)
      tag        = optional(string)
    }))
    image_pull_secrets = optional(list(string), [])
    node_environment   = optional(string)
    log_level          = optional(string)
    coordinator = optional(object({
      jvm = optional(object({
        max_heap_size    = optional(string, "")
        max_heap_percent = optional(number)
      }))
      max_query_memory_per_node = optional(string)
      heap_headroom_per_node    = optional(string, "")
      include_in_scheduling     = optional(bool, false)
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
    }))
    workers = optional(object({
      replicas = optional(number)
      jvm = optional(object({
        max_heap_size    = optional(string, "")
        max_heap_percent = optional(number)
      }))
      max_query_memory_per_node = optional(string)
      heap_headroom_per_node    = optional(string, "")
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
      graceful_shutdown = optional(object({
        enabled              = optional(bool, false)
        grace_period_seconds = optional(number)
      }))
      hpa = optional(object({
        max_replicas                      = optional(number, 0)
        target_cpu_utilization_percent    = optional(number)
        target_memory_utilization_percent = optional(number)
      }))
      keda = optional(object({
        min_replicas             = optional(number)
        max_replicas             = optional(number, 0)
        polling_interval_seconds = optional(number)
        cooldown_period_seconds  = optional(number)
        triggers                 = string
      }))
    }))
    max_query_memory = optional(string)
    auth = optional(object({
      enabled        = optional(bool)
      admin_username = optional(string)
      existing_password_db_secret = optional(object({
        secret_name = string
      }))
      groups_secret = optional(object({
        secret_name = string
      }))
    }))
    https = optional(object({
      enabled = optional(bool, false)
      port    = optional(number)
      keystore_secret = optional(object({
        secret_name = string
        secret_key  = optional(string)
      }))
    }))
    catalogs = optional(object({
      sample_catalogs_enabled = optional(bool)
      postgres = optional(list(object({
        name     = string
        host     = string
        port     = optional(number)
        database = string
        username = optional(string)
        password_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
        additional_properties = optional(list(string), [])
      })), [])
      mysql = optional(list(object({
        name     = string
        host     = string
        port     = optional(number)
        username = optional(string)
        password_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
        additional_properties = optional(list(string), [])
      })), [])
      custom = optional(map(string), {})
    }))
    fault_tolerant_execution = optional(object({
      retry_policy = string
      exchange_manager = object({
        base_directories      = list(string)
        additional_properties = optional(list(string), [])
      })
    }))
    access_control_rules         = optional(string, "")
    resource_groups_config       = optional(string, "")
    session_properties_config    = optional(string, "")
    event_listener_properties    = optional(list(string), [])
    additional_config_properties = optional(list(string), [])
    extra_env                    = optional(map(string), {})
    extra_env_from_secret = optional(map(object({
      secret_name = string
      secret_key  = string
    })), {})
    metrics = optional(object({
      enabled                 = optional(bool, false)
      service_monitor_enabled = optional(bool, false)
      exporter_image          = optional(string, "")
    }))
    network_policy_enabled = optional(bool, false)
    service = optional(object({
      type        = optional(string)
      annotations = optional(map(string), {})
    }))
    helm_values = optional(string, "")
  })
}