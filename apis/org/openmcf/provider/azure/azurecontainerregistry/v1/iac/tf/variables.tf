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
  description = "Specification for Azure Container Registry"
  type = object({
    # Required fields
    region         = string
    resource_group = string
    registry_name  = string

    # Optional fields
    sku                       = optional(string, "STANDARD")
    admin_user_enabled        = optional(bool, false)
    geo_replication_regions   = optional(list(string), [])
  })
}
