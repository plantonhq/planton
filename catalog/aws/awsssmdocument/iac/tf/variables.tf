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
  description = "AwsSsmDocument specification"
  type = object({
    region = string
    content = string
    document_type = string
    document_format = optional(string, "")
    target_type = optional(string, "")
    version_name = optional(string, "")
    attachment_sources = optional(list(object({
      key = string
      name = optional(string, "")
      values = list(string)
    })), [])
    share_with_account_ids = optional(list(string), [])
  })
}
