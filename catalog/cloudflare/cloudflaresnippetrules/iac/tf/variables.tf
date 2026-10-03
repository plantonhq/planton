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
  description = "CloudflareSnippetRules specification"
  type = object({
    zone_id = string
    rules = list(object({
      expression = string
      snippet_name = string
      description = optional(string, "")
      enabled = optional(bool)
    }))
  })
}