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
  description = "DigitalOcean Droplet specification"
  type = object({
    droplet_name        = string
    region              = string
    size                = string
    image               = string
    ssh_keys            = list(string)
    vpc                 = optional(object({
      value                       = optional(string)
      value_from_resource_output = optional(object({
        resource_id_ref = object({
          name = string
        })
        output_key = string
      }))
    }))
    enable_ipv6        = optional(bool)
    enable_backups     = optional(bool)
    disable_monitoring = optional(bool)
    volume_ids         = optional(list(object({
      value                       = optional(string)
      value_from_resource_output = optional(object({
        resource_id_ref = object({
          name = string
        })
        output_key = string
      }))
    })))
    tags      = optional(list(string))
    user_data = optional(string)
    timezone  = optional(string)
  })
}

variable "digitalocean_token" {
  description = "DigitalOcean API token for authentication"
  type        = string
  sensitive   = true
}