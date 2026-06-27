resource "aws_ecs_cluster" "this" {
  name = local.resource_name

  setting {
    name  = "containerInsights"
    value = local.container_insights_value
  }

  dynamic "configuration" {
    for_each = local.enable_execute_command ? [1] : []
    content {
      execute_command_configuration {
        logging = local.exec_logging
        kms_key_id = local.exec_kms_key_id != "" ? local.exec_kms_key_id : null

        dynamic "log_configuration" {
          for_each = local.exec_log_config != null ? [1] : []
          content {
            cloud_watch_log_group_name     = try(local.exec_log_config.cloud_watch_log_group_name, null)
            cloud_watch_encryption_enabled = try(local.exec_log_config.cloud_watch_encryption_enabled, false)
            s3_bucket_name                 = try(local.exec_log_config.s3_bucket_name, null)
            s3_key_prefix                  = try(local.exec_log_config.s3_key_prefix, null)
            s3_bucket_encryption_enabled   = try(local.exec_log_config.s3_encryption_enabled, false)
          }
        }
      }
    }
  }

  tags = local.tags
}

# Attach capacity providers to the cluster if specified
resource "aws_ecs_cluster_capacity_providers" "this" {
  count = local.has_capacity_providers ? 1 : 0

  cluster_name       = aws_ecs_cluster.this.name
  capacity_providers = local.safe_capacity_providers

  # Configure default capacity provider strategy if specified
  dynamic "default_capacity_provider_strategy" {
    for_each = local.has_default_strategy ? local.safe_default_capacity_provider_strategy : []
    content {
      capacity_provider = default_capacity_provider_strategy.value.capacity_provider
      base              = default_capacity_provider_strategy.value.base
      weight            = default_capacity_provider_strategy.value.weight
    }
  }
}


