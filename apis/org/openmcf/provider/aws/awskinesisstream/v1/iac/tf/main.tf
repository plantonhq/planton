resource "aws_kinesis_stream" "this" {
  name = local.stream_name

  # Capacity mode
  dynamic "stream_mode_details" {
    for_each = [local.stream_mode]
    content {
      stream_mode = stream_mode_details.value
    }
  }

  shard_count = local.shard_count

  # Data retention
  retention_period = local.retention_period

  # Encryption
  encryption_type = local.encryption_type
  kms_key_id      = local.kms_key_id

  # Max record size — DEFERRED: not available in pinned TF AWS provider 5.82.0

  # Enhanced monitoring
  shard_level_metrics = length(local.shard_level_metrics) > 0 ? local.shard_level_metrics : null

  # Deletion behavior
  enforce_consumer_deletion = local.enforce_consumer_deletion

  tags = local.tags
}
