resource "alicloud_fcv3_function" "main" {
  function_name = var.spec.function_name
  handler       = var.spec.handler
  runtime       = var.spec.runtime
  description   = var.spec.description != "" ? var.spec.description : null

  cpu                  = var.spec.cpu
  memory_size          = var.spec.memory_size
  timeout              = var.spec.timeout
  disk_size            = var.spec.disk_size
  instance_concurrency = var.spec.instance_concurrency

  role            = var.spec.role != "" ? var.spec.role : null
  internet_access = var.spec.internet_access
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null

  environment_variables = length(var.spec.environment_variables) > 0 ? var.spec.environment_variables : null
  layers                = length(var.spec.layers) > 0 ? var.spec.layers : null
  tags                  = local.final_tags

  dynamic "code" {
    for_each = var.spec.code != null ? [var.spec.code] : []
    content {
      oss_bucket_name = code.value.oss_bucket_name != "" ? code.value.oss_bucket_name : null
      oss_object_name = code.value.oss_object_name != "" ? code.value.oss_object_name : null
      zip_file        = code.value.zip_file != "" ? code.value.zip_file : null
      checksum        = code.value.checksum != "" ? code.value.checksum : null
    }
  }

  dynamic "vpc_config" {
    for_each = var.spec.vpc_config != null ? [var.spec.vpc_config] : []
    content {
      vpc_id            = vpc_config.value.vpc_id != "" ? vpc_config.value.vpc_id : null
      vswitch_ids       = vpc_config.value.vswitch_ids
      security_group_id = vpc_config.value.security_group_id != "" ? vpc_config.value.security_group_id : null
    }
  }

  dynamic "log_config" {
    for_each = var.spec.log_config != null ? [var.spec.log_config] : []
    content {
      project                = log_config.value.project != "" ? log_config.value.project : null
      logstore               = log_config.value.logstore != "" ? log_config.value.logstore : null
      log_begin_rule         = log_config.value.log_begin_rule
      enable_instance_metrics = log_config.value.enable_instance_metrics
      enable_request_metrics  = log_config.value.enable_request_metrics
    }
  }

  dynamic "custom_container_config" {
    for_each = var.spec.custom_container_config != null ? [var.spec.custom_container_config] : []
    content {
      image      = custom_container_config.value.image
      entrypoint = length(custom_container_config.value.entrypoint) > 0 ? custom_container_config.value.entrypoint : null
      command    = length(custom_container_config.value.command) > 0 ? custom_container_config.value.command : null
      port       = custom_container_config.value.port

      dynamic "health_check_config" {
        for_each = custom_container_config.value.health_check_config != null ? [custom_container_config.value.health_check_config] : []
        content {
          initial_delay_seconds = health_check_config.value.initial_delay_seconds
          timeout_seconds       = health_check_config.value.timeout_seconds
          http_get_url          = health_check_config.value.http_get_url != "" ? health_check_config.value.http_get_url : null
          period_seconds        = health_check_config.value.period_seconds
          failure_threshold     = health_check_config.value.failure_threshold
          success_threshold     = health_check_config.value.success_threshold
        }
      }
    }
  }

  dynamic "custom_runtime_config" {
    for_each = var.spec.custom_runtime_config != null ? [var.spec.custom_runtime_config] : []
    content {
      command = length(custom_runtime_config.value.command) > 0 ? custom_runtime_config.value.command : null
      args    = length(custom_runtime_config.value.args) > 0 ? custom_runtime_config.value.args : null
      port    = custom_runtime_config.value.port

      dynamic "health_check_config" {
        for_each = custom_runtime_config.value.health_check_config != null ? [custom_runtime_config.value.health_check_config] : []
        content {
          initial_delay_seconds = health_check_config.value.initial_delay_seconds
          timeout_seconds       = health_check_config.value.timeout_seconds
          http_get_url          = health_check_config.value.http_get_url != "" ? health_check_config.value.http_get_url : null
          period_seconds        = health_check_config.value.period_seconds
          failure_threshold     = health_check_config.value.failure_threshold
          success_threshold     = health_check_config.value.success_threshold
        }
      }
    }
  }

  dynamic "instance_lifecycle_config" {
    for_each = var.spec.instance_lifecycle_config != null ? [var.spec.instance_lifecycle_config] : []
    content {
      dynamic "initializer" {
        for_each = instance_lifecycle_config.value.initializer != null ? [instance_lifecycle_config.value.initializer] : []
        content {
          handler = initializer.value.handler != "" ? initializer.value.handler : null
          timeout = initializer.value.timeout
          command = length(initializer.value.command) > 0 ? initializer.value.command : null
        }
      }

      dynamic "pre_stop" {
        for_each = instance_lifecycle_config.value.pre_stop != null ? [instance_lifecycle_config.value.pre_stop] : []
        content {
          handler = pre_stop.value.handler != "" ? pre_stop.value.handler : null
          timeout = pre_stop.value.timeout
        }
      }
    }
  }

  dynamic "nas_config" {
    for_each = var.spec.nas_config != null ? [var.spec.nas_config] : []
    content {
      user_id  = nas_config.value.user_id
      group_id = nas_config.value.group_id

      dynamic "mount_points" {
        for_each = nas_config.value.mount_points
        content {
          server_addr = mount_points.value.server_addr
          mount_dir   = mount_points.value.mount_dir
          enable_tls  = mount_points.value.enable_tls
        }
      }
    }
  }

  dynamic "gpu_config" {
    for_each = var.spec.gpu_config != null ? [var.spec.gpu_config] : []
    content {
      gpu_memory_size = gpu_config.value.gpu_memory_size
      gpu_type        = gpu_config.value.gpu_type
    }
  }
}
