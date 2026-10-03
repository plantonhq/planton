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
  description = "KubernetesCilium specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    chart_version    = optional(string)
    cluster_name     = optional(string)

    ipam = optional(object({
      mode                        = optional(string)
      cluster_pool_ipv4_pod_cidrs = optional(list(string), [])
      cluster_pool_ipv4_mask_size = optional(number)
    }))

    routing = optional(object({
      mode                     = optional(string)
      tunnel_protocol          = optional(string)
      ipv4_native_routing_cidr = optional(string, "")
      auto_direct_node_routes  = optional(bool, false)
    }))

    kube_proxy_replacement = optional(bool, false)
    k8s_service_host       = optional(string, "")
    k8s_service_port       = optional(number)

    cni = optional(object({
      chaining_mode   = optional(string)
      chaining_target = optional(string, "")
      exclusive       = optional(bool)
    }))

    cloud = optional(object({
      aws_eni    = optional(bool, false)
      aks_byocni = optional(bool, false)
      gke        = optional(bool, false)
    }))

    hubble = optional(object({
      enabled                 = optional(bool)
      relay                   = optional(bool, false)
      ui                      = optional(bool, false)
      metrics                 = optional(list(string), [])
      metrics_service_monitor = optional(bool, false)
    }))

    encryption = optional(object({
      enabled         = optional(bool, false)
      type            = optional(string)
      node_encryption = optional(bool, false)
    }))

    policy_enforcement_mode = optional(string)
    gateway_api             = optional(bool, false)

    bandwidth_manager = optional(object({
      enabled = optional(bool, false)
      bbr     = optional(bool, false)
    }))

    operator = optional(object({
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
    }))

    agent_resources = optional(object({
      limits = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
      requests = optional(object({
        cpu    = optional(string, "")
        memory = optional(string, "")
      }))
    }))

    prometheus = optional(object({
      enabled         = optional(bool, false)
      service_monitor = optional(bool, false)
    }))

    helm_values = optional(string, "")
  })
}
