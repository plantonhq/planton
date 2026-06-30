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
    "planton.org/resource-kind" = "AwsNeptuneCluster"
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
  safe_subnet_ids       = [for s in coalesce(try(var.spec.subnet_ids, []), []) : s.value if s.value != null && s.value != ""]
  has_subnet_ids        = length(local.safe_subnet_ids) >= 2
  subnet_group_name_var = try(var.spec.neptune_subnet_group_name.value, "")
  need_subnet_group     = local.has_subnet_ids && local.subnet_group_name_var == ""

  # Security groups
  ingress_sg_ids      = [for s in coalesce(try(var.spec.security_group_ids, []), []) : s.value if s.value != null && s.value != ""]
  allowed_cidrs      = coalesce(try(var.spec.allowed_cidr_blocks, []), [])
  need_managed_sg    = length(local.ingress_sg_ids) > 0 || length(local.allowed_cidrs) > 0
  vpc_id             = try(var.spec.vpc_id.value, null)

  # Parameters
  parameters                   = coalesce(try(var.spec.cluster_parameters, []), [])
  need_cluster_parameter_group = length(local.parameters) > 0

  # Engine settings
  engine_version = coalesce(try(var.spec.engine_version, ""), "1.3.0.0")
  port           = coalesce(try(var.spec.port, 0), 8182)
  instance_count = coalesce(try(var.spec.instance_count, 0), 1)
  instance_class = coalesce(try(var.spec.instance_class, ""), "db.r6g.large")
  storage_type   = coalesce(try(var.spec.storage_type, ""), "standard")

  # Engine family for parameter group (neptune1.2, neptune1.3, etc.)
  engine_family = "neptune${substr(local.engine_version, 0, 3)}"

  # Serverless v2 scaling (only when serverless_v2_scaling is set)
  has_serverless_v2 = try(var.spec.serverless_v2_scaling, null) != null
  serverless_v2_scaling_config = local.has_serverless_v2 ? {
    min_capacity = try(var.spec.serverless_v2_scaling.min_capacity, 2.5)
    max_capacity = try(var.spec.serverless_v2_scaling.max_capacity, 128)
  } : null

  # IAM roles
  iam_role_arns = [for r in coalesce(try(var.spec.iam_roles, []), []) : r.value if r.value != null && r.value != ""]
}
