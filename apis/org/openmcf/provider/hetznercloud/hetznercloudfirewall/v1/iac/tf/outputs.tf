output "firewall_id" {
  description = "The Hetzner Cloud numeric ID of the created firewall"
  value       = hcloud_firewall.this.id
}
