# Main Snowflake Database Resource
# This resource creates a Snowflake database with all configuration parameters
# from the SnowflakeDatabaseSpec proto definition

resource "snowflake_database" "this" {
  # Required parameters
  name = var.spec.name

  # Optional string parameters - only set if non-empty
  catalog                       = var.spec.catalog != "" ? var.spec.catalog : null
  comment                       = var.spec.comment != "" ? var.spec.comment : null
  default_ddl_collation         = var.spec.default_ddl_collation != "" ? var.spec.default_ddl_collation : null
  external_volume               = var.spec.external_volume != "" ? var.spec.external_volume : null
  log_level                     = var.spec.log_level != "" ? var.spec.log_level : null
  storage_serialization_policy  = var.spec.storage_serialization_policy != "" ? var.spec.storage_serialization_policy : null
  trace_level                   = var.spec.trace_level != "" ? var.spec.trace_level : null

  # Integer parameters - only set if greater than 0
  data_retention_time_in_days      = var.spec.data_retention_time_in_days > 0 ? var.spec.data_retention_time_in_days : null
  max_data_extension_time_in_days  = var.spec.max_data_extension_time_in_days > 0 ? var.spec.max_data_extension_time_in_days : null
  suspend_task_after_num_failures  = var.spec.suspend_task_after_num_failures >= 0 ? var.spec.suspend_task_after_num_failures : null
  task_auto_retry_attempts         = var.spec.task_auto_retry_attempts >= 0 ? var.spec.task_auto_retry_attempts : null

  # Boolean parameters
  drop_public_schema_on_creation = var.spec.drop_public_schema_on_creation
  enable_console_output          = var.spec.enable_console_output
  is_transient                   = var.spec.is_transient
  quoted_identifiers_ignore_case = var.spec.quoted_identifiers_ignore_case
  replace_invalid_characters     = var.spec.replace_invalid_characters

  # User task parameters - nested configuration
  user_task_managed_initial_warehouse_size  = var.spec.user_task.managed_initial_warehouse_size != "" ? var.spec.user_task.managed_initial_warehouse_size : null
  user_task_minimum_trigger_interval_in_seconds = var.spec.user_task.minimum_trigger_interval_in_seconds > 0 ? var.spec.user_task.minimum_trigger_interval_in_seconds : null
  user_task_timeout_ms                      = var.spec.user_task.timeout_ms > 0 ? var.spec.user_task.timeout_ms : null
}




