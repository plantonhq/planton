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
  description = "DigitalOceanCdn specification"
  type = object({
    origin = string
    ttl = optional(number)
    certificate = optional(string, "")
    custom_domain = optional(string, "")
  })
}
