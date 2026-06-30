# Local variables for resource configuration
# These locals help centralize computed values and improve maintainability

locals {
  # Resource identification
  resource_id = var.metadata.id != null && var.metadata.id != "" ? var.metadata.id : var.metadata.name

  # Tags and labels for resource organization
  # Convert metadata labels to tags if needed
  tags = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "managed-by"  = "planton"
      "component"   = "snowflake-database"
      "environment" = var.metadata.env != null && var.metadata.env != "" ? var.metadata.env : "default"
      "org"         = var.metadata.org != null && var.metadata.org != "" ? var.metadata.org : "default"
    }
  )

  # Database configuration metadata
  database_name = var.spec.name
  is_transient  = var.spec.is_transient

  # Cost control indicators for monitoring
  # These locals help track cost-relevant configuration
  has_time_travel = var.spec.data_retention_time_in_days > 0
  retention_days  = var.spec.data_retention_time_in_days > 0 ? var.spec.data_retention_time_in_days : 1

  # Configuration validation flags
  has_catalog         = var.spec.catalog != ""
  has_external_volume = var.spec.external_volume != ""
  has_user_task_config = (
    var.spec.user_task.managed_initial_warehouse_size != "" ||
    var.spec.user_task.minimum_trigger_interval_in_seconds > 0 ||
    var.spec.user_task.timeout_ms > 0
  )
}




