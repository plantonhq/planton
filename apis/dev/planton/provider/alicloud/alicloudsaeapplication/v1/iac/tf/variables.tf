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
  description = "Alibaba Cloud SAE Application specification"
  type = object({
    region          = string
    app_name        = string
    app_description = optional(string, "")
    package_type    = string
    replicas        = number
    cpu             = number
    memory          = number

    vpc_id            = optional(string, "")
    vswitch_id        = optional(string, "")
    security_group_id = optional(string, "")

    namespace_id    = optional(string, "")
    image_url       = optional(string, "")
    package_url     = optional(string, "")
    package_version = optional(string, "")
    command         = optional(string, "")
    command_args    = optional(list(string), [])
    envs            = optional(map(string), {})

    jdk                               = optional(string, "")
    jar_start_options                  = optional(string, "")
    jar_start_args                     = optional(string, "")
    programming_language               = optional(string, "")
    timezone                           = optional(string, "")
    termination_grace_period_seconds   = optional(number)
    min_ready_instances                = optional(number)
    acr_instance_id                    = optional(string, "")

    liveness = optional(object({
      http_get = optional(object({
        path = optional(string, "")
        port = optional(number)
      }))
      tcp_socket = optional(object({
        port = optional(number)
      }))
      exec = optional(object({
        command = optional(string, "")
      }))
      initial_delay_seconds = optional(number)
      period_seconds        = optional(number)
      timeout_seconds       = optional(number)
      failure_threshold     = optional(number)
      success_threshold     = optional(number)
    }))

    readiness = optional(object({
      http_get = optional(object({
        path = optional(string, "")
        port = optional(number)
      }))
      tcp_socket = optional(object({
        port = optional(number)
      }))
      exec = optional(object({
        command = optional(string, "")
      }))
      initial_delay_seconds = optional(number)
      period_seconds        = optional(number)
      timeout_seconds       = optional(number)
      failure_threshold     = optional(number)
      success_threshold     = optional(number)
    }))

    custom_host_aliases = optional(list(object({
      host_name = string
      ip        = string
    })), [])

    update_strategy = optional(object({
      type = optional(string, "")
      batch_update = optional(object({
        batch           = optional(number)
        batch_wait_time = optional(number)
        release_type    = optional(string, "")
      }))
    }))

    sls_configs = optional(string, "")
    tags        = optional(map(string), {})
  })

  validation {
    condition = contains([
      "Image", "FatJar", "War", "PythonZip", "PhpZip"
    ], var.spec.package_type)
    error_message = "package_type must be one of: Image, FatJar, War, PythonZip, PhpZip."
  }

  validation {
    condition = contains([
      500, 1000, 2000, 4000, 8000, 16000, 32000
    ], var.spec.cpu)
    error_message = "cpu must be one of: 500, 1000, 2000, 4000, 8000, 16000, 32000."
  }

  validation {
    condition = contains([
      1024, 2048, 4096, 8192, 12288, 16384, 24576, 32768, 65536, 131072
    ], var.spec.memory)
    error_message = "memory must be one of: 1024, 2048, 4096, 8192, 12288, 16384, 24576, 32768, 65536, 131072."
  }
}
