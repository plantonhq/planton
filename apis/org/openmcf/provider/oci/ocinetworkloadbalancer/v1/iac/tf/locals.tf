locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.display_name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciNetworkLoadBalancer"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  policy_map = {
    "five_tuple"  = "FIVE_TUPLE"
    "three_tuple" = "THREE_TUPLE"
    "two_tuple"   = "TWO_TUPLE"
  }

  health_checker_protocol_map = {
    "http"  = "HTTP"
    "https" = "HTTPS"
    "tcp"   = "TCP"
    "udp"   = "UDP"
    "dns"   = "DNS"
  }

  listener_protocol_map = {
    "tcp"         = "TCP"
    "udp"         = "UDP"
    "tcp_and_udp" = "TCP_AND_UDP"
    "any"         = "ANY"
  }

  backend_sets_map = { for bs in var.spec.backend_sets : bs.name => bs }

  backends_flat = flatten([
    for bs in var.spec.backend_sets : [
      for be in bs.backends : {
        key        = "${bs.name}:${coalesce(be.ip_address, be.target_id, "unknown")}:${be.port}"
        bs_name    = bs.name
        ip_address = be.ip_address
        target_id  = be.target_id
        port       = be.port
        weight     = be.weight
        is_backup  = be.is_backup
        is_drain   = be.is_drain
        is_offline = be.is_offline
        name       = be.name
      }
    ]
  ])

  backends_map = { for be in local.backends_flat : be.key => be }

  listeners_map = { for ln in var.spec.listeners : ln.name => ln }
}
