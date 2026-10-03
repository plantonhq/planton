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
  description = "AwsCloudwatchLogAccountPolicy specification"
  type = object({
    region = string
    policy_name = string
    policy_type = optional(string, "")
    policy_document = any
    selection_criteria = optional(string, "")
  })
}