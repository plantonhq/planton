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
  name_prefix = "${local.resource_id}-"
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
# Replication group
# ---------------------------------------------------------------------------

resource "aws_elasticache_replication_group" "this" {
  replication_group_id = local.resource_id
  description          = var.spec.description
  engine               = local.engine
  engine_version       = local.engine_version
  node_type            = var.spec.node_type
  port                 = local.port

  # Topology — non-clustered
  num_cache_clusters = local.is_clustered ? null : local.num_cache_clusters

  # Topology — clustered
  num_node_groups         = local.is_clustered ? local.num_node_groups : null
  replicas_per_node_group = local.is_clustered ? local.replicas_per_node_group : null

  # High availability
  automatic_failover_enabled = coalesce(try(var.spec.automatic_failover_enabled, null), false)
  multi_az_enabled           = coalesce(try(var.spec.multi_az_enabled, null), false)

  # Networking
  subnet_group_name  = local.has_subnets ? aws_elasticache_subnet_group.this[0].name : null
  security_group_ids = length(local.sg_ids) > 0 ? local.sg_ids : null

  # Encryption
  at_rest_encryption_enabled = local.at_rest_encryption_enabled
  transit_encryption_enabled = local.transit_encryption_enabled
  transit_encryption_mode    = local.transit_encryption_mode
  kms_key_id                 = local.kms_key_id

  # Authentication
  auth_token     = local.auth_token
  user_group_ids = length(local.user_group_ids) > 0 ? toset(local.user_group_ids) : null

  # Maintenance and snapshots
  maintenance_window        = local.maintenance_window
  snapshot_retention_limit   = local.snapshot_retention_limit
  snapshot_window           = local.snapshot_window
  final_snapshot_identifier = local.final_snapshot_identifier
  apply_immediately         = local.apply_immediately

  # Parameter group
  parameter_group_name = local.has_parameters ? aws_elasticache_parameter_group.this[0].name : null

  # Logging
  dynamic "log_delivery_configuration" {
    for_each = local.log_configs
    content {
      destination_type = log_delivery_configuration.value.destination_type
      destination      = try(log_delivery_configuration.value.destination.value, "")
      log_format       = log_delivery_configuration.value.log_format
      log_type         = log_delivery_configuration.value.log_type
    }
  }

  # Advanced
  notification_topic_arn     = local.notification_topic_arn
  auto_minor_version_upgrade = local.auto_minor_version_upgrade ? "true" : "false"
  data_tiering_enabled       = local.data_tiering_enabled

  tags = local.tags
}
