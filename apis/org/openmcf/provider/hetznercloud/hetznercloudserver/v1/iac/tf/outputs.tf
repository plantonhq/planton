output "server_id" {
  description = "The Hetzner Cloud numeric ID of the created server"
  value       = hcloud_server.this.id
}

output "ipv4_address" {
  description = "The public IPv4 address assigned to the server"
  value       = hcloud_server.this.ipv4_address
}

output "ipv6_address" {
  description = "The first IPv6 address of the assigned /64 network"
  value       = hcloud_server.this.ipv6_address
}

output "status" {
  description = "The current status of the server"
  value       = hcloud_server.this.status
}
