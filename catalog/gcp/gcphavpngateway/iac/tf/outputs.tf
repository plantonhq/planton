# What a GcpHaVpnConnection's gateway references (its tunnels attach here),
# and what another Google Cloud gateway's connection names as its peer.
output "gateway_self_link" {
  description = "The HA VPN gateway's self link"
  value       = google_compute_ha_vpn_gateway.this.self_link
}

output "gateway_name" {
  description = "The HA VPN gateway's name in GCP"
  value       = google_compute_ha_vpn_gateway.this.name
}

# What a GcpHaVpnConnection's region references so the two never disagree.
output "region" {
  description = "The region the gateway and router live in"
  value       = google_compute_ha_vpn_gateway.this.region
}

# The two public addresses the on-premises device is configured to reach;
# picked by interface id from the read-back list (see locals.tf).
output "interface_0_ip_address" {
  description = "The public IP of gateway interface 0"
  value       = lookup(local.interface_ips, "0", "")
}

output "interface_1_ip_address" {
  description = "The public IP of gateway interface 1"
  value       = lookup(local.interface_ips, "1", "")
}

# What a GcpHaVpnConnection's router references so its interfaces and BGP
# peers land on this router.
output "router_name" {
  description = "The Cloud Router's name"
  value       = google_compute_router.this.name
}

output "router_self_link" {
  description = "The Cloud Router's self link"
  value       = google_compute_router.this.self_link
}

# The ASN the router speaks as -- what the on-premises side configures as its
# BGP neighbor's ASN.
output "router_asn" {
  description = "The Cloud Router's BGP ASN"
  value       = google_compute_router.this.bgp[0].asn
}
