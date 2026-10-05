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
  description = "KubernetesJupyterHub specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)

    hub = optional(object({
      database = optional(object({
        sqlite_pvc = optional(object({
          storage_size  = optional(string)
          storage_class = optional(string, "")
        }))
        postgres = optional(object({
          host          = string
          port          = optional(number)
          database_name = optional(string)
          username      = optional(string)
          password_secret = object({
            secret_name = string
            secret_key  = optional(string)
          })
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
      }))
      concurrent_spawn_limit      = optional(number)
      active_server_limit         = optional(number)
      allow_named_servers         = optional(bool, false)
      named_server_limit_per_user = optional(number)
      shutdown_on_logout          = optional(bool, false)
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

    authentication = optional(object({
      shared_password = optional(object({
        password_secret = optional(object({
          secret_name = string
          secret_key  = optional(string)
        }))
      }))
      native = optional(object({
        open_signup             = optional(bool, false)
        minimum_password_length = optional(number)
      }))
      github = optional(object({
        client_id = string
        client_secret_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
        oauth_callback_url    = string
        allowed_organizations = optional(list(string), [])
      }))
      google = optional(object({
        client_id = string
        client_secret_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
        oauth_callback_url = string
        hosted_domains     = optional(list(string), [])
      }))
      oidc = optional(object({
        client_id = string
        client_secret_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
        oauth_callback_url = string
        authorize_url      = string
        token_url          = string
        userdata_url       = string
        scopes             = optional(list(string), [])
        username_claim     = optional(string)
        login_service      = optional(string)
      }))
      admin_users   = optional(list(string), [])
      allowed_users = optional(list(string), [])
    }))

    proxy = optional(object({
      service_type        = optional(string)
      service_annotations = optional(map(string), {})
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

    single_user = optional(object({
      image = optional(object({
        repository = string
        tag        = string
      }))
      memory_guarantee      = optional(string)
      memory_limit          = optional(string, "")
      cpu_guarantee         = optional(string, "")
      cpu_limit             = optional(string, "")
      storage = optional(object({
        dynamic = optional(object({
          capacity      = optional(string)
          storage_class = optional(string, "")
        }))
        static = optional(object({
          pvc_name = string
          sub_path = optional(string)
        }))
        none = optional(object({}))
      }))
      default_url           = optional(string, "")
      start_timeout_seconds = optional(number)
      extra_env             = optional(map(string), {})
      profiles = optional(list(object({
        display_name = string
        description  = optional(string, "")
        default      = optional(bool, false)
        image = optional(object({
          repository = string
          tag        = string
        }))
        memory_guarantee = optional(string, "")
        memory_limit     = optional(string, "")
        cpu_guarantee    = optional(string, "")
        cpu_limit        = optional(string, "")
      })), [])
    }))

    scheduling = optional(object({
      user_scheduler_enabled    = optional(bool)
      user_placeholder_replicas = optional(number)
      core_node_selector        = optional(map(string), {})
      user_node_selector        = optional(map(string), {})
    }))

    culling = optional(object({
      enabled         = optional(bool)
      timeout_seconds = optional(number)
      every_seconds   = optional(number)
      max_age_seconds = optional(number)
      cull_users      = optional(bool, false)
    }))

    pre_puller = optional(object({
      hook_enabled       = optional(bool)
      continuous_enabled = optional(bool)
    }))

    network_policy_enabled = optional(bool)
    helm_values            = optional(string, "")
  })
}
