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
  description = "AwsEventBridgeScheduler specification"
  type = object({
    region = string
    description = optional(string, "")
    group = optional(object({
      name = string
    }))
    group_name = optional(string, "")
    schedule_expression = string
    schedule_expression_timezone = optional(string, "")
    start_date = optional(string, "")
    end_date = optional(string, "")
    state = optional(string, "")
    action_after_completion = optional(string, "")
    kms_key_arn = optional(string, "")
    flexible_time_window = object({
      mode = optional(string, "")
      maximum_window_in_minutes = optional(number)
    })
    target = object({
      arn = string
      role_arn = string
      input = optional(string, "")
      dead_letter_queue_arn = optional(string, "")
      retry_policy = optional(object({
        maximum_event_age_in_seconds = optional(number)
        maximum_retry_attempts = optional(number)
      }))
      ecs_parameters = optional(object({
        task_definition_arn = string
        task_count = optional(number)
        launch_type = optional(string, "")
        capacity_provider_strategy = optional(list(object({
          capacity_provider = string
          base = optional(number, 0)
          weight = optional(number, 0)
        })), [])
        network_configuration = optional(object({
          subnets = list(string)
          security_groups = optional(list(string), [])
          assign_public_ip = optional(bool, false)
        }))
        group = optional(string, "")
        platform_version = optional(string, "")
        placement_constraints = optional(list(object({
          type = optional(string, "")
          expression = optional(string, "")
        })), [])
        placement_strategy = optional(list(object({
          type = optional(string, "")
          field = optional(string, "")
        })), [])
        propagate_tags = optional(string, "")
        tags = optional(map(string), {})
        enable_ecs_managed_tags = optional(bool, false)
        enable_execute_command = optional(bool, false)
        reference_id = optional(string, "")
      }))
      eventbridge_parameters = optional(object({
        detail_type = string
        source = string
      }))
      kinesis_parameters = optional(object({
        partition_key = string
      }))
      sagemaker_pipeline_parameters = optional(object({
        pipeline_parameters = optional(list(object({
          name = string
          value = string
        })), [])
      }))
      sqs_parameters = optional(object({
        message_group_id = optional(string, "")
      }))
    })
  })
}