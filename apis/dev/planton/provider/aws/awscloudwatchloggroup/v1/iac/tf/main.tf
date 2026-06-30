resource "aws_cloudwatch_log_group" "this" {
  name = local.resource_name

  # Retention: 0 means never expire. null lets AWS default to 0 (never expire).
  retention_in_days = local.retention_in_days

  # KMS encryption: customer-managed key for log data at rest.
  kms_key_id = local.kms_key_id

  # Log group class: STANDARD, INFREQUENT_ACCESS, or DELIVERY.
  log_group_class = local.log_group_class

  # NOTE: deletion_protection_enabled is defined in the spec but not yet
  # available in TF AWS provider 5.82.0. When the provider version is upgraded,
  # uncomment:
  #   deletion_protection_enabled = local.deletion_protection_enabled

  tags = local.tags
}
