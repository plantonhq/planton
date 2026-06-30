# DigitalOcean Firewall Resource
#
# This module provisions a stateful, network-edge firewall for DigitalOcean Droplets.
# Firewalls enforce a default-deny security model, blocking all traffic except
# explicitly allowed inbound and outbound rules.
#
# Key Features:
# - Stateful rules (return traffic automatically allowed)
# - Tag-based targeting (production standard, scales infinitely)
# - Static Droplet IDs (dev/testing only, max 10)
# - Resource-aware sources/destinations (Load Balancer UIDs, K8s cluster IDs)
#
# Production Best Practices:
# - Use tag-based targeting (not static droplet_ids)
# - Never expose SSH (port 22) to 0.0.0.0/0 in production
# - Use Load Balancer UIDs for public services (not direct 0.0.0.0/0 HTTPS)
# - Implement explicit outbound rules for high-security tiers (DB, cache)
# - Remember the "double firewall" trap: Cloud Firewall + host firewall (ufw)

resource "digitalocean_firewall" "firewall" {
  name = local.firewall_name

  # Droplet IDs (max 10, dev/testing only)
  droplet_ids = local.droplet_ids

  # Tags (production standard, unlimited, auto-scaling friendly)
  tags = local.tags

  # Inbound rules (traffic allowed *to* Droplets)
  dynamic "inbound_rule" {
    for_each = local.inbound_rules
    content {
      protocol                   = inbound_rule.value.protocol
      port_range                 = inbound_rule.value.port_range != "" ? inbound_rule.value.port_range : null
      source_addresses           = length(inbound_rule.value.source_addresses) > 0 ? inbound_rule.value.source_addresses : null
      source_droplet_ids         = length(inbound_rule.value.source_droplet_ids) > 0 ? inbound_rule.value.source_droplet_ids : null
      source_tags                = length(inbound_rule.value.source_tags) > 0 ? inbound_rule.value.source_tags : null
      source_kubernetes_ids      = length(inbound_rule.value.source_kubernetes_ids) > 0 ? inbound_rule.value.source_kubernetes_ids : null
      source_load_balancer_uids  = length(inbound_rule.value.source_load_balancer_uids) > 0 ? inbound_rule.value.source_load_balancer_uids : null
    }
  }

  # Outbound rules (traffic allowed *from* Droplets)
  dynamic "outbound_rule" {
    for_each = local.outbound_rules
    content {
      protocol                        = outbound_rule.value.protocol
      port_range                      = outbound_rule.value.port_range != "" ? outbound_rule.value.port_range : null
      destination_addresses           = length(outbound_rule.value.destination_addresses) > 0 ? outbound_rule.value.destination_addresses : null
      destination_droplet_ids         = length(outbound_rule.value.destination_droplet_ids) > 0 ? outbound_rule.value.destination_droplet_ids : null
      destination_tags                = length(outbound_rule.value.destination_tags) > 0 ? outbound_rule.value.destination_tags : null
      destination_kubernetes_ids      = length(outbound_rule.value.destination_kubernetes_ids) > 0 ? outbound_rule.value.destination_kubernetes_ids : null
      destination_load_balancer_uids  = length(outbound_rule.value.destination_load_balancer_uids) > 0 ? outbound_rule.value.destination_load_balancer_uids : null
    }
  }
}

