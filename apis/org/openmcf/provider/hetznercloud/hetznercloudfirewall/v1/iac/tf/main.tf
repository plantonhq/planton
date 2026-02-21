resource "hcloud_firewall" "this" {
  name   = local.firewall_name
  labels = local.standard_labels

  dynamic "rule" {
    for_each = var.spec.rules != null ? var.spec.rules : []
    content {
      direction       = rule.value.direction
      protocol        = rule.value.protocol
      port            = rule.value.port
      source_ips      = rule.value.source_ips
      destination_ips = rule.value.destination_ips
      description     = rule.value.description
    }
  }
}
