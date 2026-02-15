locals {
  resource_name = coalesce(try(var.metadata.name, null), "awscloudwatchloggroup")

  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Retention: 0 (default) means never expire. Null when not configured.
  retention_in_days = try(var.spec.retention_in_days, 0) != 0 ? try(var.spec.retention_in_days, null) : null

  # KMS encryption — null when not configured.
  kms_key_id = try(var.spec.kms_key_id, null) != "" ? try(var.spec.kms_key_id, null) : null

  # Log group class — null when not configured (defaults to STANDARD).
  log_group_class = try(var.spec.log_group_class, null) != "" ? try(var.spec.log_group_class, null) : null

  # Deletion protection — false when not configured.
  deletion_protection_enabled = try(var.spec.deletion_protection_enabled, false)
}
