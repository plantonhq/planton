output "load_balancer_id" {
  description = "The Hetzner Cloud numeric ID of the created load balancer"
  value       = hcloud_load_balancer.this.id
}

output "ipv4_address" {
  description = "The public IPv4 address assigned to the load balancer"
  value       = hcloud_load_balancer.this.ipv4
}

output "ipv6_address" {
  description = "The public IPv6 address assigned to the load balancer"
  value       = hcloud_load_balancer.this.ipv6
}
