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
  description = "AwsBackupFramework specification"
  type = object({
    region = string
    framework_name = string
    description = optional(string, "")
    controls = list(object({
      name = string
      input_parameters = optional(list(object({
        name = string
        value = string
      })), [])
      scope = optional(object({
        compliance_resource_ids = optional(list(string), [])
        compliance_resource_types = optional(list(string), [])
        tags = optional(map(string), {})
      }))
    }))
  })
}