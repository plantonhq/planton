resource "aws_rds_cluster" "this" {
  cluster_identifier = local.resource_id

  engine         = var.spec.engine
  engine_version = var.spec.engine_version
  database_name  = try(var.spec.database_name, null)
  port           = try(var.spec.port, null)

  deletion_protection          = try(var.spec.deletion_protection, null)
  preferred_maintenance_window = try(var.spec.preferred_maintenance_window, null)
  backup_retention_period      = try(var.spec.backup_retention_period, null)
  preferred_backup_window      = try(var.spec.preferred_backup_window, null)
  copy_tags_to_snapshot        = try(var.spec.copy_tags_to_snapshot, null)
  skip_final_snapshot          = try(var.spec.skip_final_snapshot, null)
  final_snapshot_identifier    = try(var.spec.final_snapshot_identifier, null)
  iam_database_authentication_enabled = try(var.spec.iam_database_authentication_enabled, null)
  enable_http_endpoint                = try(var.spec.enable_http_endpoint, null)
  tags                                = local.final_labels

  enabled_cloudwatch_logs_exports = try(var.spec.enabled_cloudwatch_logs_exports, null)
  kms_key_id                      = try(var.spec.kms_key_id.value, null)
  storage_encrypted               = try(var.spec.storage_encrypted, null)
  replication_source_identifier   = try(var.spec.replication_source_identifier, null)
  snapshot_identifier             = try(var.spec.snapshot_identifier, null)
  engine_mode                     = try(var.spec.engine_mode, null)

  db_subnet_group_name = (
    local.need_subnet_group ? aws_db_subnet_group.this[0].name : try(var.spec.db_subnet_group_name.value, null)
  )

  vpc_security_group_ids = compact(concat(local.associate_sg_ids, [
    for i in aws_security_group.cluster : i.id
  ]))

  db_cluster_parameter_group_name = (
    local.need_cluster_parameter_group ? aws_rds_cluster_parameter_group.this[0].name : try(var.spec.db_cluster_parameter_group_name, null)
  )
}


