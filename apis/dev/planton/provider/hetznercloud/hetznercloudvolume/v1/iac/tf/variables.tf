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
  description = "HetznerCloudVolume specification"
  type = object({
    size              = number
    location          = string
    format            = optional(string)
    server_id         = optional(string)
    automount         = optional(bool)
    delete_protection = optional(bool)
  })
}

variable "hcloud_token" {
  description = "Hetzner Cloud API token for authentication"
  type        = string
  sensitive   = true
}
