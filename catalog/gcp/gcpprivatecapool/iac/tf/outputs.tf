output "name" {
  description = "Full resource name of the pool (projects/{project}/locations/{location}/caPools/{ca_pool_id}) -- what authorities, certificates, and TLS consumers reference"
  value       = google_privateca_ca_pool.this.id
}

output "ca_pool_id" {
  description = "The pool's ID"
  value       = google_privateca_ca_pool.this.name
}

output "location" {
  description = "The region the pool lives in"
  value       = google_privateca_ca_pool.this.location
}
