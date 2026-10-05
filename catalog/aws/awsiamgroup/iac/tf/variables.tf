variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsIamGroup specification"
  type = object({
    region              = string
    path                = optional(string, "")
    users               = optional(list(string), [])
    managed_policy_arns = optional(list(string), [])
    inline_policies     = optional(any, {})
  })
}