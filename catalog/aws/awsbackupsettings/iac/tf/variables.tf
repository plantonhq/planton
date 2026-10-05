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
  description = "AwsBackupSettings specification"
  type = object({
    region = string
    global = optional(object({
      settings = optional(map(string), {})
    }))
    region_settings = optional(object({
      resource_type_opt_in_preference = optional(map(bool), {})
      resource_type_management_preference = optional(map(bool), {})
    }))
  })
}