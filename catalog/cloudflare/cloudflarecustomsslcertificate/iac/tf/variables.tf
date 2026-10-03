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
  description = "CloudflareCustomSslCertificate specification"
  type = object({
    zone_id = string
    certificate = string
    private_key = string
    type = optional(string)
    bundle_method = optional(string)
    policy = optional(string, "")
    geo_restrictions = optional(object({
      label = optional(string)
    }))
    custom_csr_id = optional(string, "")
    deploy = optional(string)
  })
}
