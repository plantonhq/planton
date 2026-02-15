locals {
  # Name and tags
  stream_name = coalesce(try(var.metadata.name, null), "awskinesisstream")

  tags = merge({
    "Name" = local.stream_name
  }, try(var.metadata.labels, {}))

  # Capacity mode
  stream_mode = try(var.spec.stream_mode, "PROVISIONED")
  shard_count = local.stream_mode == "PROVISIONED" ? coalesce(try(var.spec.shard_count, null), 1) : null

  # Data retention — null lets AWS use its default (24h).
  retention_period = try(var.spec.retention_period_hours, null) != 0 ? try(var.spec.retention_period_hours, null) : null

  # Encryption — presence of kms_key_id implies KMS encryption.
  kms_key_id      = try(var.spec.kms_key_id.value, null)
  encryption_type = local.kms_key_id != null ? "KMS" : "NONE"

  # Max record size — DEFERRED: max_record_size_in_kib is defined in the spec
  # but is not available in the pinned TF AWS provider 5.82.0. The field was
  # added in a newer provider version. Will be wired when the provider is upgraded.

  # Enhanced monitoring
  shard_level_metrics = try(var.spec.shard_level_metrics, [])

  # Deletion behavior
  enforce_consumer_deletion = coalesce(try(var.spec.enforce_consumer_deletion, null), false)
}
