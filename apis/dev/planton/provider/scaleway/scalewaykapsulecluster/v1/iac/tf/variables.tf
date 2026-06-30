variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Scaleway Kapsule Cluster specification"
  type = object({
    # Region where the cluster will be created (e.g., "fr-par")
    region = string

    # Kubernetes version (e.g., "1.32" or "1.32.3")
    kubernetes_version = string

    # CNI plugin: "cilium" or "calico"
    cni = string

    # Private Network ID (resolved from StringValueOrRef before Terraform runs)
    private_network_id = string

    # Cluster type: "kapsule", "kapsule-dedicated-4/8/16"
    type = optional(string, "kapsule")

    # Human-readable description
    description = optional(string, "")

    # Delete K8s-created resources (LBs, volumes) on cluster deletion
    delete_additional_resources = optional(bool, true)

    # Auto-upgrade configuration (null = disabled)
    auto_upgrade = optional(object({
      enable                         = bool
      maintenance_window_start_hour  = number
      maintenance_window_day         = string
    }))

    # Cluster-wide autoscaler configuration (null = defaults)
    autoscaler_config = optional(object({
      disable_scale_down               = optional(bool, false)
      scale_down_delay_after_add       = optional(string)
      scale_down_unneeded_time         = optional(string)
      estimator                        = optional(string)
      expander                         = optional(string)
      scale_down_utilization_threshold = optional(number)
      max_graceful_termination_sec     = optional(number)
      ignore_daemonsets_utilization    = optional(bool, false)
      balance_similar_node_groups      = optional(bool, false)
      expendable_pods_priority_cutoff  = optional(number)
    }))

    # Kubernetes feature gates to enable
    feature_gates = optional(list(string), [])

    # Kubernetes admission plugins to enable
    admission_plugins = optional(list(string), [])

    # Pod CIDR (ForceNew, default: "100.64.0.0/15")
    pod_cidr = optional(string)

    # Service CIDR (ForceNew, default: "10.32.0.0/20")
    service_cidr = optional(string)

    # Default node pool configuration
    default_node_pool = object({
      name                 = optional(string, "")
      node_type            = string
      size                 = number
      auto_scale           = optional(bool, false)
      min_size             = optional(number, 1)
      max_size             = optional(number, 1)
      autohealing          = optional(bool, false)
      container_runtime    = optional(string, "containerd")
      root_volume_type     = optional(string)
      root_volume_size_in_gb = optional(number)
      public_ip_disabled   = optional(bool, false)
      upgrade_policy = optional(object({
        max_surge       = optional(number, 0)
        max_unavailable = optional(number, 1)
      }))
    })
  })
}

variable "scaleway_access_key" {
  description = "Scaleway access key for API authentication"
  type        = string
  sensitive   = true
}

variable "scaleway_secret_key" {
  description = "Scaleway secret key for API authentication"
  type        = string
  sensitive   = true
}

variable "scaleway_project_id" {
  description = "Scaleway project ID (optional, defaults from provider)"
  type        = string
  default     = ""
}

variable "scaleway_organization_id" {
  description = "Scaleway organization ID (optional, defaults from provider)"
  type        = string
  default     = ""
}
