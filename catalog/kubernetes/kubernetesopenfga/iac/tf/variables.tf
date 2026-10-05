# Typed mirror of KubernetesOpenFgaSpec (spec.proto).
# The spec arrives from the proto->tfvars converter in snake_case;
# StringValueOrRef fields arrive resolved to their literal string values.

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
  description = "KubernetesOpenFga specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string, "0.3.10")
    replicas         = optional(number)
    datastore = object({
      postgres = optional(object({
        host     = string
        port     = optional(number)
        database = string
        username = string
        password_secret = object({
          secret_name = string
          secret_key  = optional(string, "")
        })
        ssl_mode = optional(string, "")
      }))
      mysql = optional(object({
        host     = string
        port     = optional(number)
        database = string
        username = string
        password_secret = object({
          secret_name = string
          secret_key  = optional(string, "")
        })
      }))
      memory             = optional(object({}))
      migration_timeout  = optional(string, "")
      max_open_conns     = optional(number)
      max_idle_conns     = optional(number)
      conn_max_idle_time = optional(string, "")
      conn_max_lifetime  = optional(string, "")
    })
    authn = optional(object({
      preshared = optional(object({
        keys                      = optional(list(string), [])
        existing_keys_secret_name = optional(string, "")
      }))
      oidc = optional(object({
        issuer   = string
        audience = string
      }))
    }))
    metrics = optional(object({
      enabled                 = optional(bool)
      service_monitor_enabled = optional(bool, false)
      enable_rpc_histograms   = optional(bool, false)
    }))
    tracing = optional(object({
      enabled       = optional(bool, false)
      otlp_endpoint = optional(string, "")
      sample_ratio  = optional(string, "")
    }))
    log = optional(object({
      level  = optional(string, "")
      format = optional(string, "")
    }))
    tuning = optional(object({
      max_tuples_per_write              = optional(number)
      max_types_per_authorization_model = optional(number)
      max_checks_per_batch_check        = optional(number)
      list_objects_deadline             = optional(string, "")
      list_objects_max_results          = optional(number)
      list_users_deadline               = optional(string, "")
      list_users_max_results            = optional(number)
      request_timeout                   = optional(string, "")
      check_query_cache = optional(object({
        enabled = optional(bool, false)
        limit   = optional(number)
        ttl     = optional(string, "")
      }))
      experimentals = optional(list(string), [])
    }))
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
    hpa = optional(object({
      enabled                           = optional(bool, false)
      min_replicas                      = optional(number)
      max_replicas                      = optional(number)
      target_cpu_utilization_percent    = optional(number)
      target_memory_utilization_percent = optional(number)
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
    service_account_annotations = optional(map(string), {})
    helm_values                 = optional(string, "")
  })
}
