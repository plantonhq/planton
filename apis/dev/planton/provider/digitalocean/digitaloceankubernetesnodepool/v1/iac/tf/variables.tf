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
  description = "DigitalOcean Kubernetes Node Pool specification"
  type = object({
    node_pool_name = string
    cluster        = object({
      value = string
    })
    size       = string
    node_count = number
    auto_scale = optional(bool, false)
    min_nodes  = optional(number, 0)
    max_nodes  = optional(number, 0)
    labels     = optional(map(string), {})
    taints     = optional(list(object({
      key    = string
      value  = string
      effect = string
    })), [])
    tags = optional(list(string), [])
  })
}

variable "digitalocean_token" {
  description = "DigitalOcean API token for authentication"
  type        = string
  sensitive   = true
}