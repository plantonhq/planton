variable "metadata" {
  description = "Metadata for the GCP GKE Cluster resource"
  type = object({
    name = string
    id   = string
    org  = string
    env = object({
      id = string
    })
  })
}

variable "spec" {
  description = "Specification for the GCP GKE Cluster"
  type = object({
    project_id = object({
      value = string
    })
    network_self_link = object({
      value = string
    })
    location                     = string
    subnetwork_self_link        = object({
      value = string
    })
    cluster_secondary_range_name = object({
      value = string
    })
    services_secondary_range_name = object({
      value = string
    })
    master_ipv4_cidr_block      = string
    enable_public_nodes         = optional(bool, false)
    release_channel             = optional(number, 2) # 0=unspecified, 1=RAPID, 2=REGULAR, 3=STABLE, 4=NONE
    disable_network_policy      = optional(bool, false)
    disable_workload_identity   = optional(bool, false)
    router_nat_name             = object({
      value = string
    })
    cluster_name = string
  })
  
  validation {
    condition     = can(regex("^[a-z]([a-z0-9-]{0,38}[a-z0-9])?$", var.spec.cluster_name))
    error_message = "Cluster name must be 1-40 characters, lowercase letters, numbers, or hyphens, starting with a letter and ending with a letter or number."
  }
}
