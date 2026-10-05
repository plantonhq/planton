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
  description = "AwsBedrockProvisionedThroughput specification"
  type = object({
    region = string
    model_arn = string
    model_units = optional(number, 0)
    commitment_duration = optional(string, "")
  })
}