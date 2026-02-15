locals {
  # Stable resource ID from metadata
  resource_id = coalesce(try(var.metadata.id, null), var.metadata.name)

  tags = merge({
    "Name" = local.resource_id
  }, try(var.metadata.labels, {}))

  # Engine
  engine         = var.spec.engine
  engine_version = try(var.spec.engine_version, null)

  # Topology
  num_cache_clusters      = try(var.spec.num_cache_clusters, 0)
  num_node_groups         = try(var.spec.num_node_groups, 0)
  replicas_per_node_group = try(var.spec.replicas_per_node_group, null)
  is_clustered            = local.num_node_groups > 0

  # Networking
  subnet_ids    = [for s in coalesce(try(var.spec.subnet_ids, []), []) : s.value]
  has_subnets   = length(local.subnet_ids) > 0
  sg_ids        = [for s in coalesce(try(var.spec.security_group_ids, []), []) : s.value]

  # Encryption
  at_rest_encryption_enabled  = coalesce(try(var.spec.at_rest_encryption_enabled, null), false)
  transit_encryption_enabled  = coalesce(try(var.spec.transit_encryption_enabled, null), false)
  transit_encryption_mode     = try(var.spec.transit_encryption_mode, null)
  kms_key_id                  = try(var.spec.kms_key_id.value, null)

  # Authentication
  auth_token     = try(var.spec.auth_token.value, null)
  user_group_ids = coalesce(try(var.spec.user_group_ids, []), [])

  # Maintenance and snapshots
  maintenance_window         = try(var.spec.maintenance_window, null)
  snapshot_retention_limit    = try(var.spec.snapshot_retention_limit, null) != 0 ? try(var.spec.snapshot_retention_limit, null) : null
  snapshot_window            = try(var.spec.snapshot_window, null)
  final_snapshot_identifier  = try(var.spec.final_snapshot_identifier, null)
  apply_immediately          = coalesce(try(var.spec.apply_immediately, null), false)

  # Parameters
  parameter_group_family = try(var.spec.parameter_group_family, "")
  parameters             = coalesce(try(var.spec.parameters, []), [])
  has_parameters         = length(local.parameters) > 0 && local.parameter_group_family != ""

  # Logging
  log_configs = coalesce(try(var.spec.log_delivery_configurations, []), [])

  # Advanced
  notification_topic_arn      = try(var.spec.notification_topic_arn.value, null)
  auto_minor_version_upgrade  = coalesce(try(var.spec.auto_minor_version_upgrade, null), false)
  data_tiering_enabled        = coalesce(try(var.spec.data_tiering_enabled, null), false)
  port                        = try(var.spec.port, null)
}
