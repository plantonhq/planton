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
  description = "CloudflareIpAccessRule specification"
  type = object({
    account_id = optional(string, "")
    zone_id = optional(string, "")
    mode = string
    configuration = object({
      target = string
      value = string
    })
    notes = optional(string, "")
  })
}