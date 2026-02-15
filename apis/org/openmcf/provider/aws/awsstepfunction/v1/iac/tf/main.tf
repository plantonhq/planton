resource "aws_sfn_state_machine" "this" {
  name       = local.resource_name
  definition = local.definition
  role_arn   = local.role_arn
  type       = local.sm_type

  tags = local.tags

  # Tracing configuration — enable X-Ray tracing when requested.
  dynamic "tracing_configuration" {
    for_each = local.tracing_enabled ? [1] : []
    content {
      enabled = true
    }
  }

  # Logging configuration — only when a non-OFF level is specified.
  dynamic "logging_configuration" {
    for_each = local.logging_enabled ? [1] : []
    content {
      level                  = local.logging_level
      include_execution_data = local.include_execution_data
      log_destination        = local.log_destination
    }
  }

  # Encryption configuration — only when a customer-managed KMS key is provided.
  dynamic "encryption_configuration" {
    for_each = local.has_encryption && local.kms_key_id != "" ? [1] : []
    content {
      type                              = "CUSTOMER_MANAGED_KMS_KEY"
      kms_key_id                        = local.kms_key_id
      kms_data_key_reuse_period_seconds = local.kms_data_key_reuse_period
    }
  }
}
