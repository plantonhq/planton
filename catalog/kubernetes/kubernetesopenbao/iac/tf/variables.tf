variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "KubernetesOpenBao specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    server = optional(object({
      dev = optional(object({}))
      raft = optional(object({
        data_storage = optional(object({
          size          = optional(string)
          storage_class = optional(string, "")
        }))
      }))
      postgresql = optional(object({
        host     = string
        port     = optional(number)
        database = string
        username = optional(string)
        password_secret = object({
          secret_name = string
          secret_key  = optional(string)
        })
        ssl_mode     = optional(string)
        max_parallel = optional(number)
      }))
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
      audit_storage = optional(object({
        size          = optional(string)
        storage_class = optional(string, "")
      }))
      log_level  = optional(string)
      log_format = optional(string)
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
    tls = optional(object({
      enabled          = optional(bool, false)
      cert_secret_name = optional(string, "")
    }))
    auto_unseal = optional(object({
      aws_kms = optional(object({
        region            = string
        kms_key_id        = string
        access_key_id     = optional(string, "")
        secret_access_key = optional(string, "")
      }))
      gcp_kms = optional(object({
        project                           = string
        region                            = string
        key_ring                          = string
        crypto_key                        = string
        workload_identity_service_account = optional(string, "")
      }))
      azure_key_vault = optional(object({
        vault_name    = string
        key_name      = string
        tenant_id     = string
        client_id     = optional(string, "")
        client_secret = optional(string, "")
      }))
      transit = optional(object({
        address    = string
        key_name   = string
        mount_path = optional(string)
        token      = optional(string, "")
      }))
    }))
    injector = optional(object({
      enabled        = optional(bool, false)
      replicas       = optional(number)
      failure_policy = optional(string)
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
    ui_enabled             = optional(bool)
    network_policy_enabled = optional(bool, false)
    metrics = optional(object({
      enabled                 = optional(bool, false)
      service_monitor_enabled = optional(bool, false)
    }))
    service_account = optional(object({
      annotations            = optional(map(string), {})
      auth_delegator_enabled = optional(bool)
    }))
    helm_values = optional(string, "")
    backup = optional(object({
      schedule       = optional(string)
      retention_days = optional(number)
      object_store = object({
        prefix = optional(string, "")
        s3 = optional(object({
          bucket           = string
          region           = optional(string, "")
          endpoint_url     = optional(string, "")
          force_path_style = optional(bool, false)
          ca_pem           = optional(string, "")
          keyless          = optional(bool, false)
          access_keys = optional(object({
            access_key_id     = string
            secret_access_key = string
          }))
        }))
        gcs = optional(object({
          bucket              = string
          keyless             = optional(bool, false)
          service_account_key = optional(string, "")
        }))
        azure_blob = optional(object({
          storage_account   = optional(string, "")
          container         = string
          keyless           = optional(bool, false)
          storage_key       = optional(string, "")
          connection_string = optional(string, "")
        }))
        r2 = optional(object({
          bucket       = string
          account_id   = string
          jurisdiction = optional(string, "")
          credentials = object({
            access_key_id     = string
            secret_access_key = string
          })
        }))
      })
      workload_identity = optional(object({
        gke = optional(object({
          service_account_email = string
        }))
        eks = optional(object({
          role_arn = string
        }))
        aks = optional(object({
          client_id = string
          tenant_id = optional(string)
        }))
      }))
      auth = optional(object({
        mount_path = optional(string)
        role       = optional(string, "")
      }))
      images = optional(object({
        openbao = optional(object({
          repo             = optional(string, "")
          tag              = optional(string, "")
          pull_secret_name = optional(string, "")
        }))
        rclone = optional(object({
          repo             = optional(string, "")
          tag              = optional(string, "")
          pull_secret_name = optional(string, "")
        }))
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
    restore = optional(object({
      snapshot_key = optional(string, "")
      latest       = optional(bool, false)
      root_token = object({
        name = optional(string, "")
        key  = optional(string, "")
      })
    }))
  })
}
