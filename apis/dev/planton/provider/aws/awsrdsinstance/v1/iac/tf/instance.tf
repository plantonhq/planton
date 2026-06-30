resource "aws_db_instance" "this" {
  identifier             = local.resource_id
  engine                 = var.spec.engine
  engine_version         = var.spec.engine_version
  instance_class         = var.spec.instance_class
  allocated_storage      = var.spec.allocated_storage_gb
  storage_encrypted      = try(var.spec.storage_encrypted, false)
  kms_key_id             = try(var.spec.kms_key_id.value, null)
  username               = var.spec.username
  password               = var.spec.password
  port                   = try(var.spec.port, null)
  publicly_accessible    = try(var.spec.publicly_accessible, false)
  multi_az               = try(var.spec.multi_az, false)
  parameter_group_name   = try(var.spec.parameter_group_name, null)
  option_group_name      = try(var.spec.option_group_name, null)
  db_subnet_group_name   = coalesce(try(var.spec.db_subnet_group_name.value, null), try(aws_db_subnet_group.this[0].name, null))
  vpc_security_group_ids = local.ingress_sg_ids

  tags = local.final_labels
}
