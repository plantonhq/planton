variable "provider_config" {
  description = "AWS provider configuration"
  type = object({
    region            = string
    access_key_id     = optional(string)
    secret_access_key = optional(string)
    session_token     = optional(string)
  })
}

variable "metadata" {
  description = "Resource metadata"
  type = object({
    id   = string
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
  })
}

variable "spec" {
  description = "AwsMwaaEnvironmentSpec - desired state of the MWAA environment"
  type = object({
    # The AWS region where the MWAA environment will be created.
    region = string

    airflow_version                  = optional(string, "")
    airflow_configuration_options    = optional(map(string), {})

    source_bucket_arn                = string
    dag_s3_path                      = string
    plugins_s3_path                  = optional(string, "")
    plugins_s3_object_version        = optional(string, "")
    requirements_s3_path             = optional(string, "")
    requirements_s3_object_version   = optional(string, "")
    startup_script_s3_path           = optional(string, "")
    startup_script_s3_object_version = optional(string, "")

    execution_role_arn = string

    subnet_ids                   = list(string)
    security_group_ids           = optional(list(string), [])
    allowed_cidr_blocks          = optional(list(string), [])
    associate_security_group_ids = optional(list(string), [])
    vpc_id                       = optional(string, "")

    kms_key_arn = optional(string, "")

    environment_class = optional(string, "")
    min_workers       = optional(number, 0)
    max_workers       = optional(number, 0)
    min_webservers    = optional(number, 0)
    max_webservers    = optional(number, 0)
    schedulers        = optional(number, 0)

    webserver_access_mode = optional(string, "PRIVATE_ONLY")
    endpoint_management   = optional(string, "")

    logging_configuration = optional(object({
      dag_processing_logs = optional(object({
        enabled   = bool
        log_level = optional(string, "INFO")
      }), null)
      scheduler_logs = optional(object({
        enabled   = bool
        log_level = optional(string, "INFO")
      }), null)
      task_logs = optional(object({
        enabled   = bool
        log_level = optional(string, "INFO")
      }), null)
      webserver_logs = optional(object({
        enabled   = bool
        log_level = optional(string, "INFO")
      }), null)
      worker_logs = optional(object({
        enabled   = bool
        log_level = optional(string, "INFO")
      }), null)
    }), null)

    weekly_maintenance_window_start = optional(string, "")
    worker_replacement_strategy     = optional(string, "")
  })
}
