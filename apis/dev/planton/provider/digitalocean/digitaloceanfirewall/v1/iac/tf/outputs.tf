output "firewall_id" {
  description = "The unique identifier (ID) of the created firewall"
  value       = digitalocean_firewall.firewall.id
}

output "firewall_name" {
  description = "The name of the firewall"
  value       = digitalocean_firewall.firewall.name
}

output "firewall_status" {
  description = "The status of the firewall"
  value       = digitalocean_firewall.firewall.status
}

output "created_at" {
  description = "The timestamp when the firewall was created"
  value       = digitalocean_firewall.firewall.created_at
}

output "pending_changes" {
  description = "List of pending changes to the firewall"
  value       = digitalocean_firewall.firewall.pending_changes
}

