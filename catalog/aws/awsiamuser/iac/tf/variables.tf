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
  description = "AwsIamUser specification"
  type = object({
    region = string
    user_name = string
    path = optional(string, "")
    managed_policy_arns = optional(list(string), [])
    inline_policies = optional(any, {})
    permissions_boundary = optional(string, "")
    disable_access_keys = optional(bool, false)
    force_destroy = optional(bool, false)
    access_key_status = optional(string, "")
  })
}