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
  description = "AwsIamAccountSettings specification"
  type = object({
    region = string
    account_alias = optional(string, "")
    password_policy = optional(object({
      minimum_password_length = optional(number)
      require_lowercase_characters = optional(bool, false)
      require_numbers = optional(bool, false)
      require_symbols = optional(bool, false)
      require_uppercase_characters = optional(bool, false)
      allow_users_to_change_password = optional(bool)
      max_password_age = optional(number)
      password_reuse_prevention = optional(number)
      hard_expiry = optional(bool, false)
    }))
    sts = optional(object({
      global_endpoint_token_version = optional(string, "")
    }))
  })
}