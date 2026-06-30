locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.display_name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciDynamicRoutingGateway"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  network_type_map = {
    "vcn"                       = "VCN"
    "ipsec_tunnel"              = "IPSEC_TUNNEL"
    "remote_peering_connection" = "REMOTE_PEERING_CONNECTION"
    "virtual_circuit"           = "VIRTUAL_CIRCUIT"
    "loopback"                  = "LOOPBACK"
  }

  vcn_route_type_map = {
    "vcn_cidrs"    = "VCN_CIDRS"
    "subnet_cidrs" = "SUBNET_CIDRS"
  }

  distribution_type_map = {
    "import_routes" = "IMPORT"
    "export_routes" = "EXPORT"
  }

  match_type_map = {
    "match_all"            = "MATCH_ALL"
    "drg_attachment_type"  = "DRG_ATTACHMENT_TYPE"
    "drg_attachment_id"    = "DRG_ATTACHMENT_ID"
  }

  attachments_map = { for att in var.spec.attachments : att.display_name => att }

  route_tables_map = { for rt in var.spec.route_tables : rt.display_name => rt }

  route_distributions_map = { for rd in var.spec.route_distributions : rd.display_name => rd }

  static_route_rules_flat = flatten([
    for rt in var.spec.route_tables : [
      for rule in rt.static_route_rules : {
        key                     = "${rt.display_name}:${rule.destination}"
        rt_name                 = rt.display_name
        destination             = rule.destination
        next_hop_attachment_name = rule.next_hop_attachment_name
      }
    ]
  ])

  static_route_rules_map = { for rule in local.static_route_rules_flat : rule.key => rule }

  distribution_statements_flat = flatten([
    for rd in var.spec.route_distributions : [
      for stmt in rd.statements : {
        key                  = "${rd.display_name}:${stmt.priority}"
        dist_name            = rd.display_name
        priority             = stmt.priority
        match_type           = stmt.match_criteria.match_type
        attachment_type      = stmt.match_criteria.attachment_type
        drg_attachment_name  = stmt.match_criteria.drg_attachment_name
      }
    ]
  ])

  distribution_statements_map = { for stmt in local.distribution_statements_flat : stmt.key => stmt }
}
