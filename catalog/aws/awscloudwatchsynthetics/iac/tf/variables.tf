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
  description = "AwsCloudwatchSynthetics specification"
  type = object({
    region = string
    canary = optional(object({
      artifact_bucket = string
      artifact_prefix = optional(string, "")
      execution_role_arn = string
      handler = string
      runtime_version = string
      code = object({
        s3_bucket = string
        s3_key = string
        s3_version = optional(string, "")
      })
      schedule = object({
        expression = string
        duration_in_seconds = optional(number, 0)
        max_retries = optional(number)
      })
      run_config = optional(object({
        active_tracing = optional(bool, false)
        environment_variables = optional(map(string), {})
        ephemeral_storage = optional(number)
        memory_in_mb = optional(number)
        timeout_in_seconds = optional(number)
      }))
      vpc_config = optional(object({
        subnet_ids = list(string)
        security_group_ids = optional(list(string), [])
        ipv6_allowed_for_dual_stack = optional(bool, false)
      }))
      artifact_encryption_mode = optional(string, "")
      artifact_encryption_kms_key_arn = optional(string, "")
      failure_retention_period = optional(number)
      success_retention_period = optional(number)
      start_canary = optional(bool, false)
      delete_lambda = optional(bool, false)
    }))
    groups = optional(list(object({
      name = string
    })), [])
    group_names = optional(list(string), [])
  })
}