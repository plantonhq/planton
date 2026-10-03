variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesMysql specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    image_name = optional(string, "")
    instances = optional(number)
    storage = object({
      size = string
      storage_class = optional(string, "")
    })
    resources = optional(object({
      limits = optional(object({
        cpu = optional(string, "")
        memory = optional(string, "")
      }))
      requests = optional(object({
        cpu = optional(string, "")
        memory = optional(string, "")
      }))
    }))
    mysql_config = optional(string, "")
    auto_recovery = optional(bool)
    proxy = optional(object({
      haproxy = optional(object({
        replicas = optional(number)
        resources = optional(object({
          limits = optional(object({
            cpu = optional(string, "")
            memory = optional(string, "")
          }))
          requests = optional(object({
            cpu = optional(string, "")
            memory = optional(string, "")
          }))
        }))
        config = optional(string, "")
        expose_primary = optional(object({
          type = optional(string)
          annotations = optional(map(string), {})
        }))
        expose_replicas = optional(object({
          enabled = optional(bool)
          only_readers = optional(bool, false)
          type = optional(string)
          annotations = optional(map(string), {})
        }))
      }))
      proxysql = optional(object({
        replicas = optional(number)
        resources = optional(object({
          limits = optional(object({
            cpu = optional(string, "")
            memory = optional(string, "")
          }))
          requests = optional(object({
            cpu = optional(string, "")
            memory = optional(string, "")
          }))
        }))
        storage = object({
          size = string
          storage_class = optional(string, "")
        })
        config = optional(string, "")
        expose_primary = optional(object({
          type = optional(string)
          annotations = optional(map(string), {})
        }))
      }))
    }))
    tls = optional(object({
      enabled = optional(bool)
      issuer = optional(string, "")
      issuer_kind = optional(string)
      sans = optional(list(string), [])
    }))
    users = optional(list(object({
      name = string
      dbs = optional(list(string), [])
      hosts = optional(list(string), [])
      grants = optional(list(string), [])
      with_grant_option = optional(bool, false)
      password = optional(string, "")
    })), [])
    backup = optional(object({
      storages = list(object({
        name = string
        s3 = optional(object({
          bucket = string
          region = optional(string, "")
          prefix = optional(string, "")
          endpoint_url = optional(string, "")
          force_path_style = optional(bool, false)
          access_keys = object({
            access_key_id = string
            secret_access_key = string
          })
        }))
        azure = optional(object({
          container = string
          prefix = optional(string, "")
          endpoint_url = optional(string, "")
          storage_account = string
          access_key = string
        }))
        pvc = optional(object({
          volume = object({
            size = string
            storage_class = optional(string, "")
          })
        }))
        verify_tls = optional(bool)
      }))
      schedules = optional(list(object({
        name = string
        schedule = string
        storage_name = string
        keep = optional(number)
        delete_from_storage = optional(bool)
      })), [])
      pitr = optional(object({
        enabled = optional(bool, false)
        storage_name = optional(string, "")
        time_between_uploads = optional(number)
      }))
    }))
    scheduling = optional(object({
      anti_affinity_topology_key = optional(string, "")
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        key = optional(string, "")
        operator = optional(string, "")
        value = optional(string, "")
        effect = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
      priority_class_name = optional(string, "")
    }))
    pod_disruption_budget = optional(object({
      max_unavailable = optional(number, 0)
      min_available = optional(number, 0)
    }))
    log_collector = optional(object({
      enabled = optional(bool)
      resources = optional(object({
        limits = optional(object({
          cpu = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu = optional(string, "")
          memory = optional(string, "")
        }))
      }))
    }))
    update_strategy = optional(string)
    unsafe = optional(object({
      cluster_size = optional(bool, false)
      tls = optional(bool, false)
      proxy_size = optional(bool, false)
      backup_if_unhealthy = optional(bool, false)
    }))
    pause = optional(bool, false)
    image_pull_secrets = optional(list(string), [])
  })
}
