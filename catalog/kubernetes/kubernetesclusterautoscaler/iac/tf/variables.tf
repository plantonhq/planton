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
  description = "KubernetesClusterAutoscaler specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)

    aws = optional(object({
      region = string
      auto_discovery = optional(object({
        cluster_name = string
        tags         = optional(list(string), [])
      }))
      node_groups = optional(list(object({
        name     = string
        min_size = optional(number, 0)
        max_size = optional(number, 0)
      })), [])
      irsa_role_arn = optional(string, "")
      # secret_access_key is a credential — it flows only into the chart
      # values, which the chart materializes into its own Secret.
      access_keys = optional(object({
        access_key_id     = string
        secret_access_key = string
      }))
    }))

    azure = optional(object({
      subscription_id = string
      resource_group  = string
      cluster_name    = optional(string, "")
      node_groups = optional(list(object({
        name     = string
        min_size = optional(number, 0)
        max_size = optional(number, 0)
      })), [])
      identity = object({
        use_workload_identity     = optional(bool, false)
        use_managed_identity      = optional(bool, false)
        user_assigned_identity_id = optional(string, "")
        # client_secret is a credential — same Secret flow as above.
        service_principal = optional(object({
          tenant_id     = string
          client_id     = string
          client_secret = string
        }))
      })
    }))

    gce = optional(object({
      instance_group_prefixes = list(object({
        name     = string
        min_size = optional(number, 0)
        max_size = optional(number, 0)
      }))
      workload_identity_service_account_email = optional(string, "")
    }))

    cluster_api = optional(object({
      mode                  = optional(string)
      kubeconfig_secret     = optional(string, "")
      namespace             = optional(string, "")
      namespace_scoped_rbac = optional(bool, false)
    }))

    # api_key is a credential — same Secret flow as above.
    civo = optional(object({
      cluster_id = string
      region     = string
      api_key    = string
      api_url    = optional(string)
    }))

    kwok = optional(object({
      config_map_name = optional(string)
    }))

    scaling = optional(object({
      expander                    = optional(string, "")
      balance_similar_node_groups = optional(bool, false)
      scale_down = optional(object({
        enabled               = optional(bool)
        utilization_threshold = optional(string, "")
        unneeded_time         = optional(string, "")
        delay_after_add       = optional(string, "")
        delay_after_delete    = optional(string, "")
        delay_after_failure   = optional(string, "")
      }))
      scan_interval                 = optional(string, "")
      max_node_provision_time       = optional(string, "")
      skip_nodes_with_local_storage = optional(bool)
      skip_nodes_with_system_pods   = optional(bool)
    }))

    extra_args = optional(map(string), {})

    deployment = optional(object({
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
      priority_class_name = optional(string)
      node_selector       = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
    }))

    prometheus = optional(object({
      service_monitor                  = optional(bool, false)
      service_monitor_selector_release = optional(string, "")
    }))

    helm_values = optional(string, "")
  })
}
