locals {
  # Firewall name from spec
  firewall_name = var.spec.name

  # Inbound rules transformation
  # Convert protobuf inbound rules to DigitalOcean Terraform format
  inbound_rules = [
    for rule in var.spec.inbound_rules : {
      protocol                   = rule.protocol
      port_range                 = rule.port_range != null ? rule.port_range : ""
      source_addresses           = rule.source_addresses != null ? rule.source_addresses : []
      source_droplet_ids         = rule.source_droplet_ids != null ? [for id in rule.source_droplet_ids : tonumber(id)] : []
      source_tags                = rule.source_tags != null ? rule.source_tags : []
      source_kubernetes_ids      = rule.source_kubernetes_ids != null ? rule.source_kubernetes_ids : []
      source_load_balancer_uids  = rule.source_load_balancer_uids != null ? rule.source_load_balancer_uids : []
    }
  ]

  # Outbound rules transformation
  # Convert protobuf outbound rules to DigitalOcean Terraform format
  outbound_rules = [
    for rule in var.spec.outbound_rules : {
      protocol                        = rule.protocol
      port_range                      = rule.port_range != null ? rule.port_range : ""
      destination_addresses           = rule.destination_addresses != null ? rule.destination_addresses : []
      destination_droplet_ids         = rule.destination_droplet_ids != null ? [for id in rule.destination_droplet_ids : tonumber(id)] : []
      destination_tags                = rule.destination_tags != null ? rule.destination_tags : []
      destination_kubernetes_ids      = rule.destination_kubernetes_ids != null ? rule.destination_kubernetes_ids : []
      destination_load_balancer_uids  = rule.destination_load_balancer_uids != null ? rule.destination_load_balancer_uids : []
    }
  ]

  # Droplet IDs transformation (convert int64 to int)
  droplet_ids = [for id in var.spec.droplet_ids : tonumber(id)]

  # Tags (pass through as-is)
  tags = var.spec.tags
}

