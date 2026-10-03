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
  description = "AwsSagemakerPipeline specification"
  type = object({
    region = string
    display_name = optional(string, "")
    description = optional(string, "")
    role_arn = string
    definition = optional(any)
    definition_s3_location = optional(object({
      bucket = string
      object_key = string
      version_id = optional(string, "")
    }))
    parallelism_max_steps = optional(number)
  })
}