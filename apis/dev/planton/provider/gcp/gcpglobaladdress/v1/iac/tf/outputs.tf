output "address" {
  description = "The reserved IP address or start of the reserved range"
  value       = google_compute_global_address.this.address
}

output "self_link" {
  description = "Self-link URL of the global address resource"
  value       = google_compute_global_address.this.self_link
}

output "creation_timestamp" {
  description = "RFC3339 creation timestamp"
  value       = google_compute_global_address.this.creation_timestamp
}
