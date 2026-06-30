# DocumentDB Cluster
resource "aws_docdb_cluster" "main" {
  cluster_identifier = local.resource_id
  engine             = "docdb"
  engine_version     = local.engine_version

  master_username = coalesce(try(var.spec.master_username, ""), "docdbadmin")
  master_password = var.spec.master_password

  port = local.port

  db_subnet_group_name = (
    local.need_subnet_group
    ? aws_docdb_subnet_group.main[0].name
    : local.subnet_group_name_var
  )

  vpc_security_group_ids = concat(
    local.ingress_sg_ids,
    local.need_managed_sg ? [aws_security_group.main[0].id] : []
  )

  db_cluster_parameter_group_name = (
    local.need_cluster_parameter_group
    ? aws_docdb_cluster_parameter_group.main[0].name
    : try(var.spec.cluster_parameter_group_name, null)
  )

  storage_encrypted = coalesce(try(var.spec.storage_encrypted, null), true)
  kms_key_id        = try(var.spec.kms_key.value, null)

  backup_retention_period      = coalesce(try(var.spec.backup_retention_period, 0), 7)
  preferred_backup_window      = try(var.spec.preferred_backup_window, null)
  preferred_maintenance_window = try(var.spec.preferred_maintenance_window, null)

  deletion_protection       = coalesce(try(var.spec.deletion_protection, null), false)
  skip_final_snapshot       = coalesce(try(var.spec.skip_final_snapshot, null), false)
  final_snapshot_identifier = try(var.spec.final_snapshot_identifier, null)

  enabled_cloudwatch_logs_exports = coalesce(try(var.spec.enabled_cloudwatch_logs_exports, []), [])
  apply_immediately               = coalesce(try(var.spec.apply_immediately, null), false)

  tags = local.final_tags
}

# DocumentDB Cluster Instances
resource "aws_docdb_cluster_instance" "instances" {
  count = local.instance_count

  identifier         = "${local.resource_id}-${count.index + 1}"
  cluster_identifier = aws_docdb_cluster.main.id
  instance_class     = local.instance_class

  auto_minor_version_upgrade = coalesce(try(var.spec.auto_minor_version_upgrade, null), true)
  apply_immediately          = coalesce(try(var.spec.apply_immediately, null), false)

  tags = local.final_tags
}
