# The external VPN gateway resource that holds the peer device's public
# addresses -- created only when the peer is an external device. For a
# Google-to-Google connection the tunnels reference the other side's HA VPN
# gateway directly and this resource does not exist.
#
# Free metadata, immutable except labels: a device's address change is a new
# external gateway, and the tunnels that reference it are recreated with it.
resource "google_compute_external_vpn_gateway" "this" {
  count = local.is_external_peer ? 1 : 0

  project         = local.project_id
  name            = local.external_gateway_name
  redundancy_type = var.spec.peer.external_gateway.redundancy_type
  description     = var.spec.peer.external_gateway.description != "" ? var.spec.peer.external_gateway.description : null
  labels          = merge(var.spec.peer.external_gateway.labels, local.platform_labels)

  # Exactly one of the two addresses is set per interface (spec CEL); the
  # other stays out of the payload.
  dynamic "interface" {
    for_each = var.spec.peer.external_gateway.interfaces
    content {
      id           = interface.value.id
      ip_address   = interface.value.ip_address != "" ? interface.value.ip_address : null
      ipv6_address = interface.value.ipv6_address != "" ? interface.value.ipv6_address : null
    }
  }

  dynamic "params" {
    for_each = length(var.spec.resource_manager_tags) > 0 ? [1] : []
    content {
      resource_manager_tags = var.spec.resource_manager_tags
    }
  }

  deletion_policy = local.deletion_policy
}

# One IPsec tunnel per tunnels[] entry, from a gateway interface to the peer.
# Exactly one of peer_external_gateway / peer_gcp_gateway is set (the
# provider's ConflictsWith, the spec's exactly-one CEL).
#
# Every argument except labels is immutable: a new secret, cipher, or
# interface pairing recreates the tunnel. ike_version is always sent (see
# locals.tf); traffic selectors are Optional+Computed and sent only when set;
# the cipher suite only when declared.
resource "google_compute_vpn_tunnel" "this" {
  for_each = local.tunnels

  project               = local.project_id
  name                  = each.value.name
  region                = local.region
  vpn_gateway           = local.gateway
  vpn_gateway_interface = each.value.vpn_gateway_interface
  router                = local.router
  shared_secret         = each.value.shared_secret
  ike_version           = each.value.ike_version
  description           = each.value.description != "" ? each.value.description : null
  labels                = merge(each.value.labels, local.platform_labels)

  # Google pairs interfaces itself for a Google peer (interface 0 to
  # interface 0), so no peer interface id is sent on that arm.
  peer_external_gateway           = local.is_external_peer ? google_compute_external_vpn_gateway.this[0].self_link : null
  peer_external_gateway_interface = local.is_external_peer ? each.value.peer_external_gateway_interface : null
  peer_gcp_gateway                = local.is_external_peer ? null : var.spec.peer.gcp_gateway

  local_traffic_selector  = length(each.value.local_traffic_selector) > 0 ? each.value.local_traffic_selector : null
  remote_traffic_selector = length(each.value.remote_traffic_selector) > 0 ? each.value.remote_traffic_selector : null

  dynamic "cipher_suite" {
    for_each = each.value.cipher_suite != null ? [each.value.cipher_suite] : []
    content {
      dynamic "phase1" {
        for_each = cipher_suite.value.phase1 != null ? [cipher_suite.value.phase1] : []
        content {
          encryption = length(phase1.value.encryption) > 0 ? phase1.value.encryption : null
          integrity  = length(phase1.value.integrity) > 0 ? phase1.value.integrity : null
          prf        = length(phase1.value.prf) > 0 ? phase1.value.prf : null
          dh         = length(phase1.value.dh) > 0 ? phase1.value.dh : null
        }
      }
      dynamic "phase2" {
        for_each = cipher_suite.value.phase2 != null ? [cipher_suite.value.phase2] : []
        content {
          encryption = length(phase2.value.encryption) > 0 ? phase2.value.encryption : null
          integrity  = length(phase2.value.integrity) > 0 ? phase2.value.integrity : null
          pfs        = length(phase2.value.pfs) > 0 ? phase2.value.pfs : null
        }
      }
    }
  }

  dynamic "params" {
    for_each = length(var.spec.resource_manager_tags) > 0 ? [1] : []
    content {
      resource_manager_tags = var.spec.resource_manager_tags
    }
  }

  deletion_policy = local.deletion_policy
}

# The Cloud Router interface bound to each tunnel: the Google end of the BGP
# session's link-local /30. Fully immutable. ip_range and ip_version are
# Optional+Computed and sent only when set (Google assigns them for IPv6-only
# sessions).
resource "google_compute_router_interface" "this" {
  for_each = local.tunnels

  project    = local.project_id
  name       = each.value.session_name
  region     = local.region
  router     = local.router
  vpn_tunnel = google_compute_vpn_tunnel.this[each.key].self_link

  ip_range   = each.value.bgp_session.interface_ip_range != "" ? each.value.bgp_session.interface_ip_range : null
  ip_version = each.value.bgp_session.ip_version

  deletion_policy = local.deletion_policy
}

# The BGP peer on each tunnel's interface. `enable` is always sent (see
# locals.tf); the optional numerics and addresses are sent only when set so
# 0 is a real priority and Google keeps assigning the addresses it owns; the
# per-session advertisement overrides ride the same CUSTOM-mode rules as the
# router's. The MD5 key rides the peer: the provider inserts it into the
# router's key table under its name and attaches it to the session (Google
# requires each key to be used by exactly one session).
resource "google_compute_router_peer" "this" {
  for_each = local.tunnels

  project   = local.project_id
  name      = each.value.session_name
  region    = local.region
  router    = local.router
  interface = google_compute_router_interface.this[each.key].name
  peer_asn  = each.value.bgp_session.peer_asn

  peer_ip_address           = each.value.bgp_session.peer_ip_address
  advertised_route_priority = each.value.bgp_session.advertised_route_priority
  advertise_mode            = each.value.bgp_session.advertise_mode != "" ? each.value.bgp_session.advertise_mode : null
  advertised_groups         = length(each.value.bgp_session.advertised_groups) > 0 ? each.value.bgp_session.advertised_groups : null

  dynamic "advertised_ip_ranges" {
    for_each = each.value.bgp_session.advertised_ip_ranges
    content {
      range       = advertised_ip_ranges.value.range
      description = advertised_ip_ranges.value.description != "" ? advertised_ip_ranges.value.description : null
    }
  }

  enable      = each.value.enable
  enable_ipv4 = each.value.bgp_session.enable_ipv4
  enable_ipv6 = each.value.bgp_session.enable_ipv6

  ipv6_nexthop_address      = each.value.bgp_session.ipv6_nexthop_address
  peer_ipv6_nexthop_address = each.value.bgp_session.peer_ipv6_nexthop_address
  ipv4_nexthop_address      = each.value.bgp_session.ipv4_nexthop_address
  peer_ipv4_nexthop_address = each.value.bgp_session.peer_ipv4_nexthop_address

  dynamic "custom_learned_ip_ranges" {
    for_each = each.value.bgp_session.custom_learned_ip_ranges
    content {
      range = custom_learned_ip_ranges.value.range
    }
  }
  custom_learned_route_priority = each.value.bgp_session.custom_learned_route_priority

  dynamic "bfd" {
    for_each = each.value.bgp_session.bfd != null ? [each.value.bgp_session.bfd] : []
    content {
      session_initialization_mode = bfd.value.session_initialization_mode
      min_receive_interval        = bfd.value.min_receive_interval > 0 ? bfd.value.min_receive_interval : null
      min_transmit_interval       = bfd.value.min_transmit_interval > 0 ? bfd.value.min_transmit_interval : null
      multiplier                  = bfd.value.multiplier > 0 ? bfd.value.multiplier : null
    }
  }

  dynamic "md5_authentication_key" {
    for_each = each.value.bgp_session.md5_authentication_key != null ? [each.value.bgp_session.md5_authentication_key] : []
    content {
      name = each.value.md5_key_name
      key  = md5_authentication_key.value.key
    }
  }

  import_policies = length(each.value.bgp_session.import_policies) > 0 ? each.value.bgp_session.import_policies : null
  export_policies = length(each.value.bgp_session.export_policies) > 0 ? each.value.bgp_session.export_policies : null

  deletion_policy = local.deletion_policy
}
