locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  policy_display_name = (
    var.spec.policy.display_name != ""
    ? var.spec.policy.display_name
    : "${local.display_name}-policy"
  )

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciNetworkFirewall"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  address_list_type_map = {
    "ip"   = "IP"
    "fqdn" = "FQDN"
  }

  service_type_map = {
    "tcp_service" = "TCP_SERVICE"
    "udp_service" = "UDP_SERVICE"
  }

  action_map = {
    "allow"   = "ALLOW"
    "drop"    = "DROP"
    "reject"  = "REJECT"
    "inspect" = "INSPECT"
  }

  inspection_map = {
    "intrusion_detection"  = "INTRUSION_DETECTION"
    "intrusion_prevention" = "INTRUSION_PREVENTION"
  }
}
