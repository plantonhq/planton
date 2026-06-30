output "floating_ip_id" {
  description = "The Hetzner Cloud numeric ID of the created Floating IP"
  value       = hcloud_floating_ip.this.id
}

output "ip_address" {
  description = "The allocated IP address"
  value       = hcloud_floating_ip.this.ip_address
}

output "ip_network" {
  description = "The allocated IPv6 /64 CIDR (empty for IPv4)"
  value       = hcloud_floating_ip.this.ip_network
}
