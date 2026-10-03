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
  description = "AwsSagemakerEndpoint specification"
  type = object({
    region = string
    production_variants = list(object({
      variant_name = optional(string, "")
      model = optional(string, "")
      instance_type = optional(string, "")
      initial_instance_count = optional(number)
      initial_variant_weight = optional(number)
      serverless = optional(object({
        max_concurrency = optional(number, 0)
        memory_size_mb = optional(number, 0)
        provisioned_concurrency = optional(number)
      }))
      managed_instance_scaling = optional(object({
        status = optional(string, "")
        min_instance_count = optional(number)
        max_instance_count = optional(number)
      }))
      routing_strategy = optional(string, "")
      volume_size_gb = optional(number)
      container_startup_health_check_timeout_seconds = optional(number)
      model_data_download_timeout_seconds = optional(number)
      enable_ssm_access = optional(bool, false)
      inference_ami_version = optional(string, "")
      accelerator_type = optional(string, "")
      core_dump = optional(object({
        destination_s3_uri = string
        kms_key_arn = optional(string, "")
      }))
      ml_capacity_reservation_arn = optional(string, "")
    }))
    shadow_variants = optional(list(object({
      variant_name = optional(string, "")
      model = optional(string, "")
      instance_type = optional(string, "")
      initial_instance_count = optional(number)
      initial_variant_weight = optional(number)
      serverless = optional(object({
        max_concurrency = optional(number, 0)
        memory_size_mb = optional(number, 0)
        provisioned_concurrency = optional(number)
      }))
      managed_instance_scaling = optional(object({
        status = optional(string, "")
        min_instance_count = optional(number)
        max_instance_count = optional(number)
      }))
      routing_strategy = optional(string, "")
      volume_size_gb = optional(number)
      container_startup_health_check_timeout_seconds = optional(number)
      model_data_download_timeout_seconds = optional(number)
      enable_ssm_access = optional(bool, false)
      inference_ami_version = optional(string, "")
      accelerator_type = optional(string, "")
      core_dump = optional(object({
        destination_s3_uri = string
        kms_key_arn = optional(string, "")
      }))
      ml_capacity_reservation_arn = optional(string, "")
    })), [])
    kms_key_arn = optional(string, "")
    execution_role_arn = optional(string, "")
    async_inference = optional(object({
      output_s3_path = string
      failure_s3_path = optional(string, "")
      kms_key_arn = optional(string, "")
      max_concurrent_invocations_per_instance = optional(number)
      success_topic_arn = optional(string, "")
      error_topic_arn = optional(string, "")
      include_inference_response_in = optional(list(string), [])
    }))
    data_capture = optional(object({
      destination_s3_uri = string
      initial_sampling_percentage = optional(number, 0)
      capture_modes = list(string)
      enable_capture = optional(bool, false)
      csv_content_types = optional(list(string), [])
      json_content_types = optional(list(string), [])
      kms_key_arn = optional(string, "")
    }))
    deployment = optional(object({
      blue_green = optional(object({
        traffic_routing_type = optional(string, "")
        wait_interval_seconds = optional(number, 0)
        canary_size = optional(object({
          type = optional(string, "")
          value = optional(number, 0)
        }))
        linear_step_size = optional(object({
          type = optional(string, "")
          value = optional(number, 0)
        }))
        termination_wait_seconds = optional(number)
        maximum_execution_timeout_seconds = optional(number)
      }))
      rolling = optional(object({
        maximum_batch_size = object({
          type = optional(string, "")
          value = optional(number, 0)
        })
        wait_interval_seconds = optional(number, 0)
        rollback_maximum_batch_size = optional(object({
          type = optional(string, "")
          value = optional(number, 0)
        }))
        maximum_execution_timeout_seconds = optional(number)
      }))
      auto_rollback_alarm_names = optional(list(string), [])
    }))
  })
}