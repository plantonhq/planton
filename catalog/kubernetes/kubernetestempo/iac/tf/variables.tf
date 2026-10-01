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
  description = "KubernetesTempo specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    replicas         = optional(number)
    # Exactly one storage backend. Absent = local.
    storage = optional(object({
      local = optional(object({}))
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
      s3 = optional(object({
        bucket           = string
        endpoint         = string
        region           = optional(string, "")
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
      gcs = optional(object({
        bucket = string
        service_account_key_secret = optional(object({
          name = string
          key  = string
        }))
      }))
      azure = optional(object({
        account_name = string
        container    = string
        account_key_secret = optional(object({
          name = string
          key  = string
        }))
      }))
    }))
    disk_size     = optional(string)
    storage_class = optional(string, "")
    ephemeral     = optional(bool, false)
    retention     = optional(string)

    jaeger_receivers_enabled = optional(bool)
    multi_tenancy_enabled    = optional(bool, false)

    metrics_generator = optional(object({
      enabled          = optional(bool, false)
      remote_write_url = optional(string, "")
      processors       = optional(list(string), [])
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

    tempo_query_enabled     = optional(bool, false)
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
