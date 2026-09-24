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
  description = "AwsEcsTaskDefinition specification"
  type = object({
    region = string
    containers = list(object({
      name = string
      image = string
      essential = optional(bool)
      cpu = optional(number, 0)
      memory = optional(number, 0)
      memory_reservation = optional(number, 0)
      port_mappings = optional(list(object({
        container_port = optional(number, 0)
        protocol = optional(string, "")
        name = optional(string, "")
        app_protocol = optional(string, "")
      })), [])
      entry_point = optional(list(string), [])
      command = optional(list(string), [])
      working_directory = optional(string, "")
      environment = optional(map(string), {})
      secrets = optional(map(string), {})
      secret_environment = optional(map(string), {})
      environment_files = optional(list(string), [])
      health_check = optional(object({
        command = list(string)
        interval_seconds = optional(number, 0)
        timeout_seconds = optional(number, 0)
        retries = optional(number, 0)
        start_period_seconds = optional(number, 0)
      }))
      depends_on = optional(list(object({
        container_name = string
        condition = optional(string, "")
      })), [])
      mount_points = optional(list(object({
        source_volume = string
        container_path = string
        read_only = optional(bool, false)
      })), [])
      log_configuration = optional(object({
        log_driver = string
        options = optional(map(string), {})
        secret_options = optional(map(string), {})
      }))
      firelens_configuration = optional(object({
        type = optional(string, "")
        options = optional(map(string), {})
      }))
      repository_credentials_secret_arn = optional(string, "")
      user = optional(string, "")
      readonly_root_filesystem = optional(bool, false)
      privileged = optional(bool, false)
      init_process_enabled = optional(bool, false)
      gpu_count = optional(number, 0)
      ulimits = optional(list(object({
        name = string
        soft_limit = optional(number, 0)
        hard_limit = optional(number, 0)
      })), [])
      docker_labels = optional(map(string), {})
      start_timeout_seconds = optional(number, 0)
      stop_timeout_seconds = optional(number, 0)
      restart_policy = optional(object({
        enabled = optional(bool, false)
        ignored_exit_codes = optional(list(number), [])
        restart_attempt_period_seconds = optional(number, 0)
      }))
    }))
    requires_compatibilities = optional(list(string), [])
    cpu = optional(number, 0)
    memory = optional(number, 0)
    network_mode = optional(string, "")
    execution_role = optional(string, "")
    task_role = optional(string, "")
    runtime_platform = optional(object({
      cpu_architecture = optional(string, "")
      operating_system_family = optional(string, "")
    }))
    ephemeral_storage_gib = optional(number, 0)
    volumes = optional(list(object({
      name = string
      efs = optional(object({
        file_system_id = string
        root_directory = optional(string, "")
        access_point_id = optional(string, "")
        iam_authorization = optional(bool, false)
        transit_encryption_port = optional(number, 0)
      }))
      host_path = optional(string, "")
      configure_at_launch = optional(bool, false)
      docker = optional(object({
        autoprovision = optional(bool, false)
        driver = optional(string, "")
        driver_opts = optional(map(string), {})
        labels = optional(map(string), {})
        scope = optional(string, "")
      }))
      s3files = optional(object({
        file_system_arn = string
        access_point_arn = optional(string, "")
        root_directory = optional(string, "")
        transit_encryption_port = optional(number, 0)
      }))
    })), [])
    logging = optional(object({
      disabled = optional(bool, false)
      log_group = optional(string, "")
      retention_days = optional(number)
    }))
    skip_destroy = optional(bool, false)
    enable_fault_injection = optional(bool, false)
    ipc_mode = optional(string, "")
    pid_mode = optional(string, "")
    placement_constraints = optional(list(object({
      type = optional(string, "")
      expression = string
    })), [])
  })
}
