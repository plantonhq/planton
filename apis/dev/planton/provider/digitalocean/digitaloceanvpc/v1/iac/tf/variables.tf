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
  description = "DigitalOcean VPC specification"
  type = object({
    description            = optional(string, "")
    region                 = string
    ip_range_cidr          = optional(string, "")  # Optional: auto-generate if omitted (80/20 principle)
    is_default_for_region  = optional(bool, false)
  })
}

variable "digitalocean_token" {
  description = "DigitalOcean API token for authentication"
  type        = string
  sensitive   = true
}