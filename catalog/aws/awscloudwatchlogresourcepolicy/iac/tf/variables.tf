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
  description = "AwsCloudwatchLogResourcePolicy specification"
  type = object({
    region = string
    policy_name = optional(string, "")
    resource_arn = optional(string, "")
    policy_document = any
  })
}