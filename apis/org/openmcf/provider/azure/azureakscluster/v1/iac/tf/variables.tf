variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for Azure AKS Cluster"
  type = object({
    # Required fields
    region         = string
    resource_group = string
    vnet_subnet_id = string

    # Kubernetes version
    kubernetes_version = optional(string, "1.30")

    # Control plane SKU (STANDARD or FREE)
    control_plane_sku = optional(string, "STANDARD")

    # Network configuration
    network_plugin      = optional(string, "AZURE_CNI")
    network_plugin_mode = optional(string, "OVERLAY")

    # API server access
    private_cluster_enabled = optional(bool, false)
    authorized_ip_ranges    = optional(list(string), [])

    # Azure AD integration
    disable_azure_ad_rbac = optional(bool, false)

    # System node pool configuration (required)
    system_node_pool = object({
      vm_size = string
      autoscaling = object({
        min_count = number
        max_count = number
      })
      availability_zones = list(string)
    })

    # User node pools (optional)
    user_node_pools = optional(list(object({
      name    = string
      vm_size = string
      autoscaling = object({
        min_count = number
        max_count = number
      })
      availability_zones = list(string)
      spot_enabled       = optional(bool, false)
    })), [])

    # Add-ons configuration (optional)
    addons = optional(object({
      enable_container_insights   = optional(bool, false)
      enable_key_vault_csi_driver = optional(bool, true)
      enable_azure_policy         = optional(bool, true)
      enable_workload_identity    = optional(bool, true)
      log_analytics_workspace_id  = optional(string, "")
    }), {
      enable_container_insights   = false
      enable_key_vault_csi_driver = true
      enable_azure_policy         = true
      enable_workload_identity    = true
      log_analytics_workspace_id  = ""
    })

    # Advanced networking (optional)
    advanced_networking = optional(object({
      pod_cidr           = optional(string, "")
      service_cidr       = optional(string, "")
      dns_service_ip     = optional(string, "")
      custom_dns_servers = optional(list(string), [])
    }), {
      pod_cidr           = ""
      service_cidr       = ""
      dns_service_ip     = ""
      custom_dns_servers = []
    })
  })
}