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
  description = "AwsSecretsManagerSecret specification"
  type = object({
    region              = string
    secret_name         = optional(string, "")
    description         = optional(string, "")
    kms_key_id          = optional(string, "")
    string_value        = optional(string, "")
    binary_value        = optional(string, "")
    version_stages      = optional(list(string), [])
    policy              = optional(any)
    block_public_policy = optional(bool)
    replica_regions = optional(list(object({
      region     = string
      kms_key_id = optional(string, "")
    })), [])
    force_overwrite_replica_secret = optional(bool, false)
    recovery_window_in_days        = optional(number)
    type                           = optional(string, "")
    rotation = optional(object({
      rotation_lambda_arn        = optional(string, "")
      external_rotation_role_arn = optional(string, "")
      external_rotation_metadata = optional(list(object({
        key   = string
        value = string
      })), [])
      automatically_after_days = optional(number)
      schedule_expression      = optional(string, "")
      duration                 = optional(string, "")
      rotate_immediately       = optional(bool)
    }))
  })
}