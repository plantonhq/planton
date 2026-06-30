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
  description = "Specification for the GCP GKE Node Pool"
  type = object({
    # Required: Parent cluster references
    cluster_project_id = object({
      value = string
      ref   = optional(object({
        kind  = optional(string)
        name  = optional(string)
        path  = optional(string)
        value = optional(string)
      }))
    })

    cluster_name = object({
      value = string
      ref   = optional(object({
        kind  = optional(string)
        name  = optional(string)
        path  = optional(string)
        value = optional(string)
      }))
    })

    # Machine configuration
    machine_type = optional(string, "e2-medium")
    disk_size_gb = optional(number, 100)
    disk_type    = optional(string, "pd-standard")
    image_type   = optional(string, "COS_CONTAINERD")

    # Security
    service_account = optional(string, "")

    # Node management
    management = optional(object({
      disable_auto_upgrade = optional(bool, false)
      disable_auto_repair  = optional(bool, false)
    }), {})

    # Cost optimization
    spot = optional(bool, false)

    # Kubernetes labels
    node_labels = optional(map(string), {})

    # Name of the node pool
    node_pool_name = string

    # Scaling configuration (mutually exclusive)
    node_count = optional(number)

    autoscaling = optional(object({
      min_nodes       = number
      max_nodes       = number
      location_policy = optional(string, "BALANCED")
    }))
  })

  validation {
    condition     = (var.spec.node_count != null) != (var.spec.autoscaling != null)
    error_message = "Exactly one of node_count or autoscaling must be specified"
  }

  validation {
    condition     = contains(["pd-standard", "pd-ssd", "pd-balanced"], var.spec.disk_type)
    error_message = "disk_type must be one of: pd-standard, pd-ssd, pd-balanced"
  }
}
