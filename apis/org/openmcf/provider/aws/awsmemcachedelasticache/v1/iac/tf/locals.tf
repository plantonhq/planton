locals {
  # Stable resource ID from metadata
  resource_id = coalesce(try(var.metadata.id, null), var.metadata.name)

  tags = merge({
    "Name" = local.resource_id
  }, try(var.metadata.labels, {}))

  # Engine
  engine_version = try(var.spec.engine_version, null)

  # Nodes
  num_cache_nodes = coalesce(try(var.spec.num_cache_nodes, null), 1)
  az_mode         = try(var.spec.az_mode, null)

  # Networking
  subnet_ids  = [for s in coalesce(try(var.spec.subnet_ids, []), []) : s.value]
  has_subnets = length(local.subnet_ids) > 0
  sg_ids      = [for s in coalesce(try(var.spec.security_group_ids, []), []) : s.value]

  # Encryption
  transit_encryption_enabled = coalesce(try(var.spec.transit_encryption_enabled, null), false)

  # Parameters
  parameter_group_family = try(var.spec.parameter_group_family, "")
  parameters             = coalesce(try(var.spec.parameters, []), [])
  has_parameters         = length(local.parameters) > 0 && local.parameter_group_family != ""

  # Maintenance
  maintenance_window         = try(var.spec.maintenance_window, null)
  apply_immediately          = coalesce(try(var.spec.apply_immediately, null), false)
  auto_minor_version_upgrade = coalesce(try(var.spec.auto_minor_version_upgrade, null), false)

  # Notifications
  notification_topic_arn = try(var.spec.notification_topic_arn.value, null)

  # Node placement
  preferred_availability_zones = coalesce(try(var.spec.preferred_availability_zones, []), [])

  # Port
  port = try(var.spec.port, null)
}
