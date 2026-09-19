# The four lists are index-aligned with spec.tunnels (see local.tunnel_order).

output "tunnel_self_links" {
  description = "The tunnels' self links, in spec order"
  value       = [for name in local.tunnel_order : google_compute_vpn_tunnel.this[name].self_link]
}

output "tunnel_names" {
  description = "The tunnels' names in GCP, in spec order"
  value       = [for name in local.tunnel_order : google_compute_vpn_tunnel.this[name].name]
}

output "router_interface_names" {
  description = "The Cloud Router interfaces' names, in spec order"
  value       = [for name in local.tunnel_order : google_compute_router_interface.this[name].name]
}

output "bgp_peer_names" {
  description = "The BGP peers' names, in spec order"
  value       = [for name in local.tunnel_order : google_compute_router_peer.this[name].name]
}

# The external VPN gateway's self link when the peer is an external device;
# empty for a Google-to-Google connection.
output "external_gateway_self_link" {
  description = "The external VPN gateway's self link (external peer only)"
  value       = local.is_external_peer ? google_compute_external_vpn_gateway.this[0].self_link : ""
}

output "gateway_self_link" {
  description = "The HA VPN gateway the tunnels leave from"
  value       = local.gateway
}

output "router_name" {
  description = "The Cloud Router the sessions run on"
  value       = local.router
}
