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
  description = "KubernetesMlflow specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)

    server = optional(object({
      replicas = optional(number)
      image = optional(object({
        repository = optional(string)
        tag        = optional(string)
      }))
      workers = optional(number)
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

    backend_store = optional(object({
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

    artifact_store = optional(object({
      pvc = optional(object({
        storage_size  = optional(string)
        storage_class = optional(string, "")
      }))
      s3_compatible = optional(object({
        endpoint = string
        bucket   = string
        prefix   = optional(string, "")
        credentials_secret = object({
          secret_name           = string
          access_key_id_key     = optional(string)
          secret_access_key_key = optional(string)
        })
      }))
      aws_s3 = optional(object({
        bucket = string
        prefix = optional(string, "")
        region = string
        credentials_secret = optional(object({
          secret_name           = string
          access_key_id_key     = optional(string)
          secret_access_key_key = optional(string)
        }))
      }))
      gcs = optional(object({
        bucket = string
        prefix = optional(string, "")
        credentials_secret = optional(object({
          secret_name = string
          secret_key  = optional(string)
        }))
      }))
      azure_blob = optional(object({
        storage_account = string
        container       = string
        prefix          = optional(string, "")
        credentials_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
      }))
    }))

    auth = optional(object({
      enabled        = optional(bool)
      admin_username = optional(string)
      admin_password_secret = optional(object({
        secret_name = string
        secret_key  = optional(string)
      }))
      default_permission = optional(string)
    }))

    gc = optional(object({
      enabled    = optional(bool, false)
      schedule   = optional(string)
      older_than = optional(string)
    }))

    service = optional(object({
      type        = optional(string)
      annotations = optional(map(string), {})
    }))

    metrics = optional(object({
      enabled                 = optional(bool, false)
      service_monitor_enabled = optional(bool, false)
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

    extra_env = optional(map(string), {})
    extra_env_from_secret = optional(map(object({
      secret_name = string
      secret_key  = optional(string)
    })), {})
    extra_args = optional(list(string), [])
  })
}
