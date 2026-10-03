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
  description = "DigitalOceanProject specification"
  type = object({
    project_name = string
    description = optional(string, "")
    purpose = optional(string, "")
    environment = optional(string, "")
    is_default = optional(bool, false)
    resources = optional(list(string), [])
  })
}
