variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud Function Compute v3 function specification"
  type = object({
    region        = string
    function_name = string
    handler       = string
    runtime       = string
    description   = optional(string, "")

    cpu                  = optional(number)
    memory_size          = optional(number)
    timeout              = optional(number)
    disk_size            = optional(number)
    instance_concurrency = optional(number)

    code = optional(object({
      oss_bucket_name = optional(string, "")
      oss_object_name = optional(string, "")
      zip_file        = optional(string, "")
      checksum        = optional(string, "")
    }))

    role            = optional(string, "")
    internet_access = optional(bool)

    vpc_config = optional(object({
      vpc_id            = optional(string, "")
      vswitch_ids       = optional(list(string), [])
      security_group_id = optional(string, "")
    }))

    log_config = optional(object({
      project                = optional(string, "")
      logstore               = optional(string, "")
      log_begin_rule         = optional(string)
      enable_instance_metrics = optional(bool)
      enable_request_metrics  = optional(bool)
    }))

    custom_container_config = optional(object({
      image      = string
      entrypoint = optional(list(string), [])
      command    = optional(list(string), [])
      port       = optional(number)
      health_check_config = optional(object({
        initial_delay_seconds = optional(number)
        timeout_seconds       = optional(number)
        http_get_url          = optional(string, "")
        period_seconds        = optional(number)
        failure_threshold     = optional(number)
        success_threshold     = optional(number)
      }))
    }))

    custom_runtime_config = optional(object({
      command = optional(list(string), [])
      args    = optional(list(string), [])
      port    = optional(number)
      health_check_config = optional(object({
        initial_delay_seconds = optional(number)
        timeout_seconds       = optional(number)
        http_get_url          = optional(string, "")
        period_seconds        = optional(number)
        failure_threshold     = optional(number)
        success_threshold     = optional(number)
      }))
    }))

    instance_lifecycle_config = optional(object({
      initializer = optional(object({
        handler = optional(string, "")
        timeout = optional(number)
        command = optional(list(string), [])
      }))
      pre_stop = optional(object({
        handler = optional(string, "")
        timeout = optional(number)
      }))
    }))

    nas_config = optional(object({
      user_id  = optional(number)
      group_id = optional(number)
      mount_points = optional(list(object({
        server_addr = string
        mount_dir   = string
        enable_tls  = optional(bool)
      })), [])
    }))

    gpu_config = optional(object({
      gpu_memory_size = number
      gpu_type        = string
    }))

    layers                 = optional(list(string), [])
    environment_variables  = optional(map(string), {})
    tags                   = optional(map(string), {})
    resource_group_id      = optional(string, "")
  })

  validation {
    condition = contains([
      "python3.12", "python3.10", "python3.9", "python3",
      "nodejs20", "nodejs18", "nodejs16", "nodejs14",
      "java11", "java8", "go1", "php7.2", "dotnetcore3.1",
      "custom", "custom.debian10", "custom.debian11", "custom.debian12",
      "custom-container"
    ], var.spec.runtime)
    error_message = "runtime must be a valid FC v3 runtime."
  }
}
