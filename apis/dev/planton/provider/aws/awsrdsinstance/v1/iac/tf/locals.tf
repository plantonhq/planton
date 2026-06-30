locals {
  resource_id = (
    (var.metadata.id != null && var.metadata.id != "")
    ? var.metadata.id
    : var.metadata.name
  )

  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "aws_rds_instance"
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
  ingress_sg_ids = [for s in coalesce(try(var.spec.security_group_ids, []), []) : s.value]
}
