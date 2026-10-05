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
  description = "CloudflareZeroTrustList specification"
  type = object({
    account_id = string
    name = string
    type = string
    description = optional(string, "")
    items = optional(list(object({
      value = string
      description = optional(string, "")
    })), [])
  })
}