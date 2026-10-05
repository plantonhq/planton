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
  description = "CloudflareAuthenticatedOriginPullsCertificate specification"
  type = object({
    zone_id = string
    scope = optional(string)
    certificate = string
    private_key = string
  })
}
