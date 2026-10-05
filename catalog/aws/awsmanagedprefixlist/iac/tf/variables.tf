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
  description = "AwsManagedPrefixList specification"
  type = object({
    region = string
    address_family = optional(string, "")
    max_entries = optional(number, 0)
    entries = optional(list(object({
      cidr = string
      description = optional(string, "")
    })), [])
  })
}