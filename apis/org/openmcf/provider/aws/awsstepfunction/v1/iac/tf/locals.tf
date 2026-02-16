locals {
  # Name and tags
  resource_name = coalesce(try(var.metadata.name, null), "awsstepfunction")

  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # State machine type — default to STANDARD when not specified.
  sm_type = coalesce(try(var.spec.type, null), "STANDARD")

  # Definition — expect the Struct to arrive as a map; serialize to JSON.
  definition = jsonencode(try(var.spec.definition, {}))

  # IAM execution role ARN (resolved from StringValueOrRef).
  role_arn = try(var.spec.role_arn.value, "")

  # Tracing
  tracing_enabled = coalesce(try(var.spec.tracing_enabled, null), false)

  # Logging
  has_logging     = try(var.spec.logging, null) != null
  logging_level   = try(var.spec.logging.level, "OFF")
  logging_enabled = local.has_logging && local.logging_level != "" && local.logging_level != "OFF"

  # Log destination — auto-append ":*" if not present (AWS requirement).
  raw_log_destination = try(var.spec.logging.log_destination.value, "")
  log_destination     = local.raw_log_destination != "" && !endswith(local.raw_log_destination, ":*") ? "${local.raw_log_destination}:*" : local.raw_log_destination

  include_execution_data = coalesce(try(var.spec.logging.include_execution_data, null), false)

  # Encryption
  has_encryption = try(var.spec.encryption, null) != null
  kms_key_id     = try(var.spec.encryption.kms_key_id.value, "")

  kms_data_key_reuse_period = try(var.spec.encryption.kms_data_key_reuse_period_seconds, null) != 0 ? try(var.spec.encryption.kms_data_key_reuse_period_seconds, null) : null
}
