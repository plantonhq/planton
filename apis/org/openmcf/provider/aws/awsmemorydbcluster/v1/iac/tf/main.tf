resource "aws_memorydb_subnet_group" "this" {
  count = local.create_subnet_group ? 1 : 0

  name        = var.metadata.id
  description = "MemoryDB subnet group for ${var.metadata.id}"
  subnet_ids  = var.spec.subnet_ids
  tags        = local.tags
}

resource "aws_memorydb_parameter_group" "this" {
  count = local.create_parameter_group ? 1 : 0

  name        = "${var.metadata.id}-custom"
  family      = var.spec.parameter_group_family
  description = "Custom parameter group for ${var.metadata.id}"
  tags        = local.tags

  dynamic "parameter" {
    for_each = var.spec.parameters
    content {
      name  = parameter.value.name
      value = parameter.value.value
    }
  }
}

resource "aws_memorydb_cluster" "this" {
  name        = var.metadata.id
  acl_name    = var.spec.acl_name
  node_type   = var.spec.node_type
  description = var.spec.description
  tags        = local.tags

  engine         = var.spec.engine
  engine_version = var.spec.engine_version
  port           = var.spec.port

  num_shards              = var.spec.num_shards
  num_replicas_per_shard  = var.spec.num_replicas_per_shard

  subnet_group_name  = local.create_subnet_group ? aws_memorydb_subnet_group.this[0].name : null
  security_group_ids = length(var.spec.security_group_ids) > 0 ? var.spec.security_group_ids : null

  tls_enabled = var.spec.tls_enabled
  kms_key_arn = var.spec.kms_key_id

  maintenance_window       = var.spec.maintenance_window
  snapshot_retention_limit = var.spec.snapshot_retention_limit
  snapshot_window          = var.spec.snapshot_window
  final_snapshot_name      = var.spec.final_snapshot_name

  snapshot_arns = length(var.spec.snapshot_arns) > 0 ? var.spec.snapshot_arns : null
  snapshot_name = var.spec.snapshot_name

  parameter_group_name = local.create_parameter_group ? aws_memorydb_parameter_group.this[0].name : null

  sns_topic_arn              = var.spec.sns_topic_arn
  auto_minor_version_upgrade = var.spec.auto_minor_version_upgrade
  data_tiering               = var.spec.data_tiering
}
