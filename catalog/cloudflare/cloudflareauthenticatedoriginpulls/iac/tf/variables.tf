variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareAuthenticatedOriginPulls specification"
  type = object({
    zone_id = string
    zone_enabled = optional(bool)
    hostname_associations = optional(list(object({
      hostname = string
      certificate_id = optional(string, "")
      enabled = optional(bool)
    })), [])
  })
}
