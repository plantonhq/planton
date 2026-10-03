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
  description = "KubernetesKubePrometheusStack specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    skip_crds        = optional(bool, false)
    crd_upgrade_job  = optional(bool, false)
    prometheus = optional(object({
      replicas       = optional(number)
      retention      = optional(string)
      retention_size = optional(string, "")
      disk_size      = optional(string)
      storage_class  = optional(string, "")
      ephemeral      = optional(bool, false)
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
      external_labels              = optional(map(string), {})
      scrape_interval              = optional(string, "")
      evaluation_interval          = optional(string, "")
      discovery                    = optional(string)
      remote_write = optional(list(object({
        url  = string
        name = optional(string, "")
        basic_auth = optional(object({
          username = string
          password_secret = object({
            name = string
            key  = string
          })
        }))
        bearer_token_secret = optional(object({
          name = string
          key  = string
        }))
        sigv4 = optional(object({
          region   = string
          role_arn = optional(string, "")
          access_key_secret = optional(object({
            name = string
            key  = string
          }))
          secret_key_secret = optional(object({
            name = string
            key  = string
          }))
        }))
        azure_ad = optional(object({
          managed_identity_client_id = string
          cloud                      = optional(string, "")
        }))
      })), [])
      enable_remote_write_receiver = optional(bool, false)
      additional_scrape_configs    = optional(string, "")
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
    }))
    alertmanager = optional(object({
      enabled       = optional(bool)
      replicas      = optional(number)
      retention     = optional(string)
      disk_size     = optional(string)
      storage_class = optional(string, "")
      ephemeral     = optional(bool, false)
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
      config_yaml = optional(string, "")
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
      # Typed alert delivery. Every credential arrives already resolved from
      # a managed-secret reference (the tfvars generation flattens each
      # StringValueOrRef to its string).
      notifications = optional(object({
        receivers = list(object({
          name = string
          discord = optional(list(object({
            webhook_url = string
          })), [])
          pushover = optional(list(object({
            token    = string
            user_key = string
            priority = optional(string)
          })), [])
          webhook = optional(list(object({
            url          = string
            bearer_token = optional(string, "")
          })), [])
        }))
        route = object({
          receiver        = string
          group_by        = optional(list(string), [])
          group_wait      = optional(string, "")
          group_interval  = optional(string, "")
          repeat_interval = optional(string, "")
          routes = optional(list(object({
            matchers = list(object({
              label    = string
              operator = optional(string)
              value    = optional(string, "")
            }))
            receiver          = string
            continue_matching = optional(bool, false)
            repeat_interval   = optional(string, "")
          })), [])
        })
        heartbeat = optional(object({
          url          = string
          bearer_token = optional(string, "")
          interval     = optional(string)
        }))
        message = optional(object({
          environment_label = optional(string)
          component_label   = optional(string)
        }))
      }))
    }))
    grafana = optional(object({
      enabled = optional(bool)
      admin_secret = optional(object({
        name         = string
        user_key     = optional(string)
        password_key = optional(string)
      }))
      default_dashboards_enabled = optional(bool)
      storage = optional(object({
        size          = optional(string)
        storage_class = optional(string, "")
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
    operator = optional(object({
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
      admission_webhooks = optional(object({
        disabled     = optional(bool, false)
        cert_manager = optional(bool, false)
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
        priority_class_name = optional(string, "")
      }))
    }))
    exporters = optional(object({
      kube_state_metrics_enabled = optional(bool)
      kube_state_metrics_resources = optional(object({
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))
      node_exporter_enabled = optional(bool)
      node_exporter_resources = optional(object({
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
    control_plane_scrapers = optional(object({
      kube_api_server         = optional(bool)
      kubelet                 = optional(bool)
      kube_controller_manager = optional(bool)
      core_dns                = optional(bool)
      kube_etcd               = optional(bool)
      kube_scheduler          = optional(bool)
      kube_proxy              = optional(bool)
    }))
    default_rules = optional(object({
      enabled         = optional(bool)
      disabled_groups = optional(list(string), [])
    }))
    image_registry     = optional(string, "")
    image_pull_secrets = optional(list(string), [])
    helm_values        = optional(string, "")
  })
}
