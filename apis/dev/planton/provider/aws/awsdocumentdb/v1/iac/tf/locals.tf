locals {
  # Stable resource ID from metadata
  resource_id = (
    (var.metadata.id != null && var.metadata.id != "")
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels/tags
  base_tags = {
    "planton.org/resource"      = "true"
    "planton.org/resource-id"   = local.resource_id
    "planton.org/resource-kind" = "AwsDocumentDb"
  }

  org_tag = (
    (var.metadata.org != null && var.metadata.org != "")
    ? { "planton.org/organization" = var.metadata.org }
    : {}
  )

  env_tag = (
    (var.metadata.env != null && var.metadata.env != "")
    ? { "planton.org/environment" = var.metadata.env }
    : {}
  )

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # Networking
  safe_subnet_ids       = [for s in coalesce(try(var.spec.subnets, []), []) : s.value if s.value != null && s.value != ""]
  has_subnet_ids        = length(local.safe_subnet_ids) >= 2
  subnet_group_name_var = try(var.spec.db_subnet_group.value, "")
  need_subnet_group     = local.has_subnet_ids && local.subnet_group_name_var == ""

  # Security groups
  ingress_sg_ids      = [for s in coalesce(try(var.spec.security_groups, []), []) : s.value if s.value != null && s.value != ""]
  allowed_cidrs       = coalesce(try(var.spec.allowed_cidrs, []), [])
  need_managed_sg     = length(local.ingress_sg_ids) > 0 || length(local.allowed_cidrs) > 0
  vpc_id              = try(var.spec.vpc.value, null)

  # Parameters
  parameters                   = coalesce(try(var.spec.cluster_parameters, []), [])
  need_cluster_parameter_group = length(local.parameters) > 0

  # Engine settings
  engine_version = coalesce(try(var.spec.engine_version, ""), "5.0.0")
  port           = coalesce(try(var.spec.port, 0), 27017)
  instance_count = coalesce(try(var.spec.instance_count, 0), 1)
  instance_class = coalesce(try(var.spec.instance_class, ""), "db.r6g.large")

  # Engine family for parameter group (docdb5.0, docdb4.0, etc.)
  engine_family = "docdb${substr(local.engine_version, 0, 3)}"
}
