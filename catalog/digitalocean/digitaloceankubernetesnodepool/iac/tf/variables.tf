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
  description = "DigitalOceanKubernetesNodePool specification"
  type = object({
    node_pool_name = string
    cluster = string
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
}
