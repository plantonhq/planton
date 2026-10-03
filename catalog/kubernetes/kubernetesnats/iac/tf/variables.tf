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
  description = "KubernetesNats specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    cluster = optional(object({
      enabled  = optional(bool, false)
      replicas = optional(number)
    }))
    jet_stream = optional(object({
      enabled               = optional(bool)
      disk_size             = optional(string)
      storage_class         = optional(string, "")
      max_file_store        = optional(string)
      memory_store_max_size = optional(string)
    }))
    auth = optional(object({
      users = optional(list(object({
        username = string
        permissions = optional(object({
          publish_allow   = optional(list(string), [])
          publish_deny    = optional(list(string), [])
          subscribe_allow = optional(list(string), [])
          subscribe_deny  = optional(list(string), [])
        }))
      })), [])
      accounts = optional(list(object({
        name = string
        users = list(object({
          username = string
          permissions = optional(object({
            publish_allow   = optional(list(string), [])
            publish_deny    = optional(list(string), [])
            subscribe_allow = optional(list(string), [])
            subscribe_deny  = optional(list(string), [])
          }))
        }))
        jet_stream_enabled = optional(bool, false)
      })), [])
      no_auth_user = optional(string, "")
    }))
    tls = optional(object({
      secret_name    = string
      verify_clients = optional(bool, false)
    }))
    websocket = optional(object({
      enabled = optional(bool, false)
      port    = optional(number)
    }))
    mqtt = optional(object({
      enabled = optional(bool, false)
      port    = optional(number)
    }))
    leafnodes = optional(object({
      enabled = optional(bool, false)
      port    = optional(number)
    }))
    metrics = optional(object({
      exporter_enabled    = optional(bool, false)
      pod_monitor_enabled = optional(bool, false)
    }))
    nats_box_enabled = optional(bool)
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
    service = optional(object({
      type        = optional(string, "")
      annotations = optional(map(string), {})
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
    images = optional(object({
      nats = optional(object({
        repo             = optional(string, "")
        tag              = optional(string, "")
        pull_secret_name = optional(string, "")
      }))
      reloader = optional(object({
        repo             = optional(string, "")
        tag              = optional(string, "")
        pull_secret_name = optional(string, "")
      }))
      exporter = optional(object({
        repo             = optional(string, "")
        tag              = optional(string, "")
        pull_secret_name = optional(string, "")
      }))
      nats_box = optional(object({
        repo             = optional(string, "")
        tag              = optional(string, "")
        pull_secret_name = optional(string, "")
      }))
    }))
    helm_values = optional(string, "")
  })
}