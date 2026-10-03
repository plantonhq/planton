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
  description = "KubernetesAirflow specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    airflow_version  = optional(string)
    executor         = optional(string)

    database = object({
      postgres = optional(object({
        host          = string
        port          = optional(number)
        database_name = optional(string)
        username      = optional(string)
        password_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
        ssl_mode = optional(string)
      }))
      mysql = optional(object({
        host          = string
        port          = optional(number)
        database_name = optional(string)
        username      = optional(string)
        password_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
      }))
    })

    broker = optional(object({
      bundled_redis = optional(object({
        persistence_size = optional(string)
        storage_class    = optional(string)
        resources = optional(object({
          requests = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
          limits = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
        }))
      }))
      valkey = optional(object({
        host     = string
        port     = optional(number)
        username = optional(string, "")
        password_secret = optional(object({
          secret_name = string
          secret_key  = optional(string)
        }))
        database_number = optional(number)
      }))
      existing_broker_url_secret = optional(object({
        secret_name = string
      }))
    }))

    dags = optional(object({
      git_sync = optional(object({
        repo               = string
        ref                = optional(string, "")
        sub_path           = optional(string, "")
        period_seconds     = optional(number)
        depth              = optional(number)
        credentials_secret = optional(string, "")
        ssh_key_secret     = optional(string, "")
        known_hosts        = optional(string, "")
        resources = optional(object({
          requests = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
          limits = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
        }))
      }))
      persistence = optional(object({
        size           = optional(string)
        storage_class  = optional(string, "")
        existing_claim = optional(string, "")
      }))
    }))

    components = optional(object({
      api_server = optional(object({
        replicas = optional(number)
        resources = optional(object({
          requests = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
          limits = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
        }))
      }))
      scheduler = optional(object({
        replicas = optional(number)
        resources = optional(object({
          requests = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
          limits = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
        }))
      }))
      dag_processor = optional(object({
        replicas = optional(number)
        resources = optional(object({
          requests = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
          limits = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
        }))
      }))
      triggerer = optional(object({
        enabled          = optional(bool)
        replicas         = optional(number)
        persistence_size = optional(string)
        resources = optional(object({
          requests = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
          limits = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
        }))
      }))
      workers = optional(object({
        replicas = optional(number)
        resources = optional(object({
          requests = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
          limits = optional(object({
            cpu    = optional(string, "")
            memory = optional(string, "")
          }))
        }))
        persistence_enabled = optional(bool)
        persistence_size    = optional(string)
        keda = optional(object({
          enabled                  = optional(bool, false)
          min_replicas             = optional(number)
          max_replicas             = optional(number)
          polling_interval_seconds = optional(number)
          cooldown_period_seconds  = optional(number)
        }))
      }))
    }))

    pgbouncer = optional(object({
      enabled                  = optional(bool, false)
      metadata_pool_size       = optional(number)
      result_backend_pool_size = optional(number)
      max_client_connections   = optional(number)
      resources = optional(object({
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
    }))

    logging = optional(object({
      persistence = optional(object({
        enabled       = optional(bool, false)
        size          = optional(string)
        storage_class = optional(string, "")
      }))
      elasticsearch = optional(object({
        host     = string
        port     = optional(number)
        scheme   = optional(string)
        username = optional(string, "")
        password_secret = optional(object({
          secret_name = string
          secret_key  = optional(string)
        }))
      }))
      opensearch = optional(object({
        host     = string
        port     = optional(number)
        scheme   = optional(string)
        username = optional(string, "")
        password_secret = optional(object({
          secret_name = string
          secret_key  = optional(string)
        }))
      }))
    }))

    admin_user = optional(object({
      create   = optional(bool)
      username = optional(string)
      email    = optional(string)
      password_secret = optional(object({
        secret_name = string
        secret_key  = optional(string)
      }))
    }))

    security = optional(object({
      fernet_key_secret_name     = optional(string, "")
      api_secret_key_secret_name = optional(string, "")
      jwt_secret_name            = optional(string, "")
    }))

    statsd_enabled = optional(bool)
    load_examples  = optional(bool, false)

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
      airflow_repository   = optional(string, "")
      airflow_tag          = optional(string, "")
      airflow_digest       = optional(string, "")
      statsd_repository    = optional(string, "")
      redis_repository     = optional(string, "")
      pgbouncer_repository = optional(string, "")
      git_sync_repository  = optional(string, "")
    }))

    helm_values = optional(string, "")
  })
}
