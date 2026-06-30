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
  description = "Specification for the DigitalOcean Container Registry"
  type = object({
    name                        = string
    subscription_tier           = string # "STARTER", "BASIC", or "PROFESSIONAL"
    region                      = string # DigitalOcean region (e.g., "nyc3", "sfo3")
    garbage_collection_enabled  = optional(bool, false)
  })
}
