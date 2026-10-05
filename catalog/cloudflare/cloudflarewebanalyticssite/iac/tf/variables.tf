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
  description = "CloudflareWebAnalyticsSite specification"
  type = object({
    account_id = string
    host = optional(string, "")
    zone_tag = optional(string, "")
    auto_install = optional(bool)
    enabled = optional(bool)
    lite = optional(bool)
    rules = optional(list(object({
      host = optional(string, "")
      paths = optional(list(string), [])
      inclusive = optional(bool)
      is_paused = optional(bool)
    })), [])
  })
}