resource "oci_core_network_security_group_security_rule" "this" {
  for_each = local.all_rules

  network_security_group_id = oci_core_network_security_group.this.id
  direction                 = each.value.direction
  protocol                  = each.value.protocol
  description               = each.value.description != "" ? each.value.description : null
  stateless                 = each.value.stateless

  source      = each.value.direction == "INGRESS" ? each.value.source : null
  source_type = each.value.direction == "INGRESS" ? each.value.source_type : null

  destination      = each.value.direction == "EGRESS" ? each.value.destination : null
  destination_type = each.value.direction == "EGRESS" ? each.value.destination_type : null

  dynamic "tcp_options" {
    for_each = each.value.tcp_options != null ? [each.value.tcp_options] : []
    content {
      dynamic "destination_port_range" {
        for_each = tcp_options.value.destination_port_range != null ? [tcp_options.value.destination_port_range] : []
        content {
          min = destination_port_range.value.min
          max = destination_port_range.value.max
        }
      }
      dynamic "source_port_range" {
        for_each = tcp_options.value.source_port_range != null ? [tcp_options.value.source_port_range] : []
        content {
          min = source_port_range.value.min
          max = source_port_range.value.max
        }
      }
    }
  }

  dynamic "udp_options" {
    for_each = each.value.udp_options != null ? [each.value.udp_options] : []
    content {
      dynamic "destination_port_range" {
        for_each = udp_options.value.destination_port_range != null ? [udp_options.value.destination_port_range] : []
        content {
          min = destination_port_range.value.min
          max = destination_port_range.value.max
        }
      }
      dynamic "source_port_range" {
        for_each = udp_options.value.source_port_range != null ? [udp_options.value.source_port_range] : []
        content {
          min = source_port_range.value.min
          max = source_port_range.value.max
        }
      }
    }
  }

  dynamic "icmp_options" {
    for_each = each.value.icmp_options != null ? [each.value.icmp_options] : []
    content {
      type = icmp_options.value.type
      code = icmp_options.value.code
    }
  }
}
