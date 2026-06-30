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
  description = "HetznerCloudPlacementGroup specification"
  type = object({
    type = optional(string)
  })
}

variable "hcloud_token" {
  description = "Hetzner Cloud API token for authentication"
  type        = string
  sensitive   = true
}
