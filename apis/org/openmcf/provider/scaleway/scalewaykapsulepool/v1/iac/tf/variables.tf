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
  description = "Scaleway Kapsule Pool specification"
  type = object({
    # Region where the pool will be created (must match cluster region)
    region = string

    # Cluster ID (resolved from StringValueOrRef before Terraform runs)
    cluster_id = string

    # Instance type for worker nodes
    node_type = string

    # Number of nodes
    size = number

    # Autoscaling configuration
    auto_scale = optional(bool, false)
    min_size   = optional(number, 1)
    max_size   = optional(number, 1)

    # Node health
    autohealing = optional(bool, false)

    # Container runtime
    container_runtime = optional(string, "containerd")

    # Root volume configuration
    root_volume_type       = optional(string)
    root_volume_size_in_gb = optional(number)

    # Network isolation
    public_ip_disabled = optional(bool, false)

    # Zone placement (optional)
    zone = optional(string)

    # Anti-affinity placement group (optional)
    placement_group_id = optional(string)

    # Kubernetes labels for node scheduling
    kubernetes_labels = optional(map(string), {})

    # Kubernetes taints for workload isolation
    taints = optional(list(object({
      key    = string
      value  = optional(string, "")
      effect = string
    })), [])

    # Upgrade policy (optional)
    upgrade_policy = optional(object({
      max_surge       = optional(number, 0)
      max_unavailable = optional(number, 1)
    }))

    # Custom kubelet arguments (power-user escape hatch)
    kubelet_args = optional(map(string), {})
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
