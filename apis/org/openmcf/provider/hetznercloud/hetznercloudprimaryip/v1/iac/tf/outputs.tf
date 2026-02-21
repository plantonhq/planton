output "primary_ip_id" {
  description = "The Hetzner Cloud numeric ID of the created Primary IP"
  value       = hcloud_primary_ip.this.id
}

output "ip_address" {
  description = "The allocated IP address"
  value       = hcloud_primary_ip.this.ip_address
}

output "ip_network" {
  description = "The allocated IPv6 /64 CIDR (empty for IPv4)"
  value       = hcloud_primary_ip.this.ip_network
}
