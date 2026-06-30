# --- Subnet Group (conditional) ---

resource "aws_redshift_subnet_group" "this" {
  count = local.create_subnet_group ? 1 : 0

  name       = "${local.cluster_identifier}-subnet-group"
  subnet_ids = var.spec.subnet_ids
  tags       = local.tags
}

# --- Security Group (conditional) ---

resource "aws_security_group" "this" {
  count = local.create_security_group ? 1 : 0

  name_prefix = "${local.cluster_identifier}-redshift-"
  description = "Managed security group for Redshift cluster ${local.cluster_identifier}"
  vpc_id      = var.spec.vpc_id
  tags        = local.tags

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_security_group_rule" "ingress_sg" {
  for_each = local.create_security_group ? toset(var.spec.security_group_ids) : toset([])

  type                     = "ingress"
  from_port                = var.spec.port
  to_port                  = var.spec.port
  protocol                 = "tcp"
  source_security_group_id = each.value
  security_group_id        = aws_security_group.this[0].id
  description              = "Allow Redshift access from ${each.value}"
}

resource "aws_security_group_rule" "ingress_cidr" {
  for_each = local.create_security_group ? toset(var.spec.allowed_cidr_blocks) : toset([])

  type              = "ingress"
  from_port         = var.spec.port
  to_port           = var.spec.port
  protocol          = "tcp"
  cidr_blocks       = [each.value]
  security_group_id = aws_security_group.this[0].id
  description       = "Allow Redshift access from ${each.value}"
}

resource "aws_security_group_rule" "egress" {
  count = local.create_security_group ? 1 : 0

  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.this[0].id
  description       = "Allow all outbound traffic"
}

# --- Parameter Group (conditional) ---

resource "aws_redshift_parameter_group" "this" {
  count = local.create_parameter_group ? 1 : 0

  name   = "${local.cluster_identifier}-params"
  family = "redshift-1.0"
  tags   = local.tags

  dynamic "parameter" {
    for_each = var.spec.parameters
    content {
      name  = parameter.value.name
      value = parameter.value.value
    }
  }
}

# --- Redshift Cluster ---

resource "aws_redshift_cluster" "this" {
  cluster_identifier = local.cluster_identifier
  cluster_type       = local.cluster_type
  node_type          = var.spec.node_type
  number_of_nodes    = var.spec.number_of_nodes > 1 ? var.spec.number_of_nodes : null

  database_name   = var.spec.database_name
  master_username = var.spec.master_username
  master_password = var.spec.master_password != "" ? var.spec.master_password : null
  port            = var.spec.port

  manage_master_password                = var.spec.manage_master_password ? true : null
  master_password_secret_kms_key_id     = var.spec.master_password_secret_kms_key_id != "" ? var.spec.master_password_secret_kms_key_id : null

  # Networking
  cluster_subnet_group_name = local.create_subnet_group ? aws_redshift_subnet_group.this[0].name : (var.spec.cluster_subnet_group_name != "" ? var.spec.cluster_subnet_group_name : null)
  vpc_security_group_ids    = length(local.all_security_group_ids) > 0 ? local.all_security_group_ids : null
  publicly_accessible       = var.spec.publicly_accessible
  enhanced_vpc_routing      = var.spec.enhanced_vpc_routing
  multi_az                  = var.spec.multi_az

  # Encryption
  encrypted  = var.spec.encrypted
  kms_key_id = var.spec.kms_key_id != "" ? var.spec.kms_key_id : null

  # IAM
  iam_roles          = length(var.spec.iam_roles) > 0 ? var.spec.iam_roles : null
  default_iam_role_arn = var.spec.default_iam_role_arn != "" ? var.spec.default_iam_role_arn : null

  # Snapshots
  automated_snapshot_retention_period = var.spec.automated_snapshot_retention_period
  skip_final_snapshot                 = var.spec.skip_final_snapshot
  final_snapshot_identifier           = !var.spec.skip_final_snapshot && var.spec.final_snapshot_identifier != "" ? var.spec.final_snapshot_identifier : null

  # Maintenance
  preferred_maintenance_window = var.spec.preferred_maintenance_window != "" ? var.spec.preferred_maintenance_window : null
  allow_version_upgrade        = var.spec.allow_version_upgrade
  maintenance_track_name       = var.spec.maintenance_track_name != "" ? var.spec.maintenance_track_name : null
  apply_immediately            = var.spec.apply_immediately

  # Parameter group
  cluster_parameter_group_name = local.effective_parameter_group_name

  tags = local.tags
}

# --- Logging (conditional) ---

resource "aws_redshift_logging" "this" {
  count = local.create_logging ? 1 : 0

  cluster_identifier   = aws_redshift_cluster.this.id
  log_destination_type = var.spec.logging.log_destination_type
  bucket_name          = var.spec.logging.s3_bucket_name != "" ? var.spec.logging.s3_bucket_name : null
  s3_key_prefix        = var.spec.logging.s3_key_prefix != "" ? var.spec.logging.s3_key_prefix : null
  log_exports          = length(var.spec.logging.log_exports) > 0 ? var.spec.logging.log_exports : null
}
