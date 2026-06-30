locals {
  resource_name = coalesce(try(var.metadata.name, null), "awseventbridgerule")

  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Event bus name — null for the default bus.
  event_bus_name = try(var.spec.event_bus_name.value, null) != "" ? try(var.spec.event_bus_name.value, null) : null

  # Event pattern — encode the struct to JSON if provided.
  event_pattern = try(var.spec.event_pattern, null) != null ? jsonencode(var.spec.event_pattern) : null

  # Schedule expression — null when not configured.
  schedule_expression = try(var.spec.schedule_expression, null) != "" ? try(var.spec.schedule_expression, null) : null

  # State — null to use AWS default (ENABLED).
  state = try(var.spec.state, null) != "" ? try(var.spec.state, null) : null

  # Build a map of targets keyed by name for for_each iteration.
  targets = {
    for target in try(var.spec.targets, []) : target.name => target
  }
}
