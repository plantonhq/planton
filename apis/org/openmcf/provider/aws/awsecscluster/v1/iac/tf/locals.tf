locals {
  # resource name and tags
  resource_name = coalesce(try(var.metadata.name, null), "aws-ecs-cluster")
  tags          = merge({
    "Name"        = local.resource_name
    "Environment" = try(var.metadata.env, "")
    "Org"         = try(var.metadata.org, "")
  }, try(var.metadata.labels, {}))

  # safe settings with defaults
  safe_enable_container_insights         = try(var.spec.enable_container_insights, true)
  safe_capacity_providers                = try(var.spec.capacity_providers, [])
  safe_default_capacity_provider_strategy = try(var.spec.default_capacity_provider_strategy, [])
  safe_exec_config                       = try(var.spec.execute_command_configuration, null)

  # computed values
  container_insights_value = local.safe_enable_container_insights ? "enabled" : "disabled"
  has_capacity_providers   = length(local.safe_capacity_providers) > 0
  has_default_strategy     = length(local.safe_default_capacity_provider_strategy) > 0

  # exec configuration - enabled if logging is not UNSPECIFIED (empty/null)
  enable_execute_command = local.safe_exec_config != null && try(local.safe_exec_config.logging, "") != ""
  exec_logging           = local.enable_execute_command ? try(local.safe_exec_config.logging, "DEFAULT") : "DEFAULT"
  exec_log_config        = local.enable_execute_command && local.exec_logging == "OVERRIDE" ? try(local.safe_exec_config.log_configuration, null) : null
  exec_kms_key_id        = local.enable_execute_command ? try(local.safe_exec_config.kms_key_id.value, "") : ""
}


