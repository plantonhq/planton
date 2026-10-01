variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "DigitalOceanKubernetesCluster specification"
  type = object({
    cluster_name = string
    region = string
    kubernetes_version = string
    vpc = string
    highly_available = optional(bool, false)
    auto_upgrade = optional(bool, false)
    registry_integration = optional(bool, false)
    tags = optional(list(string), [])
    default_node_pool = object({
      size = string
      node_count = optional(number, 0)
      auto_scale = optional(bool, false)
      min_nodes = optional(number, 0)
      max_nodes = optional(number, 0)
      labels = optional(map(string), {})
      taints = optional(list(object({
        key = string
        value = optional(string, "")
        effect = string
      })), [])
      tags = optional(list(string), [])
      gpu_partition_mode = optional(string, "")
    })
    surge_upgrade = optional(bool)
    maintenance_policy = optional(object({
      day = string
      start_time = string
    }))
    control_plane_firewall = optional(object({
      enabled = bool
      allowed_addresses = optional(list(string), [])
    }))
    cluster_subnet = optional(string, "")
    service_subnet = optional(string, "")
    worker_subnet_uuid = optional(string, "")
    isolated_workers = optional(bool, false)
    destroy_all_associated_resources = optional(bool, false)
    kubeconfig_expire_seconds = optional(number, 0)
    cluster_autoscaler_configuration = optional(object({
      scale_down_utilization_threshold = optional(number)
      scale_down_unneeded_time = optional(string, "")
      expanders = optional(list(string), [])
    }))
    sso = optional(object({
      enabled = bool
      required = optional(bool, false)
      issuer_url = optional(string, "")
      client_id = optional(string, "")
    }))
    routing_agent = optional(object({
      enabled = bool
    }))
    p2p_oci_registry_plugin = optional(object({
      enabled = bool
    }))
    amd_gpu_device_plugin = optional(object({
      enabled = bool
    }))
    amd_gpu_dra_driver = optional(object({
      enabled = bool
    }))
    amd_gpu_device_metrics_exporter_plugin = optional(object({
      enabled = bool
    }))
    nvidia_gpu_device_plugin = optional(object({
      enabled = bool
    }))
    nvidia_gpu_dra_driver = optional(object({
      enabled = bool
    }))
    rdma_shared_device_plugin = optional(object({
      enabled = bool
    }))
    coredns_autoscaler = optional(object({
      enabled = bool
    }))
  })
}
