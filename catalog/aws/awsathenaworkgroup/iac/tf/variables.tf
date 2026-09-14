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
  description = "AwsAthenaWorkgroup specification"
  type = object({
    region = string
    description = optional(string, "")
    state = optional(string)
    result_configuration = optional(object({
      output_location = optional(string, "")
      encryption_option = optional(string, "")
      kms_key_arn = optional(string, "")
      expected_bucket_owner = optional(string, "")
      s3_acl_option = optional(string, "")
    }))
    managed_query_results = optional(object({
      kms_key = optional(string, "")
      enabled = optional(bool)
    }))
    bytes_scanned_cutoff_per_query = optional(number, 0)
    enforce_workgroup_configuration = optional(bool)
    publish_cloudwatch_metrics_enabled = optional(bool)
    requester_pays_enabled = optional(bool, false)
    enable_minimum_encryption_configuration = optional(bool, false)
    selected_engine_version = optional(string, "")
    execution_role = optional(string, "")
    customer_content_encryption_kms_key = optional(string, "")
    identity_center = optional(object({
      enable_identity_center = optional(bool, false)
      identity_center_instance_arn = optional(string, "")
    }))
    s3_access_grants = optional(object({
      enable_s3_access_grants = optional(bool, false)
      authentication_type = optional(string, "")
      create_user_level_prefix = optional(bool, false)
    }))
    monitoring = optional(object({
      cloud_watch_logging = optional(object({
        log_group = optional(string, "")
        log_stream_name_prefix = optional(string, "")
        log_types = optional(list(object({
          key = string
          values = list(string)
        })), [])
        enabled = optional(bool)
      }))
      managed_logging = optional(object({
        kms_key = optional(string, "")
        enabled = optional(bool)
      }))
      s3_logging = optional(object({
        log_location = optional(string, "")
        kms_key = optional(string, "")
        enabled = optional(bool)
      }))
    }))
    force_destroy = optional(bool, false)
  })
}
