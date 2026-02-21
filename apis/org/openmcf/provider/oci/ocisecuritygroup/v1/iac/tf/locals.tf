locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.display_name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciSecurityGroup"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  protocol_map = {
    "all"    = "all"
    "icmp"   = "1"
    "tcp"    = "6"
    "udp"    = "17"
    "icmpv6" = "58"
  }

  target_type_map = {
    "cidr_block"              = "CIDR_BLOCK"
    "service_cidr_block"      = "SERVICE_CIDR_BLOCK"
    "network_security_group"  = "NETWORK_SECURITY_GROUP"
    "target_type_unspecified"  = "CIDR_BLOCK"
    ""                         = "CIDR_BLOCK"
  }

  ingress_rules = {
    for i, rule in var.spec.ingress_rules : "ingress-${i}" => merge(rule, {
      direction        = "INGRESS"
      source           = rule.source
      source_type      = lookup(local.target_type_map, rule.source_type, "CIDR_BLOCK")
      destination      = null
      destination_type = null
      protocol         = lookup(local.protocol_map, rule.protocol, "all")
    })
  }

  egress_rules = {
    for i, rule in var.spec.egress_rules : "egress-${i}" => merge(rule, {
      direction        = "EGRESS"
      source           = null
      source_type      = null
      destination      = rule.destination
      destination_type = lookup(local.target_type_map, rule.destination_type, "CIDR_BLOCK")
      protocol         = lookup(local.protocol_map, rule.protocol, "all")
    })
  }

  all_rules = merge(local.ingress_rules, local.egress_rules)
}
