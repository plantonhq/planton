variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)"
  type = object({
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
    id   = optional(string, "")
  })
}

variable "spec" {
  description = "AwsAthenaWorkgroup spec"
  type = object({
    result_configuration = optional(object({
      output_location        = optional(string, "")
      encryption_option      = optional(string, "")
      kms_key_arn            = optional(string, "")
      expected_bucket_owner  = optional(string, "")
      s3_acl_option          = optional(string, "")
    }), null)
    bytes_scanned_cutoff_per_query          = optional(number, 0)
    enforce_workgroup_configuration         = optional(bool, true)
    publish_cloudwatch_metrics_enabled      = optional(bool, true)
    requester_pays_enabled                  = optional(bool, false)
    enable_minimum_encryption_configuration = optional(bool, false)
    selected_engine_version                 = optional(string, "")
    force_destroy                           = optional(bool, false)
    execution_role                          = optional(string, "")
  })
}
