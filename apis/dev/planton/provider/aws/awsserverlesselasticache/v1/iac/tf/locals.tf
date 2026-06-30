locals {
  # Stable resource ID from metadata
  resource_id = coalesce(try(var.metadata.id, null), var.metadata.name)

  tags = merge({
    "Name" = local.resource_id
  }, try(var.metadata.labels, {}))

  # Engine
  engine               = var.spec.engine
  major_engine_version = try(var.spec.major_engine_version, null)
  description          = try(var.spec.description, null)

  # Scaling limits — data storage
  data_storage_min_gb = coalesce(try(var.spec.data_storage_min_gb, null), 0)
  data_storage_max_gb = coalesce(try(var.spec.data_storage_max_gb, null), 0)
  has_data_storage    = local.data_storage_min_gb > 0 || local.data_storage_max_gb > 0

  # Scaling limits — ECPU
  ecpu_min = coalesce(try(var.spec.ecpu_min, null), 0)
  ecpu_max = coalesce(try(var.spec.ecpu_max, null), 0)
  has_ecpu = local.ecpu_min > 0 || local.ecpu_max > 0

  # Whether to include the cache_usage_limits block
  has_limits = local.has_data_storage || local.has_ecpu

  # Networking
  subnet_ids  = [for s in coalesce(try(var.spec.subnet_ids, []), []) : s.value]
  has_subnets = length(local.subnet_ids) > 0
  sg_ids      = [for s in coalesce(try(var.spec.security_group_ids, []), []) : s.value]
  has_sgs     = length(local.sg_ids) > 0

  # Encryption
  kms_key_id = try(var.spec.kms_key_id.value, null)

  # Snapshots (Redis/Valkey only)
  daily_snapshot_time      = try(var.spec.daily_snapshot_time, null)
  snapshot_retention_limit  = coalesce(try(var.spec.snapshot_retention_limit, null), 0)

  # Authentication (Redis/Valkey only)
  user_group_id = try(var.spec.user_group_id, null)
}
