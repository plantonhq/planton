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
  description = "AwsLambdaLayer specification"
  type = object({
    region = string
    code = object({
      bucket = string
      key = string
      version = optional(string, "")
    })
    source_code_hash = optional(string, "")
    description = optional(string, "")
    compatible_runtimes = optional(list(string), [])
    compatible_architectures = optional(list(string), [])
    license_info = optional(string, "")
    skip_destroy = optional(bool, false)
    permissions = optional(list(object({
      statement_id = string
      principal = optional(string, "")
      organization_id = optional(string, "")
      skip_destroy = optional(bool, false)
    })), [])
  })
}