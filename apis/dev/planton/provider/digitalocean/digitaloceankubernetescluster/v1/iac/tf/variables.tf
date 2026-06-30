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
  description = "DigitalOcean Kubernetes Cluster specification"
  type = object({
    cluster_name     = string
    region           = string
    kubernetes_version = string
    vpc              = object({
      value = string
    })
    highly_available           = optional(bool, false)
    auto_upgrade               = optional(bool, false)
    disable_surge_upgrade      = optional(bool, false)
    maintenance_window         = optional(string, "")
    registry_integration       = optional(bool, false)
    control_plane_firewall_allowed_ips = optional(list(string), [])
    tags                       = optional(list(string), [])
    default_node_pool = object({
      size       = string
      node_count = number
      auto_scale = optional(bool, false)
      min_nodes  = optional(number, 0)
      max_nodes  = optional(number, 0)
    })
  })
}

variable "digitalocean_token" {
  description = "DigitalOcean API token for authentication"
  type        = string
  sensitive   = true
}