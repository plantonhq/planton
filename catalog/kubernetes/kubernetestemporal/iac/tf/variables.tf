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
  description = "KubernetesTemporal specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    database = object({
      postgres = optional(object({
        host     = string
        port     = optional(number)
        username = string
        password_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
        max_conns         = optional(number)
        max_idle_conns    = optional(number)
        max_conn_lifetime = optional(string)
        tls = optional(object({
          enabled           = optional(bool, false)
          host_verification = optional(bool, false)
          server_name       = optional(string, "")
        }))
      }))
      mysql = optional(object({
        host     = string
        port     = optional(number)
        username = string
        password_secret = object({
          secret_name = string
          secret_key  = string
        })
        max_conns         = optional(number)
        max_idle_conns    = optional(number)
        max_conn_lifetime = optional(string)
        tls = optional(object({
          enabled           = optional(bool, false)
          host_verification = optional(bool, false)
          server_name       = optional(string, "")
        }))
      }))
      cassandra = optional(object({
        hosts    = list(string)
        port     = optional(number)
        username = string
        password_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
        replication_factor = optional(number)
        datacenter         = optional(string, "")
        tls = optional(object({
          enabled           = optional(bool, false)
          host_verification = optional(bool, false)
          server_name       = optional(string, "")
        }))
      }))
      database_name            = optional(string)
      visibility_database_name = optional(string)
      visibility = optional(object({
        postgres = optional(object({
          host     = string
          port     = optional(number)
          username = string
          password_secret = object({
            secret_name = string
            secret_key  = optional(string)
          })
          max_conns         = optional(number)
          max_idle_conns    = optional(number)
          max_conn_lifetime = optional(string)
          tls = optional(object({
            enabled           = optional(bool, false)
            host_verification = optional(bool, false)
            server_name       = optional(string, "")
          }))
        }))
        mysql = optional(object({
          host     = string
          port     = optional(number)
          username = string
          password_secret = object({
            secret_name = string
            secret_key  = string
          })
          max_conns         = optional(number)
          max_idle_conns    = optional(number)
          max_conn_lifetime = optional(string)
          tls = optional(object({
            enabled           = optional(bool, false)
            host_verification = optional(bool, false)
            server_name       = optional(string, "")
          }))
        }))
        database_name = optional(string)
      }))
      create_databases  = optional(bool, false)
      skip_schema_setup = optional(bool, false)
    })
    num_history_shards = optional(number)
    services = optional(object({
      frontend = optional(object({
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
      history = optional(object({
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
      matching = optional(object({
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
      worker = optional(object({
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
    internal_frontend_enabled = optional(bool, false)
    web_ui = optional(object({
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
    admin_tools_enabled = optional(bool)
    temporal_namespaces = optional(list(object({
      name      = string
      retention = optional(string)
    })), [])
    dynamic_config = optional(object({
      history_size_limit_error  = optional(number)
      history_size_limit_warn   = optional(number)
      history_count_limit_error = optional(number)
      history_count_limit_warn  = optional(number)
      blob_size_limit_error     = optional(number)
      blob_size_limit_warn      = optional(number)
    }))
    archival = optional(object({
      s3 = optional(object({
        region = string
      }))
      gcs            = optional(object({}))
      filestore      = optional(object({}))
      history_uri    = string
      visibility_uri = string
    }))
    service_monitor_enabled = optional(bool, false)
    log_level               = optional(string)
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
    images = optional(object({
      server = optional(object({
        repo             = optional(string, "")
        tag              = optional(string, "")
        pull_secret_name = optional(string, "")
      }))
      web_ui = optional(object({
        repo             = optional(string, "")
        tag              = optional(string, "")
        pull_secret_name = optional(string, "")
      }))
      admin_tools = optional(object({
        repo             = optional(string, "")
        tag              = optional(string, "")
        pull_secret_name = optional(string, "")
      }))
    }))
    helm_values = optional(string, "")
  })
}