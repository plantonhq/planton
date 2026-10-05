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
  description = "AwsSsmMaintenanceWindow specification"
  type = object({
    region = string
    schedule = string
    duration = number
    cutoff = number
    description = optional(string, "")
    enabled = optional(bool)
    allow_unassociated_targets = optional(bool, false)
    schedule_timezone = optional(string, "")
    schedule_offset = optional(number, 0)
    start_date = optional(string, "")
    end_date = optional(string, "")
    targets = optional(list(object({
      name = string
      resource_type = string
      targets = list(object({
        key = string
        values = list(string)
      }))
      description = optional(string, "")
      owner_information = optional(string, "")
    })), [])
    tasks = optional(list(object({
      name = string
      task_type = string
      task_arn = string
      description = optional(string, "")
      service_role_arn = optional(string, "")
      priority = optional(number, 0)
      max_concurrency = optional(string, "")
      max_errors = optional(string, "")
      cutoff_behavior = optional(string, "")
      targets = optional(list(object({
        key = string
        values = list(string)
      })), [])
      invocation = optional(object({
        run_command = optional(object({
          comment = optional(string, "")
          document_hash = optional(string, "")
          document_hash_type = optional(string, "")
          document_version = optional(string, "")
          output_s3_bucket = optional(string, "")
          output_s3_key_prefix = optional(string, "")
          service_role_arn = optional(string, "")
          timeout_seconds = optional(number, 0)
          parameters = optional(list(object({
            name = string
            values = list(string)
          })), [])
          cloudwatch_config = optional(object({
            cloudwatch_log_group_name = optional(string, "")
            cloudwatch_output_enabled = optional(bool, false)
          }))
          notification_config = optional(object({
            notification_arn = optional(string, "")
            notification_events = optional(list(string), [])
            notification_type = optional(string, "")
          }))
        }))
        automation = optional(object({
          document_version = optional(string, "")
          parameters = optional(list(object({
            name = string
            values = list(string)
          })), [])
        }))
        lambda = optional(object({
          client_context = optional(string, "")
          payload = optional(string, "")
          qualifier = optional(string, "")
        }))
        step_functions = optional(object({
          input = optional(string, "")
          name = optional(string, "")
        }))
      }))
    })), [])
  })
}
