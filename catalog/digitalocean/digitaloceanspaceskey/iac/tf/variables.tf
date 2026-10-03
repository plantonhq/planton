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
  description = "DigitalOceanSpacesKey specification"
  type = object({
    key_name = string
    grants = optional(list(object({
      bucket = optional(string, "")
      permission = string
    })), [])
  })
}
