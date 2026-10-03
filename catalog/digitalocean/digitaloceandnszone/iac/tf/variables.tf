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
  description = "DigitalOceanDnsZone specification"
  type = object({
    domain_name = string
    records = optional(list(object({
      name = string
      values = list(string)
      ttl_seconds = optional(number, 0)
      type = string
      priority = optional(number)
      weight = optional(number)
      port = optional(number)
      flags = optional(number)
      tag = optional(string, "")
    })), [])
    ip_address = optional(string, "")
  })
}