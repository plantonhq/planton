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
  description = "DigitalOceanDnsRecord specification"
  type = object({
    domain = string
    name = string
    type = string
    value = string
    ttl_seconds = optional(number)
    priority = optional(number)
    weight = optional(number)
    port = optional(number)
    flags = optional(number)
    tag = optional(string, "")
  })
}