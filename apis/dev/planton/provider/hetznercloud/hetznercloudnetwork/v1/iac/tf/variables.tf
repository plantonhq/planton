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
  description = "HetznerCloudNetwork specification"
  type = object({
    ip_range = string
    subnets = list(object({
      type         = string
      network_zone = string
      ip_range     = string
      vswitch_id   = optional(number)
    }))
    routes = optional(list(object({
      destination = string
      gateway     = string
    })))
    delete_protection         = optional(bool)
    expose_routes_to_vswitch  = optional(bool)
  })
}

variable "hcloud_token" {
  description = "Hetzner Cloud API token for authentication"
  type        = string
  sensitive   = true
}
