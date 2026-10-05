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
  description = "KubernetesLoki specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    # Exactly one deployment mode. Absent = a single-replica monolithic
    # instance (the documented default).
    monolithic = optional(object({
      replicas      = optional(number)
      disk_size     = optional(string)
      storage_class = optional(string, "")
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
    simple_scalable = optional(object({
      write_replicas   = optional(number)
      read_replicas    = optional(number)
      backend_replicas = optional(number)
      disk_size        = optional(string)
      storage_class    = optional(string, "")
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
    # Exactly one storage backend. Absent = filesystem.
    storage = optional(object({
      filesystem = optional(object({}))
      s3 = optional(object({
        bucket           = string
        ruler_bucket     = optional(string, "")
        region           = optional(string, "")
        endpoint         = optional(string, "")
        force_path_style = optional(bool, false)
        insecure         = optional(bool, false)
        credentials = optional(object({
          access_key_id_secret = object({
            name = string
            key  = string
          })
          secret_access_key_secret = object({
            name = string
            key  = string
          })
        }))
      }))
      # The r2 arm; references arrive resolved to their values.
      r2 = optional(object({
        account_id   = string
        jurisdiction = optional(string, "")
        bucket       = string
        credentials = object({
          access_key_id     = string
          secret_access_key = string
        })
      }))
      gcs = optional(object({
        bucket       = string
        ruler_bucket = optional(string, "")
        service_account_key_secret = optional(object({
          name = string
          key  = string
        }))
      }))
      azure = optional(object({
        account_name    = string
        container       = string
        ruler_container = optional(string, "")
        account_key_secret = optional(object({
          name = string
          key  = string
        }))
      }))
    }))
    schema_from_date = optional(string, "")
    retention_period = optional(string, "")
    limits = optional(object({
      ingestion_rate_mb           = optional(number)
      ingestion_burst_size_mb     = optional(number)
      max_global_streams_per_user = optional(number)
      max_query_lookback          = optional(string, "")
    }))
    multi_tenancy = optional(object({
      enabled = optional(bool, false)
      tenants = optional(list(object({
        name          = string
        password_hash = string
      })), [])
      existing_htpasswd_secret = optional(string, "")
    }))
    gateway = optional(object({
      enabled  = optional(bool)
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
    caching = optional(object({
      chunks_cache_enabled    = optional(bool)
      chunks_cache_memory_mb  = optional(number)
      results_cache_enabled   = optional(bool)
      results_cache_memory_mb = optional(number)
    }))
    canary_enabled = optional(bool)
    ruler = optional(object({
      enabled          = optional(bool, false)
      alertmanager_url = optional(string, "")
    }))
    service_monitor_enabled = optional(bool, false)
    usage_reporting         = optional(bool)
    image_registry          = optional(string, "")
    image_pull_secrets      = optional(list(string), [])
    scheduling = optional(object({
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
      priority_class_name = optional(string, "")
    }))
    helm_values = optional(string, "")
  })
}
