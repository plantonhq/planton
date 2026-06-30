# =============================================================================
# Outputs (matching GcpRouterNatStackOutputs proto)
# =============================================================================

output "name" {
  description = "Name of the Cloud NAT gateway"
  value       = google_compute_router_nat.nat.name
}

output "router_self_link" {
  description = "Self-link URL of the Cloud Router"
  value       = google_compute_router.router.self_link
}

output "nat_ip_addresses" {
  description = "List of external IP addresses utilized by this NAT (auto-allocated or static)"
  value       = local.nat_ip_allocate_option == "MANUAL_ONLY" ? google_compute_address.nat_ips[*].address : []
}

