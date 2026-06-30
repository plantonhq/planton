resource "alicloud_sae_application" "main" {
  app_name     = var.spec.app_name
  package_type = var.spec.package_type
  replicas     = var.spec.replicas
  cpu          = var.spec.cpu
  memory       = var.spec.memory

  app_description = var.spec.app_description != "" ? var.spec.app_description : null

  vpc_id            = var.spec.vpc_id != "" ? var.spec.vpc_id : null
  vswitch_id        = var.spec.vswitch_id != "" ? var.spec.vswitch_id : null
  security_group_id = var.spec.security_group_id != "" ? var.spec.security_group_id : null

  namespace_id    = var.spec.namespace_id != "" ? var.spec.namespace_id : null
  image_url       = var.spec.image_url != "" ? var.spec.image_url : null
  package_url     = var.spec.package_url != "" ? var.spec.package_url : null
  package_version = var.spec.package_version != "" ? var.spec.package_version : null

  command          = var.spec.command != "" ? var.spec.command : null
  command_args_v2  = length(var.spec.command_args) > 0 ? var.spec.command_args : null
  envs             = local.envs_json

  jdk              = var.spec.jdk != "" ? var.spec.jdk : null
  jar_start_options = var.spec.jar_start_options != "" ? var.spec.jar_start_options : null
  jar_start_args   = var.spec.jar_start_args != "" ? var.spec.jar_start_args : null

  programming_language              = var.spec.programming_language != "" ? var.spec.programming_language : null
  timezone                          = var.spec.timezone != "" ? var.spec.timezone : null
  termination_grace_period_seconds  = var.spec.termination_grace_period_seconds
  min_ready_instances               = var.spec.min_ready_instances
  acr_instance_id                   = var.spec.acr_instance_id != "" ? var.spec.acr_instance_id : null

  sls_configs = var.spec.sls_configs != "" ? var.spec.sls_configs : null

  tags = local.final_tags

  dynamic "liveness_v2" {
    for_each = var.spec.liveness != null ? [var.spec.liveness] : []
    content {
      dynamic "http_get" {
        for_each = liveness_v2.value.http_get != null ? [liveness_v2.value.http_get] : []
        content {
          path = http_get.value.path != "" ? http_get.value.path : null
          port = http_get.value.port
        }
      }

      dynamic "tcp_socket" {
        for_each = liveness_v2.value.tcp_socket != null ? [liveness_v2.value.tcp_socket] : []
        content {
          port = tcp_socket.value.port
        }
      }

      dynamic "exec" {
        for_each = liveness_v2.value.exec != null ? [liveness_v2.value.exec] : []
        content {
          command = exec.value.command != "" ? [exec.value.command] : null
        }
      }

      initial_delay_seconds = liveness_v2.value.initial_delay_seconds
      period_seconds        = liveness_v2.value.period_seconds
      timeout_seconds       = liveness_v2.value.timeout_seconds
      failure_threshold     = liveness_v2.value.failure_threshold
    }
  }

  dynamic "readiness_v2" {
    for_each = var.spec.readiness != null ? [var.spec.readiness] : []
    content {
      dynamic "http_get" {
        for_each = readiness_v2.value.http_get != null ? [readiness_v2.value.http_get] : []
        content {
          path = http_get.value.path != "" ? http_get.value.path : null
          port = http_get.value.port
        }
      }

      dynamic "tcp_socket" {
        for_each = readiness_v2.value.tcp_socket != null ? [readiness_v2.value.tcp_socket] : []
        content {
          port = tcp_socket.value.port
        }
      }

      dynamic "exec" {
        for_each = readiness_v2.value.exec != null ? [readiness_v2.value.exec] : []
        content {
          command = exec.value.command != "" ? [exec.value.command] : null
        }
      }

      initial_delay_seconds = readiness_v2.value.initial_delay_seconds
      period_seconds        = readiness_v2.value.period_seconds
      timeout_seconds       = readiness_v2.value.timeout_seconds
      failure_threshold     = readiness_v2.value.failure_threshold
    }
  }

  dynamic "custom_host_alias_v2" {
    for_each = var.spec.custom_host_aliases
    content {
      host_name = custom_host_alias_v2.value.host_name
      ip        = custom_host_alias_v2.value.ip
    }
  }

  dynamic "update_strategy_v2" {
    for_each = var.spec.update_strategy != null ? [var.spec.update_strategy] : []
    content {
      type = update_strategy_v2.value.type != "" ? update_strategy_v2.value.type : null

      dynamic "batch_update" {
        for_each = update_strategy_v2.value.batch_update != null ? [update_strategy_v2.value.batch_update] : []
        content {
          batch           = batch_update.value.batch
          batch_wait_time = batch_update.value.batch_wait_time
          release_type    = batch_update.value.release_type != "" ? batch_update.value.release_type : null
        }
      }
    }
  }
}
