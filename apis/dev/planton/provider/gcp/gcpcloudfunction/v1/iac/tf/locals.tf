locals {
  # Derive function name: use spec.function_name if provided, otherwise metadata.name
  function_name = coalesce(
    var.spec.function_name,
    var.metadata.name
  )

  # Derive stable resource ID for labels
  resource_id = coalesce(
    var.metadata.id,
    var.metadata.name
  )

  # Base GCP labels
  base_gcp_labels = {
    "resource"      = "true"
    "resource_kind" = "gcp-cloud-function"
    "resource_name" = var.metadata.name
    "resource_id"   = local.resource_id
  }

  # Organization label (only if non-empty)
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ? { "organization" = var.metadata.org }
    : {}
  )

  # Environment label (only if non-empty)
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
    ? { "environment" = var.metadata.env }
    : {}
  )

  # User-provided labels from metadata (only if non-null)
  user_labels = var.metadata.labels != null ? var.metadata.labels : {}

  # Merge all labels
  final_gcp_labels = merge(
    local.base_gcp_labels,
    local.org_label,
    local.env_label,
    local.user_labels
  )

  # Determine if trigger is HTTP or Event-driven
  # Default to HTTP (0) if not specified
  is_http_trigger = (
    var.spec.trigger == null
    || var.spec.trigger.trigger_type == null
    || var.spec.trigger.trigger_type == 0
  )

  # Service config with defaults
  service_config = var.spec.service_config != null ? var.spec.service_config : {}

  # Memory: default 256MB if not specified
  available_memory_mb = coalesce(
    try(local.service_config.available_memory_mb, null),
    256
  )

  # Timeout: default 60 seconds if not specified
  timeout_seconds = coalesce(
    try(local.service_config.timeout_seconds, null),
    60
  )

  # Concurrency: default 80 if not specified
  max_instance_request_concurrency = coalesce(
    try(local.service_config.max_instance_request_concurrency, null),
    80
  )

  # Ingress settings: default to ALLOW_ALL
  ingress_settings_map = {
    0 = "ALLOW_ALL"
    1 = "ALLOW_INTERNAL_ONLY"
    2 = "ALLOW_INTERNAL_AND_GCLB"
  }

  ingress_settings = try(
    local.ingress_settings_map[local.service_config.ingress_settings],
    "ALLOW_ALL"
  )

  # VPC egress settings: default to PRIVATE_RANGES_ONLY
  vpc_egress_settings_map = {
    0 = "PRIVATE_RANGES_ONLY"
    1 = "ALL_TRAFFIC"
  }

  vpc_egress_settings = try(
    local.vpc_egress_settings_map[local.service_config.vpc_connector_egress_settings],
    "PRIVATE_RANGES_ONLY"
  )

  # Scaling config
  min_instance_count = try(
    local.service_config.scaling.min_instance_count,
    0
  )

  max_instance_count = try(
    local.service_config.scaling.max_instance_count,
    100
  )

  # Secret environment variables formatted for Terraform
  # Converts map to list of objects with key, project_id, secret, version
  secret_environment_variables = try(
    [
      for key, secret_name in local.service_config.secret_environment_variables : {
        key        = key
        project_id = var.spec.project_id
        secret     = secret_name
        version    = "latest"
      }
    ],
    []
  )

  # Retry policy mapping
  retry_policy_map = {
    0 = "RETRY_POLICY_DO_NOT_RETRY"
    1 = "RETRY_POLICY_RETRY"
  }

  event_retry_policy = try(
    local.retry_policy_map[var.spec.trigger.event_trigger.retry_policy],
    "RETRY_POLICY_DO_NOT_RETRY"
  )
}

