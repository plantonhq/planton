resource "google_compute_firewall" "this" {
  name      = local.rule_name
  network   = local.network
  project   = local.project_id
  direction = local.direction
  priority  = local.priority
  disabled  = var.spec.disabled

  description = var.spec.description

  # Allow rules -- populated when action is "ALLOW".
  dynamic "allow" {
    for_each = local.action == "ALLOW" ? local.rules : []
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  # Deny rules -- populated when action is "DENY".
  dynamic "deny" {
    for_each = local.action == "DENY" ? local.rules : []
    content {
      protocol = deny.value.protocol
      ports    = deny.value.ports
    }
  }

  source_ranges           = length(var.spec.source_ranges) > 0 ? var.spec.source_ranges : null
  destination_ranges      = length(var.spec.destination_ranges) > 0 ? var.spec.destination_ranges : null
  source_tags             = length(var.spec.source_tags) > 0 ? var.spec.source_tags : null
  target_tags             = length(var.spec.target_tags) > 0 ? var.spec.target_tags : null
  source_service_accounts = length(var.spec.source_service_accounts) > 0 ? var.spec.source_service_accounts : null
  target_service_accounts = length(var.spec.target_service_accounts) > 0 ? var.spec.target_service_accounts : null

  dynamic "log_config" {
    for_each = var.spec.log_config != null ? [var.spec.log_config] : []
    content {
      metadata = log_config.value.metadata
    }
  }
}
