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
  description = "AwsSagemakerImage specification"
  type = object({
    region = string
    role_arn = string
    display_name = optional(string, "")
    description = optional(string, "")
    versions = optional(list(object({
      base_image = string
      aliases = optional(list(string), [])
      horovod = optional(bool, false)
      job_type = optional(string, "")
      ml_framework = optional(string, "")
      processor = optional(string, "")
      programming_lang = optional(string, "")
      release_notes = optional(string, "")
      vendor_guidance = optional(string, "")
    })), [])
  })
}