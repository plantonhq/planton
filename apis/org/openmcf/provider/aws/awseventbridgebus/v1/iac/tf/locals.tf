locals {
  resource_name = coalesce(try(var.metadata.name, null), "awseventbridgebus")

  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Partner event source — null when not configured.
  event_source_name = try(var.spec.event_source_name, null) != "" ? try(var.spec.event_source_name, null) : null
}
