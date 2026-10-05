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
  description = "CloudflareSnippet specification"
  type = object({
    zone_id = string
    snippet_name = string
    files = list(object({
      name = string
      content = string
    }))
    main_module = string
  })
}