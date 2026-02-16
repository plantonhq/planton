output "firewall_self_link" {
  description = "Self-link URI of the created firewall rule"
  value       = google_compute_firewall.this.self_link
}

output "firewall_name" {
  description = "Name of the firewall rule"
  value       = google_compute_firewall.this.name
}

output "creation_timestamp" {
  description = "RFC3339 creation timestamp"
  value       = google_compute_firewall.this.creation_timestamp
}
