# ---------------------------------------------------------------------------
# ElastiCache Serverless cache
# ---------------------------------------------------------------------------

resource "aws_elasticache_serverless_cache" "this" {
  name   = local.resource_id
  engine = local.engine

  description          = local.description
  major_engine_version = local.major_engine_version

  # Networking
  subnet_ids         = local.has_subnets ? local.subnet_ids : null
  security_group_ids = local.has_sgs ? local.sg_ids : null

  # Encryption
  kms_key_id = local.kms_key_id

  # Snapshots (Redis/Valkey only — ignored for Memcached by AWS)
  daily_snapshot_time      = local.daily_snapshot_time
  snapshot_retention_limit  = local.snapshot_retention_limit > 0 ? local.snapshot_retention_limit : null

  # Authentication (Redis/Valkey only)
  user_group_id = local.user_group_id

  # Scaling limits
  dynamic "cache_usage_limits" {
    for_each = local.has_limits ? [1] : []
    content {
      dynamic "data_storage" {
        for_each = local.has_data_storage ? [1] : []
        content {
          unit    = "GB"
          minimum = local.data_storage_min_gb > 0 ? local.data_storage_min_gb : null
          maximum = local.data_storage_max_gb > 0 ? local.data_storage_max_gb : null
        }
      }

      dynamic "ecpu_per_second" {
        for_each = local.has_ecpu ? [1] : []
        content {
          minimum = local.ecpu_min > 0 ? local.ecpu_min : null
          maximum = local.ecpu_max > 0 ? local.ecpu_max : null
        }
      }
    }
  }

  tags = local.tags
}
