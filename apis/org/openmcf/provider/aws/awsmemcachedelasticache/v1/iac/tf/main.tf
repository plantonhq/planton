# ---------------------------------------------------------------------------
# Subnet group (conditional)
# ---------------------------------------------------------------------------

resource "aws_elasticache_subnet_group" "this" {
  count      = local.has_subnets ? 1 : 0
  name       = local.resource_id
  subnet_ids = local.subnet_ids
  tags       = local.tags
}

# ---------------------------------------------------------------------------
# Parameter group (conditional)
# ---------------------------------------------------------------------------

resource "aws_elasticache_parameter_group" "this" {
  count       = local.has_parameters ? 1 : 0
  name        = "${local.resource_id}-custom"
  family      = local.parameter_group_family
  description = "Custom parameter group for ${local.resource_id}"

  dynamic "parameter" {
    for_each = local.parameters
    content {
      name  = parameter.value.name
      value = parameter.value.value
    }
  }

  tags = local.tags
}

# ---------------------------------------------------------------------------
# Memcached cluster
# ---------------------------------------------------------------------------

resource "aws_elasticache_cluster" "this" {
  cluster_id      = local.resource_id
  engine          = "memcached"
  engine_version  = local.engine_version
  node_type       = var.spec.node_type
  num_cache_nodes = local.num_cache_nodes
  port            = local.port

  # AZ mode
  az_mode                      = local.az_mode
  preferred_availability_zones = length(local.preferred_availability_zones) > 0 ? local.preferred_availability_zones : null

  # Encryption
  transit_encryption_enabled = local.transit_encryption_enabled

  # Networking
  subnet_group_name  = local.has_subnets ? aws_elasticache_subnet_group.this[0].name : null
  security_group_ids = length(local.sg_ids) > 0 ? local.sg_ids : null

  # Parameter group
  parameter_group_name = local.has_parameters ? aws_elasticache_parameter_group.this[0].name : null

  # Maintenance
  maintenance_window         = local.maintenance_window
  apply_immediately          = local.apply_immediately
  auto_minor_version_upgrade = local.auto_minor_version_upgrade ? "true" : "false"

  # Notifications
  notification_topic_arn = local.notification_topic_arn

  tags = local.tags
}
