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
  description = "KubernetesIstio specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    version          = optional(string)
    revision         = optional(string)
    dataplane_mode   = optional(string)
    istiod = optional(object({
      replicas = optional(number)
      autoscale = optional(object({
        enabled                        = optional(bool)
        min_replicas                   = optional(number)
        max_replicas                   = optional(number)
        target_cpu_utilization_percent = optional(number)
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
      log_level             = optional(string)
      pod_disruption_budget = optional(bool)
      priority_class_name   = optional(string, "")
      node_selector         = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
    }))
    mesh_config = optional(object({
      trust_domain                 = optional(string)
      outbound_traffic_policy_mode = optional(string)
      access_log_file              = optional(string, "")
      cluster_name                 = optional(string, "")
      network                      = optional(string, "")
      mesh_id                      = optional(string, "")
      enable_prometheus_merge      = optional(bool)
    }))
    proxy = optional(object({
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
      log_level      = optional(string)
      auto_inject    = optional(string)
      cluster_domain = optional(string)
    }))
    sidecar_injector = optional(object({
      enable_namespaces_by_default = optional(bool, false)
      rewrite_app_http_probe       = optional(bool)
    }))
    cni = optional(object({
      enabled            = optional(bool, false)
      exclude_namespaces = optional(list(string), [])
      cni_bin_dir        = optional(string)
      cni_conf_dir       = optional(string)
      chained            = optional(bool)
    }))
    ztunnel = optional(object({
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
      log_level = optional(string)
    }))
    gateway_defaults = optional(object({
      service_type = optional(string)
    }))
    images = optional(object({
      hub                = optional(string, "")
      variant            = optional(string)
      image_pull_secrets = optional(list(string), [])
    }))
    helm_values = optional(object({
      base    = optional(string, "")
      istiod  = optional(string, "")
      cni     = optional(string, "")
      ztunnel = optional(string, "")
    }))
  })
}