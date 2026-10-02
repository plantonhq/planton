locals {
  # The cache name is metadata.name — create-only in AWS, and the basis
  # both engines share so a manifest deploys identically on either.
  cache_name = var.metadata.name

  # Resource-identity tags match the Pulumi module key-for-key.
  aws_tags = {
    "Name"                     = var.metadata.name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsServerlessElasticache"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # Engine
  engine               = var.spec.engine
  major_engine_version = try(var.spec.major_engine_version, null) != "" ? var.spec.major_engine_version : null
  description          = try(var.spec.description, null) != "" ? var.spec.description : null

  # Scaling limits — data storage
  data_storage_min_gb = coalesce(try(var.spec.data_storage_min_gb, null), 0)
  data_storage_max_gb = coalesce(try(var.spec.data_storage_max_gb, null), 0)
  has_data_storage    = local.data_storage_min_gb > 0 || local.data_storage_max_gb > 0

  # Scaling limits — ECPU
  ecpu_min = coalesce(try(var.spec.ecpu_min, null), 0)
  ecpu_max = coalesce(try(var.spec.ecpu_max, null), 0)
  has_ecpu = local.ecpu_min > 0 || local.ecpu_max > 0

  has_limits = local.has_data_storage || local.has_ecpu

  # Networking — references arrive pre-flattened to plain strings (the
  # generator contract lowers StringValueOrRef to string).
  subnet_ids   = coalesce(try(var.spec.subnet_ids, []), [])
  has_subnets  = length(local.subnet_ids) > 0
  sg_ids       = coalesce(try(var.spec.security_group_ids, []), [])
  has_sgs      = length(local.sg_ids) > 0
  network_type = try(var.spec.network_type, null) != "" ? var.spec.network_type : null

  # Encryption
  kms_key_id = try(var.spec.kms_key_id, "") != "" ? var.spec.kms_key_id : null

  # Snapshots (Redis/Valkey only — CEL guards prevent Memcached usage)
  daily_snapshot_time      = try(var.spec.daily_snapshot_time, null) != "" ? var.spec.daily_snapshot_time : null
  snapshot_retention_limit = coalesce(try(var.spec.snapshot_retention_limit, null), 0)
  snapshot_arns_to_restore = coalesce(try(var.spec.snapshot_arns_to_restore, []), [])

  # Authentication (Redis/Valkey only — CEL guards prevent Memcached usage)
  user_group_id = try(var.spec.user_group_id, "") != "" ? var.spec.user_group_id : null
}
