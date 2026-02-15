output "key_ring_id" {
  description = "Fully qualified key ring resource path (projects/{project}/locations/{location}/keyRings/{name})"
  value       = google_kms_key_ring.this.id
}

output "key_ring_name" {
  description = "The short name of the key ring"
  value       = google_kms_key_ring.this.name
}
