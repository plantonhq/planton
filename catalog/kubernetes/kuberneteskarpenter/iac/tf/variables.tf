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
  description = "KubernetesKarpenter specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    crds = optional(object({
      install           = optional(bool)
      keep_on_uninstall = optional(bool)
    }))
    cluster = object({
      name              = string
      endpoint          = optional(string, "")
      eks_control_plane = optional(bool, false)
      ca_bundle         = optional(string, "")
    })
    aws = optional(object({
      irsa_role_arn              = optional(string, "")
      interruption_queue         = optional(string, "")
      isolated_vpc               = optional(bool, false)
      reserved_enis              = optional(number)
      enable_zonal_shift         = optional(bool, false)
      vm_memory_overhead_percent = optional(string)
    }))
    controller = optional(object({
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
      log_level = optional(string)
    }))
    batching = optional(object({
      max_duration  = optional(string)
      idle_duration = optional(string)
    }))
    scheduling = optional(object({
      preference_policy = optional(string)
      min_values_policy = optional(string)
    }))
    feature_gates = optional(object({
      node_repair                = optional(bool, false)
      node_overlay               = optional(bool, false)
      reserved_capacity          = optional(bool)
      spot_to_spot_consolidation = optional(bool, false)
      static_capacity            = optional(bool, false)
      capacity_buffer            = optional(bool, false)
    }))
    controller_scheduling = optional(object({
      priority_class_name = optional(string)
      node_selector       = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
      host_network = optional(bool, false)
    }))
    prometheus = optional(object({
      service_monitor = optional(bool, false)
    }))
    helm_values = optional(string, "")
  })
}