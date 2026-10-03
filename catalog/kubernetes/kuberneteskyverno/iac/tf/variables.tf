# Typed mirror of KubernetesKyvernoSpec (spec.proto).
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

  validation {
    condition     = length(var.metadata.name) <= 47
    error_message = "metadata.name must be at most 47 characters: the kyverno chart derives the webhook Service, the runtime ConfigMap and the pre-delete hook Job (longest suffix: -hook-pre-delete) from it and truncates past the Kubernetes 63-character limit, breaking the chart's own name-based wiring."
  }
}

variable "spec" {
  description = "KubernetesKyverno specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string, "3.8.2")
    crds = optional(object({
      install           = optional(bool, true)
      keep_on_uninstall = optional(bool, false)
      migration_enabled = optional(bool, true)
    }))
    config = optional(object({
      webhook_exclude_namespaces       = optional(list(string), [])
      resource_filters_include         = optional(list(string), [])
      resource_filters_exclude         = optional(list(string), [])
      exclude_groups                   = optional(list(string), [])
      exclude_usernames                = optional(list(string), [])
      default_registry                 = optional(string, "")
      enable_default_registry_mutation = optional(bool)
    }))
    features = optional(object({
      force_failure_policy_ignore = optional(bool, false)
      background_scan = optional(object({
        enabled  = optional(bool, true)
        workers  = optional(number)
        interval = optional(string, "")
      }))
      generate_validating_admission_policy = optional(bool)
      admission_reports                    = optional(bool)
      aggregate_reports                    = optional(bool)
      policy_reports                       = optional(bool)
      logging_format                       = optional(string, "")
      logging_verbosity                    = optional(number)
      omit_event_types                     = optional(list(string), [])
    }))
    admission_controller = optional(object({
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
      autoscaling = optional(object({
        min_replicas                      = optional(number)
        max_replicas                      = number
        target_cpu_utilization_percentage = optional(number)
      }))
    }))
    background_controller = optional(object({
      enabled  = optional(bool)
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
    cleanup_controller = optional(object({
      enabled  = optional(bool)
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
    reports_controller = optional(object({
      enabled  = optional(bool)
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
    certificates = optional(object({
      cert_manager = optional(object({
        issuer_name = optional(string, "")
        issuer_kind = optional(string, "")
      }))
    }))
    metrics = optional(object({
      service_monitor = optional(bool, false)
    }))
    image_registry           = optional(string, "")
    image_pull_secrets       = optional(list(string), [])
    webhooks_cleanup_enabled = optional(bool)
    helm_values              = optional(string, "")
  })
}
