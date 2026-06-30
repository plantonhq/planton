output "network_id" {
  description = "The Hetzner Cloud numeric ID of the created network"
  value       = hcloud_network.this.id
}
