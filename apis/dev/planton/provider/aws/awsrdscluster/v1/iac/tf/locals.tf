locals {
  # Stable resource ID from metadata
  resource_id = (
    (var.metadata.id != null && var.metadata.id != "")
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "aws_rds_cluster"
  }

  org_label = (
    (var.metadata.org != null && var.metadata.org != "")
    ? { "organization" = var.metadata.org }
    : {}
  )

  env_label = (
    (var.metadata.env != null && try(var.metadata.env, "") != "")
    ? { "environment" = var.metadata.env }
    : {}
  )

  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Networking
  safe_subnet_ids = [for s in coalesce(try(var.spec.subnet_ids, []), []) : s.value]
  has_subnet_ids  = length(local.safe_subnet_ids) >= 2
  subnet_group_name_from_var = try(var.spec.db_subnet_group_name.value, "")
  need_subnet_group = local.has_subnet_ids && local.subnet_group_name_from_var == ""

  # Security groups
  ingress_sg_ids        = [for s in coalesce(try(var.spec.security_group_ids, []), []) : s.value]
  associate_sg_ids      = [for s in coalesce(try(var.spec.associate_security_group_ids, []), []) : s.value]
  allowed_cidr_blocks   = coalesce(try(var.spec.allowed_cidr_blocks, []), [])
  need_managed_sg       = length(local.ingress_sg_ids) > 0 || length(local.allowed_cidr_blocks) > 0
  vpc_id                = try(var.spec.vpc_id.value, null)

  # Parameters
  parameters = coalesce(try(var.spec.parameters, []), [])
  need_cluster_parameter_group = length(local.parameters) > 0 || try(var.spec.db_cluster_parameter_group_name, "") == ""

  # Engine family derivation for parameter group family
  engine        = var.spec.engine
  engine_version = var.spec.engine_version
  engine_family = (
    startswith(local.engine, "aurora-mysql") ? "aurora-mysql${split(".", local.engine_version)[0]}" : (
    startswith(local.engine, "aurora-postgresql") ? "aurora-postgresql${split(".", local.engine_version)[0]}" : ""
    )
  )
}


