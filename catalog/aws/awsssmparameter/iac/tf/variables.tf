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
  description = "AwsSsmParameter specification"
  type = object({
    region = string
    parameter_name = string
    type = string
    value = optional(string, "")
    secure_value = optional(string, "")
    description = optional(string, "")
    allowed_pattern = optional(string, "")
    tier = optional(string, "")
    key_id = optional(string, "")
    data_type = optional(string, "")
    overwrite = optional(bool, false)
  })
}
