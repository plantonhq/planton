# ---------------------------------------------------------------------------
# Global Accelerator
# ---------------------------------------------------------------------------

resource "aws_globalaccelerator_accelerator" "this" {
  name            = local.name
  enabled         = var.spec.enabled
  ip_address_type = var.spec.ip_address_type
  ip_addresses    = length(var.spec.ip_addresses) > 0 ? var.spec.ip_addresses : null

  dynamic "attributes" {
    for_each = var.spec.flow_logs != null && var.spec.flow_logs.enabled ? [var.spec.flow_logs] : []
    content {
      flow_logs_enabled   = true
      flow_logs_s3_bucket = attributes.value.s3_bucket
      flow_logs_s3_prefix = attributes.value.s3_prefix
    }
  }

  tags = local.tags
}

# ---------------------------------------------------------------------------
# Listeners
# ---------------------------------------------------------------------------

resource "aws_globalaccelerator_listener" "this" {
  for_each = local.listeners_map

  accelerator_arn = aws_globalaccelerator_accelerator.this.id
  protocol        = each.value.protocol
  client_affinity = each.value.client_affinity

  dynamic "port_range" {
    for_each = each.value.port_ranges
    content {
      from_port = port_range.value.from_port
      to_port   = port_range.value.to_port
    }
  }
}

# ---------------------------------------------------------------------------
# Endpoint Groups
# ---------------------------------------------------------------------------

resource "aws_globalaccelerator_endpoint_group" "this" {
  for_each = local.endpoint_groups_map

  listener_arn = aws_globalaccelerator_listener.this[each.value.listener_name].id

  endpoint_group_region          = each.value.group.endpoint_group_region != "" ? each.value.group.endpoint_group_region : null
  health_check_port              = each.value.group.health_check_port
  health_check_protocol          = each.value.group.health_check_protocol
  health_check_path              = each.value.group.health_check_path != "" ? each.value.group.health_check_path : null
  health_check_interval_seconds  = each.value.group.health_check_interval_seconds
  threshold_count                = each.value.group.threshold_count
  traffic_dial_percentage        = each.value.group.traffic_dial_percentage

  dynamic "endpoint_configuration" {
    for_each = each.value.group.endpoints
    content {
      endpoint_id                    = endpoint_configuration.value.endpoint_id
      weight                         = endpoint_configuration.value.weight
      client_ip_preservation_enabled = endpoint_configuration.value.client_ip_preservation_enabled
    }
  }

  dynamic "port_override" {
    for_each = each.value.group.port_overrides
    content {
      listener_port = port_override.value.listener_port
      endpoint_port = port_override.value.endpoint_port
    }
  }
}
