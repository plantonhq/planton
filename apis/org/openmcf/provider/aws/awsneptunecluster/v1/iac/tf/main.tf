# Neptune Cluster
resource "aws_neptune_cluster" "main" {
  cluster_identifier = local.resource_id
  engine             = "neptune"
  engine_version     = local.engine_version

  port = local.port

  neptune_subnet_group_name = (
    local.need_subnet_group
    ? aws_neptune_subnet_group.main[0].name
    : local.subnet_group_name_var
  )

  vpc_security_group_ids = concat(
    local.ingress_sg_ids,
    local.need_managed_sg ? [aws_security_group.main[0].id] : []
  )

  neptune_cluster_parameter_group_name = (
    local.need_cluster_parameter_group
    ? aws_neptune_cluster_parameter_group.main[0].name
    : try(var.spec.cluster_parameter_group_name, null)
  )

  storage_type     = local.storage_type
  storage_encrypted = coalesce(try(var.spec.storage_encrypted, null), true)
  kms_key_arn      = try(var.spec.kms_key_id.value, null)

  iam_database_authentication_enabled = coalesce(try(var.spec.iam_database_authentication_enabled, null), false)
  iam_roles                           = length(local.iam_role_arns) > 0 ? local.iam_role_arns : null

  backup_retention_period      = coalesce(try(var.spec.backup_retention_period, 0), 7)
  preferred_backup_window      = try(var.spec.preferred_backup_window, null)
  preferred_maintenance_window = try(var.spec.preferred_maintenance_window, null)

  deletion_protection       = coalesce(try(var.spec.deletion_protection, null), false)
  skip_final_snapshot        = coalesce(try(var.spec.skip_final_snapshot, null), false)
  final_snapshot_identifier  = try(var.spec.final_snapshot_identifier, null)

  enable_cloudwatch_logs_exports = coalesce(try(var.spec.enabled_cloudwatch_logs_exports, []), [])
  apply_immediately             = coalesce(try(var.spec.apply_immediately, null), false)
  copy_tags_to_snapshot        = coalesce(try(var.spec.copy_tags_to_snapshot, null), false)
  allow_major_version_upgrade  = coalesce(try(var.spec.allow_major_version_upgrade, null), false)

  dynamic "serverless_v2_scaling_configuration" {
    for_each = local.serverless_v2_scaling_config != null ? [local.serverless_v2_scaling_config] : []
    content {
      min_capacity = serverless_v2_scaling_configuration.value.min_capacity
      max_capacity = serverless_v2_scaling_configuration.value.max_capacity
    }
  }

  tags = local.final_tags
}

# Neptune Cluster Instances
resource "aws_neptune_cluster_instance" "instances" {
  count = local.instance_count

  identifier         = "${local.resource_id}-${count.index + 1}"
  cluster_identifier = aws_neptune_cluster.main.id
  engine            = "neptune"
  instance_class    = local.instance_class

  neptune_subnet_group_name = (
    local.need_subnet_group
    ? aws_neptune_subnet_group.main[0].name
    : local.subnet_group_name_var
  )

  apply_immediately = coalesce(try(var.spec.apply_immediately, null), false)

  tags = local.final_tags
}
