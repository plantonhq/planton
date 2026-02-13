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
  description = "Specification for Azure AKS Node Pool"
  type = object({
    # Required fields
    cluster_name       = string
    vm_size            = string
    initial_node_count = number

    # Optional autoscaling
    autoscaling = optional(object({
      min_nodes = number
      max_nodes = number
    }))

    # Optional availability zones (min 2 if specified)
    availability_zones = optional(list(string), [])

    # Optional OS type (defaults to Linux)
    os_type = optional(string, "LINUX")

    # Optional mode (defaults to User)
    mode = optional(string, "USER")

    # Optional Spot instances
    spot_enabled = optional(bool, false)

    # The Azure Resource Group where the node pool's parent cluster resides
    resource_group = string
  })
}
