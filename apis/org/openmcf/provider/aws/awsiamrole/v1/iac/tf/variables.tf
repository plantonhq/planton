variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AwsIamRole specification"
  type = object({
    region = string
    description = optional(string, "")
    path = optional(string, "")
    trust_policy = any
    managed_policy_arns = optional(list(string), [])
    inline_policies = optional(map(any), {})
  })
}
