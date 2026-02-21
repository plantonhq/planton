resource "alicloud_vpn_customer_gateway" "connections" {
  for_each = local.connections_map

  customer_gateway_name = "${each.key}-cg"
  ip_address            = each.value.customer_gateway_ip
  asn                   = each.value.customer_gateway_asn != "" ? each.value.customer_gateway_asn : null
}

resource "alicloud_vpn_connection" "connections" {
  for_each = local.connections_map

  vpn_gateway_id      = alicloud_vpn_gateway.main.id
  customer_gateway_id = alicloud_vpn_customer_gateway.connections[each.key].id
  vpn_connection_name = each.key
  local_subnet        = each.value.local_subnets
  remote_subnet       = each.value.remote_subnets
  enable_dpd          = each.value.enable_dpd
  enable_nat_traversal = each.value.enable_nat_traversal
  effect_immediately  = each.value.effect_immediately

  dynamic "ike_config" {
    for_each = each.value.ike_config != null ? [each.value.ike_config] : []
    content {
      psk          = ike_config.value.psk != "" ? ike_config.value.psk : null
      ike_version  = ike_config.value.ike_version
      ike_mode     = ike_config.value.ike_mode
      ike_enc_alg  = ike_config.value.ike_enc_alg
      ike_auth_alg = ike_config.value.ike_auth_alg
      ike_pfs      = ike_config.value.ike_pfs
      ike_lifetime = ike_config.value.ike_lifetime
    }
  }

  dynamic "ipsec_config" {
    for_each = each.value.ipsec_config != null ? [each.value.ipsec_config] : []
    content {
      ipsec_enc_alg  = ipsec_config.value.ipsec_enc_alg
      ipsec_auth_alg = ipsec_config.value.ipsec_auth_alg
      ipsec_pfs      = ipsec_config.value.ipsec_pfs
      ipsec_lifetime = ipsec_config.value.ipsec_lifetime
    }
  }

  dynamic "health_check_config" {
    for_each = each.value.health_check_config != null ? [each.value.health_check_config] : []
    content {
      enable   = health_check_config.value.enable
      sip      = health_check_config.value.sip != "" ? health_check_config.value.sip : null
      dip      = health_check_config.value.dip != "" ? health_check_config.value.dip : null
      interval = health_check_config.value.interval
      retry    = health_check_config.value.retry
    }
  }
}
