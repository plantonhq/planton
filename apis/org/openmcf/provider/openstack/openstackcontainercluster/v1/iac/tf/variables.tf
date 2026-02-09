variable "metadata" {
  description = "Metadata for the resource"
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
  description = "OpenStackContainerClusterSpec"
  type = object({
    cluster_template    = object({ value = string })
    master_count        = optional(number)
    node_count          = optional(number)
    keypair             = optional(object({ value = string }))
    flavor              = optional(string, "")
    master_flavor       = optional(string, "")
    docker_volume_size  = optional(number)
    labels              = optional(map(string), {})
    create_timeout      = optional(number)
    floating_ip_enabled = optional(bool)
    region              = optional(string, "")
  })
}
