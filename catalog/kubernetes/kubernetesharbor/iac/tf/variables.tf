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
  description = "KubernetesHarbor specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    external_url     = string
    expose = optional(object({
      service_type = optional(string, "")
      tls = optional(object({
        enabled          = optional(bool, false)
        cert_secret_name = optional(string, "")
      }))
      node_ports = optional(object({
        http  = optional(number)
        https = optional(number)
      }))
      service_annotations = optional(map(string), {})
      source_ranges       = optional(list(string), [])
      load_balancer_ip    = optional(string, "")
    }))
    admin_auth = optional(object({
      existing_secret_name = optional(string, "")
      existing_secret_key  = optional(string, "")
    }))
    database = object({
      internal = optional(object({
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
        disk_size      = optional(string, "")
        storage_class  = optional(string, "")
        shm_size_limit = optional(string, "")
      }))
      external = optional(object({
        host                 = string
        port                 = optional(number)
        username             = string
        password_secret_name = string
        core_database        = optional(string, "")
        ssl_mode             = optional(string, "")
      }))
    })
    cache = object({
      internal = optional(object({
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
        disk_size     = optional(string, "")
        storage_class = optional(string, "")
      }))
      external = optional(object({
        addr                 = string
        sentinel_master_set  = optional(string, "")
        username             = optional(string, "")
        password             = optional(string, "")
        existing_secret_name = optional(string, "")
        tls_enabled          = optional(bool, false)
      }))
    })
    storage = object({
      filesystem = optional(object({
        disk_size     = optional(string, "")
        storage_class = optional(string, "")
        access_mode   = optional(string, "")
      }))
      s3 = optional(object({
        bucket   = string
        region   = string
        endpoint = optional(string, "")
        credentials = optional(object({
          access_key           = optional(string, "")
          secret_key           = optional(string, "")
          existing_secret_name = optional(string, "")
        }))
        disable_redirect = optional(bool, false)
        encrypt          = optional(bool, false)
        secure           = optional(bool)
        skip_verify      = optional(bool, false)
        root_directory   = optional(string, "")
        storage_class    = optional(string, "")
      }))
      gcs = optional(object({
        bucket                = string
        use_workload_identity = optional(bool, false)
        key_data              = optional(string, "")
        existing_secret_name  = optional(string, "")
        root_directory        = optional(string, "")
        chunk_size            = optional(number)
      }))
      azure = optional(object({
        account_name         = string
        container            = string
        account_key          = optional(string, "")
        existing_secret_name = optional(string, "")
        realm                = optional(string, "")
      }))
    })
    trivy = optional(object({
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
      disk_size            = optional(string, "")
      skip_update          = optional(bool, false)
      skip_java_db_update  = optional(bool, false)
      offline_scan         = optional(bool, false)
      db_repositories      = optional(list(string), [])
      java_db_repositories = optional(list(string), [])
      github_token         = optional(string, "")
      severity             = optional(string, "")
      ignore_unfixed       = optional(bool, false)
      timeout              = optional(string, "")
    }))
    core = optional(object({
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
    portal = optional(object({
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
    registry = optional(object({
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
    jobservice = optional(object({
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
      max_job_workers = optional(number)
      log_disk_size   = optional(string, "")
    }))
    nginx = optional(object({
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
    internal_tls = optional(object({
      enabled = optional(bool, false)
      cert_secrets = optional(object({
        core_secret_name       = string
        jobservice_secret_name = string
        registry_secret_name   = string
        portal_secret_name     = string
        trivy_secret_name      = optional(string, "")
      }))
      strong_ssl_ciphers = optional(bool, false)
    }))
    metrics = optional(object({
      enabled                  = optional(bool, false)
      service_monitor_enabled  = optional(bool, false)
      service_monitor_interval = optional(string, "")
      service_monitor_labels   = optional(map(string), {})
    }))
    cache_layer = optional(object({
      enabled      = optional(bool, false)
      expire_hours = optional(number)
    }))
    outbound_proxy = optional(object({
      http_proxy  = optional(string, "")
      https_proxy = optional(string, "")
      no_proxy    = optional(string, "")
    }))
    log_level                 = optional(string, "")
    image_registry            = optional(string, "")
    image_pull_secrets        = optional(list(string), [])
    ca_bundle_secret_name     = optional(string, "")
    keep_volumes_on_uninstall = optional(bool)
    update_strategy           = optional(string, "")
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
