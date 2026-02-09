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
  description = "OpenStackContainerClusterTemplateSpec"
  type = object({
    coe   = string
    image = object({ value = string })

    keypair          = optional(object({ value = string }))
    external_network = optional(object({ value = string }))
    fixed_network    = optional(object({ value = string }))
    fixed_subnet     = optional(object({ value = string }))

    network_driver     = optional(string, "")
    volume_driver      = optional(string, "")
    dns_nameserver     = optional(string, "")
    docker_volume_size = optional(number)
    flavor             = optional(string, "")
    master_flavor      = optional(string, "")
    floating_ip_enabled = optional(bool)
    master_lb_enabled   = optional(bool)
    tls_disabled        = optional(bool)
    labels = optional(map(string), {})
    region = optional(string, "")
  })
}
